package requests

import "time"

// CreateSupplierRequest represents the request to create a new supplier
type CreateSupplierRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	ContactPerson string  `json:"contact_person" validate:"required,min=2,max=100"`
	Email         string  `json:"email" validate:"required,email"`
	Phone         string  `json:"phone" validate:"required,min=10,max=20"`
	Address       string  `json:"address" validate:"required,max=500"`
	TaxID         string  `json:"tax_id" validate:"max=50"`
	PaymentTerms  string  `json:"payment_terms" validate:"required"`
	Rating        float64 `json:"rating" validate:"min=0,max=5"`
	IsActive      *bool   `json:"is_active"`
}

// UpdateSupplierRequest represents the request to update a supplier
type UpdateSupplierRequest struct {
	Name          *string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	ContactPerson *string  `json:"contact_person,omitempty" validate:"omitempty,min=2,max=100"`
	Email         *string  `json:"email,omitempty" validate:"omitempty,email"`
	Phone         *string  `json:"phone,omitempty" validate:"omitempty,min=10,max=20"`
	Address       *string  `json:"address,omitempty" validate:"omitempty,max=500"`
	TaxID         *string  `json:"tax_id,omitempty" validate:"omitempty,max=50"`
	PaymentTerms  *string  `json:"payment_terms,omitempty" validate:"omitempty"`
	Rating        *float64 `json:"rating,omitempty" validate:"omitempty,min=0,max=5"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

// CreatePurchaseOrderRequest represents the request to create a purchase order
type CreatePurchaseOrderRequest struct {
	SupplierID       int                        `json:"supplier_id" validate:"required"`
	OrderNumber      string                     `json:"order_number" validate:"required,max=50"`
	ExpectedDelivery time.Time                  `json:"expected_delivery" validate:"required"`
	Items            []PurchaseOrderItemRequest `json:"items" validate:"required,min=1"`
	Notes            string                     `json:"notes" validate:"max=1000"`
}

// UpdatePurchaseOrderRequest represents the request to update a purchase order
type UpdatePurchaseOrderRequest struct {
	SupplierID       int                        `json:"supplier_id" validate:"required"`
	OrderNumber      string                     `json:"order_number" validate:"required,max=50"`
	Status           string                     `json:"status" validate:"required"`
	ExpectedDelivery time.Time                  `json:"expected_delivery" validate:"required"`
	ActualDelivery   *time.Time                 `json:"actual_delivery,omitempty"`
	Items            []PurchaseOrderItemRequest `json:"items" validate:"required,min=1"`
	Notes            string                     `json:"notes" validate:"max=1000"`
}

// PurchaseOrderItemRequest represents a purchase order item
type PurchaseOrderItemRequest struct {
	InventoryItemID int     `json:"inventory_item_id" validate:"required"`
	Quantity        float64 `json:"quantity" validate:"required,min=0"`
	UnitPrice       float64 `json:"unit_price" validate:"required,min=0"`
	Notes           string  `json:"notes" validate:"max=500"`
}

// UpdatePurchaseOrderStatusRequest represents the request to update purchase order status
type UpdatePurchaseOrderStatusRequest struct {
	Status         string     `json:"status" validate:"required"`
	ActualDelivery *time.Time `json:"actual_delivery,omitempty"`
	Notes          string     `json:"notes" validate:"max=500"`
}

// ReceivePurchaseOrderRequest represents the request to receive a purchase order
type ReceivePurchaseOrderRequest struct {
	Items []ReceiveItemRequest `json:"items" validate:"required,min=1"`
	Notes string               `json:"notes" validate:"max=500"`
}

// ReceiveItemRequest represents an item being received
type ReceiveItemRequest struct {
	PurchaseOrderItemID int     `json:"purchase_order_item_id" validate:"required"`
	ReceivedQuantity    float64 `json:"received_quantity" validate:"required,min=0"`
	Notes               string  `json:"notes" validate:"max=500"`
}

// SupplierFilterRequest represents filters for supplier queries
type SupplierFilterRequest struct {
	IsActive     *bool    `json:"is_active"`
	MinRating    *float64 `json:"min_rating" validate:"min=0,max=5"`
	PaymentTerms string   `json:"payment_terms"`
	SearchTerm   string   `json:"search_term" validate:"max=100"`
}
