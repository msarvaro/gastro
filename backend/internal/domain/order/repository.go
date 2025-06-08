package order

import "context"

// Repository defines the interface for order data operations
type Repository interface {
	// GetActiveOrdersWithItems retrieves all active orders along with their items
	GetActiveOrdersWithItems(ctx context.Context, businessID int) ([]Order, error)

	// GetOrderByID retrieves a specific order by its ID, including its items
	GetOrderByID(ctx context.Context, id int, businessID ...int) (*Order, error)

	// CreateOrderAndItems creates a new order and its associated items in a transaction
	CreateOrderAndItems(ctx context.Context, order *Order, businessID int) (*Order, error)

	// UpdateOrder updates an existing order's status and relevant timestamps
	UpdateOrder(ctx context.Context, order *Order) error

	// GetOrderStatus retrieves order statistics
	GetOrderStatus(ctx context.Context, businessID int) (*OrderStats, error)

	// GetOrderHistoryWithItems retrieves completed or cancelled orders along with their items
	GetOrderHistoryWithItems(ctx context.Context, businessID int) ([]Order, error)

	// GetOrdersByStatus retrieves all orders with a specific status along with their items and dish categories
	GetOrdersByStatus(ctx context.Context, status string, businessID int) ([]Order, error)

	// IsLastActiveOrderForTable checks if the given orderID is the last active order for the tableID
	IsLastActiveOrderForTable(ctx context.Context, tableID, currentOrderID int) (bool, error)

	// GetDishByID retrieves a specific dish by its ID
	GetDishByID(ctx context.Context, id int) (*Dish, error)
}
