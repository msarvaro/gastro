package service

import (
	"context"
	"log"
	"restaurant-management/internal/domain/business"
	"strings"
)

type BusinessService struct {
	repo business.Repository
}

func NewBusinessService(repo business.Repository) business.Service {
	return &BusinessService{repo: repo}
}

func (s *BusinessService) CreateBusiness(ctx context.Context, b *business.Business) error {
	// Validation
	if strings.TrimSpace(b.Name) == "" {
		return business.ErrInvalidBusinessData
	}

	// Set default status if not provided
	if b.Status == "" {
		b.Status = "active"
	}

	// Validate status
	if b.Status != "active" && b.Status != "inactive" && b.Status != "suspended" {
		return business.ErrInvalidBusinessData
	}

	return s.repo.CreateBusiness(ctx, b)
}

func (s *BusinessService) GetBusinessByID(ctx context.Context, id int) (*business.Business, error) {
	if id <= 0 {
		return nil, business.ErrInvalidBusinessID
	}

	return s.repo.GetBusinessByID(ctx, id)
}

func (s *BusinessService) GetAllBusinesses(ctx context.Context) ([]business.Business, *business.BusinessStats, error) {
	businesses, err := s.repo.GetAllBusinesses(ctx)
	if err != nil {
		return nil, nil, err
	}

	stats, err := s.repo.GetBusinessStats(ctx)
	if err != nil {
		log.Printf("Error getting business stats: %v", err)
		// Continue with empty stats if fetching stats fails
		stats = &business.BusinessStats{}
	}

	return businesses, stats, nil
}

func (s *BusinessService) UpdateBusiness(ctx context.Context, b *business.Business) error {
	if b.ID <= 0 {
		return business.ErrInvalidBusinessID
	}

	// Validation
	if strings.TrimSpace(b.Name) == "" {
		return business.ErrInvalidBusinessData
	}

	// Validate status
	if b.Status != "" && b.Status != "active" && b.Status != "inactive" && b.Status != "suspended" {
		return business.ErrInvalidBusinessData
	}

	// Check if business exists
	_, err := s.repo.GetBusinessByID(ctx, b.ID)
	if err != nil {
		return business.ErrBusinessNotFound
	}

	return s.repo.UpdateBusiness(ctx, b)
}

func (s *BusinessService) DeleteBusiness(ctx context.Context, id int) error {
	if id <= 0 {
		return business.ErrInvalidBusinessID
	}

	// Check if business exists
	_, err := s.repo.GetBusinessByID(ctx, id)
	if err != nil {
		return business.ErrBusinessNotFound
	}

	return s.repo.DeleteBusiness(ctx, id)
}

func (s *BusinessService) SetBusinessCookie(ctx context.Context, businessID int) error {
	if businessID <= 0 {
		return business.ErrInvalidBusinessID
	}

	// Check if business exists
	_, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return business.ErrBusinessNotFound
	}

	// The actual cookie setting will be handled by the handler
	// This service method validates the business exists
	return nil
}
