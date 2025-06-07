package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type OrderService interface {
	// Order management
	CreateOrder(ctx context.Context, order *entity.Order) error
	GetOrderByID(ctx context.Context, orderID int) (*entity.Order, error)
	GetActiveOrders(ctx context.Context, businessID int) ([]*entity.Order, error)
	GetOrdersByTable(ctx context.Context, tableID int) ([]*entity.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int, status string) error

	// Order items
	AddItemToOrder(ctx context.Context, orderID int, item *entity.OrderItem) error
	UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error
	RemoveItemFromOrder(ctx context.Context, orderID, itemID int) error
	UpdateItemStatus(ctx context.Context, orderID, itemID int, status string) error

	// Order completion
	CalculateOrderTotal(ctx context.Context, orderID int) (float64, error)
	CompleteOrder(ctx context.Context, orderID int) error
	CancelOrder(ctx context.Context, orderID int, reason string) error
}
