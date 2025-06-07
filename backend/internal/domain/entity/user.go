package entity

import "time"

type User struct {
	ID           int
	Username     string
	Email        string
	Password     string // Hashed password
	Name         string
	Role         string // admin, manager, waiter, kitchen
	Status       string // active, inactive
	BusinessID   *int   // Null for admin users
	LastActiveAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserProfile represents extended user information including related data
type UserProfile struct {
	User                *User
	AssignedTables      []*Table       // Tables assigned to a waiter
	CurrentShift        *Shift         // Current active shift
	UpcomingShifts      []*Shift       // Future shifts
	OrderStats          map[string]int // Statistics about orders (new, accepted, etc.)
	PerformanceData     map[string]int // Performance metrics
	CurrentShiftManager string         // Name of the manager for current shift
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) CanAccessBusiness(businessID int) bool {
	if u.IsAdmin() {
		return true
	}
	return u.BusinessID != nil && *u.BusinessID == businessID
}
