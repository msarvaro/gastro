package request

import (
	"database/sql/driver"
	"time"
)

// Request represents a request entity
type Request struct {
	ID          int        `json:"id"`
	SupplierID  int        `json:"supplier_id"`
	Items       []string   `json:"items"`
	Priority    string     `json:"priority"`
	Comment     string     `json:"comment"`
	Status      string     `json:"status"`
	BusinessID  int        `json:"business_id"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// CreateRequestRequest represents the request for creating a request
type CreateRequestRequest struct {
	SupplierID int      `json:"supplier_id" validate:"required"`
	Items      []string `json:"items" validate:"required"`
	Priority   string   `json:"priority"`
	Comment    string   `json:"comment"`
	Status     string   `json:"status"`
}

// UpdateRequestRequest represents the request for updating a request
type UpdateRequestRequest struct {
	SupplierID int      `json:"supplier_id"`
	Items      []string `json:"items"`
	Priority   string   `json:"priority"`
	Comment    string   `json:"comment"`
	Status     string   `json:"status"`
}

// Value implements the driver.Valuer interface for database serialization
func (r Request) Value() (driver.Value, error) {
	return r, nil
}
