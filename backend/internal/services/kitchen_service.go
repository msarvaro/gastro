package services

import (
	"context"
	"errors"
	"restaurant-management/internal/domain/consts"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// KitchenService implements the kitchen service interface
type KitchenService struct {
	orderRepo repository.OrderRepository
	userRepo  repository.UserRepository
}

func createKitchenService(
	orderRepo repository.OrderRepository,
	userRepo repository.UserRepository,
) *KitchenService {
	return &KitchenService{
		orderRepo: orderRepo,
		userRepo:  userRepo,
	}
}

// GetPendingOrders retrieves all pending orders for a business
func (s *KitchenService) GetPendingOrders(ctx context.Context, businessID int) ([]*entity.Order, error) {
	// Get orders with status "new" or "confirmed"
	orders, err := s.orderRepo.GetByStatus(ctx, businessID, consts.OrderStatusConfirmed)
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

// GetOrderItems retrieves all order items with a specific status
func (s *KitchenService) GetOrderItems(ctx context.Context, status string) ([]*entity.OrderItem, error) {
	// This would require a more complex implementation
	// For now, we'll just return a not implemented error
	return nil, errors.New("not implemented")
}

// StartPreparingItem marks an order item as preparing
func (s *KitchenService) StartPreparingItem(ctx context.Context, itemID int, chefID int) error {
	// Verify chef exists and has kitchen role
	chef, err := s.userRepo.GetByID(ctx, chefID)
	if err != nil {
		return err
	}

	if chef.Role != consts.RoleKitchen {
		return errors.New("user is not a kitchen staff")
	}

	// This would require a more complex implementation to manage item status
	// For now, we'll update the order item with the chef ID
	return errors.New("not implemented")
}

// MarkItemReady marks an order item as ready
func (s *KitchenService) MarkItemReady(ctx context.Context, itemID int) error {
	// This would require a more complex implementation to manage item status
	return errors.New("not implemented")
}

// GetKitchenDisplay retrieves all orders for the kitchen display
func (s *KitchenService) GetKitchenDisplay(ctx context.Context, businessID int) ([]*entity.Order, error) {
	// Get orders with relevant statuses for kitchen display
	// This typically includes confirmed, preparing, and ready orders
	confirmedOrders, _ := s.orderRepo.GetByStatus(ctx, businessID, consts.OrderStatusConfirmed)
	preparingOrders, _ := s.orderRepo.GetByStatus(ctx, businessID, consts.OrderStatusPreparing)
	readyOrders, _ := s.orderRepo.GetByStatus(ctx, businessID, consts.OrderStatusReady)

	// Combine all orders
	allOrders := make([]*entity.Order, 0)
	allOrders = append(allOrders, confirmedOrders...)
	allOrders = append(allOrders, preparingOrders...)
	allOrders = append(allOrders, readyOrders...)

	// Get items for each order
	for _, order := range allOrders {
		items, err := s.orderRepo.GetOrderItems(ctx, order.ID)
		if err == nil {
			order.Items = items
		}
	}

	return allOrders, nil
}

// GetPreparationQueue retrieves all order items in the preparation queue
func (s *KitchenService) GetPreparationQueue(ctx context.Context, businessID int) ([]*entity.OrderItem, error) {
	// This would require a more complex implementation to manage item queue
	return nil, errors.New("not implemented")
}
