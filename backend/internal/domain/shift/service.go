package shift

import "context"

// Service defines the shift service interface
type Service interface {
	// GetAllShifts retrieves all shifts with pagination
	GetAllShifts(ctx context.Context, page, limit int, businessID int) ([]ShiftWithEmployees, int, error)

	// GetShiftByID retrieves a specific shift by its ID
	GetShiftByID(ctx context.Context, shiftID int, businessID int) (*ShiftWithEmployees, error)

	// GetShiftEmployees retrieves employees assigned to a shift
	GetShiftEmployees(ctx context.Context, shiftID int, businessID int) ([]User, error)

	// CreateShift creates a new shift with employee assignments
	CreateShift(ctx context.Context, shift *Shift, employeeIDs []int, businessID int) (*Shift, error)

	// UpdateShift updates shift information and employee assignments
	UpdateShift(ctx context.Context, shift *Shift, employeeIDs []int, businessID int) error

	// DeleteShift deletes a shift by ID
	DeleteShift(ctx context.Context, shiftID int, businessID int) error

	// GetCurrentAndUpcomingShifts retrieves current and upcoming shifts
	GetCurrentAndUpcomingShifts(ctx context.Context, businessID int) ([]ShiftWithEmployees, error)
}
