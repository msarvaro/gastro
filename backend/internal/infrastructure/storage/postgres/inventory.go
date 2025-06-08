package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"restaurant-management/internal/domain/inventory"
	"time"
)

type InventoryRepository struct {
	db *DB
}

func NewInventoryRepository(db *DB) inventory.Repository {
	return &InventoryRepository{db: db}
}

func (r *InventoryRepository) GetAllInventory(ctx context.Context, businessID int) ([]inventory.Inventory, error) {
	query := `
		SELECT id, name, category, quantity, unit, min_quantity, business_id, created_at, updated_at
		FROM inventory 
		WHERE business_id = $1 OR business_id IS NULL
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		log.Printf("Error querying inventory: %v", err)
		return nil, err
	}
	defer rows.Close()

	var items []inventory.Inventory
	for rows.Next() {
		var item inventory.Inventory
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Category,
			&item.Quantity,
			&item.Unit,
			&item.MinQuantity,
			&item.BusinessID,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning inventory item: %v", err)
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating inventory rows: %v", err)
		return nil, err
	}

	return items, nil
}

func (r *InventoryRepository) GetInventoryByID(ctx context.Context, id int, businessID int) (*inventory.Inventory, error) {
	query := `
		SELECT id, name, category, quantity, unit, min_quantity, business_id, created_at, updated_at
		FROM inventory 
		WHERE id = $1 AND (business_id = $2 OR business_id IS NULL)`

	var item inventory.Inventory
	err := r.db.QueryRowContext(ctx, query, id, businessID).Scan(
		&item.ID,
		&item.Name,
		&item.Category,
		&item.Quantity,
		&item.Unit,
		&item.MinQuantity,
		&item.BusinessID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("inventory item with ID %d not found", id)
	}
	if err != nil {
		log.Printf("Error scanning inventory item by ID %d: %v", id, err)
		return nil, err
	}

	return &item, nil
}

func (r *InventoryRepository) CreateInventory(ctx context.Context, item *inventory.Inventory) error {
	query := `
		INSERT INTO inventory (name, category, quantity, unit, min_quantity, business_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		item.Name,
		item.Category,
		item.Quantity,
		item.Unit,
		item.MinQuantity,
		item.BusinessID,
		item.CreatedAt,
		item.UpdatedAt,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		log.Printf("Error creating inventory item: %v", err)
		return err
	}

	return nil
}

func (r *InventoryRepository) UpdateInventory(ctx context.Context, item *inventory.Inventory) error {
	query := `
		UPDATE inventory 
		SET name = $1, category = $2, quantity = $3, unit = $4, min_quantity = $5, updated_at = $6
		WHERE id = $7 AND (business_id = $8 OR business_id IS NULL)`

	item.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		item.Name,
		item.Category,
		item.Quantity,
		item.Unit,
		item.MinQuantity,
		item.UpdatedAt,
		item.ID,
		item.BusinessID,
	)

	if err != nil {
		log.Printf("Error updating inventory item ID %d: %v", item.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting affected rows for inventory item %d: %v", item.ID, err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("inventory item with ID %d not found", item.ID)
	}

	return nil
}

func (r *InventoryRepository) DeleteInventory(ctx context.Context, id int, businessID int) error {
	query := `DELETE FROM inventory WHERE id = $1 AND (business_id = $2 OR business_id IS NULL)`

	result, err := r.db.ExecContext(ctx, query, id, businessID)
	if err != nil {
		log.Printf("Error deleting inventory item ID %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting affected rows for inventory deletion %d: %v", id, err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("inventory item with ID %d not found", id)
	}

	return nil
}
