package table

import "context"

// TableStatusUpdateRequest represents a request to update table status
type TableStatusUpdateRequest struct {
	Status string `json:"status"` // "free", "occupied", or "reserved"
}

// Service defines the table service interface
type Service interface {
	// GetTables retrieves all tables with their current status and orders
	GetTables(ctx context.Context, businessID int) ([]Table, error)

	// GetTableByID retrieves a specific table by its ID
	GetTableByID(ctx context.Context, id int) (*Table, error)

	// UpdateTableStatus updates a table's status with business logic validation
	UpdateTableStatus(ctx context.Context, tableID int, req TableStatusUpdateRequest, businessID int) error

	// GetTableStats retrieves table statistics
	GetTableStats(ctx context.Context, businessID int) (*TableStats, error)
}
