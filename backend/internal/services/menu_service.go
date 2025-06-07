package services

import (
	"context"
	"errors"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// MenuService implements the menu service interface
type MenuService struct {
	menuRepo repository.MenuRepository
}

func createMenuService(menuRepo repository.MenuRepository) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
	}
}

// GetMenuByBusinessID retrieves the active menu for a business
func (s *MenuService) GetMenuByBusinessID(ctx context.Context, businessID int) (*entity.Menu, error) {
	return s.menuRepo.GetActiveByBusinessID(ctx, businessID)
}

// CreateMenu creates a new menu
func (s *MenuService) CreateMenu(ctx context.Context, menu *entity.Menu) error {
	return s.menuRepo.Create(ctx, menu)
}

// UpdateMenu updates an existing menu
func (s *MenuService) UpdateMenu(ctx context.Context, menu *entity.Menu) error {
	// Verify menu exists
	_, err := s.menuRepo.GetByID(ctx, menu.ID)
	if err != nil {
		return err
	}
	return s.menuRepo.Update(ctx, menu)
}

// AddCategory adds a new category to a menu
func (s *MenuService) AddCategory(ctx context.Context, category *entity.Category) error {
	return s.menuRepo.CreateCategory(ctx, category)
}

// UpdateCategory updates an existing category
func (s *MenuService) UpdateCategory(ctx context.Context, category *entity.Category) error {
	return s.menuRepo.UpdateCategory(ctx, category)
}

// RemoveCategory removes a category from a menu
func (s *MenuService) RemoveCategory(ctx context.Context, categoryID int) error {
	return s.menuRepo.DeleteCategory(ctx, categoryID)
}

// ReorderCategories reorders categories within a menu
func (s *MenuService) ReorderCategories(ctx context.Context, menuID int, categoryIDs []int) error {
	// This would require implementing category ordering logic
	// For now, return not implemented
	return errors.New("not implemented")
}

// AddDish adds a new dish to a category
func (s *MenuService) AddDish(ctx context.Context, dish *entity.Dish) error {
	return s.menuRepo.CreateDish(ctx, dish)
}

// UpdateDish updates an existing dish
func (s *MenuService) UpdateDish(ctx context.Context, dish *entity.Dish) error {
	return s.menuRepo.UpdateDish(ctx, dish)
}

// RemoveDish removes a dish from the menu
func (s *MenuService) RemoveDish(ctx context.Context, dishID int) error {
	return s.menuRepo.DeleteDish(ctx, dishID)
}

// SetDishAvailability sets the availability status of a dish
func (s *MenuService) SetDishAvailability(ctx context.Context, dishID int, available bool) error {
	return s.menuRepo.SetDishAvailability(ctx, dishID, available)
}

// GetAvailableDishes retrieves all available dishes for a business
func (s *MenuService) GetAvailableDishes(ctx context.Context, businessID int) ([]*entity.Dish, error) {
	// Get active menu for business
	menu, err := s.menuRepo.GetActiveByBusinessID(ctx, businessID)
	if err != nil {
		return nil, err
	}

	// Get categories for the menu
	categories, err := s.menuRepo.GetCategoriesByMenuID(ctx, menu.ID)
	if err != nil {
		return nil, err
	}

	// Get all dishes for all categories and filter available ones
	var availableDishes []*entity.Dish
	for _, category := range categories {
		dishes, err := s.menuRepo.GetDishesByCategoryID(ctx, category.ID)
		if err != nil {
			continue
		}

		for _, dish := range dishes {
			if dish.IsAvailable {
				availableDishes = append(availableDishes, dish)
			}
		}
	}

	return availableDishes, nil
}
