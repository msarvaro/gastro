package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type RequestService interface {
	// Request operations
	GetRequestByID(ctx context.Context, id int) (*entity.ServiceRequest, error)
	GetActiveRequestsByBusinessID(ctx context.Context, businessID int) ([]*entity.ServiceRequest, error)
	GetRequestsByTableID(ctx context.Context, tableID int) ([]*entity.ServiceRequest, error)
	GetRequestsByWaiterID(ctx context.Context, waiterID int) ([]*entity.ServiceRequest, error)
	CreateRequest(ctx context.Context, request *entity.ServiceRequest) error
	UpdateRequest(ctx context.Context, request *entity.ServiceRequest) error
	UpdateRequestStatus(ctx context.Context, id int, status string) error

	// Request assignment and completion
	AssignRequestToWaiter(ctx context.Context, requestID int, waiterID int) error
	AcknowledgeRequest(ctx context.Context, requestID int) error
	CompleteRequest(ctx context.Context, requestID int) error
}
