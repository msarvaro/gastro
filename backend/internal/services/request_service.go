package services

import (
	"context"
	"restaurant-management/internal/domain/consts" // Added import
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"time"
)

// RequestService implements service request operations
type RequestService struct {
	requestRepo repository.RequestRepository
}

func createRequestService(requestRepo repository.RequestRepository) *RequestService {
	return &RequestService{
		requestRepo: requestRepo,
	}
}

// GetRequestByID retrieves a service request by ID
func (s *RequestService) GetRequestByID(ctx context.Context, id int) (*entity.ServiceRequest, error) {
	return s.requestRepo.GetByID(ctx, id)
}

// GetActiveRequestsByBusinessID retrieves all active service requests for a business
func (s *RequestService) GetActiveRequestsByBusinessID(ctx context.Context, businessID int) ([]*entity.ServiceRequest, error) {
	return s.requestRepo.GetActiveByBusinessID(ctx, businessID)
}

// GetRequestsByTableID retrieves all service requests for a specific table
func (s *RequestService) GetRequestsByTableID(ctx context.Context, tableID int) ([]*entity.ServiceRequest, error) {
	return s.requestRepo.GetByTableID(ctx, tableID)
}

// GetRequestsByWaiterID retrieves all service requests assigned to a specific waiter
func (s *RequestService) GetRequestsByWaiterID(ctx context.Context, waiterID int) ([]*entity.ServiceRequest, error) {
	return s.requestRepo.GetByWaiterID(ctx, waiterID)
}

// CreateRequest creates a new service request
func (s *RequestService) CreateRequest(ctx context.Context, request *entity.ServiceRequest) error {
	// Set default values
	request.CreatedAt = time.Now()
	if request.Status == "" {
		request.Status = consts.RequestStatusPending // Changed "pending"
	}
	if request.Priority == "" {
		request.Priority = consts.RequestPriorityMedium // Changed "medium"
	}
	return s.requestRepo.Create(ctx, request)
}

// UpdateRequest updates an existing service request
func (s *RequestService) UpdateRequest(ctx context.Context, request *entity.ServiceRequest) error {
	return s.requestRepo.Update(ctx, request)
}

// UpdateRequestStatus updates the status of a service request
func (s *RequestService) UpdateRequestStatus(ctx context.Context, id int, status string) error {
	return s.requestRepo.UpdateStatus(ctx, id, status)
}

// AssignRequestToWaiter assigns a service request to a specific waiter
func (s *RequestService) AssignRequestToWaiter(ctx context.Context, requestID int, waiterID int) error {
	return s.requestRepo.AssignToWaiter(ctx, requestID, waiterID)
}

// AcknowledgeRequest marks a service request as acknowledged
func (s *RequestService) AcknowledgeRequest(ctx context.Context, requestID int) error {
	now := time.Now()

	// Get the request first to update acknowledged time
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	request.Status = consts.RequestStatusAcknowledged // Changed "acknowledged"
	request.AcknowledgedAt = &now

	return s.requestRepo.Update(ctx, request)
}

// CompleteRequest marks a service request as completed
func (s *RequestService) CompleteRequest(ctx context.Context, requestID int) error {
	now := time.Now()

	// Get the request first to update completion time
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	request.Status = consts.RequestStatusCompleted // Changed "completed"
	request.CompletedAt = &now

	return s.requestRepo.Update(ctx, request)
}
