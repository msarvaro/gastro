package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type InventoryService interface {
	// Item management
	GetInventoryItems(ctx context.Context, businessID int) ([]*entity.InventoryItem, error)
	AddInventoryItem(ctx context.Context, item *entity.InventoryItem) error
	UpdateInventoryItem(ctx context.Context, item *entity.InventoryItem) error
	RemoveInventoryItem(ctx context.Context, itemID int) error

	// Stock management
	UpdateStock(ctx context.Context, itemID int, quantity float64, reason string) error
	GetLowStockItems(ctx context.Context, businessID int) ([]*entity.InventoryItem, error)
	GetExpiringItems(ctx context.Context, businessID int, days int) ([]*entity.InventoryItem, error)

	// Stock movements
	RecordStockMovement(ctx context.Context, movement *entity.StockMovement) error
	GetStockHistory(ctx context.Context, itemID int) ([]*entity.StockMovement, error)
}
