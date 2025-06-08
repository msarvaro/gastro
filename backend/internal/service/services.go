package service

import (
	"restaurant-management/internal/domain/business"
	"restaurant-management/internal/domain/inventory"
	"restaurant-management/internal/domain/menu"
	"restaurant-management/internal/domain/order"
	"restaurant-management/internal/domain/request"
	"restaurant-management/internal/domain/shift"
	"restaurant-management/internal/domain/supplier"
	"restaurant-management/internal/domain/table"
	"restaurant-management/internal/domain/user"
	"restaurant-management/internal/domain/waiter"
)

// Services contains all application services
type Services struct {
	Business  business.Service
	User      user.Service
	Menu      menu.Service
	Order     order.Service
	Table     table.Service
	Inventory inventory.Service
	Shift     shift.Service
	Supplier  supplier.Service
	Request   request.Service
	Waiter    waiter.Service
}

// NewServices creates a new instance of Services with all dependencies
func NewServices(
	businessRepo business.Repository,
	userRepo user.Repository,
	menuRepo menu.Repository,
	orderRepo order.Repository,
	tableRepo table.Repository,
	inventoryRepo inventory.Repository,
	shiftRepo shift.Repository,
	supplierRepo supplier.Repository,
	requestRepo request.Repository,
	waiterRepo waiter.Repository,
	jwtKey string,
) *Services {
	return &Services{
		Business:  NewBusinessService(businessRepo),
		User:      NewUserService(userRepo, jwtKey),
		Menu:      NewMenuService(menuRepo),
		Order:     NewOrderService(orderRepo),
		Table:     NewTableService(tableRepo),
		Inventory: NewInventoryService(inventoryRepo),
		Shift:     NewShiftService(shiftRepo),
		Supplier:  NewSupplierService(supplierRepo),
		Request:   NewRequestService(requestRepo),
		Waiter:    NewWaiterService(waiterRepo),
	}
}
