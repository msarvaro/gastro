package handler

import (
	"restaurant-management/internal/domain/business"
	"restaurant-management/internal/domain/inventory"
	"restaurant-management/internal/domain/menu"
	"restaurant-management/internal/domain/notification"
	"restaurant-management/internal/domain/order"
	"restaurant-management/internal/domain/request"
	"restaurant-management/internal/domain/shift"
	"restaurant-management/internal/domain/supplier"
	"restaurant-management/internal/domain/table"
	"restaurant-management/internal/domain/user"
	"restaurant-management/internal/domain/waiter"
)

// Controllers contains all application controllers
type Controllers struct {
	Auth      *AuthController
	Admin     *AdminController
	Business  *BusinessController
	Dashboard *DashboardController
	Inventory *InventoryController
	Menu      *MenuController
	Manager   *ManagerController
	Shift     *ShiftController

	Waiter       *WaiterController
	Kitchen      *KitchenController
	Notification *NotificationController

	// Controllers now using services
	Supplier *SupplierController
	Request  *RequestController
}

// NewControllers creates a new instance of Controllers with all dependencies
func NewControllers(
	businessService business.Service,
	userService user.Service,
	menuService menu.Service,
	orderService order.Service,
	tableService table.Service,
	inventoryService inventory.Service,
	shiftService shift.Service,
	supplierService supplier.Service,
	requestService request.Service,
	waiterService waiter.Service,
	notificationService notification.Service,
) *Controllers {
	return &Controllers{
		Auth:         NewAuthController(userService),
		Admin:        NewAdminController(userService),
		Business:     NewBusinessController(businessService),
		Dashboard:    NewDashboardController(businessService, userService, orderService, tableService, inventoryService),
		Inventory:    NewInventoryController(inventoryService),
		Menu:         NewMenuController(menuService),
		Manager:      NewManagerController(orderService),
		Shift:        NewShiftController(shiftService),
		Waiter:       NewWaiterController(orderService, tableService, userService, waiterService),
		Kitchen:      NewKitchenController(orderService, inventoryService),
		Notification: NewNotificationController(notificationService),
		Supplier:     NewSupplierController(supplierService),
		Request:      NewRequestController(requestService),
	}
}
