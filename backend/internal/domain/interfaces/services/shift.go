package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type ShiftService interface {
	// Shift operations
	GetShiftByID(ctx context.Context, id int) (*entity.Shift, error)
	GetShiftsByBusinessID(ctx context.Context, businessID int) ([]*entity.Shift, error)
	GetShiftsByDate(ctx context.Context, businessID int, date time.Time) ([]*entity.Shift, error)
	CreateShift(ctx context.Context, shift *entity.Shift) error
	UpdateShift(ctx context.Context, shift *entity.Shift) error
	DeleteShift(ctx context.Context, id int) error

	// Employee shift operations
	GetShiftsByEmployeeID(ctx context.Context, employeeID int) ([]*entity.Shift, error)
	GetCurrentShiftForEmployee(ctx context.Context, employeeID int) (*entity.Shift, error)
	GetUpcomingShiftsForEmployee(ctx context.Context, employeeID int) ([]*entity.Shift, error)

	// Shift assignment
	AssignEmployeeToShift(ctx context.Context, shiftID, employeeID int) error
	RemoveEmployeeFromShift(ctx context.Context, shiftID, employeeID int) error
	GetEmployeesByShiftID(ctx context.Context, shiftID int) ([]*entity.User, error)
}
