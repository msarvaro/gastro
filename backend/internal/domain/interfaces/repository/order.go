package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type OrderRepository interface {
	// Basic CRUD
	GetByID(ctx context.Context, id int) (*entity.Order, error)
	GetByBusinessID(ctx context.Context, businessID int, limit, offset int) ([]*entity.Order, error)
	Create(ctx context.Context, order *entity.Order) error
	Update(ctx context.Context, order *entity.Order) error
	Delete(ctx context.Context, id int) error

	// Status operations
	UpdateStatus(ctx context.Context, id int, status string) error
	GetByStatus(ctx context.Context, businessID int, status string) ([]*entity.Order, error)

	// Waiter operations
	GetByWaiterID(ctx context.Context, waiterID int) ([]*entity.Order, error)
	GetWaiterOrderStatistics(ctx context.Context, waiterID int) (map[string]int, error)
	GetWaiterTablesServedCount(ctx context.Context, waiterID int) (int, error)
	GetWaiterCompletedOrdersCount(ctx context.Context, waiterID int) (int, error)

	// Date range operations
	GetByDateRange(ctx context.Context, businessID int, startDate, endDate time.Time) ([]*entity.Order, error)

	// Table operations
	GetByTableID(ctx context.Context, tableID int) ([]*entity.Order, error)
	GetActiveByTableID(ctx context.Context, tableID int) (*entity.Order, error)

	// Item operations
	AddOrderItem(ctx context.Context, orderID int, item *entity.OrderItem) error
	RemoveOrderItem(ctx context.Context, orderID, itemID int) error
	UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error
	GetOrderItems(ctx context.Context, orderID int) ([]*entity.OrderItem, error)
}
