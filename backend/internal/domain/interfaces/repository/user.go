package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type UserRepository interface {
	// Basic CRUD
	GetByID(ctx context.Context, id int) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByBusinessID(ctx context.Context, businessID int) ([]*entity.User, error)
	GetByRole(ctx context.Context, businessID int, role string) ([]*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int) error

	// Specific operations
	UpdateLastActiveAt(ctx context.Context, userID int) error
	UpdatePassword(ctx context.Context, userID int, hashedPassword string) error
	GetActiveWaiters(ctx context.Context, businessID int) ([]*entity.User, error)
}
