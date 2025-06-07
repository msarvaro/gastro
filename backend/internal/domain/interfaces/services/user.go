package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

// UserService defines the interface for user-related operations
type UserService interface {
	// User management
	GetUserByID(ctx context.Context, id int) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) (int, error)
	UpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, id int) error
	ChangeUserStatus(ctx context.Context, id int, status string) error

	// User listing and filtering
	ListUsers(ctx context.Context, businessID int, filter map[string]interface{}) ([]*entity.User, error)
	GetUsersByRole(ctx context.Context, businessID int, role string) ([]*entity.User, error)

	// Profile management
	GetUserProfile(ctx context.Context, id int) (*entity.UserProfile, error)
	UpdateUserProfile(ctx context.Context, profile *entity.UserProfile) error

	// Business associations
	AssignUserToBusiness(ctx context.Context, userID, businessID int) error
	RemoveUserFromBusiness(ctx context.Context, userID, businessID int) error

	// Activity tracking
	UpdateUserActivity(ctx context.Context, userID int) error
}
