package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type ShiftsService interface {
	// Shift management
	StartShift(ctx context.Context, userID int) (*entity.Shift, error)
	EndShift(ctx context.Context, shiftID int) error
	GetActiveShift(ctx context.Context, userID int) (*entity.Shift, error)

	// Break management
	StartBreak(ctx context.Context, shiftID int) error
	EndBreak(ctx context.Context, shiftID int) error

	// Shift reports
	GetShiftSummary(ctx context.Context, shiftID int) (*entity.ShiftSummary, error)
	GetShiftsByDate(ctx context.Context, businessID int, date time.Time) ([]*entity.Shift, error)
	GetUserShiftHistory(ctx context.Context, userID int, start, end time.Time) ([]*entity.Shift, error)
}
