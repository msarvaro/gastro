package entity

import "time"

type Supplier struct {
	ID            int
	BusinessID    int
	Name          string
	ContactPerson string
	Email         string
	Phone         string
	Address       string
	TaxID         string
	PaymentTerms  string  // e.g., "Net 30", "COD"
	Rating        float64 // 0-5
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PurchaseOrder struct {
	ID               int
	BusinessID       int
	SupplierID       int
	OrderNumber      string
	Status           string // draft, sent, confirmed, delivered, canceled
	TotalAmount      float64
	ExpectedDelivery time.Time
	ActualDelivery   *time.Time
	Items            []PurchaseOrderItem
	CreatedBy        int
	ApprovedBy       *int
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type PurchaseOrderItem struct {
	ID              int
	PurchaseOrderID int
	InventoryItemID int
	Quantity        float64
	UnitPrice       float64
	TotalPrice      float64
	ReceivedQty     float64
	Notes           string
}

// Business methods
func (p *PurchaseOrder) CalculateTotal() float64 {
	var total float64
	for _, item := range p.Items {
		total += item.TotalPrice
	}
	return total
}

func (p *PurchaseOrder) IsComplete() bool {
	return p.Status == "delivered" && p.ActualDelivery != nil
}
