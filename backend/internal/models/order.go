package models

import "time"

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

// OrderItem reflects the structure of the 'order_items' table.
// It represents a single item within an order.
type OrderItem struct {
	ID       int     `json:"id"`              // Corresponds to 'order_items.id'
	OrderID  int     `json:"order_id"`        // Corresponds to 'order_items.order_id'
	DishID   int     `json:"dish_id"`         // Corresponds to 'order_items.dish_id', which is a foreign key to 'dishes.id'
	Name     string  `json:"name,omitempty"`  // Name of the dish, can be stored for convenience or fetched via DishID
	Quantity int     `json:"quantity"`        // Corresponds to 'order_items.quantity'
	Price    float64 `json:"price"`           // Price of one unit AT THE TIME OF ORDER. Corresponds to 'order_items.price'
	Total    float64 `json:"total"`           // Subtotal for this item (Quantity * Price). Can be calculated or stored.
	Notes    string  `json:"notes,omitempty"` // Corresponds to 'order_items.notes'
	// If menu_item_options are used and affect price/details, they should be represented here too.
	// For example: SelectedOptions []SelectedOrderItemOption `json:"selected_options,omitempty"`
}

// Order reflects the structure of the 'orders' table.
type Order struct {
	ID          int         `json:"id"`                     // Corresponds to 'orders.id'
	TableID     int         `json:"table_id"`               // Corresponds to 'orders.table_id'
	WaiterID    int         `json:"waiter_id"`              // Corresponds to 'orders.waiter_id'
	Status      OrderStatus `json:"status"`                 // Corresponds to 'orders.status'
	TotalAmount float64     `json:"total_amount"`           // Corresponds to 'orders.total_amount'
	Comment     string      `json:"comment,omitempty"`      // Corresponds to 'orders.comment'
	CreatedAt   time.Time   `json:"created_at"`             // Corresponds to 'orders.created_at'
	UpdatedAt   time.Time   `json:"updated_at"`             // Corresponds to 'orders.updated_at'
	CompletedAt *time.Time  `json:"completed_at,omitempty"` // Corresponds to 'orders.completed_at' (needs to be added to DB table)
	CancelledAt *time.Time  `json:"cancelled_at,omitempty"` // Corresponds to 'orders.cancelled_at' (needs to be added to DB table)
	Items       []OrderItem `json:"items,omitempty"`        // Populated from 'order_items' table
}

// OrderStats provides statistics about orders.
type OrderStats struct {
	TotalActiveOrders int     `json:"total_active_orders"`
	New               int     `json:"new"`
	Accepted          int     `json:"accepted"`
	Preparing         int     `json:"preparing"`
	Ready             int     `json:"ready"`
	Served            int     `json:"served"`
	CompletedTotal    int     `json:"completed_total,omitempty"`  // Total completed, not just today
	CancelledTotal    int     `json:"cancelled_total,omitempty"`  // Total cancelled
	TotalAmountAll    float64 `json:"total_amount_all,omitempty"` // Sum of total_amount for all orders (or active ones)
}

// CreateOrderRequest defines the expected structure for creating a new order from the client.
type CreateOrderRequest struct {
	TableID int              `json:"table_id" binding:"required"`
	Comment string           `json:"comment,omitempty"`
	Items   []OrderItemInput `json:"items" binding:"required,min=1"`
	// WaiterID will be extracted from the auth token on the backend
}

// OrderItemInput defines the structure for items when creating a new order.
type OrderItemInput struct {
	DishID   int    `json:"dish_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,gt=0"`
	Notes    string `json:"notes,omitempty"`
	// If options are selectable per item:
	// SelectedOptionIDs []int `json:"selected_option_ids,omitempty"`
}

// UpdateOrderStatusRequest defines the structure for updating an order's status.
type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}

// Dish represents the 'dishes' table structure (simplified for context here).
// You likely have a more complete model in models/menu.go or similar.
type Dish struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	IsAvailable bool    `json:"is_available"`
	// ... other fields like is_available, category_id etc.
}
