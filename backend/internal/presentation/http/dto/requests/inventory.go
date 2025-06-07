package requests

import "time"

// CreateInventoryItemRequest represents the request to create a new inventory item
type CreateInventoryItemRequest struct {
	Name            string     `json:"name" validate:"required,max=100"`
	SKU             string     `json:"sku" validate:"required,max=50"`
	Category        string     `json:"category" validate:"required,max=50"`
	Unit            string     `json:"unit" validate:"required,max=20"`
	CurrentStock    float64    `json:"current_stock" validate:"min=0"`
	MinimumStock    float64    `json:"minimum_stock" validate:"min=0"`
	MaximumStock    float64    `json:"maximum_stock" validate:"min=0"`
	ReorderPoint    float64    `json:"reorder_point" validate:"min=0"`
	Cost            float64    `json:"cost" validate:"min=0"`
	SupplierID      *int       `json:"supplier_id,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	StorageLocation string     `json:"storage_location,omitempty" validate:"max=100"`
}

// UpdateInventoryItemRequest represents the request to update an existing inventory item
type UpdateInventoryItemRequest struct {
	Name            *string    `json:"name,omitempty" validate:"omitempty,max=100"`
	SKU             *string    `json:"sku,omitempty" validate:"omitempty,max=50"`
	Category        *string    `json:"category,omitempty" validate:"omitempty,max=50"`
	Unit            *string    `json:"unit,omitempty" validate:"omitempty,max=20"`
	CurrentStock    *float64   `json:"current_stock,omitempty" validate:"omitempty,min=0"`
	MinimumStock    *float64   `json:"minimum_stock,omitempty" validate:"omitempty,min=0"`
	MaximumStock    *float64   `json:"maximum_stock,omitempty" validate:"omitempty,min=0"`
	ReorderPoint    *float64   `json:"reorder_point,omitempty" validate:"omitempty,min=0"`
	Cost            *float64   `json:"cost,omitempty" validate:"omitempty,min=0"`
	SupplierID      *int       `json:"supplier_id,omitempty"`
	ExpiryDate      *time.Time `json:"expiry_date,omitempty"`
	StorageLocation *string    `json:"storage_location,omitempty" validate:"omitempty,max=100"`
	IsActive        *bool      `json:"is_active,omitempty"`
}

// UpdateStockRequest represents a request to update stock quantity
type UpdateStockRequest struct {
	Quantity float64 `json:"quantity" validate:"required,min=0"`
	Reason   string  `json:"reason,omitempty" validate:"max=200"`
}

// StockMovementRequest represents a request to record stock movement
type StockMovementRequest struct {
	MovementType  string  `json:"movement_type" validate:"required,oneof=in out adjustment waste transfer"`
	Quantity      float64 `json:"quantity" validate:"required"`
	Reason        string  `json:"reason" validate:"required,max=200"`
	ReferenceType string  `json:"reference_type,omitempty" validate:"max=50"`
	ReferenceID   *int    `json:"reference_id,omitempty"`
	Notes         string  `json:"notes,omitempty" validate:"max=500"`
}
