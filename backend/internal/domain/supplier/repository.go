package supplier

import "context"

// Repository defines the interface for supplier data operations
type Repository interface {
	GetAll(ctx context.Context, businessID int) ([]Supplier, error)
	GetByID(ctx context.Context, id int, businessID int) (*Supplier, error)
	Create(ctx context.Context, supplier CreateSupplierRequest, businessID int) (*Supplier, error)
	Update(ctx context.Context, id int, supplier UpdateSupplierRequest, businessID int) (*Supplier, error)
	Delete(ctx context.Context, id int, businessID int) error
} 