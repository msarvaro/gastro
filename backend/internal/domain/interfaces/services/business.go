package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type BusinessService interface {
	GetBusinessByID(ctx context.Context, id int) (*entity.Business, error)
	GetAllBusinesses(ctx context.Context) ([]*entity.Business, error)
	GetActiveBusinesses(ctx context.Context) ([]*entity.Business, error)
	CreateBusiness(ctx context.Context, business *entity.Business) error
	UpdateBusiness(ctx context.Context, business *entity.Business) error
	DeleteBusiness(ctx context.Context, id int) error

	// User management for business
	GetBusinessUsers(ctx context.Context, businessID int) ([]*entity.User, error)
	AddUserToBusiness(ctx context.Context, user *entity.User, businessID int) error
	RemoveUserFromBusiness(ctx context.Context, userID int, businessID int) error
}
