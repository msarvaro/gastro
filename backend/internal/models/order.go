package models

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "new"
	OrderStatusInProgress OrderStatus = "in_progress"
	OrderStatusReady      OrderStatus = "ready"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type OrderItem struct {
	ID       int     `json:"id" db:"id"`
	OrderID  int     `json:"order_id" db:"order_id"`
	DishID   int     `json:"dish_id" db:"dish_id"`
	Quantity int     `json:"quantity" db:"quantity"`
	Price    float64 `json:"price" db:"price"`
	Notes    string  `json:"notes,omitempty" db:"notes"`
}

type Order struct {
	ID          int         `json:"id" db:"id"`
	TableID     int         `json:"table_id" db:"table_id"`
	WaiterID    int         `json:"waiter_id" db:"waiter_id"`
	Status      OrderStatus `json:"status" db:"status"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`
	Comment     string      `json:"comment,omitempty" db:"comment"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty" db:"completed_at"`
}

type OrderStats struct {
	Total       int     `json:"total"`
	New         int     `json:"new"`
	Accepted    int     `json:"accepted"`
	Preparing   int     `json:"preparing"`
	Ready       int     `json:"ready"`
	Served      int     `json:"served"`
	Completed   int     `json:"completed"`
	Cancelled   int     `json:"cancelled"`
	TotalAmount float64 `json:"total_amount"`
}
