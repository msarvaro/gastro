package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type MenuRepository interface {
	// Menu operations
	GetByID(ctx context.Context, id int) (*entity.Menu, error)
	GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Menu, error)
	GetActiveByBusinessID(ctx context.Context, businessID int) (*entity.Menu, error)
	Create(ctx context.Context, menu *entity.Menu) error
	Update(ctx context.Context, menu *entity.Menu) error
	Delete(ctx context.Context, id int) error

	// Category operations
	GetCategoriesByMenuID(ctx context.Context, menuID int) ([]*entity.Category, error)
	CreateCategory(ctx context.Context, category *entity.Category) error
	UpdateCategory(ctx context.Context, category *entity.Category) error
	DeleteCategory(ctx context.Context, id int) error

	// Dish operations
	GetDishesByCategoryID(ctx context.Context, categoryID int) ([]*entity.Dish, error)
	GetDishByID(ctx context.Context, id int) (*entity.Dish, error)
	CreateDish(ctx context.Context, dish *entity.Dish) error
	UpdateDish(ctx context.Context, dish *entity.Dish) error
	DeleteDish(ctx context.Context, id int) error
	SetDishAvailability(ctx context.Context, id int, isAvailable bool) error
}
