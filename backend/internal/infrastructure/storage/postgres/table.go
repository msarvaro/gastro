package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"restaurant-management/internal/domain/table"
	"time"
)

type TableRepository struct {
	db *DB
}

func NewTableRepository(db *DB) table.Repository {
	return &TableRepository{db: db}
}

func (r *TableRepository) GetAllTables(ctx context.Context, businessID int) ([]table.Table, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, number, seats, status, reserved_at, occupied_at FROM tables WHERE business_id = $1 OR business_id IS NULL ORDER BY number ASC", businessID)
	if err != nil {
		log.Printf("Error GetAllTables - querying tables: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tables []table.Table
	for rows.Next() {
		var t table.Table
		if err := rows.Scan(&t.ID, &t.Number, &t.Seats, &t.Status, &t.ReservedAt, &t.OccupiedAt); err != nil {
			log.Printf("Error GetAllTables - scanning table row: %v", err)
			return nil, err
		}

		// Fetch active orders for this table
		orderRows, err := r.db.QueryContext(ctx, `
            SELECT id, status, created_at, comment 
            FROM orders 
            WHERE table_id = $1 AND status NOT IN ('completed', 'cancelled') AND (business_id = $2 OR business_id IS NULL)
            ORDER BY created_at ASC`,
			t.ID, businessID,
		)
		if err != nil {
			log.Printf("Error GetAllTables - querying active orders for table %d: %v", t.ID, err)
			tables = append(tables, t)
			continue
		}
		defer orderRows.Close()

		var tableOrders []table.TableOrderInfo
		for orderRows.Next() {
			var toi table.TableOrderInfo
			var comment sql.NullString
			if err := orderRows.Scan(&toi.ID, &toi.Status, &toi.Time, &comment); err != nil {
				log.Printf("Error GetAllTables - scanning order row for table %d: %v", t.ID, err)
				continue
			}
			if comment.Valid {
				toi.Comment = &comment.String
			}
			tableOrders = append(tableOrders, toi)
		}
		orderRows.Close()

		if err := orderRows.Err(); err != nil {
			log.Printf("Error GetAllTables - iterating order rows for table %d: %v", t.ID, err)
		}

		t.Orders = tableOrders
		tables = append(tables, t)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error GetAllTables - iterating table rows: %v", err)
		return nil, err
	}

	return tables, nil
}

func (r *TableRepository) GetTableByID(ctx context.Context, id int) (*table.Table, error) {
	query := `SELECT id, number, seats, status, reserved_at, occupied_at FROM tables WHERE id = $1`
	var t table.Table
	err := r.db.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Number, &t.Seats, &t.Status, &t.ReservedAt, &t.OccupiedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("table with ID %d not found", id)
		}
		log.Printf("Error GetTableByID - scanning table %d: %v", id, err)
		return nil, err
	}
	return &t, nil
}

func (r *TableRepository) GetTableStats(ctx context.Context, businessID int) (*table.TableStats, error) {
	stats := &table.TableStats{}
	query := `
        SELECT
            COUNT(*) as total_tables,
            SUM(CASE WHEN status = 'free' THEN 1 ELSE 0 END) as free_tables,
            SUM(CASE WHEN status = 'occupied' THEN 1 ELSE 0 END) as occupied_tables,
            SUM(CASE WHEN status = 'reserved' THEN 1 ELSE 0 END) as reserved_tables,
            (SUM(CASE WHEN status = 'occupied' THEN 1 ELSE 0 END) * 100.0 / CASE WHEN COUNT(*) = 0 THEN 1 ELSE COUNT(*) END) as occupancy_percentage
        FROM tables
        WHERE business_id = $1
    `
	err := r.db.QueryRowContext(ctx, query, businessID).Scan(
		&stats.Total,
		&stats.Free,
		&stats.Occupied,
		&stats.Reserved,
		&stats.Occupancy,
	)
	if err != nil {
		log.Printf("Error GetTableStats query: %v", err)
		return nil, err
	}
	return stats, nil
}

func (r *TableRepository) UpdateTableStatus(ctx context.Context, tableID int, status string) error {
	var occupiedAt sql.NullTime
	var reservedAt sql.NullTime

	switch table.TableStatus(status) { // Assuming table.TableStatus is defined for status constants
	case table.TableStatusOccupied:
		occupiedAt.Time = time.Now()
		occupiedAt.Valid = true
		reservedAt.Valid = false // Clear reservation if it becomes occupied
	case table.TableStatusReserved:
		reservedAt.Time = time.Now()
		reservedAt.Valid = true
		occupiedAt.Valid = false // Clear occupation if it becomes reserved (e.g. future reservation)
	case table.TableStatusFree:
		occupiedAt.Valid = false
		reservedAt.Valid = false
	default:
		// For any other status, don't explicitly change occupied_at or reserved_at
		// or handle as an error if status is unexpected
		log.Printf("Warning: UpdateTableStatus called with unhandled status '%s' for table %d. occupied_at and reserved_at will not be changed.", status, tableID)
		// Depending on strictness, you might want to return an error here or just update status and updated_at for unhandled cases.
		// For now, let's proceed to update only status and updated_at for unhandled cases.
		_, err := r.db.ExecContext(ctx, "UPDATE tables SET status = $1 WHERE id = $2", status, tableID)
		if err != nil {
			log.Printf("Ошибка обновления статуса (без occupied_at/reserved_at) стола для ID %d: %v", tableID, err)
			return err
		}
		return nil
	}

	// updated_at is always set
	result, err := r.db.ExecContext(ctx, `UPDATE tables 
						   SET status = $1, occupied_at = $2, reserved_at = $3 
						   WHERE id = $4`,
		status, occupiedAt, reservedAt, tableID)

	if err != nil {
		log.Printf("Ошибка обновления статуса стола для ID %d: %v", tableID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Ошибка получения количества затронутых строк для обновления статуса стола ID %d: %v", tableID, err)
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("стол с ID %d для обновления статуса не найден", tableID)
	}
	return nil
}

// UpdateTableStatusWithTimes updates a table's status and timestamp fields
func (r *TableRepository) UpdateTableStatusWithTimes(ctx context.Context, tableID int, status string, reservedAt, occupiedAt *time.Time) error {
	query := `
		UPDATE tables 
		SET status = $1, reserved_at = $2, occupied_at = $3
		WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query, status, reservedAt, occupiedAt, tableID)
	if err != nil {
		log.Printf("Error updating table %d status and timestamps: %v", tableID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting affected rows for table %d status update: %v", tableID, err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("table with ID %d not found", tableID)
	}

	return nil
}

// TableHasActiveOrders checks if a table has any active orders
func (r *TableRepository) TableHasActiveOrders(ctx context.Context, tableID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM orders 
		WHERE table_id = $1 
		AND status NOT IN ('completed', 'cancelled')`,
		tableID,
	).Scan(&count)

	if err != nil {
		log.Printf("Error checking for active orders for table %d: %v", tableID, err)
		return false, err
	}
	return count > 0, nil
}
