package requests

// CreateOrderRequest represents the data needed to create a new order
type CreateOrderRequest struct {
	TableID int                      `json:"tableId"`
	Comment string                   `json:"comment,omitempty"`
	Items   []CreateOrderItemRequest `json:"items"`
}

// CreateOrderItemRequest represents an item in a new order
type CreateOrderItemRequest struct {
	DishID   int    `json:"dishId"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes,omitempty"`
}

// UpdateOrderStatusRequest represents the data needed to update an order's status
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

// UpdateTableStatusRequest represents the data needed to update a table's status
type UpdateTableStatusRequest struct {
	Status string `json:"status"`
}
