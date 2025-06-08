package inventory

import "time"

// Inventory represents an inventory item entity
type Inventory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Quantity    float64   `json:"quantity"`
	Unit        string    `json:"unit"`
	MinQuantity float64   `json:"min_quantity"`
	BusinessID  int       `json:"business_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
