package models

import (
	"time"
)

// Category represents a menu category
type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MenuItem represents a menu item
type MenuItem struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CategoryID  int       `json:"category_id"`
	Category    Category  `json:"category,omitempty"`
	Description string    `json:"description,omitempty"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url,omitempty"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MenuItemCreate represents data needed to create a menu item
type MenuItemCreate struct {
	Name        string  `json:"name" validate:"required"`
	CategoryID  int     `json:"category_id" validate:"required"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	ImageURL    string  `json:"image_url,omitempty"`
	IsAvailable bool    `json:"is_available"`
}

// MenuItemUpdate represents data needed to update a menu item
type MenuItemUpdate struct {
	Name        string  `json:"name,omitempty"`
	CategoryID  int     `json:"category_id,omitempty"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	ImageURL    string  `json:"image_url,omitempty"`
	IsAvailable *bool   `json:"is_available,omitempty"`
}

// CategoryCreate represents data needed to create a category
type CategoryCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description,omitempty"`
}

// CategoryUpdate represents data needed to update a category
type CategoryUpdate struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
