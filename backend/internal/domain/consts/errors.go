package consts

import "errors"

// Authentication errors
var (
	ErrInvalidCredentials      = errors.New("invalid username or password")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserInactive            = errors.New("user account is inactive")
	ErrInvalidToken            = errors.New("invalid or expired token")
	ErrUnauthorized            = errors.New("unauthorized access")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)

// Business errors
var (
	ErrBusinessNotFound    = errors.New("business not found")
	ErrBusinessInactive    = errors.New("business is inactive")
	ErrBusinessExists      = errors.New("business already exists")
	ErrInvalidBusinessData = errors.New("invalid business data")
)

// User errors
var (
	ErrUsernameExists    = errors.New("username already exists")
	ErrInvalidPassword   = errors.New("invalid password format")
	ErrPasswordMismatch  = errors.New("passwords do not match")
	ErrCannotDeleteAdmin = errors.New("cannot delete admin user")
)

// Menu errors
var (
	ErrMenuNotFound      = errors.New("menu not found")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrDishNotFound      = errors.New("dish not found")
	ErrDishNotAvailable  = errors.New("dish is not available")
	ErrInvalidPrice      = errors.New("invalid price")
	ErrMenuAlreadyExists = errors.New("active menu already exists for this business")
)

// Order errors
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderNotModifiable = errors.New("order cannot be modified")
	ErrOrderAlreadyPaid   = errors.New("order is already paid")
	ErrOrderItemNotFound  = errors.New("order item not found")
	ErrInvalidOrderStatus = errors.New("invalid order status")
	ErrEmptyOrder         = errors.New("order must contain at least one item")
)

// Table errors
var (
	ErrTableNotFound       = errors.New("table not found")
	ErrTableOccupied       = errors.New("table is already occupied")
	ErrTableNotAvailable   = errors.New("table is not available")
	ErrInvalidTableNumber  = errors.New("invalid table number")
	ErrTableHasActiveOrder = errors.New("table has active orders")
)

// Shift errors
var (
	ErrShiftNotFound      = errors.New("shift not found")
	ErrShiftAlreadyActive = errors.New("user already has an active shift")
	ErrNoActiveShift      = errors.New("no active shift found")
	ErrShiftAlreadyEnded  = errors.New("shift has already ended")
	ErrCannotEndShift     = errors.New("cannot end shift with active orders")
)

// Inventory errors
var (
	ErrInventoryItemNotFound = errors.New("inventory item not found")
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrSKUExists             = errors.New("SKU already exists")
	ErrItemExpired           = errors.New("item has expired")
)

// Supplier errors
var (
	ErrSupplierNotFound      = errors.New("supplier not found")
	ErrPurchaseOrderNotFound = errors.New("purchase order not found")
	ErrOrderAlreadyDelivered = errors.New("order already delivered")
)

// Request errors
var (
	ErrRequestNotFound    = errors.New("service request not found")
	ErrRequestCompleted   = errors.New("request already completed")
	ErrInvalidRequestType = errors.New("invalid request type")
)

// General errors
var (
	ErrInvalidInput    = errors.New("invalid input data")
	ErrDatabaseError   = errors.New("database error")
	ErrInternalError   = errors.New("internal server error")
	ErrNotImplemented  = errors.New("feature not implemented")
	ErrOperationFailed = errors.New("operation failed")
)
