package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type BusinessRepository interface {
	// Basic CRUD
	GetByID(ctx context.Context, id int) (*entity.Business, error)
	GetAll(ctx context.Context) ([]*entity.Business, error)
	Create(ctx context.Context, business *entity.Business) error
	Update(ctx context.Context, business *entity.Business) error
	Delete(ctx context.Context, id int) error

	// User operations
	GetByUserID(ctx context.Context, userID int) ([]*entity.Business, error)
	GetUserBusinesses(ctx context.Context, userID int) ([]*entity.Business, error)

	// Status operations
	UpdateStatus(ctx context.Context, id int, status string) error
	GetByStatus(ctx context.Context, status string) ([]*entity.Business, error)
}
