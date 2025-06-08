package business

import "context"

// Repository defines the interface for business data operations
type Repository interface {
	// CreateBusiness creates a new business
	CreateBusiness(ctx context.Context, business *Business) error

	// GetBusinessByID retrieves a business by its ID
	GetBusinessByID(ctx context.Context, id int) (*Business, error)

	// GetAllBusinesses retrieves all businesses
	GetAllBusinesses(ctx context.Context) ([]Business, error)

	// UpdateBusiness updates an existing business
	UpdateBusiness(ctx context.Context, business *Business) error

	// DeleteBusiness deletes a business by ID
	DeleteBusiness(ctx context.Context, id int) error

	// GetBusinessStats retrieves business statistics
	GetBusinessStats(ctx context.Context) (*BusinessStats, error)
}
