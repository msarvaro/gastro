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

type Table struct {
	ID           int         `json:"id" db:"id"`
	Number       int         `json:"number" db:"number"`
	Seats        int         `json:"seats" db:"seats"`
	Status       TableStatus `json:"status" db:"status"`
	ReservedAt   *time.Time  `json:"reserved_at,omitempty" db:"reserved_at"`
	OccupiedAt   *time.Time  `json:"occupied_at,omitempty" db:"occupied_at"`
	CurrentOrder *int        `json:"current_order,omitempty" db:"current_order"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

type TableStats struct {
	Total    int `json:"total"`
	Free     int `json:"free"`
	Occupied int `json:"occupied"`
	Reserved int `json:"reserved"`
}
