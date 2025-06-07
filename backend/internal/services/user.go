package services

import (
	"context"
	"errors"
	"time"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"restaurant-management/internal/domain/interfaces/services"

	"golang.org/x/crypto/bcrypt"
)

// userService implements the UserService interface
type userService struct {
	userRepo     repository.UserRepository
	tableRepo    repository.TableRepository
	shiftRepo    repository.ShiftRepository
	orderRepo    repository.OrderRepository
	businessRepo repository.BusinessRepository
}

// NewUserService creates a new instance of the user service
func NewUserService(
	userRepo repository.UserRepository,
	tableRepo repository.TableRepository,
	shiftRepo repository.ShiftRepository,
	orderRepo repository.OrderRepository,
	businessRepo repository.BusinessRepository,
) services.UserService {
	return &userService{
		userRepo:     userRepo,
		tableRepo:    tableRepo,
		shiftRepo:    shiftRepo,
		orderRepo:    orderRepo,
		businessRepo: businessRepo,
	}
}

// GetUserByID retrieves a user by their ID
func (s *userService) GetUserByID(ctx context.Context, id int) (*entity.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// GetUserByUsername retrieves a user by their username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// GetUserByEmail retrieves a user by their email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	// Implementation would depend on a new repository method
	// For now, we can return an error
	return nil, errors.New("not implemented")
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, user *entity.User) (int, error) {
	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	return user.ID, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, user *entity.User) error {
	// Get the existing user to verify it exists
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not found")
	}

	// If password is provided, hash it
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	} else {
		// Keep the existing password
		user.Password = existingUser.Password
	}

	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

// DeleteUser deletes a user by their ID
func (s *userService) DeleteUser(ctx context.Context, id int) error {
	return s.userRepo.Delete(ctx, id)
}

// ChangeUserStatus changes a user's status (active/inactive)
func (s *userService) ChangeUserStatus(ctx context.Context, id int, status string) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	user.Status = status
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// ListUsers lists users based on filters
func (s *userService) ListUsers(ctx context.Context, businessID int, filter map[string]interface{}) ([]*entity.User, error) {
	// Basic implementation that can be extended based on filters
	if role, ok := filter["role"].(string); ok && role != "" {
		return s.userRepo.GetByRole(ctx, businessID, role)
	}
	return s.userRepo.GetByBusinessID(ctx, businessID)
}

// GetUsersByRole retrieves users by their role
func (s *userService) GetUsersByRole(ctx context.Context, businessID int, role string) ([]*entity.User, error) {
	return s.userRepo.GetByRole(ctx, businessID, role)
}

// GetUserProfile retrieves a user's profile with additional data
func (s *userService) GetUserProfile(ctx context.Context, id int) (*entity.UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	profile := &entity.UserProfile{
		User: user,
	}

	// If this is a waiter, get assigned tables
	if user.Role == "waiter" && user.BusinessID != nil {
		// This would require a repository method to get tables assigned to a waiter
		assignedTables, err := s.tableRepo.GetTablesByWaiterID(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		profile.AssignedTables = assignedTables

		// Get current shift
		currentShift, err := s.shiftRepo.GetCurrentShiftForEmployee(ctx, user.ID)
		if err != nil {
			// Log the error but continue
			// We don't want to fail the whole profile load if just one part fails
		} else if currentShift != nil {
			profile.CurrentShift = currentShift

			// Get the manager for the shift
			manager, err := s.userRepo.GetByID(ctx, currentShift.ManagerID)
			if err == nil && manager != nil {
				profile.CurrentShiftManager = manager.Name
			}
		}

		// Get upcoming shifts
		upcomingShifts, err := s.shiftRepo.GetUpcomingShiftsForEmployee(ctx, user.ID)
		if err == nil {
			profile.UpcomingShifts = upcomingShifts
		}

		// Get order statistics
		orderStats, err := s.orderRepo.GetWaiterOrderStatistics(ctx, user.ID)
		if err == nil {
			profile.OrderStats = orderStats
		}

		// Performance data - this might be calculated or retrieved from a reporting service
		performanceData := make(map[string]int)
		tablesServed, _ := s.orderRepo.GetWaiterTablesServedCount(ctx, user.ID)
		ordersCompleted, _ := s.orderRepo.GetWaiterCompletedOrdersCount(ctx, user.ID)

		performanceData["tables_served"] = tablesServed
		performanceData["orders_completed"] = ordersCompleted
		profile.PerformanceData = performanceData
	}

	return profile, nil
}

// UpdateUserProfile updates a user's profile
func (s *userService) UpdateUserProfile(ctx context.Context, profile *entity.UserProfile) error {
	if profile.User == nil {
		return errors.New("user is required")
	}

	return s.UpdateUser(ctx, profile.User)
}

// AssignUserToBusiness assigns a user to a business
func (s *userService) AssignUserToBusiness(ctx context.Context, userID, businessID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if business exists
	business, err := s.businessRepo.GetByID(ctx, businessID)
	if err != nil {
		return err
	}
	if business == nil {
		return errors.New("business not found")
	}

	user.BusinessID = &businessID
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// RemoveUserFromBusiness removes a user from a business
func (s *userService) RemoveUserFromBusiness(ctx context.Context, userID, businessID int) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Check if user is assigned to the specified business
	if user.BusinessID == nil || *user.BusinessID != businessID {
		return errors.New("user is not assigned to this business")
	}

	// Remove business association
	user.BusinessID = nil
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// UpdateUserActivity updates a user's last active timestamp
func (s *userService) UpdateUserActivity(ctx context.Context, userID int) error {
	return s.userRepo.UpdateLastActiveAt(ctx, userID)
}
