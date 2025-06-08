package menu

import "errors"

var (
	// ErrMenuItemNotFound is returned when a menu item is not found
	ErrMenuItemNotFound = errors.New("menu item not found")

	// ErrCategoryNotFound is returned when a category is not found
	ErrCategoryNotFound = errors.New("category not found")

	// ErrInvalidMenuData is returned when menu data validation fails
	ErrInvalidMenuData = errors.New("invalid menu data")

	// ErrMenuItemAlreadyExists is returned when trying to create a menu item that already exists
	ErrMenuItemAlreadyExists = errors.New("menu item already exists")

	// ErrCategoryAlreadyExists is returned when trying to create a category that already exists
	ErrCategoryAlreadyExists = errors.New("category already exists")

	// ErrCategoryHasMenuItems is returned when trying to delete a category that has menu items
	ErrCategoryHasMenuItems = errors.New("cannot delete category with menu items")

	// ErrInvalidMenuItemID is returned when an invalid menu item ID is provided
	ErrInvalidMenuItemID = errors.New("invalid menu item ID")

	// ErrInvalidCategoryID is returned when an invalid category ID is provided
	ErrInvalidCategoryID = errors.New("invalid category ID")

	// ErrInvalidMenuItemData is returned when menu item data validation fails
	ErrInvalidMenuItemData = errors.New("invalid menu item data")

	// ErrInvalidCategoryData is returned when category data validation fails
	ErrInvalidCategoryData = errors.New("invalid category data")

	// ErrMenuItemUnavailable is returned when trying to access an unavailable menu item
	ErrMenuItemUnavailable = errors.New("menu item is unavailable")
)
