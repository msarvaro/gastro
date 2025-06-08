package supplier

import "time"

// Supplier represents a supplier entity
type Supplier struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Categories []string  `json:"categories"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Address    string    `json:"address"`
	Status     string    `json:"status"`
	BusinessID int       `json:"business_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateSupplierRequest represents the request for creating a supplier
type CreateSupplierRequest struct {
	Name       string   `json:"name" validate:"required"`
	Categories []string `json:"categories" validate:"required"`
	Phone      string   `json:"phone"`
	Email      string   `json:"email"`
	Address    string   `json:"address"`
	Status     string   `json:"status"`
}

// UpdateSupplierRequest represents the request for updating a supplier
type UpdateSupplierRequest struct {
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	Phone      string   `json:"phone"`
	Email      string   `json:"email"`
	Address    string   `json:"address"`
	Status     string   `json:"status"`
}
