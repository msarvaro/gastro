package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// InventoryService implements inventory service operations
type InventoryService struct {
	inventoryRepo repository.InventoryRepository
}

func createInventoryService(inventoryRepo repository.InventoryRepository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
	}
}

// GetInventoryItems retrieves all inventory items for a business
func (s *InventoryService) GetInventoryItems(ctx context.Context, businessID int) ([]*entity.InventoryItem, error) {
	return s.inventoryRepo.GetByBusinessID(ctx, businessID)
}

// AddInventoryItem adds a new inventory item
func (s *InventoryService) AddInventoryItem(ctx context.Context, item *entity.InventoryItem) error {
	return s.inventoryRepo.Create(ctx, item)
}

// UpdateInventoryItem updates an existing inventory item
func (s *InventoryService) UpdateInventoryItem(ctx context.Context, item *entity.InventoryItem) error {
	// Verify item exists
	_, err := s.inventoryRepo.GetByID(ctx, item.ID)
	if err != nil {
		return err
	}
	return s.inventoryRepo.Update(ctx, item)
}

// RemoveInventoryItem removes an inventory item
func (s *InventoryService) RemoveInventoryItem(ctx context.Context, itemID int) error {
	return s.inventoryRepo.Delete(ctx, itemID)
}

// UpdateStock updates the stock quantity for an inventory item
func (s *InventoryService) UpdateStock(ctx context.Context, itemID int, quantity float64, reason string) error {
	return s.inventoryRepo.UpdateStock(ctx, itemID, quantity)
}

// GetLowStockItems retrieves items that are low in stock
func (s *InventoryService) GetLowStockItems(ctx context.Context, businessID int) ([]*entity.InventoryItem, error) {
	return s.inventoryRepo.GetLowStock(ctx, businessID)
}

// GetExpiringItems retrieves items that are expiring within specified days
func (s *InventoryService) GetExpiringItems(ctx context.Context, businessID int, days int) ([]*entity.InventoryItem, error) {
	return s.inventoryRepo.GetExpiring(ctx, businessID, days)
}

// RecordStockMovement records a stock movement
func (s *InventoryService) RecordStockMovement(ctx context.Context, movement *entity.StockMovement) error {
	return s.inventoryRepo.CreateStockMovement(ctx, movement)
}

// GetStockHistory retrieves stock movement history for an item
func (s *InventoryService) GetStockHistory(ctx context.Context, itemID int) ([]*entity.StockMovement, error) {
	return s.inventoryRepo.GetStockMovements(ctx, itemID)
}
