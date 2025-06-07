package services

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type SupplierService interface {
	// Supplier operations
	GetSupplierByID(ctx context.Context, id int) (*entity.Supplier, error)
	GetSuppliersByBusinessID(ctx context.Context, businessID int) ([]*entity.Supplier, error)
	GetActiveSuppliers(ctx context.Context, businessID int) ([]*entity.Supplier, error)
	CreateSupplier(ctx context.Context, supplier *entity.Supplier) error
	UpdateSupplier(ctx context.Context, supplier *entity.Supplier) error
	DeleteSupplier(ctx context.Context, id int) error

	// Purchase order operations
	GetPurchaseOrderByID(ctx context.Context, id int) (*entity.PurchaseOrder, error)
	GetPurchaseOrdersBySupplier(ctx context.Context, supplierID int) ([]*entity.PurchaseOrder, error)
	GetPendingPurchaseOrders(ctx context.Context, businessID int) ([]*entity.PurchaseOrder, error)
	CreatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error
	UpdatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error
	UpdatePurchaseOrderStatus(ctx context.Context, id int, status string) error
}
