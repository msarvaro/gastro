package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// tableRepository implements the TableRepository interface using PostgreSQL
type tableRepository struct {
	db *sql.DB
}

// NewTableRepository creates a new table repository
func NewTableRepository(db *sql.DB) repository.TableRepository {
	return &tableRepository{
		db: db,
	}
}

// GetByID retrieves a table by ID
func (r *tableRepository) GetByID(ctx context.Context, id int) (*entity.Table, error) {
	query := `
		SELECT id, number, seats, status, reserved_at, occupied_at, business_id
		FROM tables 
		WHERE id = $1
	`

	var table entity.Table
	var reservedAt, occupiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&table.ID,
		&table.Number,
		&table.Seats,
		&table.Status,
		&reservedAt,
		&occupiedAt,
		&table.BusinessID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if reservedAt.Valid {
		table.ReservedAt = &reservedAt.Time
	}

	if occupiedAt.Valid {
		table.OccupiedAt = &occupiedAt.Time
	}

	// Get waiter ID if assigned through active orders
	waiterQuery := `
		SELECT DISTINCT waiter_id 
		FROM orders 
		WHERE table_id = $1 AND status NOT IN ('completed', 'cancelled')
		LIMIT 1
	`

	var waiterID sql.NullInt64
	err = r.db.QueryRowContext(ctx, waiterQuery, id).Scan(&waiterID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if waiterID.Valid {
		wid := int(waiterID.Int64)
		table.WaiterID = &wid
	}

	return &table, nil
}

// GetByBusinessID retrieves all tables for a business
func (r *tableRepository) GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Table, error) {
	query := `
		SELECT id, number, seats, status, reserved_at, occupied_at, business_id
		FROM tables 
		WHERE business_id = $1
		ORDER BY number
	`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*entity.Table

	for rows.Next() {
		var table entity.Table
		var reservedAt, occupiedAt sql.NullTime

		err := rows.Scan(
			&table.ID,
			&table.Number,
			&table.Seats,
			&table.Status,
			&reservedAt,
			&occupiedAt,
			&table.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if reservedAt.Valid {
			table.ReservedAt = &reservedAt.Time
		}

		if occupiedAt.Valid {
			table.OccupiedAt = &occupiedAt.Time
		}

		tables = append(tables, &table)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Get waiter assignments for all tables
	for _, table := range tables {
		waiterQuery := `
			SELECT DISTINCT waiter_id 
			FROM orders 
			WHERE table_id = $1 AND status NOT IN ('completed', 'cancelled')
			LIMIT 1
		`

		var waiterID sql.NullInt64
		err = r.db.QueryRowContext(ctx, waiterQuery, table.ID).Scan(&waiterID)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if waiterID.Valid {
			wid := int(waiterID.Int64)
			table.WaiterID = &wid
		}
	}

	return tables, nil
}

// GetByNumber retrieves a table by its number
func (r *tableRepository) GetByNumber(ctx context.Context, businessID, number int) (*entity.Table, error) {
	query := `
		SELECT id, number, seats, status, reserved_at, occupied_at, business_id
		FROM tables 
		WHERE business_id = $1 AND number = $2
	`

	var table entity.Table
	var reservedAt, occupiedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, businessID, number).Scan(
		&table.ID,
		&table.Number,
		&table.Seats,
		&table.Status,
		&reservedAt,
		&occupiedAt,
		&table.BusinessID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if reservedAt.Valid {
		table.ReservedAt = &reservedAt.Time
	}

	if occupiedAt.Valid {
		table.OccupiedAt = &occupiedAt.Time
	}

	// Get waiter ID if assigned through active orders
	waiterQuery := `
		SELECT DISTINCT waiter_id 
		FROM orders 
		WHERE table_id = $1 AND status NOT IN ('completed', 'cancelled')
		LIMIT 1
	`

	var waiterID sql.NullInt64
	err = r.db.QueryRowContext(ctx, waiterQuery, table.ID).Scan(&waiterID)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if waiterID.Valid {
		wid := int(waiterID.Int64)
		table.WaiterID = &wid
	}

	return &table, nil
}

// Create adds a new table
func (r *tableRepository) Create(ctx context.Context, table *entity.Table) error {
	query := `
		INSERT INTO tables (number, seats, status, business_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		table.Number,
		table.Seats,
		table.Status,
		table.BusinessID,
	).Scan(&table.ID)

	return err
}

// Update updates an existing table
func (r *tableRepository) Update(ctx context.Context, table *entity.Table) error {
	query := `
		UPDATE tables
		SET number = $1, seats = $2, status = $3, reserved_at = $4, occupied_at = $5
		WHERE id = $6 AND business_id = $7
	`

	var reservedAt, occupiedAt sql.NullTime

	if table.ReservedAt != nil {
		reservedAt.Time = *table.ReservedAt
		reservedAt.Valid = true
	}

	if table.OccupiedAt != nil {
		occupiedAt.Time = *table.OccupiedAt
		occupiedAt.Valid = true
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		table.Number,
		table.Seats,
		table.Status,
		reservedAt,
		occupiedAt,
		table.ID,
		table.BusinessID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("table not found")
	}

	return nil
}

// Delete removes a table
func (r *tableRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM tables WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("table not found")
	}

	return nil
}

// UpdateTableStatus updates a table's status
func (r *tableRepository) UpdateTableStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE tables
		SET status = $1
	`

	params := []interface{}{status}
	paramCount := 2 // Starting parameter count (after status)

	// Add timestamps based on status
	if status == "reserved" {
		query += ", reserved_at = $" + strconv.Itoa(paramCount)
		params = append(params, time.Now())
		paramCount++
	} else if status == "occupied" {
		query += ", occupied_at = $" + strconv.Itoa(paramCount)
		params = append(params, time.Now())
		paramCount++
	} else if status == "free" {
		query += ", reserved_at = NULL, occupied_at = NULL"
	}

	query += " WHERE id = $" + strconv.Itoa(paramCount)
	params = append(params, id)

	result, err := r.db.ExecContext(ctx, query, params...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("table not found")
	}

	return nil
}

// GetTablesByStatus retrieves tables by status
func (r *tableRepository) GetTablesByStatus(ctx context.Context, businessID int, status string) ([]*entity.Table, error) {
	query := `
		SELECT id, number, seats, status, reserved_at, occupied_at, business_id
		FROM tables 
		WHERE business_id = $1 AND status = $2
		ORDER BY number
	`

	rows, err := r.db.QueryContext(ctx, query, businessID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*entity.Table

	for rows.Next() {
		var table entity.Table
		var reservedAt, occupiedAt sql.NullTime

		err := rows.Scan(
			&table.ID,
			&table.Number,
			&table.Seats,
			&table.Status,
			&reservedAt,
			&occupiedAt,
			&table.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if reservedAt.Valid {
			table.ReservedAt = &reservedAt.Time
		}

		if occupiedAt.Valid {
			table.OccupiedAt = &occupiedAt.Time
		}

		tables = append(tables, &table)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

// GetTablesByWaiterID retrieves tables assigned to a waiter through active orders
func (r *tableRepository) GetTablesByWaiterID(ctx context.Context, waiterID int) ([]*entity.Table, error) {
	query := `
		SELECT DISTINCT t.id, t.number, t.seats, t.status, t.reserved_at, t.occupied_at, t.business_id
		FROM tables t
		JOIN orders o ON t.id = o.table_id
		WHERE o.waiter_id = $1 AND o.status NOT IN ('completed', 'cancelled')
		ORDER BY t.number
	`

	rows, err := r.db.QueryContext(ctx, query, waiterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []*entity.Table

	for rows.Next() {
		var table entity.Table
		var reservedAt, occupiedAt sql.NullTime

		err := rows.Scan(
			&table.ID,
			&table.Number,
			&table.Seats,
			&table.Status,
			&reservedAt,
			&occupiedAt,
			&table.BusinessID,
		)

		if err != nil {
			return nil, err
		}

		if reservedAt.Valid {
			table.ReservedAt = &reservedAt.Time
		}

		if occupiedAt.Valid {
			table.OccupiedAt = &occupiedAt.Time
		}

		// Set the waiter ID since we know it from the query
		table.WaiterID = &waiterID

		tables = append(tables, &table)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

// AssignTableToWaiter assigns a table to a waiter (simplified - assignments happen through orders)
func (r *tableRepository) AssignTableToWaiter(ctx context.Context, tableID, waiterID int) error {
	// First, check if the table exists
	tableQuery := `SELECT id FROM tables WHERE id = $1`
	var id int
	err := r.db.QueryRowContext(ctx, tableQuery, tableID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("table not found")
		}
		return err
	}

	// Then, check if the waiter exists
	waiterQuery := `SELECT id FROM users WHERE id = $1 AND role = 'waiter'`
	err = r.db.QueryRowContext(ctx, waiterQuery, waiterID).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("waiter not found")
		}
		return err
	}

	// In this system, table assignments happen through orders
	// This is a placeholder that validates the inputs exist
	// Actual assignment happens when an order is created
	return nil
}

// UnassignTableFromWaiter removes a waiter assignment from a table (simplified)
func (r *tableRepository) UnassignTableFromWaiter(ctx context.Context, tableID, waiterID int) error {
	// In this system, table assignments happen through orders
	// To "unassign" a table, we would complete or cancel active orders
	// This is a placeholder function
	return nil
}

// GetTableOccupancyRate calculates the occupancy rate for a business
func (r *tableRepository) GetTableOccupancyRate(ctx context.Context, businessID int) (float64, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'free' THEN 0 ELSE 1 END) as occupied
		FROM tables
		WHERE business_id = $1
	`

	var total, occupied int
	err := r.db.QueryRowContext(ctx, query, businessID).Scan(&total, &occupied)
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 0, nil
	}

	return float64(occupied) / float64(total), nil
}

// GetReservationsByDate retrieves reservations for a specific date
func (r *tableRepository) GetReservationsByDate(ctx context.Context, businessID int, date time.Time) ([]*entity.TableReservation, error) {
	query := `
		SELECT id, table_id, customer_name, customer_phone, reservation_date,
		       duration, party_size, status, notes, created_at, updated_at
		FROM table_reservations
		WHERE business_id = $1 AND DATE(reservation_date) = DATE($2)
		ORDER BY reservation_date
	`

	rows, err := r.db.QueryContext(ctx, query, businessID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []*entity.TableReservation

	for rows.Next() {
		var reservation entity.TableReservation

		err := rows.Scan(
			&reservation.ID,
			&reservation.TableID,
			&reservation.CustomerName,
			&reservation.CustomerPhone,
			&reservation.ReservationDate,
			&reservation.Duration,
			&reservation.PartySize,
			&reservation.Status,
			&reservation.Notes,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

// GetReservationByID retrieves a specific reservation by ID
func (r *tableRepository) GetReservationByID(ctx context.Context, id int) (*entity.TableReservation, error) {
	query := `
		SELECT id, table_id, customer_name, customer_phone, reservation_date,
		       duration, party_size, status, notes, created_at, updated_at
		FROM table_reservations
		WHERE id = $1
	`

	var reservation entity.TableReservation

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&reservation.ID,
		&reservation.TableID,
		&reservation.CustomerName,
		&reservation.CustomerPhone,
		&reservation.ReservationDate,
		&reservation.Duration,
		&reservation.PartySize,
		&reservation.Status,
		&reservation.Notes,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("reservation not found")
		}
		return nil, err
	}

	return &reservation, nil
}

// CreateReservation creates a new table reservation
func (r *tableRepository) CreateReservation(ctx context.Context, reservation *entity.TableReservation) error {
	query := `
		INSERT INTO table_reservations (table_id, customer_name, customer_phone, 
		                              reservation_date, duration, party_size, 
		                              status, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id
	`

	now := time.Now()

	err := r.db.QueryRowContext(
		ctx, query,
		reservation.TableID,
		reservation.CustomerName,
		reservation.CustomerPhone,
		reservation.ReservationDate,
		reservation.Duration,
		reservation.PartySize,
		reservation.Status,
		reservation.Notes,
		now,
	).Scan(&reservation.ID)

	if err != nil {
		return err
	}

	reservation.CreatedAt = now
	reservation.UpdatedAt = now

	return nil
}

// UpdateReservation updates an existing table reservation
func (r *tableRepository) UpdateReservation(ctx context.Context, reservation *entity.TableReservation) error {
	query := `
		UPDATE table_reservations
		SET table_id = $1, customer_name = $2, customer_phone = $3,
		    reservation_date = $4, duration = $5, party_size = $6,
		    status = $7, notes = $8, updated_at = $9
		WHERE id = $10
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx, query,
		reservation.TableID,
		reservation.CustomerName,
		reservation.CustomerPhone,
		reservation.ReservationDate,
		reservation.Duration,
		reservation.PartySize,
		reservation.Status,
		reservation.Notes,
		now,
		reservation.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("reservation not found")
	}

	reservation.UpdatedAt = now

	return nil
}

// CancelReservation cancels a table reservation
func (r *tableRepository) CancelReservation(ctx context.Context, id int) error {
	query := `
		UPDATE table_reservations
		SET status = 'canceled', updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("reservation not found")
	}

	return nil
}
