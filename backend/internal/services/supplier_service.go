package services

import (
	"context"
	"restaurant-management/internal/domain/consts" // Added import
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// SupplierService implements supplier service operations
type SupplierService struct {
	supplierRepo repository.SupplierRepository
}

func createSupplierService(supplierRepo repository.SupplierRepository) *SupplierService {
	return &SupplierService{
		supplierRepo: supplierRepo,
	}
}

// GetSupplierByID retrieves a supplier by ID
func (s *SupplierService) GetSupplierByID(ctx context.Context, id int) (*entity.Supplier, error) {
	return s.supplierRepo.GetByID(ctx, id)
}

// GetSuppliersByBusinessID retrieves all suppliers for a business
func (s *SupplierService) GetSuppliersByBusinessID(ctx context.Context, businessID int) ([]*entity.Supplier, error) {
	return s.supplierRepo.GetByBusinessID(ctx, businessID)
}

// GetActiveSuppliers retrieves all active suppliers for a business
func (s *SupplierService) GetActiveSuppliers(ctx context.Context, businessID int) ([]*entity.Supplier, error) {
	return s.supplierRepo.GetActive(ctx, businessID)
}

// CreateSupplier creates a new supplier
func (s *SupplierService) CreateSupplier(ctx context.Context, supplier *entity.Supplier) error {
	// Set default status if not provided
	if !supplier.IsActive {
		supplier.IsActive = true
	}
	return s.supplierRepo.Create(ctx, supplier)
}

// UpdateSupplier updates an existing supplier
func (s *SupplierService) UpdateSupplier(ctx context.Context, supplier *entity.Supplier) error {
	// Verify supplier exists
	_, err := s.supplierRepo.GetByID(ctx, supplier.ID)
	if err != nil {
		return err
	}
	return s.supplierRepo.Update(ctx, supplier)
}

// DeleteSupplier deletes a supplier
func (s *SupplierService) DeleteSupplier(ctx context.Context, id int) error {
	return s.supplierRepo.Delete(ctx, id)
}

// GetPurchaseOrderByID retrieves a purchase order by ID
func (s *SupplierService) GetPurchaseOrderByID(ctx context.Context, id int) (*entity.PurchaseOrder, error) {
	return s.supplierRepo.GetPurchaseOrderByID(ctx, id)
}

// GetPurchaseOrdersBySupplier retrieves all purchase orders for a supplier
func (s *SupplierService) GetPurchaseOrdersBySupplier(ctx context.Context, supplierID int) ([]*entity.PurchaseOrder, error) {
	return s.supplierRepo.GetPurchaseOrdersBySupplierID(ctx, supplierID)
}

// GetPendingPurchaseOrders retrieves all pending purchase orders for a business
func (s *SupplierService) GetPendingPurchaseOrders(ctx context.Context, businessID int) ([]*entity.PurchaseOrder, error) {
	return s.supplierRepo.GetPendingPurchaseOrders(ctx, businessID)
}

// CreatePurchaseOrder creates a new purchase order
func (s *SupplierService) CreatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error {
	// Set default status if not provided
	if order.Status == "" {
		order.Status = consts.PurchaseOrderStatusDraft // Changed "draft"
	}
	// Calculate total amount
	order.TotalAmount = order.CalculateTotal()
	return s.supplierRepo.CreatePurchaseOrder(ctx, order)
}

// UpdatePurchaseOrder updates an existing purchase order
func (s *SupplierService) UpdatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error {
	// Recalculate total amount
	order.TotalAmount = order.CalculateTotal()
	return s.supplierRepo.UpdatePurchaseOrder(ctx, order)
}

// UpdatePurchaseOrderStatus updates the status of a purchase order
func (s *SupplierService) UpdatePurchaseOrderStatus(ctx context.Context, id int, status string) error {
	return s.supplierRepo.UpdatePurchaseOrderStatus(ctx, id, status)
}
