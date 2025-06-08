package service

import (
	"context"
	"log"
	"restaurant-management/internal/domain/order"
	"time"
)

type OrderService struct {
	repo order.Repository
}

func NewOrderService(repo order.Repository) order.Service {
	return &OrderService{repo: repo}
}

func (s *OrderService) GetActiveOrders(ctx context.Context, businessID int) ([]order.Order, error) {
	if businessID <= 0 {
		return nil, order.ErrInvalidOrderData
	}

	return s.repo.GetActiveOrdersWithItems(ctx, businessID)
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int, businessID int) (*order.Order, error) {
	if id <= 0 {
		return nil, order.ErrOrderNotFound
	}
	if businessID <= 0 {
		return nil, order.ErrInvalidOrderData
	}

	return s.repo.GetOrderByID(ctx, id, businessID)
}

func (s *OrderService) CreateOrder(ctx context.Context, req order.CreateOrderRequest, waiterID, businessID int) (*order.Order, error) {
	if waiterID <= 0 || businessID <= 0 {
		return nil, order.ErrInvalidOrderData
	}

	// Validation
	if req.TableID <= 0 {
		return nil, order.ErrInvalidOrderData
	}
	if len(req.Items) == 0 {
		return nil, order.ErrInvalidOrderData
	}

	// Create order object
	o := &order.Order{
		TableID:  req.TableID,
		WaiterID: waiterID,
		Status:   order.OrderStatusNew,
		Comment:  req.Comment,
		Items:    make([]order.OrderItem, len(req.Items)),
	}

	// Calculate total amount and validate items
	var totalAmount float64
	for i, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, order.ErrInvalidOrderData
		}

		// Get dish details
		dish, err := s.repo.GetDishByID(ctx, item.DishID)
		if err != nil {
			log.Printf("Error getting dish %d: %v", item.DishID, err)
			return nil, order.ErrDishNotFound
		}

		if !dish.IsAvailable {
			return nil, order.ErrDishNotAvailable
		}

		itemTotal := float64(item.Quantity) * dish.Price
		totalAmount += itemTotal

		o.Items[i] = order.OrderItem{
			DishID:   item.DishID,
			Name:     dish.Name,
			Quantity: item.Quantity,
			Price:    dish.Price,
			Total:    itemTotal,
			Notes:    item.Notes,
		}
	}

	o.TotalAmount = totalAmount

	return s.repo.CreateOrderAndItems(ctx, o, businessID)
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id int, req order.UpdateOrderStatusRequest, businessID int) error {
	if id <= 0 {
		return order.ErrOrderNotFound
	}
	if businessID <= 0 {
		return order.ErrInvalidOrderData
	}

	// Get current order
	o, err := s.repo.GetOrderByID(ctx, id, businessID)
	if err != nil {
		return order.ErrOrderNotFound
	}

	// Validate status transition
	if !s.isValidStatusTransition(o.Status, req.Status) {
		return order.ErrInvalidStatusTransition
	}

	// Update order status
	o.Status = req.Status

	// Set timestamps based on status
	now := time.Now()
	switch req.Status {
	case order.OrderStatusCompleted:
		o.CompletedAt = &now
	case order.OrderStatusCancelled:
		o.CancelledAt = &now
	}

	return s.repo.UpdateOrder(ctx, o)
}

func (s *OrderService) GetOrderStats(ctx context.Context, businessID int) (*order.OrderStats, error) {
	if businessID <= 0 {
		return nil, order.ErrInvalidOrderData
	}

	return s.repo.GetOrderStatus(ctx, businessID)
}

func (s *OrderService) GetOrderHistory(ctx context.Context, businessID int) ([]order.Order, error) {
	if businessID <= 0 {
		return nil, order.ErrInvalidOrderData
	}

	return s.repo.GetOrderHistoryWithItems(ctx, businessID)
}

func (s *OrderService) GetKitchenOrders(ctx context.Context, businessID int) ([]order.Order, error) {
	if businessID <= 0 {
		return nil, order.ErrInvalidOrderData
	}

	return s.repo.GetOrdersByStatus(ctx, string(order.OrderStatusPreparing), businessID)
}

func (s *OrderService) UpdateOrderStatusByCook(ctx context.Context, id int, req order.UpdateOrderStatusRequest, businessID int) error {
	if id <= 0 {
		return order.ErrOrderNotFound
	}
	if businessID <= 0 {
		return order.ErrInvalidOrderData
	}

	// Get current order
	o, err := s.repo.GetOrderByID(ctx, id, businessID)
	if err != nil {
		return order.ErrOrderNotFound
	}

	// Kitchen can only update from 'preparing' to 'ready'
	if o.Status != order.OrderStatusPreparing {
		return order.ErrInvalidStatusTransition
	}
	if req.Status != order.OrderStatusReady {
		return order.ErrInvalidStatusTransition
	}

	// Update order status
	o.Status = req.Status

	return s.repo.UpdateOrder(ctx, o)
}

// Helper method to validate status transitions
func (s *OrderService) isValidStatusTransition(currentStatus, newStatus order.OrderStatus) bool {
	validTransitions := map[order.OrderStatus][]order.OrderStatus{
		order.OrderStatusNew: {
			order.OrderStatusAccepted,
			order.OrderStatusCancelled,
		},
		order.OrderStatusAccepted: {
			order.OrderStatusPreparing,
			order.OrderStatusCancelled,
		},
		order.OrderStatusPreparing: {
			order.OrderStatusReady,
			order.OrderStatusCancelled,
		},
		order.OrderStatusReady: {
			order.OrderStatusServed,
			order.OrderStatusCancelled,
		},
		order.OrderStatusServed: {
			order.OrderStatusCompleted,
		},
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedStatuses {
		if allowedStatus == newStatus {
			return true
		}
	}

	return false
}
