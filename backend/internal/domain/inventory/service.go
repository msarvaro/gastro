package inventory

import "context"

// Service defines the inventory service interface
type Service interface {
	// GetAllInventory retrieves all inventory items with low stock alerts
	GetAllInventory(ctx context.Context, businessID int) ([]Inventory, error)

	// GetInventoryByID retrieves an inventory item by its ID
	GetInventoryByID(ctx context.Context, id int, businessID int) (*Inventory, error)

	// CreateInventory creates a new inventory item with validation
	CreateInventory(ctx context.Context, item *Inventory, businessID int) error

	// UpdateInventory updates an existing inventory item with validation
	UpdateInventory(ctx context.Context, item *Inventory, businessID int) error

	// DeleteInventory deletes an inventory item
	DeleteInventory(ctx context.Context, id int, businessID int) error

	// CheckLowStockLevels checks for items with low stock and returns them
	CheckLowStockLevels(ctx context.Context, businessID int) ([]Inventory, error)
}
