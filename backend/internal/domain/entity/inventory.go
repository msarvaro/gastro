package entity

import "time"

type InventoryItem struct {
	ID              int
	BusinessID      int
	Name            string
	SKU             string // Stock Keeping Unit
	Category        string // e.g., "Vegetables", "Dairy", "Beverages"
	Unit            string // e.g., "kg", "liters", "pieces"
	CurrentStock    float64
	MinimumStock    float64
	MaximumStock    float64
	ReorderPoint    float64
	Cost            float64 // Cost per unit
	SupplierID      *int
	ExpiryDate      *time.Time
	StorageLocation string
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StockMovement struct {
	ID            int
	InventoryID   int
	MovementType  string // in, out, adjustment, waste
	Quantity      float64
	Reason        string
	ReferenceType string // order, delivery, manual
	ReferenceID   *int   // ID of order or delivery
	PerformedBy   int    // User ID
	Notes         string
	CreatedAt     time.Time
}

// Business methods
func (i *InventoryItem) NeedsReorder() bool {
	return i.CurrentStock <= i.ReorderPoint
}

func (i *InventoryItem) IsExpired() bool {
	if i.ExpiryDate == nil {
		return false
	}
	return time.Now().After(*i.ExpiryDate)
}

func (i *InventoryItem) IsLowStock() bool {
	return i.CurrentStock <= i.MinimumStock
}
