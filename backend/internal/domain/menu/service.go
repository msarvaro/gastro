package menu

import "context"

// Service defines the menu service interface
type Service interface {
	// Menu Items
	GetMenuItems(ctx context.Context, categoryID *int, businessID int) ([]MenuItem, error)
	GetMenuItemByID(ctx context.Context, id int, businessID int) (*MenuItem, error)
	CreateMenuItem(ctx context.Context, item MenuItemCreate, businessID int) (*MenuItem, error)
	UpdateMenuItem(ctx context.Context, id int, item MenuItemUpdate, businessID int) (*MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int, businessID int) error

	// Categories
	GetCategories(ctx context.Context, businessID int) ([]Category, error)
	GetCategoryByID(ctx context.Context, id int, businessID int) (*Category, error)
	CreateCategory(ctx context.Context, category CategoryCreate, businessID int) (*Category, error)
	UpdateCategory(ctx context.Context, id int, category CategoryUpdate, businessID int) (*Category, error)
	DeleteCategory(ctx context.Context, id int, businessID int) error

	// Menu Summary
	GetMenuSummary(ctx context.Context, businessID int) (interface{}, error)

	// GetDishByID retrieves a specific dish by its ID
	GetDishByID(ctx context.Context, id int) (*MenuItem, error)
}
