package repository

import (
	"context"
	"database/sql"
	"restaurant-management/internal/domain/entity"
)

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// Inventory item operations
func (r *InventoryRepository) GetByID(ctx context.Context, id int) (*entity.InventoryItem, error) {
	query := `SELECT id, business_id, name, sku, category, unit, current_stock, minimum_stock,
	         maximum_stock, reorder_point, cost, supplier_id, expiry_date, storage_location,
	         is_active, created_at, updated_at
	         FROM inventory_items WHERE id = $1`

	item := &entity.InventoryItem{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.BusinessID, &item.Name, &item.SKU, &item.Category, &item.Unit,
		&item.CurrentStock, &item.MinimumStock, &item.MaximumStock, &item.ReorderPoint,
		&item.Cost, &item.SupplierID, &item.ExpiryDate, &item.StorageLocation,
		&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
	)

	return item, err
}

func (r *InventoryRepository) GetByBusinessID(ctx context.Context, businessID int) ([]*entity.InventoryItem, error) {
	query := `SELECT id, business_id, name, sku, category, unit, current_stock, minimum_stock,
	         maximum_stock, reorder_point, cost, supplier_id, expiry_date, storage_location,
	         is_active, created_at, updated_at
	         FROM inventory_items WHERE business_id = $1 AND is_active = true ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.InventoryItem
	for rows.Next() {
		item := &entity.InventoryItem{}
		err := rows.Scan(
			&item.ID, &item.BusinessID, &item.Name, &item.SKU, &item.Category, &item.Unit,
			&item.CurrentStock, &item.MinimumStock, &item.MaximumStock, &item.ReorderPoint,
			&item.Cost, &item.SupplierID, &item.ExpiryDate, &item.StorageLocation,
			&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *InventoryRepository) GetBySKU(ctx context.Context, businessID int, sku string) (*entity.InventoryItem, error) {
	query := `SELECT id, business_id, name, sku, category, unit, current_stock, minimum_stock,
	         maximum_stock, reorder_point, cost, supplier_id, expiry_date, storage_location,
	         is_active, created_at, updated_at
	         FROM inventory_items WHERE business_id = $1 AND sku = $2`

	item := &entity.InventoryItem{}
	err := r.db.QueryRowContext(ctx, query, businessID, sku).Scan(
		&item.ID, &item.BusinessID, &item.Name, &item.SKU, &item.Category, &item.Unit,
		&item.CurrentStock, &item.MinimumStock, &item.MaximumStock, &item.ReorderPoint,
		&item.Cost, &item.SupplierID, &item.ExpiryDate, &item.StorageLocation,
		&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
	)

	return item, err
}

func (r *InventoryRepository) GetLowStock(ctx context.Context, businessID int) ([]*entity.InventoryItem, error) {
	query := `SELECT id, business_id, name, sku, category, unit, current_stock, minimum_stock,
	         maximum_stock, reorder_point, cost, supplier_id, expiry_date, storage_location,
	         is_active, created_at, updated_at
	         FROM inventory_items WHERE business_id = $1 AND is_active = true 
	         AND current_stock <= minimum_stock ORDER BY current_stock ASC`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.InventoryItem
	for rows.Next() {
		item := &entity.InventoryItem{}
		err := rows.Scan(
			&item.ID, &item.BusinessID, &item.Name, &item.SKU, &item.Category, &item.Unit,
			&item.CurrentStock, &item.MinimumStock, &item.MaximumStock, &item.ReorderPoint,
			&item.Cost, &item.SupplierID, &item.ExpiryDate, &item.StorageLocation,
			&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *InventoryRepository) GetExpiring(ctx context.Context, businessID int, days int) ([]*entity.InventoryItem, error) {
	query := `SELECT id, business_id, name, sku, category, unit, current_stock, minimum_stock,
	         maximum_stock, reorder_point, cost, supplier_id, expiry_date, storage_location,
	         is_active, created_at, updated_at
	         FROM inventory_items WHERE business_id = $1 AND is_active = true 
	         AND expiry_date IS NOT NULL AND expiry_date <= NOW() + INTERVAL '%d days'
	         ORDER BY expiry_date ASC`

	rows, err := r.db.QueryContext(ctx, query, businessID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*entity.InventoryItem
	for rows.Next() {
		item := &entity.InventoryItem{}
		err := rows.Scan(
			&item.ID, &item.BusinessID, &item.Name, &item.SKU, &item.Category, &item.Unit,
			&item.CurrentStock, &item.MinimumStock, &item.MaximumStock, &item.ReorderPoint,
			&item.Cost, &item.SupplierID, &item.ExpiryDate, &item.StorageLocation,
			&item.IsActive, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *InventoryRepository) Create(ctx context.Context, item *entity.InventoryItem) error {
	query := `INSERT INTO inventory_items (business_id, name, sku, category, unit, current_stock,
	         minimum_stock, maximum_stock, reorder_point, cost, supplier_id, expiry_date,
	         storage_location, is_active, created_at, updated_at)
	         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
	         RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		item.BusinessID, item.Name, item.SKU, item.Category, item.Unit,
		item.CurrentStock, item.MinimumStock, item.MaximumStock, item.ReorderPoint,
		item.Cost, item.SupplierID, item.ExpiryDate, item.StorageLocation, item.IsActive,
	).Scan(&item.ID)

	return err
}

func (r *InventoryRepository) Update(ctx context.Context, item *entity.InventoryItem) error {
	query := `UPDATE inventory_items SET name = $1, sku = $2, category = $3, unit = $4,
	         current_stock = $5, minimum_stock = $6, maximum_stock = $7, reorder_point = $8,
	         cost = $9, supplier_id = $10, expiry_date = $11, storage_location = $12,
	         is_active = $13, updated_at = NOW() WHERE id = $14`

	_, err := r.db.ExecContext(ctx, query,
		item.Name, item.SKU, item.Category, item.Unit, item.CurrentStock,
		item.MinimumStock, item.MaximumStock, item.ReorderPoint, item.Cost,
		item.SupplierID, item.ExpiryDate, item.StorageLocation, item.IsActive, item.ID,
	)

	return err
}

func (r *InventoryRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE inventory_items SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *InventoryRepository) UpdateStock(ctx context.Context, id int, quantity float64) error {
	query := `UPDATE inventory_items SET current_stock = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, quantity, id)
	return err
}

// Stock movement operations
func (r *InventoryRepository) GetStockMovements(ctx context.Context, inventoryID int) ([]*entity.StockMovement, error) {
	query := `SELECT id, inventory_id, movement_type, quantity, reason, reference_type,
	         reference_id, performed_by, notes, created_at
	         FROM stock_movements WHERE inventory_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, inventoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []*entity.StockMovement
	for rows.Next() {
		movement := &entity.StockMovement{}
		err := rows.Scan(
			&movement.ID, &movement.InventoryID, &movement.MovementType, &movement.Quantity,
			&movement.Reason, &movement.ReferenceType, &movement.ReferenceID,
			&movement.PerformedBy, &movement.Notes, &movement.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}

	return movements, nil
}

func (r *InventoryRepository) CreateStockMovement(ctx context.Context, movement *entity.StockMovement) error {
	query := `INSERT INTO stock_movements (inventory_id, movement_type, quantity, reason,
	         reference_type, reference_id, performed_by, notes, created_at)
	         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW()) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		movement.InventoryID, movement.MovementType, movement.Quantity, movement.Reason,
		movement.ReferenceType, movement.ReferenceID, movement.PerformedBy, movement.Notes,
	).Scan(&movement.ID)

	return err
}
