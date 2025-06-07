package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type ShiftRepository interface {
	// Basic CRUD
	GetByID(ctx context.Context, id int) (*entity.Shift, error)
	GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Shift, error)
	Create(ctx context.Context, shift *entity.Shift) error
	Update(ctx context.Context, shift *entity.Shift) error
	Delete(ctx context.Context, id int) error

	// Shift management
	GetShiftsByDate(ctx context.Context, businessID int, date time.Time) ([]*entity.Shift, error)
	GetShiftsByDateRange(ctx context.Context, businessID int, startDate, endDate time.Time) ([]*entity.Shift, error)

	// Employee shifts
	GetShiftsByEmployeeID(ctx context.Context, employeeID int) ([]*entity.Shift, error)
	GetCurrentShiftForEmployee(ctx context.Context, employeeID int) (*entity.Shift, error)
	GetUpcomingShiftsForEmployee(ctx context.Context, employeeID int) ([]*entity.Shift, error)

	// Shift assignments
	AssignEmployeeToShift(ctx context.Context, shiftID, employeeID int) error
	RemoveEmployeeFromShift(ctx context.Context, shiftID, employeeID int) error
	GetEmployeesByShiftID(ctx context.Context, shiftID int) ([]*entity.User, error)
}
