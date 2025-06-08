package request

import "context"

// Service defines the interface for request business logic
type Service interface {
	GetAll(ctx context.Context, businessID int) ([]Request, error)
	GetByID(ctx context.Context, id int, businessID int) (*Request, error)
	Create(ctx context.Context, request CreateRequestRequest, businessID int) (*Request, error)
	Update(ctx context.Context, id int, request UpdateRequestRequest, businessID int) (*Request, error)
	Delete(ctx context.Context, id int, businessID int) error
}
