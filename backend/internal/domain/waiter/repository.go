package waiter

import (
	"context"
)

// Repository defines the interface for waiter data operations
type Repository interface {
	GetWaiterProfile(ctx context.Context, waiterID int, businessID int) (*WaiterProfile, error)
	GetWaiterCurrentAndUpcomingShifts(ctx context.Context, waiterID int, businessID int) (*ShiftWithEmployees, []ShiftWithEmployees, error)
	GetTablesAssignedToWaiter(ctx context.Context, waiterID int, businessID int) ([]Table, error)
	GetWaiterOrderStats(ctx context.Context, waiterID int, businessID int) (OrderStatusCounts, error)
	GetWaiterPerformanceMetrics(ctx context.Context, waiterID int, businessID int) (PerformanceMetrics, error)
}
