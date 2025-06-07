package requests

// KitchenUpdateOrderStatusRequest represents a request to update order status from kitchen
type KitchenUpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=preparing ready"`
}

// UpdateInventoryFromKitchenRequest represents a request to update inventory from kitchen
type UpdateInventoryFromKitchenRequest struct {
	Quantity float64 `json:"quantity" validate:"required,min=0"`
}

// KitchenOrderFilterRequest represents filters for kitchen orders
type KitchenOrderFilterRequest struct {
	Status   string `json:"status,omitempty"`
	TableID  *int   `json:"table_id,omitempty"`
	WaiterID *int   `json:"waiter_id,omitempty"`
}
