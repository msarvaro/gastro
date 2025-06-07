package services

import (
	"context"
	"errors"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// BusinessService implements the business service interface
type BusinessService struct {
	businessRepo repository.BusinessRepository
	userRepo     repository.UserRepository
}

func createBusinessService(
	businessRepo repository.BusinessRepository,
	userRepo repository.UserRepository,
) *BusinessService {
	return &BusinessService{
		businessRepo: businessRepo,
		userRepo:     userRepo,
	}
}

// GetBusinessByID retrieves a business by ID
func (s *BusinessService) GetBusinessByID(ctx context.Context, id int) (*entity.Business, error) {
	return s.businessRepo.GetByID(ctx, id)
}

// GetAllBusinesses retrieves all businesses
func (s *BusinessService) GetAllBusinesses(ctx context.Context) ([]*entity.Business, error) {
	return s.businessRepo.GetAll(ctx)
}

// GetActiveBusinesses retrieves all active businesses
func (s *BusinessService) GetActiveBusinesses(ctx context.Context) ([]*entity.Business, error) {
	return s.businessRepo.GetByStatus(ctx, "active")
}

// CreateBusiness creates a new business
func (s *BusinessService) CreateBusiness(ctx context.Context, business *entity.Business) error {
	// Set default status if not provided
	if business.Status == "" {
		business.Status = "active"
	}
	return s.businessRepo.Create(ctx, business)
}

// UpdateBusiness updates an existing business
func (s *BusinessService) UpdateBusiness(ctx context.Context, business *entity.Business) error {
	// Verify business exists
	_, err := s.businessRepo.GetByID(ctx, business.ID)
	if err != nil {
		return err
	}
	return s.businessRepo.Update(ctx, business)
}

// DeleteBusiness deletes a business
func (s *BusinessService) DeleteBusiness(ctx context.Context, id int) error {
	// Verify business exists
	_, err := s.businessRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return s.businessRepo.Delete(ctx, id)
}

// GetBusinessUsers retrieves all users for a business
func (s *BusinessService) GetBusinessUsers(ctx context.Context, businessID int) ([]*entity.User, error) {
	// Verify business exists
	_, err := s.businessRepo.GetByID(ctx, businessID)
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetByBusinessID(ctx, businessID)
}

// AddUserToBusiness adds a user to a business
func (s *BusinessService) AddUserToBusiness(ctx context.Context, user *entity.User, businessID int) error {
	// Verify business exists
	_, err := s.businessRepo.GetByID(ctx, businessID)
	if err != nil {
		return err
	}

	// If user already has a business ID, return error
	if user.BusinessID != nil {
		return errors.New("user already assigned to a business")
	}

	// Set business ID and update user
	user.BusinessID = &businessID
	return s.userRepo.Update(ctx, user)
}

// RemoveUserFromBusiness removes a user from a business
func (s *BusinessService) RemoveUserFromBusiness(ctx context.Context, userID int, businessID int) error {
	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify user belongs to the business
	if user.BusinessID == nil || *user.BusinessID != businessID {
		return errors.New("user does not belong to this business")
	}

	// Remove business ID and update user
	user.BusinessID = nil
	return s.userRepo.Update(ctx, user)
}
