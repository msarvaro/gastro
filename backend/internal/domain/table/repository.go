package table

import (
	"context"
	"time"
)

// Repository defines the interface for table data operations
type Repository interface {
	// GetAllTables retrieves all tables with their order information
	GetAllTables(ctx context.Context, businessID int) ([]Table, error)

	// GetTableByID retrieves a table by its ID
	GetTableByID(ctx context.Context, id int) (*Table, error)

	// GetTableStats retrieves table statistics
	GetTableStats(ctx context.Context, businessID int) (*TableStats, error)

	// UpdateTableStatus updates a table's status
	UpdateTableStatus(ctx context.Context, tableID int, status string) error

	// UpdateTableStatusWithTimes updates a table's status and timestamp fields
	UpdateTableStatusWithTimes(ctx context.Context, tableID int, status string, reservedAt, occupiedAt *time.Time) error

	// TableHasActiveOrders checks if a table has any active orders
	TableHasActiveOrders(ctx context.Context, tableID int) (bool, error)
}
