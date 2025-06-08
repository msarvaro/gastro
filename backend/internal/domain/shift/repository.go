package shift

import "context"

// Repository defines the interface for shift data operations
type Repository interface {
	// GetAllShifts returns all shifts with limited data and pagination
	GetAllShifts(ctx context.Context, page, limit int, businessID int) ([]ShiftWithEmployees, int, error)

	// GetShiftByID returns information about a specific shift
	GetShiftByID(ctx context.Context, shiftID int, businessID int) (*ShiftWithEmployees, error)

	// GetShiftEmployees returns the list of employees assigned to a shift
	GetShiftEmployees(ctx context.Context, shiftID int, businessID int) ([]User, error)

	// CreateShift creates a new shift and associates it with employees
	CreateShift(ctx context.Context, shift *Shift, employeeIDs []int, businessID int) (*Shift, error)

	// UpdateShift updates shift information and redistributes employees
	UpdateShift(ctx context.Context, shift *Shift, employeeIDs []int, businessID int) error

	// DeleteShift deletes a shift and all related records
	DeleteShift(ctx context.Context, shiftID int, businessID int) error

	// GetEmployeeShifts returns shifts for an employee
	GetEmployeeShifts(ctx context.Context, employeeID int, businessID int) ([]ShiftWithEmployees, error)

	// GetCurrentAndUpcomingShifts returns current shift and list of upcoming shifts for an employee
	GetCurrentAndUpcomingShifts(ctx context.Context, employeeID int, businessID int) (*ShiftWithEmployees, []ShiftWithEmployees, error)
}
