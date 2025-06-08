package service

import (
	"context"
	"fmt"
	"log"
	"restaurant-management/internal/domain/request"
)

type RequestService struct {
	requestRepo request.Repository
}

func NewRequestService(requestRepo request.Repository) request.Service {
	return &RequestService{
		requestRepo: requestRepo,
	}
}

func (s *RequestService) GetAll(ctx context.Context, businessID int) ([]request.Request, error) {
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	requests, err := s.requestRepo.GetAll(ctx, businessID)
	if err != nil {
		log.Printf("Error retrieving requests for business %d: %v", businessID, err)
		return nil, err
	}

	if requests == nil {
		requests = []request.Request{}
	}

	return requests, nil
}

func (s *RequestService) GetByID(ctx context.Context, id int, businessID int) (*request.Request, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid request ID: %d", id)
	}
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	request, err := s.requestRepo.GetByID(ctx, id, businessID)
	if err != nil {
		log.Printf("Error retrieving request %d for business %d: %v", id, businessID, err)
		return nil, err
	}

	return request, nil
}

func (s *RequestService) Create(ctx context.Context, requestReq request.CreateRequestRequest, businessID int) (*request.Request, error) {
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	// Validate required fields
	if requestReq.SupplierID <= 0 {
		return nil, fmt.Errorf("supplier ID is required")
	}
	if len(requestReq.Items) == 0 {
		return nil, fmt.Errorf("request items are required")
	}

	// Set default status if not provided
	if requestReq.Status == "" {
		requestReq.Status = "pending"
	}

	createdRequest, err := s.requestRepo.Create(ctx, requestReq, businessID)
	if err != nil {
		log.Printf("Error creating request for business %d: %v", businessID, err)
		return nil, err
	}

	log.Printf("Successfully created request %d for supplier %d in business %d", createdRequest.ID, createdRequest.SupplierID, businessID)
	return createdRequest, nil
}

func (s *RequestService) Update(ctx context.Context, id int, requestReq request.UpdateRequestRequest, businessID int) (*request.Request, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid request ID: %d", id)
	}
	if businessID <= 0 {
		return nil, fmt.Errorf("invalid business ID: %d", businessID)
	}

	// Validate that at least one field is being updated
	if requestReq.SupplierID == 0 && len(requestReq.Items) == 0 &&
		requestReq.Priority == "" && requestReq.Comment == "" &&
		requestReq.Status == "" {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	updatedRequest, err := s.requestRepo.Update(ctx, id, requestReq, businessID)
	if err != nil {
		log.Printf("Error updating request %d for business %d: %v", id, businessID, err)
		return nil, err
	}

	log.Printf("Successfully updated request %d for business %d", id, businessID)
	return updatedRequest, nil
}

func (s *RequestService) Delete(ctx context.Context, id int, businessID int) error {
	if id <= 0 {
		return fmt.Errorf("invalid request ID: %d", id)
	}
	if businessID <= 0 {
		return fmt.Errorf("invalid business ID: %d", businessID)
	}

	// Check if request exists before deleting
	_, err := s.requestRepo.GetByID(ctx, id, businessID)
	if err != nil {
		log.Printf("Request %d not found for business %d: %v", id, businessID, err)
		return fmt.Errorf("request not found")
	}

	err = s.requestRepo.Delete(ctx, id, businessID)
	if err != nil {
		log.Printf("Error deleting request %d for business %d: %v", id, businessID, err)
		return err
	}

	log.Printf("Successfully deleted request %d for business %d", id, businessID)
	return nil
}
