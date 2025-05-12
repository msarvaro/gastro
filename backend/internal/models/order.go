package models

import "time"

type OrderItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Total    float64 `json:"total"`
}

type Order struct {
	ID          int         `json:"id"`
	TableID     int         `json:"table_id"`
	WaiterID    int         `json:"waiter_id"`
	Items       []OrderItem `json:"items"`
	Status      string      `json:"status"` // new, accepted, preparing, ready, served, completed, cancelled
	Comment     string      `json:"comment"`
	Total       float64     `json:"total"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	CompletedAt *time.Time  `json:"completed_at,omitempty"`
	CancelledAt *time.Time  `json:"cancelled_at,omitempty"`
}

type OrderStatus struct {
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
