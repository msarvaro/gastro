package services

import (
	"context"
	"errors"
	"restaurant-management/internal/domain/consts"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"time"
)

// OrderService implements the order service interface
type OrderService struct {
	orderRepo repository.OrderRepository
	tableRepo repository.TableRepository
}

func createOrderService(
	orderRepo repository.OrderRepository,
	tableRepo repository.TableRepository,
) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
		tableRepo: tableRepo,
	}
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(ctx context.Context, order *entity.Order) error {
	// Validate table exists and is available
	table, err := s.tableRepo.GetByID(ctx, order.TableID)
	if err != nil {
		return err
	}

	// Check if table is occupied or reserved
	if table.Status != consts.TableStatusAvailable && table.Status != consts.TableStatusOccupied {
		return errors.New("table is not available for orders")
	}

	// Set initial order properties
	order.Status = consts.OrderStatusPending
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	// Calculate initial total
	order.TotalAmount = order.CalculateTotal()

	// Mark table as occupied
	err = s.tableRepo.UpdateTableStatus(ctx, table.ID, consts.TableStatusOccupied)
	if err != nil {
		return err
	}

	// Create the order
	return s.orderRepo.Create(ctx, order)
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(ctx context.Context, orderID int) (*entity.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Get order items
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	order.Items = items
	return order, nil
}

// GetActiveOrders retrieves all active orders for a business
func (s *OrderService) GetActiveOrders(ctx context.Context, businessID int) ([]*entity.Order, error) {
	// Get orders that are not completed or cancelled
	// We'll get all orders and filter them
	orders, err := s.orderRepo.GetByBusinessID(ctx, businessID, 100, 0)
	if err != nil {
		return nil, err
	}

	activeOrders := make([]*entity.Order, 0)
	for _, order := range orders {
		if order.IsActive() {
			// Get order items for each active order
			items, err := s.orderRepo.GetOrderItems(ctx, order.ID)
			if err == nil {
				order.Items = items
			}
			activeOrders = append(activeOrders, order)
		}
	}

	return activeOrders, nil
}

// GetOrdersByTable retrieves all orders for a specific table
func (s *OrderService) GetOrdersByTable(ctx context.Context, tableID int) ([]*entity.Order, error) {
	// Validate table exists
	_, err := s.tableRepo.GetByID(ctx, tableID)
	if err != nil {
		return nil, err
	}

	orders, err := s.orderRepo.GetByTableID(ctx, tableID)
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

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Update the status
	order.UpdateStatus(status)

	// If the order is completed or cancelled, free up the table
	if order.IsCompleted() || order.IsCancelled() {
		err = s.tableRepo.UpdateTableStatus(ctx, order.TableID, consts.TableStatusAvailable)
		if err != nil {
			return err
		}
	}

	return s.orderRepo.Update(ctx, order)
}

// AddItemToOrder adds an item to an order
func (s *OrderService) AddItemToOrder(ctx context.Context, orderID int, item *entity.OrderItem) error {
	// Ensure the order exists and can be modified
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.CanBeModified() {
		return errors.New("order cannot be modified")
	}

	// Set order ID and current time
	item.OrderID = orderID
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()
	item.BusinessID = order.BusinessID

	// Add the item
	err = s.orderRepo.AddOrderItem(ctx, orderID, item)
	if err != nil {
		return err
	}

	// Recalculate order total
	order.Items, _ = s.orderRepo.GetOrderItems(ctx, orderID)
	order.TotalAmount = order.CalculateTotal()
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// UpdateOrderItem updates an item in an order
func (s *OrderService) UpdateOrderItem(ctx context.Context, item *entity.OrderItem) error {
	// Ensure the order exists and can be modified
	order, err := s.orderRepo.GetByID(ctx, item.OrderID)
	if err != nil {
		return err
	}

	if !order.CanBeModified() {
		return errors.New("order cannot be modified")
	}

	// Update the item
	item.UpdatedAt = time.Now()
	err = s.orderRepo.UpdateOrderItem(ctx, item)
	if err != nil {
		return err
	}

	// Recalculate order total
	order.Items, _ = s.orderRepo.GetOrderItems(ctx, order.ID)
	order.TotalAmount = order.CalculateTotal()
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// RemoveItemFromOrder removes an item from an order
func (s *OrderService) RemoveItemFromOrder(ctx context.Context, orderID, itemID int) error {
	// Ensure the order exists and can be modified
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	if !order.CanBeModified() {
		return errors.New("order cannot be modified")
	}

	// Remove the item
	err = s.orderRepo.RemoveOrderItem(ctx, orderID, itemID)
	if err != nil {
		return err
	}

	// Recalculate order total
	order.Items, _ = s.orderRepo.GetOrderItems(ctx, orderID)
	order.TotalAmount = order.CalculateTotal()
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// UpdateItemStatus updates the status of an item in an order
func (s *OrderService) UpdateItemStatus(ctx context.Context, orderID, itemID int, status string) error {
	// This method would typically require a more complex implementation
	// with a separate status field for each order item
	// For simplicity, we'll just update the order item
	items, err := s.orderRepo.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}

	var item *entity.OrderItem
	for _, i := range items {
		if i.ID == itemID {
			item = i
			break
		}
	}

	if item == nil {
		return errors.New("item not found in order")
	}

	// In a real implementation, we would update the status field of the item
	// item.Status = status

	// Since there's no status field in the current entity, we'll just update the item
	item.UpdatedAt = time.Now()
	return s.orderRepo.UpdateOrderItem(ctx, item)
}

// CalculateOrderTotal calculates the total for an order
func (s *OrderService) CalculateOrderTotal(ctx context.Context, orderID int) (float64, error) {
	order, err := s.GetOrderByID(ctx, orderID)
	if err != nil {
		return 0, err
	}

	totalAmount := order.CalculateTotal()

	// Update the order with the new total
	order.TotalAmount = totalAmount
	order.UpdatedAt = time.Now()

	err = s.orderRepo.Update(ctx, order)
	if err != nil {
		return 0, err
	}

	return float64(totalAmount), nil
}

// CompleteOrder marks an order as completed
func (s *OrderService) CompleteOrder(ctx context.Context, orderID int) error {
	return s.UpdateOrderStatus(ctx, orderID, consts.OrderStatusPaid)
}

// CancelOrder cancels an order
func (s *OrderService) CancelOrder(ctx context.Context, orderID int, reason string) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	// Add reason to comment
	if reason != "" {
		if order.Comment != "" {
			order.Comment += " | "
		}
		order.Comment += "Cancelled: " + reason
	}

	order.UpdateStatus(consts.OrderStatusCanceled)

	// Free up the table
	err = s.tableRepo.UpdateTableStatus(ctx, order.TableID, consts.TableStatusAvailable)
	if err != nil {
		return err
	}

	return s.orderRepo.Update(ctx, order)
}
