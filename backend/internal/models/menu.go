package models

import (
	"time"
)

// Category represents a menu category
type Category struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	BusinessID int       `json:"business_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MenuItem represents a menu item
type MenuItem struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	CategoryID      int       `json:"category_id"`
	Category        Category  `json:"category,omitempty"`
	Price           float64   `json:"price"`
	ImageURL        string    `json:"image_url,omitempty"`
	IsAvailable     bool      `json:"is_available"`
	PreparationTime int       `json:"preparation_time,omitempty"`
	Calories        int       `json:"calories,omitempty"`
	Allergens       []string  `json:"allergens,omitempty"`
	Description     string    `json:"description,omitempty"`
	BusinessID      int       `json:"business_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// MenuItemCreate represents data needed to create a menu item
type MenuItemCreate struct {
	Name            string   `json:"name" validate:"required"`
	CategoryID      int      `json:"category_id" validate:"required"`
	Price           float64  `json:"price" validate:"required,gt=0"`
	ImageURL        string   `json:"image_url,omitempty"`
	IsAvailable     bool     `json:"is_available"`
	PreparationTime int      `json:"preparation_time,omitempty"`
	Calories        int      `json:"calories,omitempty"`
	Allergens       []string `json:"allergens,omitempty"`
	Description     string   `json:"description,omitempty"`
	BusinessID      int      `json:"business_id,omitempty"`
}

// MenuItemUpdate represents data needed to update a menu item
type MenuItemUpdate struct {
	Name            string   `json:"name,omitempty"`
	CategoryID      int      `json:"category_id,omitempty"`
	Price           float64  `json:"price,omitempty" validate:"omitempty,gt=0"`
	ImageURL        string   `json:"image_url,omitempty"`
	IsAvailable     *bool    `json:"is_available,omitempty"`
	PreparationTime int      `json:"preparation_time,omitempty"`
	Calories        int      `json:"calories,omitempty"`
	Allergens       []string `json:"allergens,omitempty"`
	Description     string   `json:"description,omitempty"`
	BusinessID      int      `json:"business_id,omitempty"`
}

// CategoryCreate represents data needed to create a category
type CategoryCreate struct {
	Name       string `json:"name" validate:"required"`
	BusinessID int    `json:"business_id,omitempty"`
}

// CategoryUpdate represents data needed to update a category
type CategoryUpdate struct {
	Name       string `json:"name,omitempty"`
	BusinessID int    `json:"business_id,omitempty"`
}
