package requests

// CreateMenuRequest represents the request to create a new menu
type CreateMenuRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

// UpdateMenuRequest represents the request to update a menu
type UpdateMenuRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	IsActive    *bool  `json:"is_active"`
}

// CreateCategoryRequest represents the request to create a new category
type CreateCategoryRequest struct {
	MenuID int    `json:"menu_id" validate:"required"`
	Name   string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

// CreateDishRequest represents the request to create a new dish
type CreateDishRequest struct {
	CategoryID      *int    `json:"category_id"`
	Name            string  `json:"name" validate:"required,min=2,max=100"`
	Description     string  `json:"description" validate:"max=1000"`
	Price           float64 `json:"price" validate:"required,min=0"`
	ImageURL        string  `json:"image_url" validate:"url"`
	IsAvailable     *bool   `json:"is_available"`
	PreparationTime int     `json:"preparation_time" validate:"min=0,max=240"`
	Calories        int     `json:"calories" validate:"min=0"`
	Allergens       string  `json:"allergens" validate:"max=500"`
}

// UpdateDishRequest represents the request to update a dish
type UpdateDishRequest struct {
	CategoryID      *int    `json:"category_id"`
	Name            string  `json:"name" validate:"required,min=2,max=100"`
	Description     string  `json:"description" validate:"max=1000"`
	Price           float64 `json:"price" validate:"required,min=0"`
	ImageURL        string  `json:"image_url" validate:"url"`
	IsAvailable     *bool   `json:"is_available"`
	PreparationTime int     `json:"preparation_time" validate:"min=0,max=240"`
	Calories        int     `json:"calories" validate:"min=0"`
	Allergens       string  `json:"allergens" validate:"max=500"`
}

// UpdateDishAvailabilityRequest represents the request to update dish availability
type UpdateDishAvailabilityRequest struct {
	IsAvailable bool `json:"is_available"`
}

// MenuItemOptionRequest represents a menu item option
type MenuItemOptionRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=100"`
	PriceModifier float64 `json:"price_modifier"`
	IsAvailable   *bool   `json:"is_available"`
}

// CreateMenuItemOptionRequest represents the request to create a menu item option
type CreateMenuItemOptionRequest struct {
	MenuItemOptionRequest
}

// UpdateMenuItemOptionRequest represents the request to update a menu item option
type UpdateMenuItemOptionRequest struct {
	MenuItemOptionRequest
}

// MenuFilterRequest represents filters for menu queries
type MenuFilterRequest struct {
	CategoryID  *int     `json:"category_id"`
	IsAvailable *bool    `json:"is_available"`
	MinPrice    *float64 `json:"min_price" validate:"min=0"`
	MaxPrice    *float64 `json:"max_price" validate:"min=0"`
	SearchTerm  string   `json:"search_term" validate:"max=100"`
}

// ReorderCategoriesRequest represents the request to reorder categories
type ReorderCategoriesRequest struct {
	CategoryIDs []int `json:"category_ids" validate:"required,min=1"`
}
