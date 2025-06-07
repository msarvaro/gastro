package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

// TableService defines the interface for table-related operations
type TableService interface {
	// Table management
	GetTableByID(ctx context.Context, tableID int) (*entity.Table, error)
	GetTablesByBusinessID(ctx context.Context, businessID int) ([]*entity.Table, error)
	CreateTable(ctx context.Context, table *entity.Table) error
	UpdateTable(ctx context.Context, table *entity.Table) error
	DeleteTable(ctx context.Context, tableID int) error
	UpdateTableStatus(ctx context.Context, tableID int, status string) error

	// Waiter assignments
	AssignTableToWaiter(ctx context.Context, tableID, waiterID int) error
	UnassignTableFromWaiter(ctx context.Context, tableID, waiterID int) error
	GetTablesByWaiter(ctx context.Context, waiterID int) ([]*entity.Table, error)
	GetTablesByStatus(ctx context.Context, businessID int, status string) ([]*entity.Table, error)

	// Reservations
	CreateTableReservation(ctx context.Context, reservation *entity.TableReservation) error
	CancelTableReservation(ctx context.Context, reservationID int) error
	GetTableReservations(ctx context.Context, businessID int, date time.Time) ([]*entity.TableReservation, error)
}
