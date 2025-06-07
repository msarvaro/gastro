package repository

import (
	"context"
	"restaurant-management/internal/domain/entity"
)

type SupplierRepository interface {
	// Supplier operations
	GetByID(ctx context.Context, id int) (*entity.Supplier, error)
	GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Supplier, error)
	GetActive(ctx context.Context, businessID int) ([]*entity.Supplier, error)
	Create(ctx context.Context, supplier *entity.Supplier) error
	Update(ctx context.Context, supplier *entity.Supplier) error
	Delete(ctx context.Context, id int) error

	// Purchase order operations
	GetPurchaseOrderByID(ctx context.Context, id int) (*entity.PurchaseOrder, error)
	GetPurchaseOrdersBySupplierID(ctx context.Context, supplierID int) ([]*entity.PurchaseOrder, error)
	GetPendingPurchaseOrders(ctx context.Context, businessID int) ([]*entity.PurchaseOrder, error)
	CreatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error
	UpdatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error
	UpdatePurchaseOrderStatus(ctx context.Context, id int, status string) error
}
