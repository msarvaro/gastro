package services

import (
	"restaurant-management/internal/domain/interfaces/repository"
	"restaurant-management/internal/domain/interfaces/services"
)

// Factory is a factory for creating service instances
type Factory struct {
	userRepo      repository.UserRepository
	businessRepo  repository.BusinessRepository
	tableRepo     repository.TableRepository
	shiftRepo     repository.ShiftRepository
	orderRepo     repository.OrderRepository
	requestRepo   repository.RequestRepository
	supplierRepo  repository.SupplierRepository
	menuRepo      repository.MenuRepository
	inventoryRepo repository.InventoryRepository
	jwtSecret     string
}

// NewFactory creates a new service factory
func NewFactory(
	userRepo repository.UserRepository,
	businessRepo repository.BusinessRepository,
	tableRepo repository.TableRepository,
	shiftRepo repository.ShiftRepository,
	orderRepo repository.OrderRepository,
	requestRepo repository.RequestRepository,
	supplierRepo repository.SupplierRepository,
	menuRepo repository.MenuRepository,
	inventoryRepo repository.InventoryRepository,
	jwtSecret string,
) *Factory {
	return &Factory{
		userRepo:      userRepo,
		businessRepo:  businessRepo,
		tableRepo:     tableRepo,
		shiftRepo:     shiftRepo,
		orderRepo:     orderRepo,
		requestRepo:   requestRepo,
		supplierRepo:  supplierRepo,
		menuRepo:      menuRepo,
		inventoryRepo: inventoryRepo,
		jwtSecret:     jwtSecret,
	}
}

// NewUserService creates a new user service
func (f *Factory) NewUserService() services.UserService {
	return NewUserService(
		f.userRepo,
		f.tableRepo,
		f.shiftRepo,
		f.orderRepo,
		f.businessRepo,
	)
}

// NewAuthService creates a new auth service
func (f *Factory) NewAuthService() services.AuthService {
	return CreateAuthService(f.userRepo, f.jwtSecret)
}

// NewBusinessService creates a new business service
func (f *Factory) NewBusinessService() services.BusinessService {
	return createBusinessService(f.businessRepo, f.userRepo)
}

// NewOrderService creates a new order service
func (f *Factory) NewOrderService() services.OrderService {
	return createOrderService(f.orderRepo, f.tableRepo)
}

// NewTableService creates a new table service
func (f *Factory) NewTableService() services.TableService {
	return createTableService(f.tableRepo, f.orderRepo, f.userRepo)
}

// NewWaiterService creates a new waiter service
func (f *Factory) NewWaiterService() services.WaiterService {
	return createWaiterService(f.userRepo, f.tableRepo, f.orderRepo, f.requestRepo)
}

// NewKitchenService creates a new kitchen service
func (f *Factory) NewKitchenService() services.KitchenService {
	return createKitchenService(f.orderRepo, f.userRepo)
}

// NewMenuService creates a new menu service
func (f *Factory) NewMenuService() services.MenuService {
	return createMenuService(f.menuRepo)
}

// NewSupplierService creates a new supplier service
func (f *Factory) NewSupplierService() services.SupplierService {
	return createSupplierService(f.supplierRepo)
}

// NewRequestService creates a new request service
func (f *Factory) NewRequestService() services.RequestService {
	return createRequestService(f.requestRepo)
}

// NewShiftService creates a new shift service
func (f *Factory) NewShiftService() services.ShiftService {
	return createShiftService(f.shiftRepo, f.userRepo)
}

// NewManagerService creates a new manager service
func (f *Factory) NewManagerService() services.ManagerService {
	return createManagerService(f.userRepo, f.orderRepo, f.businessRepo)
}

// NewInventoryService creates a new inventory service
func (f *Factory) NewInventoryService() services.InventoryService {
	return createInventoryService(f.inventoryRepo)
}
