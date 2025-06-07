package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type RequestRepository interface {
	GetByID(ctx context.Context, id int) (*entity.ServiceRequest, error)
	GetActiveByBusinessID(ctx context.Context, businessID int) ([]*entity.ServiceRequest, error)
	GetByTableID(ctx context.Context, tableID int) ([]*entity.ServiceRequest, error)
	GetByWaiterID(ctx context.Context, waiterID int) ([]*entity.ServiceRequest, error)
	Create(ctx context.Context, request *entity.ServiceRequest) error
	Update(ctx context.Context, request *entity.ServiceRequest) error
	UpdateStatus(ctx context.Context, id int, status string) error
	AssignToWaiter(ctx context.Context, id int, waiterID int) error
	MarkCompleted(ctx context.Context, id int, completedAt time.Time) error
}
