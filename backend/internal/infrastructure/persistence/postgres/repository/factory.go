package repository

import (
	"database/sql"

	"restaurant-management/internal/domain/interfaces/repository"
)

// Factory is a factory for creating repository instances
type Factory struct {
	db *sql.DB
}

// NewFactory creates a new repository factory
func NewFactory(db *sql.DB) *Factory {
	return &Factory{
		db: db,
	}
}

// NewUserRepository creates a new user repository
func (f *Factory) NewUserRepository() repository.UserRepository {
	return NewUserRepository(f.db)
}

// NewBusinessRepository creates a new business repository
func (f *Factory) NewBusinessRepository() repository.BusinessRepository {
	return NewBusinessRepository(f.db)
}

// NewTableRepository creates a new table repository
func (f *Factory) NewTableRepository() repository.TableRepository {
	return NewTableRepository(f.db)
}

// NewShiftRepository creates a new shift repository
func (f *Factory) NewShiftRepository() repository.ShiftRepository {
	return NewShiftRepository(f.db)
}

// NewOrderRepository creates a new order repository
func (f *Factory) NewOrderRepository() repository.OrderRepository {
	return NewOrderRepository(f.db)
}

// NewMenuRepository creates a new menu repository
func (f *Factory) NewMenuRepository() repository.MenuRepository {
	return NewMenuRepository(f.db)
}

// NewSupplierRepository creates a new supplier repository
func (f *Factory) NewSupplierRepository() repository.SupplierRepository {
	return NewSupplierRepository(f.db)
}

// NewRequestRepository creates a new request repository
func (f *Factory) NewRequestRepository() repository.RequestRepository {
	return NewRequestRepository(f.db)
}

// NewInventoryRepository creates a new inventory repository
func (f *Factory) NewInventoryRepository() repository.InventoryRepository {
	return NewInventoryRepository(f.db)
}

// NewWaiterRepository creates a new waiter repository
func (f *Factory) NewWaiterRepository() repository.WaiterRepository {
	return NewWaiterRepository(f.db)
}
