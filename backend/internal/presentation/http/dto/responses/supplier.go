package responses

import "time"

// SupplierResponse represents a supplier in API responses
type SupplierResponse struct {
	ID            int       `json:"id"`
	BusinessID    int       `json:"business_id"`
	Name          string    `json:"name"`
	ContactPerson string    `json:"contact_person"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address"`
	TaxID         string    `json:"tax_id"`
	PaymentTerms  string    `json:"payment_terms"`
	Rating        float64   `json:"rating"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SuppliersListResponse represents a list of suppliers
type SuppliersListResponse struct {
	Suppliers []SupplierResponse `json:"suppliers"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

// PurchaseOrderResponse represents a purchase order in API responses
type PurchaseOrderResponse struct {
	ID               int                         `json:"id"`
	BusinessID       int                         `json:"business_id"`
	SupplierID       int                         `json:"supplier_id"`
	SupplierName     string                      `json:"supplier_name"`
	OrderNumber      string                      `json:"order_number"`
	Status           string                      `json:"status"`
	TotalAmount      float64                     `json:"total_amount"`
	ExpectedDelivery time.Time                   `json:"expected_delivery"`
	ActualDelivery   *time.Time                  `json:"actual_delivery,omitempty"`
	Items            []PurchaseOrderItemResponse `json:"items"`
	CreatedBy        int                         `json:"created_by"`
	CreatedByName    string                      `json:"created_by_name"`
	ApprovedBy       *int                        `json:"approved_by,omitempty"`
	ApprovedByName   *string                     `json:"approved_by_name,omitempty"`
	Notes            string                      `json:"notes"`
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
}

// PurchaseOrderItemResponse represents a purchase order item
type PurchaseOrderItemResponse struct {
	ID              int     `json:"id"`
	PurchaseOrderID int     `json:"purchase_order_id"`
	InventoryItemID int     `json:"inventory_item_id"`
	ItemName        string  `json:"item_name"`
	Quantity        float64 `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	TotalPrice      float64 `json:"total_price"`
	ReceivedQty     float64 `json:"received_qty"`
	Notes           string  `json:"notes"`
}

// PurchaseOrdersListResponse represents a list of purchase orders
type PurchaseOrdersListResponse struct {
	Orders   []PurchaseOrderResponse `json:"orders"`
	Total    int                     `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"page_size"`
}
