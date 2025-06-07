package services

import (
	"context"
	"errors"
	"restaurant-management/internal/domain/consts"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"time"
)

// WaiterService implements the waiter service interface
type WaiterService struct {
	waiterRepo  repository.UserRepository
	tableRepo   repository.TableRepository
	orderRepo   repository.OrderRepository
	requestRepo repository.RequestRepository
}

func createWaiterService(
	waiterRepo repository.UserRepository,
	tableRepo repository.TableRepository,
	orderRepo repository.OrderRepository,
	requestRepo repository.RequestRepository,
) *WaiterService {
	return &WaiterService{
		waiterRepo:  waiterRepo,
		tableRepo:   tableRepo,
		orderRepo:   orderRepo,
		requestRepo: requestRepo,
	}
}

// GetAssignedTables retrieves all tables assigned to a waiter
func (s *WaiterService) GetAssignedTables(ctx context.Context, waiterID int) ([]*entity.Table, error) {
	// Verify waiter exists and has waiter role
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	if waiter.Role != consts.RoleWaiter {
		return nil, errors.New("user is not a waiter")
	}

	return s.tableRepo.GetTablesByWaiterID(ctx, waiterID)
}

// AssignTable assigns a table to a waiter
func (s *WaiterService) AssignTable(ctx context.Context, waiterID, tableID int) error {
	// Verify waiter exists and has waiter role
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return err
	}

	if waiter.Role != consts.RoleWaiter {
		return errors.New("user is not a waiter")
	}

	// Verify table exists
	table, err := s.tableRepo.GetByID(ctx, tableID)
	if err != nil {
		return err
	}

	// Verify table belongs to the same business as the waiter
	if waiter.BusinessID == nil || *waiter.BusinessID != table.BusinessID {
		return errors.New("waiter and table are not in the same business")
	}

	return s.tableRepo.AssignTableToWaiter(ctx, tableID, waiterID)
}

// ReleaseTable removes a waiter assignment from a table
func (s *WaiterService) ReleaseTable(ctx context.Context, tableID int) error {
	// Verify table exists
	table, err := s.tableRepo.GetByID(ctx, tableID)
	if err != nil {
		return err
	}

	// Check if table has a waiter assigned
	if table.WaiterID == nil {
		return errors.New("table has no waiter assigned")
	}

	// Unassign waiter from table
	return s.tableRepo.UnassignTableFromWaiter(ctx, tableID, *table.WaiterID)
}

// TakeOrder creates a new order for a table by a waiter
func (s *WaiterService) TakeOrder(ctx context.Context, waiterID int, order *entity.Order) error {
	// Verify waiter exists and has waiter role
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return err
	}

	if waiter.Role != consts.RoleWaiter {
		return errors.New("user is not a waiter")
	}

	// Verify table exists
	table, err := s.tableRepo.GetByID(ctx, order.TableID)
	if err != nil {
		return err
	}

	// Verify table belongs to the same business as the waiter
	if waiter.BusinessID == nil || *waiter.BusinessID != table.BusinessID {
		return errors.New("waiter and table are not in the same business")
	}

	// Set waiter ID in the order
	order.WaiterID = waiterID

	// Set business ID in the order
	order.BusinessID = *waiter.BusinessID

	// Set initial order properties
	order.Status = "new"
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Calculate total
	order.TotalAmount = order.CalculateTotal()

	// Update table status
	err = s.tableRepo.UpdateTableStatus(ctx, table.ID, "occupied")
	if err != nil {
		return err
	}

	// Create the order
	return s.orderRepo.Create(ctx, order)
}

// GetWaiterOrders retrieves all orders for a waiter
func (s *WaiterService) GetWaiterOrders(ctx context.Context, waiterID int) ([]*entity.Order, error) {
	// Verify waiter exists and has waiter role
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	if waiter.Role != consts.RoleWaiter {
		return nil, errors.New("user is not a waiter")
	}

	orders, err := s.orderRepo.GetByWaiterID(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	// Get items for each order
	for _, order := range orders {
		items, err := s.orderRepo.GetOrderItems(ctx, order.ID)
		if err == nil {
			order.Items = items
		}
	}

	return orders, nil
}

// GetActiveRequests retrieves all active service requests for a waiter
func (s *WaiterService) GetActiveRequests(ctx context.Context, waiterID int) ([]*entity.ServiceRequest, error) {
	// Verify waiter exists and has waiter role
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	if waiter.Role != consts.RoleWaiter {
		return nil, errors.New("user is not a waiter")
	}

	// Get assigned tables for this waiter
	tables, err := s.tableRepo.GetTablesByWaiterID(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	// Get requests for all assigned tables
	var requests []*entity.ServiceRequest
	for _, table := range tables {
		tableRequests, err := s.requestRepo.GetByTableID(ctx, table.ID)
		if err == nil {
			// Filter for active requests only
			for _, req := range tableRequests {
				if req.IsActive() {
					requests = append(requests, req)
				}
			}
		}
	}

	return requests, nil
}

// AcknowledgeRequest marks a service request as acknowledged
func (s *WaiterService) AcknowledgeRequest(ctx context.Context, requestID int) error {
	// Get the request
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Check if request is in pending status
	if request.Status != consts.RequestStatusPending {
		return errors.New("request is not in pending status")
	}

	// Update request status
	request.Status = consts.RequestStatusAcknowledged

	return s.requestRepo.Update(ctx, request)
}

// CompleteRequest marks a service request as completed
func (s *WaiterService) CompleteRequest(ctx context.Context, requestID int) error {
	// Get the request
	request, err := s.requestRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}

	// Check if request can be completed
	if request.Status == consts.RequestStatusCompleted {
		return errors.New("request is already completed")
	}

	// Update request status
	now := time.Now()
	request.Status = consts.RequestStatusCompleted
	request.CompletedAt = &now

	return s.requestRepo.Update(ctx, request)
}

// GetPerformanceStats retrieves performance statistics for a waiter
func (s *WaiterService) GetPerformanceStats(ctx context.Context, waiterID int, date time.Time) (*entity.WaiterStats, error) {
	// Verify waiter exists and has waiter role
	waiter, err := s.waiterRepo.GetByID(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	if waiter.Role != consts.RoleWaiter {
		return nil, errors.New("user is not a waiter")
	}

	// Get completed orders count - this is all we need for basic stats
	ordersCompleted, err := s.orderRepo.GetWaiterCompletedOrdersCount(ctx, waiterID)
	if err != nil {
		return nil, err
	}

	// Create waiter stats object
	stats := &entity.WaiterStats{
		UserID:            waiterID,
		TotalOrders:       ordersCompleted,
		TotalRevenue:      0, // Would calculate from actual orders
		AverageOrderValue: 0, // Would calculate from actual orders
		Date:              date,
		Period:            "daily", // Default period
	}

	return stats, nil
}
