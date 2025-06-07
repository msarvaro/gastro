package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"time"
)

type TableRepository interface {
	// Basic CRUD
	GetByID(ctx context.Context, id int) (*entity.Table, error)
	GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Table, error)
	GetByNumber(ctx context.Context, businessID, number int) (*entity.Table, error)
	Create(ctx context.Context, table *entity.Table) error
	Update(ctx context.Context, table *entity.Table) error
	Delete(ctx context.Context, id int) error

	// Table status operations
	UpdateTableStatus(ctx context.Context, id int, status string) error
	GetTablesByStatus(ctx context.Context, businessID int, status string) ([]*entity.Table, error)

	// Waiter operations
	GetTablesByWaiterID(ctx context.Context, waiterID int) ([]*entity.Table, error)
	AssignTableToWaiter(ctx context.Context, tableID, waiterID int) error
	UnassignTableFromWaiter(ctx context.Context, tableID, waiterID int) error

	// Statistics
	GetTableOccupancyRate(ctx context.Context, businessID int) (float64, error)

	// Reservation operations
	GetReservationsByDate(ctx context.Context, businessID int, date time.Time) ([]*entity.TableReservation, error)
	GetReservationByID(ctx context.Context, id int) (*entity.TableReservation, error)
	CreateReservation(ctx context.Context, reservation *entity.TableReservation) error
	UpdateReservation(ctx context.Context, reservation *entity.TableReservation) error
	CancelReservation(ctx context.Context, id int) error
}
