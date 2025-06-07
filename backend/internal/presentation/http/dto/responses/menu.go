package responses

import "time"

// MenuEntityResponse represents a complete menu entity response
type MenuEntityResponse struct {
	ID          int                 `json:"id"`
	BusinessID  int                 `json:"business_id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	IsActive    bool                `json:"is_active"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Categories  []*CategoryResponse `json:"categories,omitempty"`
}

// CategoryResponse represents a category response
type CategoryResponse struct {
	ID         int             `json:"id"`
	MenuID     int             `json:"menu_id"`
	Name       string          `json:"name"`
	BusinessID int             `json:"business_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	Dishes     []*DishResponse `json:"dishes,omitempty"`
	DishCount  int             `json:"dish_count"`
}

// DishResponse represents a dish response
type DishResponse struct {
	ID              int                       `json:"id"`
	CategoryID      *int                      `json:"category_id"`
	BusinessID      int                       `json:"business_id"`
	Name            string                    `json:"name"`
	Description     string                    `json:"description"`
	Price           float64                   `json:"price"`
	ImageURL        string                    `json:"image_url"`
	IsAvailable     bool                      `json:"is_available"`
	PreparationTime int                       `json:"preparation_time"`
	Calories        int                       `json:"calories"`
	Allergens       string                    `json:"allergens"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	Options         []*MenuItemOptionResponse `json:"options,omitempty"`
	CanBeOrdered    bool                      `json:"can_be_ordered"`
}

// MenuItemOptionResponse represents a menu item option response
type MenuItemOptionResponse struct {
	ID            int     `json:"id"`
	DishID        int     `json:"dish_id"`
	Name          string  `json:"name"`
	PriceModifier float64 `json:"price_modifier"`
	IsAvailable   bool    `json:"is_available"`
}

// MenuListResponse represents a paginated list of menus
type MenuListResponse struct {
	Menus []*MenuEntityResponse `json:"menus"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// CategoryListResponse represents a list of categories
type CategoryListResponse struct {
	Categories []*CategoryResponse `json:"categories"`
	Total      int                 `json:"total"`
}

// DishListResponse represents a paginated list of dishes
type DishListResponse struct {
	Dishes []*DishResponse `json:"dishes"`
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
}

// MenuSummaryResponse represents a menu summary with statistics
type MenuSummaryResponse struct {
	Menu            *MenuEntityResponse `json:"menu"`
	TotalCategories int                 `json:"total_categories"`
	TotalDishes     int                 `json:"total_dishes"`
	AvailableDishes int                 `json:"available_dishes"`
	AveragePrice    float64             `json:"average_price"`
	PriceRange      PriceRange          `json:"price_range"`
}

// PriceRange represents the price range of dishes
type PriceRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// MenuStatsResponse represents menu statistics
type MenuStatsResponse struct {
	TotalMenus      int     `json:"total_menus"`
	ActiveMenus     int     `json:"active_menus"`
	TotalCategories int     `json:"total_categories"`
	TotalDishes     int     `json:"total_dishes"`
	AvailableDishes int     `json:"available_dishes"`
	AveragePrice    float64 `json:"average_price"`
}
