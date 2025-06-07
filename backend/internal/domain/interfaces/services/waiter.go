package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type WaiterService interface {
	// Table management
	GetAssignedTables(ctx context.Context, waiterID int) ([]*entity.Table, error)
	AssignTable(ctx context.Context, waiterID, tableID int) error
	ReleaseTable(ctx context.Context, tableID int) error

	// Order management
	TakeOrder(ctx context.Context, waiterID int, order *entity.Order) error
	GetWaiterOrders(ctx context.Context, waiterID int) ([]*entity.Order, error)

	// Service requests
	GetActiveRequests(ctx context.Context, waiterID int) ([]*entity.ServiceRequest, error)
	AcknowledgeRequest(ctx context.Context, requestID int) error
	CompleteRequest(ctx context.Context, requestID int) error

	// Performance
	GetPerformanceStats(ctx context.Context, waiterID int, date time.Time) (*entity.WaiterStats, error)
}
