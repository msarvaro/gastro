package responses

import "time"

// InventoryItemResponse represents an inventory item in API responses
type InventoryItemResponse struct {
	ID              int        `json:"id"`
	BusinessID      int        `json:"business_id"`
	Name            string     `json:"name"`
	SKU             string     `json:"sku"`
	Category        string     `json:"category"`
	Unit            string     `json:"unit"`
	CurrentStock    float64    `json:"current_stock"`
	MinimumStock    float64    `json:"minimum_stock"`
	MaximumStock    float64    `json:"maximum_stock"`
	ReorderPoint    float64    `json:"reorder_point"`
	Cost            float64    `json:"cost"`
	SupplierID      *int       `json:"supplier_id"`
	ExpiryDate      *time.Time `json:"expiry_date"`
	StorageLocation string     `json:"storage_location"`
	IsActive        bool       `json:"is_active"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	// Computed fields
	NeedsReorder bool   `json:"needs_reorder"`
	IsExpired    bool   `json:"is_expired"`
	IsLowStock   bool   `json:"is_low_stock"`
	StockStatus  string `json:"stock_status"` // "critical", "low", "normal", "high"
}

// InventoryListResponse represents a list of inventory items
type InventoryListResponse struct {
	Items []*InventoryItemResponse `json:"items"`
	Total int                      `json:"total"`
	Page  int                      `json:"page,omitempty"`
	Limit int                      `json:"limit,omitempty"`
}

// StockMovementResponse represents a stock movement record
type StockMovementResponse struct {
	ID            int       `json:"id"`
	InventoryID   int       `json:"inventory_id"`
	MovementType  string    `json:"movement_type"`
	Quantity      float64   `json:"quantity"`
	Reason        string    `json:"reason"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   *int      `json:"reference_id"`
	PerformedBy   int       `json:"performed_by"`
	Notes         string    `json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
}

// InventoryStatsResponse represents inventory statistics
type InventoryStatsResponse struct {
	TotalItems      int                      `json:"total_items"`
	LowStockItems   int                      `json:"low_stock_items"`
	ExpiredItems    int                      `json:"expired_items"`
	ExpiringItems   int                      `json:"expiring_items"`
	ItemsByCategory map[string]int           `json:"items_by_category"`
	TotalValue      float64                  `json:"total_value"`
	RecentMovements []*StockMovementResponse `json:"recent_movements"`
	TopUsedItems    []*InventoryItemResponse `json:"top_used_items"`
	CriticalItems   []*InventoryItemResponse `json:"critical_items"`
}

// InventoryDashboardResponse represents inventory dashboard data
type InventoryDashboardResponse struct {
	Stats          InventoryStatsResponse   `json:"stats"`
	RecentActivity []*StockMovementResponse `json:"recent_activity"`
	AlertsCount    int                      `json:"alerts_count"`
	PendingOrders  int                      `json:"pending_orders"`
}
