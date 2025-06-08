package business

import "context"

// Service defines the business service interface
type Service interface {
	// CreateBusiness creates a new business with validation
	CreateBusiness(ctx context.Context, business *Business) error

	// GetBusinessByID retrieves a business by its ID
	GetBusinessByID(ctx context.Context, id int) (*Business, error)

	// GetAllBusinesses retrieves all businesses with statistics
	GetAllBusinesses(ctx context.Context) ([]Business, *BusinessStats, error)

	// UpdateBusiness updates an existing business with validation
	UpdateBusiness(ctx context.Context, business *Business) error

	// DeleteBusiness deletes a business by ID
	DeleteBusiness(ctx context.Context, id int) error

	// SetBusinessCookie sets a business cookie for user session
	SetBusinessCookie(ctx context.Context, businessID int) error
}
