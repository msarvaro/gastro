package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type InventoryRepository interface {
	// Inventory item operations
	GetByID(ctx context.Context, id int) (*entity.InventoryItem, error)
	GetByBusinessID(ctx context.Context, businessID int) ([]*entity.InventoryItem, error)
	GetBySKU(ctx context.Context, businessID int, sku string) (*entity.InventoryItem, error)
	GetLowStock(ctx context.Context, businessID int) ([]*entity.InventoryItem, error)
	GetExpiring(ctx context.Context, businessID int, days int) ([]*entity.InventoryItem, error)
	Create(ctx context.Context, item *entity.InventoryItem) error
	Update(ctx context.Context, item *entity.InventoryItem) error
	Delete(ctx context.Context, id int) error
	UpdateStock(ctx context.Context, id int, quantity float64) error

	// Stock movement operations
	GetStockMovements(ctx context.Context, inventoryID int) ([]*entity.StockMovement, error)
	CreateStockMovement(ctx context.Context, movement *entity.StockMovement) error
}
