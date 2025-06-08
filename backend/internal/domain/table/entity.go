package table

import "time"

// TableStatus represents the status of a table
type TableStatus string

const (
	TableStatusFree     TableStatus = "free"
	TableStatusOccupied TableStatus = "occupied"
	TableStatusReserved TableStatus = "reserved"
)

// TableOrderInfo represents order information for a table
type TableOrderInfo struct {
	ID      int       `json:"id"`
	Time    time.Time `json:"time"`
	Status  string    `json:"status"`
	Comment *string   `json:"comment,omitempty"`
}

// Table represents a table entity
type Table struct {
	ID           int              `json:"id"`
	Number       int              `json:"number"`
	Seats        int              `json:"seats"`
	Status       TableStatus      `json:"status"`
	Orders       []TableOrderInfo `json:"orders,omitempty"` // Active orders associated with the table
	ReservedAt   *time.Time       `json:"reserved_at,omitempty"`
	OccupiedAt   *time.Time       `json:"occupied_at,omitempty"`
	CurrentOrder *int             `json:"current_order,omitempty"`
}

// TableStats represents table statistics
type TableStats struct {
	Total     int     `json:"total"`
	Free      int     `json:"free"`
	Occupied  int     `json:"occupied"`
	Reserved  int     `json:"reserved"`
	Occupancy float64 `json:"occupancy"`
}
