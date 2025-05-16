package models

import (
	"time"
)

type TableStatus string

const (
	TableStatusFree     TableStatus = "free"
	TableStatusOccupied TableStatus = "occupied"
	TableStatusReserved TableStatus = "reserved"
)

// TableOrderInfo holds simplified order information for display with tables.
type TableOrderInfo struct {
	ID      int       `json:"id"`
	Time    time.Time `json:"time"` // This will be 'created_at' from the orders table
	Comment *string   `json:"comment,omitempty"`
}

// Table represents a restaurant table.
type Table struct {
	ID           int              `json:"id"`
	Number       int              `json:"number"`
	Seats        int              `json:"seats"`
	Status       TableStatus      `json:"status"`
	Orders       []TableOrderInfo `json:"orders,omitempty"` // Active orders associated with the table
	ReservedAt   *time.Time       `json:"reserved_at,omitempty"`
	OccupiedAt   *time.Time       `json:"occupied_at,omitempty"`
	CurrentOrder *int             `json:"current_order,omitempty"`
	// CreatedAt and UpdatedAt should be removed as they are not used for tables
	// CreatedAt    time.Time `json:"created_at"`
	// UpdatedAt    time.Time `json:"updated_at"`
}

// TableStats holds statistics about the tables.
// Ensure this matches what GetTableStats in postgres.go returns and waiter.js expects.
type TableStats struct {
	Total     int     `json:"total"`
	Free      int     `json:"free"`
	Occupied  int     `json:"occupied"`  // If you calculate and return this
	Reserved  int     `json:"reserved"`  // If you calculate and return this
	Occupancy float64 `json:"occupancy"` // This was in a previous summary, ensure it's correct
}
