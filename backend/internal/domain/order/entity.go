package order

import "time"

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "new"
	OrderStatusAccepted  OrderStatus = "accepted"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusServed    OrderStatus = "served"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents an item within an order
type OrderItem struct {
	ID       int     `json:"id"`       // Corresponds to 'order_items.id'
	OrderID  int     `json:"order_id"` // Corresponds to 'order_items.order_id'
	DishID   int     `json:"dish_id"`  // Corresponds to 'order_items.dish_id', which is a foreign key to 'dishes.id'
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Quantity int     `json:"quantity"`        // Corresponds to 'order_items.quantity'
	Price    float64 `json:"price"`           // Price of one unit AT THE TIME OF ORDER. Corresponds to 'order_items.price'
	Total    float64 `json:"total"`           // Subtotal for this item (Quantity * Price). Can be calculated or stored.
	Notes    string  `json:"notes,omitempty"` // Corresponds to 'order_items.notes'
}

// Order represents an order entity
type Order struct {
	ID          int         `json:"id"`                     // Corresponds to 'orders.id'
	TableID     int         `json:"table_id"`               // Corresponds to 'orders.table_id'
	WaiterID    int         `json:"waiter_id"`              // Corresponds to 'orders.waiter_id'
	Status      OrderStatus `json:"status"`                 // Corresponds to 'orders.status'
	TotalAmount float64     `json:"total_amount"`           // Corresponds to 'orders.total_amount'
	Comment     string      `json:"comment,omitempty"`      // Corresponds to 'orders.comment'
	CreatedAt   time.Time   `json:"created_at"`             // Corresponds to 'orders.created_at'
	UpdatedAt   time.Time   `json:"updated_at"`             // Corresponds to 'orders.updated_at'
	CompletedAt *time.Time  `json:"completed_at,omitempty"` // Corresponds to 'orders.completed_at'
	CancelledAt *time.Time  `json:"cancelled_at,omitempty"` // Corresponds to 'orders.cancelled_at'
	Items       []OrderItem `json:"items,omitempty"`        // Populated from 'order_items' table
}

// OrderStats represents order statistics
type OrderStats struct {
	TotalActiveOrders    int     `json:"total_active_orders"`
	New                  int     `json:"new"`
	Accepted             int     `json:"accepted"`
	Preparing            int     `json:"preparing"`
	Ready                int     `json:"ready"`
	Served               int     `json:"served"`
	CompletedTotal       int     `json:"completed_total,omitempty"`        // Total completed, not just today
	CancelledTotal       int     `json:"cancelled_total,omitempty"`        // Total cancelled
	CompletedAmountTotal float64 `json:"completed_amount_total,omitempty"` // Sum of total_amount for all orders (or active ones)
}

// CreateOrderRequest represents data for creating an order
type CreateOrderRequest struct {
	TableID int              `json:"tableId" binding:"required"`
	Comment string           `json:"comment,omitempty"`
	Items   []OrderItemInput `json:"items" binding:"required,min=1"`
	// WaiterID will be extracted from the auth token on the backend
}

// OrderItemInput represents input data for an order item
type OrderItemInput struct {
	DishID   int    `json:"dishId" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
	Notes    string `json:"notes,omitempty"`
}

// UpdateOrderStatusRequest represents data for updating order status
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}

// Dish represents a dish entity (simplified for orders)
type Dish struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	CategoryID  int     `json:"category_id"`
	IsAvailable bool    `json:"is_available"`
}
