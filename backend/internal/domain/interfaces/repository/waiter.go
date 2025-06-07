package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type WaiterRepository interface {
	// Stats operations
	GetStats(ctx context.Context, userID int, period string, date time.Time) (*entity.WaiterStats, error)
	GetStatsByBusinessID(ctx context.Context, businessID int, period string, date time.Time) ([]*entity.WaiterStats, error)
	CreateOrUpdateStats(ctx context.Context, stats *entity.WaiterStats) error

	// Performance operations
	GetPerformance(ctx context.Context, userID int, shiftID int) (*entity.WaiterPerformance, error)
	GetPerformanceByDateRange(ctx context.Context, userID int, start, end time.Time) ([]*entity.WaiterPerformance, error)
	CreatePerformanceRecord(ctx context.Context, performance *entity.WaiterPerformance) error
	UpdatePerformanceRecord(ctx context.Context, performance *entity.WaiterPerformance) error
}
