package service

import (
	"context"
	"restaurant-management/internal/domain/menu"
	"strings"
)

type MenuService struct {
	repo menu.Repository
}

func NewMenuService(repo menu.Repository) menu.Service {
	return &MenuService{repo: repo}
}

func (s *MenuService) GetMenuItems(ctx context.Context, categoryID *int, businessID int) ([]menu.MenuItem, error) {
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	return s.repo.GetMenuItems(ctx, categoryID, businessID)
}

func (s *MenuService) GetMenuItemByID(ctx context.Context, id int, businessID int) (*menu.MenuItem, error) {
	if id <= 0 {
		return nil, menu.ErrMenuItemNotFound
	}
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	return s.repo.GetMenuItemByID(ctx, id, businessID)
}

func (s *MenuService) CreateMenuItem(ctx context.Context, item menu.MenuItemCreate, businessID int) (*menu.MenuItem, error) {
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	// Validation
	if strings.TrimSpace(item.Name) == "" {
		return nil, menu.ErrInvalidMenuData
	}
	if item.CategoryID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}
	if item.Price <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	// Set business ID
	item.BusinessID = businessID

	// Verify category exists
	category, err := s.repo.GetCategoryByID(ctx, item.CategoryID, businessID)
	if err != nil || category == nil {
		return nil, menu.ErrCategoryNotFound
	}

	return s.repo.CreateMenuItem(ctx, item)
}

func (s *MenuService) UpdateMenuItem(ctx context.Context, id int, item menu.MenuItemUpdate, businessID int) (*menu.MenuItem, error) {
	if id <= 0 {
		return nil, menu.ErrMenuItemNotFound
	}
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	// Validation for provided fields
	if item.Name != "" && strings.TrimSpace(item.Name) == "" {
		return nil, menu.ErrInvalidMenuData
	}
	if item.Price > 0 && item.Price <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	// Verify menu item exists
	existing, err := s.repo.GetMenuItemByID(ctx, id, businessID)
	if err != nil || existing == nil {
		return nil, menu.ErrMenuItemNotFound
	}

	// Verify category exists if category is being updated
	if item.CategoryID > 0 {
		category, err := s.repo.GetCategoryByID(ctx, item.CategoryID, businessID)
		if err != nil || category == nil {
			return nil, menu.ErrCategoryNotFound
		}
	}

	// Set business ID
	item.BusinessID = businessID

	return s.repo.UpdateMenuItem(ctx, id, item)
}

func (s *MenuService) DeleteMenuItem(ctx context.Context, id int, businessID int) error {
	if id <= 0 {
		return menu.ErrMenuItemNotFound
	}
	if businessID <= 0 {
		return menu.ErrInvalidMenuData
	}

	// Verify menu item exists
	existing, err := s.repo.GetMenuItemByID(ctx, id, businessID)
	if err != nil || existing == nil {
		return menu.ErrMenuItemNotFound
	}

	return s.repo.DeleteMenuItem(ctx, id, businessID)
}

func (s *MenuService) GetCategories(ctx context.Context, businessID int) ([]menu.Category, error) {
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	return s.repo.GetCategories(ctx, businessID)
}

func (s *MenuService) GetCategoryByID(ctx context.Context, id int, businessID int) (*menu.Category, error) {
	if id <= 0 {
		return nil, menu.ErrCategoryNotFound
	}
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	return s.repo.GetCategoryByID(ctx, id, businessID)
}

func (s *MenuService) CreateCategory(ctx context.Context, category menu.CategoryCreate, businessID int) (*menu.Category, error) {
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	// Validation
	if strings.TrimSpace(category.Name) == "" {
		return nil, menu.ErrInvalidMenuData
	}

	// Set business ID
	category.BusinessID = businessID

	return s.repo.CreateCategory(ctx, category)
}

func (s *MenuService) UpdateCategory(ctx context.Context, id int, category menu.CategoryUpdate, businessID int) (*menu.Category, error) {
	if id <= 0 {
		return nil, menu.ErrCategoryNotFound
	}
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	// Validation
	if category.Name != "" && strings.TrimSpace(category.Name) == "" {
		return nil, menu.ErrInvalidMenuData
	}

	// Verify category exists
	existing, err := s.repo.GetCategoryByID(ctx, id, businessID)
	if err != nil || existing == nil {
		return nil, menu.ErrCategoryNotFound
	}

	// Set business ID
	category.BusinessID = businessID

	return s.repo.UpdateCategory(ctx, id, category)
}

func (s *MenuService) DeleteCategory(ctx context.Context, id int, businessID int) error {
	if id <= 0 {
		return menu.ErrCategoryNotFound
	}
	if businessID <= 0 {
		return menu.ErrInvalidMenuData
	}

	// Verify category exists
	existing, err := s.repo.GetCategoryByID(ctx, id, businessID)
	if err != nil || existing == nil {
		return menu.ErrCategoryNotFound
	}

	return s.repo.DeleteCategory(ctx, id, businessID)
}

func (s *MenuService) GetMenuSummary(ctx context.Context, businessID int) (interface{}, error) {
	if businessID <= 0 {
		return nil, menu.ErrInvalidMenuData
	}

	categories, err := s.repo.GetCategories(ctx, businessID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetMenuItems(ctx, nil, businessID)
	if err != nil {
		return nil, err
	}

	// Just return categories and items separately like the old handler
	return map[string]interface{}{
		"categories": categories,
		"items":      items,
	}, nil
}

func (s *MenuService) GetDishByID(ctx context.Context, id int) (*menu.MenuItem, error) {
	if id <= 0 {
		return nil, menu.ErrMenuItemNotFound
	}

	return s.repo.GetDishByID(ctx, id)
}
