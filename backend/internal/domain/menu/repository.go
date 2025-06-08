package menu

import "context"

// Repository defines the interface for menu data operations
type Repository interface {
	// Menu Items
	GetMenuItems(ctx context.Context, categoryID *int, businessID int) ([]MenuItem, error)
	GetMenuItemByID(ctx context.Context, id int, businessID int) (*MenuItem, error)
	CreateMenuItem(ctx context.Context, item MenuItemCreate) (*MenuItem, error)
	UpdateMenuItem(ctx context.Context, id int, item MenuItemUpdate) (*MenuItem, error)
	DeleteMenuItem(ctx context.Context, id int, businessID int) error

	// Categories
	GetCategories(ctx context.Context, businessID int) ([]Category, error)
	GetCategoryByID(ctx context.Context, id int, businessID int) (*Category, error)
	CreateCategory(ctx context.Context, category CategoryCreate) (*Category, error)
	UpdateCategory(ctx context.Context, id int, category CategoryUpdate) (*Category, error)
	DeleteCategory(ctx context.Context, id int, businessID int) error

	// GetDishByID retrieves a specific dish by its ID
	GetDishByID(ctx context.Context, id int) (*MenuItem, error)
}
