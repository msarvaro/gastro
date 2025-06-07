package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type KitchenService interface {
	// Order management
	GetPendingOrders(ctx context.Context, businessID int) ([]*entity.Order, error)
	GetOrderItems(ctx context.Context, status string) ([]*entity.OrderItem, error)

	// Order item processing
	StartPreparingItem(ctx context.Context, itemID int, chefID int) error
	MarkItemReady(ctx context.Context, itemID int) error

	// Kitchen display
	GetKitchenDisplay(ctx context.Context, businessID int) ([]*entity.Order, error)
	GetPreparationQueue(ctx context.Context, businessID int) ([]*entity.OrderItem, error)
}
