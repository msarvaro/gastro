package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type ManagerService interface {
	// Staff management
	GetStaffList(ctx context.Context, businessID int) ([]*entity.User, error)
	CreateStaffMember(ctx context.Context, user *entity.User) error
	UpdateStaffMember(ctx context.Context, user *entity.User) error
	DeactivateStaffMember(ctx context.Context, userID int) error

	// Reports
	GetDailyReport(ctx context.Context, businessID int, date time.Time) (interface{}, error)
	GetRevenueReport(ctx context.Context, businessID int, start, end time.Time) (interface{}, error)
	GetStaffPerformanceReport(ctx context.Context, businessID int, period string) (interface{}, error)

	// Business operations
	UpdateBusinessHours(ctx context.Context, businessID int, openTime, closeTime string) error
	GetBusinessStatistics(ctx context.Context, businessID int) (interface{}, error)
}
