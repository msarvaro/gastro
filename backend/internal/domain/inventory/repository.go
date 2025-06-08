package inventory

import "context"

// Repository defines the interface for inventory data operations
type Repository interface {
	// GetAllInventory retrieves all inventory items for a business
	GetAllInventory(ctx context.Context, businessID int) ([]Inventory, error)

	// GetInventoryByID retrieves an inventory item by its ID
	GetInventoryByID(ctx context.Context, id int, businessID int) (*Inventory, error)

	// CreateInventory creates a new inventory item
	CreateInventory(ctx context.Context, item *Inventory) error

	// UpdateInventory updates an existing inventory item
	UpdateInventory(ctx context.Context, item *Inventory) error

	// DeleteInventory deletes an inventory item
	DeleteInventory(ctx context.Context, id int, businessID int) error
}
