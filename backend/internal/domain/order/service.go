package order

import "context"

// Service defines the order service interface
type Service interface {
	// GetActiveOrders retrieves all active orders
	GetActiveOrders(ctx context.Context, businessID int) ([]Order, error)

	// GetOrderByID retrieves a specific order by its ID
	GetOrderByID(ctx context.Context, id int, businessID int) (*Order, error)

	// CreateOrder creates a new order with validation
	CreateOrder(ctx context.Context, req CreateOrderRequest, waiterID, businessID int) (*Order, error)

	// UpdateOrderStatus updates an order's status with business rules
	UpdateOrderStatus(ctx context.Context, id int, req UpdateOrderStatusRequest, businessID int) error

	// GetOrderStats retrieves order statistics
	GetOrderStats(ctx context.Context, businessID int) (*OrderStats, error)

	// GetOrderHistory retrieves completed or cancelled orders
	GetOrderHistory(ctx context.Context, businessID int) ([]Order, error)

	// GetKitchenOrders retrieves orders for kitchen display
	GetKitchenOrders(ctx context.Context, businessID int) ([]Order, error)

	// UpdateOrderStatusByCook updates order status by kitchen staff
	UpdateOrderStatusByCook(ctx context.Context, id int, req UpdateOrderStatusRequest, businessID int) error
}
