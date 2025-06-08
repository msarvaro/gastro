package service

import (
	"context"
	"log"
	"restaurant-management/internal/domain/inventory"
	"strings"
)

type InventoryService struct {
	repo inventory.Repository
}

func NewInventoryService(repo inventory.Repository) inventory.Service {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) GetAllInventory(ctx context.Context, businessID int) ([]inventory.Inventory, error) {
	if businessID <= 0 {
		return nil, inventory.ErrInvalidInventoryData
	}

	items, err := s.repo.GetAllInventory(ctx, businessID)
	if err != nil {
		return nil, err
	}

	// Check for low stock items and log warnings
	for _, item := range items {
		if item.Quantity <= item.MinQuantity {
			log.Printf("Warning: Low stock for item %s (ID: %d). Current: %.2f, Minimum: %.2f",
				item.Name, item.ID, item.Quantity, item.MinQuantity)
		}
	}

	return items, nil
}

func (s *InventoryService) GetInventoryByID(ctx context.Context, id int, businessID int) (*inventory.Inventory, error) {
	if id <= 0 {
		return nil, inventory.ErrInventoryItemNotFound
	}
	if businessID <= 0 {
		return nil, inventory.ErrInvalidInventoryData
	}

	return s.repo.GetInventoryByID(ctx, id, businessID)
}

func (s *InventoryService) CreateInventory(ctx context.Context, item *inventory.Inventory, businessID int) error {
	if businessID <= 0 {
		return inventory.ErrInvalidInventoryData
	}

	// Validation
	if strings.TrimSpace(item.Name) == "" {
		return inventory.ErrInvalidInventoryData
	}
	if strings.TrimSpace(item.Category) == "" {
		return inventory.ErrInvalidInventoryData
	}
	if strings.TrimSpace(item.Unit) == "" {
		return inventory.ErrInvalidInventoryData
	}
	if item.Quantity < 0 {
		return inventory.ErrInvalidInventoryData
	}
	if item.MinQuantity < 0 {
		return inventory.ErrInvalidInventoryData
	}

	// Set business ID
	item.BusinessID = businessID

	// Check for low stock and log warning
	if item.Quantity <= item.MinQuantity {
		log.Printf("Warning: Creating inventory item %s with low stock. Current: %.2f, Minimum: %.2f",
			item.Name, item.Quantity, item.MinQuantity)
	}

	return s.repo.CreateInventory(ctx, item)
}

func (s *InventoryService) UpdateInventory(ctx context.Context, item *inventory.Inventory, businessID int) error {
	if item.ID <= 0 {
		return inventory.ErrInventoryItemNotFound
	}
	if businessID <= 0 {
		return inventory.ErrInvalidInventoryData
	}

	// Validation
	if strings.TrimSpace(item.Name) == "" {
		return inventory.ErrInvalidInventoryData
	}
	if strings.TrimSpace(item.Category) == "" {
		return inventory.ErrInvalidInventoryData
	}
	if strings.TrimSpace(item.Unit) == "" {
		return inventory.ErrInvalidInventoryData
	}
	if item.Quantity < 0 {
		return inventory.ErrInvalidInventoryData
	}
	if item.MinQuantity < 0 {
		return inventory.ErrInvalidInventoryData
	}

	// Verify item exists
	existing, err := s.repo.GetInventoryByID(ctx, item.ID, businessID)
	if err != nil {
		return inventory.ErrInventoryItemNotFound
	}

	// Set business ID to maintain consistency
	item.BusinessID = existing.BusinessID

	// Check for low stock and log warning
	if item.Quantity <= item.MinQuantity {
		log.Printf("Warning: Updating inventory item %s to low stock. Current: %.2f, Minimum: %.2f",
			item.Name, item.Quantity, item.MinQuantity)
	}

	return s.repo.UpdateInventory(ctx, item)
}

func (s *InventoryService) DeleteInventory(ctx context.Context, id int, businessID int) error {
	if id <= 0 {
		return inventory.ErrInventoryItemNotFound
	}
	if businessID <= 0 {
		return inventory.ErrInvalidInventoryData
	}

	// Verify item exists
	_, err := s.repo.GetInventoryByID(ctx, id, businessID)
	if err != nil {
		return inventory.ErrInventoryItemNotFound
	}

	return s.repo.DeleteInventory(ctx, id, businessID)
}
