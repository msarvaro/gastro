package services

import (
	"context"
	"errors"
	"restaurant-management/internal/domain/consts"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"time"
)

// TableService implements the table service interface
type TableService struct {
	tableRepo  repository.TableRepository
	orderRepo  repository.OrderRepository
	waiterRepo repository.UserRepository
}

func createTableService(
	tableRepo repository.TableRepository,
	orderRepo repository.OrderRepository,
	waiterRepo repository.UserRepository,
) *TableService {
	return &TableService{
		tableRepo:  tableRepo,
		orderRepo:  orderRepo,
		waiterRepo: waiterRepo,
	}
}

// GetTableByID retrieves a table by ID
func (s *TableService) GetTableByID(ctx context.Context, tableID int) (*entity.Table, error) {
	return s.tableRepo.GetByID(ctx, tableID)
}

// GetTablesByBusinessID retrieves all tables for a business
func (s *TableService) GetTablesByBusinessID(ctx context.Context, businessID int) ([]*entity.Table, error) {
	return s.tableRepo.GetByBusinessID(ctx, businessID)
}

// CreateTable creates a new table
func (s *TableService) CreateTable(ctx context.Context, table *entity.Table) error {
	// Set default status if not provided
	if table.Status == "" {
		table.Status = consts.TableStatusAvailable
	}
	return s.tableRepo.Create(ctx, table)
}

// UpdateTable updates an existing table
func (s *TableService) UpdateTable(ctx context.Context, table *entity.Table) error {
	// Verify table exists
	_, err := s.tableRepo.GetByID(ctx, table.ID)
	if err != nil {
		return err
	}
	return s.tableRepo.Update(ctx, table)
}

// DeleteTable deletes a table
func (s *TableService) DeleteTable(ctx context.Context, tableID int) error {
	// Verify table exists
	table, err := s.tableRepo.GetByID(ctx, tableID)
	if err != nil {
		return err
	}

	// Check if table has active orders
	activeOrder, err := s.orderRepo.GetActiveByTableID(ctx, tableID)
	if err == nil && activeOrder != nil {
		return errors.New("cannot delete table with active orders")
	}

	// Set status to maintenance before deletion
	table.Status = consts.TableStatusMaintenance
	if err := s.tableRepo.Update(ctx, table); err != nil {
		return err
	}

	return s.tableRepo.Delete(ctx, tableID)
}

// UpdateTableStatus updates the status of a table
func (s *TableService) UpdateTableStatus(ctx context.Context, tableID int, status string) error {
	// Verify valid status
	validStatuses := map[string]bool{
		consts.TableStatusAvailable:   true,
		consts.TableStatusOccupied:    true,
		consts.TableStatusReserved:    true,
		consts.TableStatusMaintenance: true,
	}

	if !validStatuses[status] {
		return errors.New("invalid table status")
	}

	// If changing to occupied, check if there are any active orders
	if status == consts.TableStatusAvailable {
		activeOrder, err := s.orderRepo.GetActiveByTableID(ctx, tableID)
		if err == nil && activeOrder != nil {
			return errors.New("cannot set table to available with active orders")
		}
	}

	return s.tableRepo.UpdateTableStatus(ctx, tableID, status)
}

// AssignTableToWaiter assigns a table to a waiter
func (s *TableService) AssignTableToWaiter(ctx context.Context, tableID, waiterID int) error {
	// Verify table exists
	_, err := s.tableRepo.GetByID(ctx, tableID)
	if err != nil {
		return err
	}

	// Verify waiter exists and is active
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return err
	}

	if waiter.Role != consts.RoleWaiter {
		return errors.New("user is not a waiter")
	}

	if waiter.Status != consts.UserStatusActive { // Changed "active"
		return errors.New("waiter is not active")
	}

	return s.tableRepo.AssignTableToWaiter(ctx, tableID, waiterID)
}

// UnassignTableFromWaiter removes a waiter assignment from a table
func (s *TableService) UnassignTableFromWaiter(ctx context.Context, tableID, waiterID int) error {
	return s.tableRepo.UnassignTableFromWaiter(ctx, tableID, waiterID)
}

// GetTablesByWaiter retrieves all tables assigned to a waiter
func (s *TableService) GetTablesByWaiter(ctx context.Context, waiterID int) ([]*entity.Table, error) {
	// Verify waiter exists
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	if waiter.Role != consts.RoleWaiter {
		return nil, errors.New("user is not a waiter")
	}

	return s.tableRepo.GetTablesByWaiterID(ctx, waiterID)
}

// GetTablesByStatus retrieves all tables with a specific status
func (s *TableService) GetTablesByStatus(ctx context.Context, businessID int, status string) ([]*entity.Table, error) {
	return s.tableRepo.GetTablesByStatus(ctx, businessID, status)
}

// CreateTableReservation creates a reservation for a table
func (s *TableService) CreateTableReservation(ctx context.Context, reservation *entity.TableReservation) error {
	// Verify table exists
	table, err := s.tableRepo.GetByID(ctx, reservation.TableID)
	if err != nil {
		return err
	}

	// Check for overlapping reservations
	reservations, err := s.tableRepo.GetReservationsByDate(ctx, table.BusinessID, reservation.ReservationDate)
	if err != nil {
		return err
	}

	// This part would need a proper time overlap check which depends on how reservation times are stored
	// In this simplified version, we're just checking for same date reservations for the same table
	for _, r := range reservations {
		if r.TableID == reservation.TableID && r.Status == consts.ReservationStatusConfirmed { // Changed "confirmed"
			return errors.New("table already has a reservation for this date")
		}
	}

	// Set table status to reserved
	err = s.tableRepo.UpdateTableStatus(ctx, table.ID, consts.TableStatusReserved)
	if err != nil {
		return err
	}

	return s.tableRepo.CreateReservation(ctx, reservation)
}

// CancelTableReservation cancels a table reservation
func (s *TableService) CancelTableReservation(ctx context.Context, reservationID int) error {
	// Get the reservation
	reservation, err := s.tableRepo.GetReservationByID(ctx, reservationID)
	if err != nil {
		return err
	}

	// Check if the table has any other reservations for this date
	table, err := s.tableRepo.GetByID(ctx, reservation.TableID)
	if err != nil {
		return err
	}

	reservations, err := s.tableRepo.GetReservationsByDate(ctx, table.BusinessID, reservation.ReservationDate)
	if err != nil {
		return err
	}

	hasOtherReservations := false
	for _, r := range reservations {
		if r.ID != reservationID && r.TableID == reservation.TableID && r.Status == consts.ReservationStatusConfirmed { // Changed "confirmed"
			hasOtherReservations = true
			break
		}
	}

	if !hasOtherReservations {
		err = s.tableRepo.UpdateTableStatus(ctx, reservation.TableID, consts.TableStatusAvailable)
		if err != nil {
			return err
		}
	}

	return s.tableRepo.CancelReservation(ctx, reservationID)
}

// GetTableReservations retrieves all reservations for a date
func (s *TableService) GetTableReservations(ctx context.Context, businessID int, date time.Time) ([]*entity.TableReservation, error) {
	return s.tableRepo.GetReservationsByDate(ctx, businessID, date)
}
