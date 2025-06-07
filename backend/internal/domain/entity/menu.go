package entity

import "time"

// Menu represents a restaurant menu
type Menu struct {
	ID          int
	BusinessID  int
	Name        string
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Categories  []*Category // Categories in this menu
}

// Category represents a menu category
type Category struct {
	ID         int
	MenuID     int
	Name       string
	BusinessID int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Dishes     []*Dish // Dishes in this category
}

// Dish represents a menu item/dish
type Dish struct {
	ID              int
	CategoryID      *int
	BusinessID      int
	Name            string
	Description     string
	Price           float64
	ImageURL        string
	IsAvailable     bool
	PreparationTime int // in minutes
	Calories        int
	Allergens       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Options         []*MenuItemOption // Available options for this dish
}

// MenuItemOption represents an option for a menu item
type MenuItemOption struct {
	ID            int
	DishID        int
	Name          string
	PriceModifier float64
	IsAvailable   bool
}

// GetFullPrice returns the price with modifiers
func (d *Dish) GetFullPrice(selectedOptions []*MenuItemOption) float64 {
	price := d.Price

	for _, option := range selectedOptions {
		price += option.PriceModifier
	}

	return price
}

// Business methods
func (d *Dish) CanBeOrdered() bool {
	return d.IsAvailable && d.Price > 0
}
