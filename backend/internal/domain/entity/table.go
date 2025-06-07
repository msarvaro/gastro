package entity

import "time"

// Table represents a dining table in a restaurant
type Table struct {
	ID         int
	Number     int
	Seats      int
	Status     string // "free", "occupied", "reserved"
	ReservedAt *time.Time
	OccupiedAt *time.Time
	BusinessID int
	WaiterID   *int     // Optional assignment to a waiter
	Orders     []*Order // Current and past orders for this table
}

// IsAvailable checks if the table is currently available
func (t *Table) IsAvailable() bool {
	return t.Status == "free"
}

// IsReserved checks if the table is currently reserved
func (t *Table) IsReserved() bool {
	return t.Status == "reserved"
}

// IsOccupied checks if the table is currently occupied
func (t *Table) IsOccupied() bool {
	return t.Status == "occupied"
}

// SetStatus updates the table status and related timestamps
func (t *Table) SetStatus(status string) {
	now := time.Now()
	t.Status = status

	switch status {
	case "reserved":
		t.ReservedAt = &now
	case "occupied":
		t.OccupiedAt = &now
	case "free":
		// Reset timestamps when table becomes free
		t.ReservedAt = nil
		t.OccupiedAt = nil
	}
}

// AssignWaiter assigns a waiter to this table
func (t *Table) AssignWaiter(waiterID int) {
	t.WaiterID = &waiterID
}

// UnassignWaiter removes waiter assignment
func (t *Table) UnassignWaiter() {
	t.WaiterID = nil
}

type TableReservation struct {
	ID              int
	TableID         int
	CustomerName    string
	CustomerPhone   string
	ReservationDate time.Time
	Duration        int // in minutes
	PartySize       int
	Status          string // confirmed, canceled, completed
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Business methods
func (t *Table) CanAccommodate(partySize int) bool {
	return t.Seats >= partySize
}
