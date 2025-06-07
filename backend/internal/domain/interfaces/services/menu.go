package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type MenuService interface {
	// Menu operations
	GetMenuByBusinessID(ctx context.Context, businessID int) (*entity.Menu, error)
	CreateMenu(ctx context.Context, menu *entity.Menu) error
	UpdateMenu(ctx context.Context, menu *entity.Menu) error

	// Category operations
	AddCategory(ctx context.Context, category *entity.Category) error
	UpdateCategory(ctx context.Context, category *entity.Category) error
	RemoveCategory(ctx context.Context, categoryID int) error
	ReorderCategories(ctx context.Context, menuID int, categoryIDs []int) error

	// Dish operations
	AddDish(ctx context.Context, dish *entity.Dish) error
	UpdateDish(ctx context.Context, dish *entity.Dish) error
	RemoveDish(ctx context.Context, dishID int) error
	SetDishAvailability(ctx context.Context, dishID int, available bool) error
	GetAvailableDishes(ctx context.Context, businessID int) ([]*entity.Dish, error)
}
