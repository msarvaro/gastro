package service

import (
	"context"
	"restaurant-management/internal/domain/table"
	"time"
)

type TableService struct {
	repo table.Repository
}

func NewTableService(repo table.Repository) table.Service {
	return &TableService{repo: repo}
}

func (s *TableService) GetTables(ctx context.Context, businessID int) ([]table.Table, error) {
	if businessID <= 0 {
		return nil, table.ErrInvalidTableData
	}

	return s.repo.GetAllTables(ctx, businessID)
}

func (s *TableService) GetTableByID(ctx context.Context, id int) (*table.Table, error) {
	if id <= 0 {
		return nil, table.ErrTableNotFound
	}

	return s.repo.GetTableByID(ctx, id)
}

func (s *TableService) UpdateTableStatus(ctx context.Context, tableID int, req table.TableStatusUpdateRequest, businessID int) error {
	if tableID <= 0 {
		return table.ErrTableNotFound
	}
	if businessID <= 0 {
		return table.ErrInvalidTableData
	}

	// Validate status
	validStatuses := map[string]bool{
		"free":     true,
		"occupied": true,
		"reserved": true,
	}
	if !validStatuses[req.Status] {
		return table.ErrInvalidTableData
	}

	// Get current table to check status transition rules
	currentTable, err := s.repo.GetTableByID(ctx, tableID)
	if err != nil {
		return table.ErrTableNotFound
	}

	// Business rules for status transitions
	if req.Status == "free" {
		// Check if table has active orders before marking as free
		hasActiveOrders, err := s.repo.TableHasActiveOrders(ctx, tableID)
		if err != nil {
			return err
		}
		if hasActiveOrders {
			return table.ErrTableHasActiveOrders
		}
	}

	// Update table status with appropriate timestamps
	var reservedAt, occupiedAt *time.Time
	now := time.Now()

	switch table.TableStatus(req.Status) {
	case table.TableStatusOccupied:
		occupiedAt = &now
		// Keep existing reserved_at if transitioning from reserved to occupied
		if currentTable.Status == table.TableStatusReserved {
			reservedAt = currentTable.ReservedAt
		}
	case table.TableStatusReserved:
		reservedAt = &now
		// Clear occupied_at when reserving (future reservation)
		occupiedAt = nil
	case table.TableStatusFree:
		// Clear both timestamps when table becomes free
		reservedAt = nil
		occupiedAt = nil
	}

	return s.repo.UpdateTableStatusWithTimes(ctx, tableID, req.Status, reservedAt, occupiedAt)
}

func (s *TableService) GetTableStats(ctx context.Context, businessID int) (*table.TableStats, error) {
	if businessID <= 0 {
		return nil, table.ErrInvalidTableData
	}

	return s.repo.GetTableStats(ctx, businessID)
}
