package order

import "errors"

var (
	// ErrOrderNotFound is returned when an order is not found
	ErrOrderNotFound = errors.New("order not found")

	// ErrInvalidOrderID is returned when an invalid order ID is provided
	ErrInvalidOrderID = errors.New("invalid order ID")

	// ErrInvalidOrderData is returned when order data validation fails
	ErrInvalidOrderData = errors.New("invalid order data")

	// ErrInvalidOrderStatus is returned when an invalid order status is provided
	ErrInvalidOrderStatus = errors.New("invalid order status")

	// ErrOrderAlreadyCompleted is returned when trying to modify a completed order
	ErrOrderAlreadyCompleted = errors.New("order is already completed")

	// ErrOrderAlreadyCancelled is returned when trying to modify a cancelled order
	ErrOrderAlreadyCancelled = errors.New("order is already cancelled")

	// ErrInvalidTableID is returned when an invalid table ID is provided
	ErrInvalidTableID = errors.New("invalid table ID")

	// ErrEmptyOrderItems is returned when trying to create an order without items
	ErrEmptyOrderItems = errors.New("order must contain at least one item")

	// ErrDishNotFound is returned when a dish is not found
	ErrDishNotFound = errors.New("dish not found")

	// ErrDishUnavailable is returned when a dish is not available
	ErrDishUnavailable = errors.New("dish is not available")

	// ErrInvalidQuantity is returned when an invalid quantity is provided
	ErrInvalidQuantity = errors.New("invalid quantity")

	// ErrOrderCreationFailed is returned when order creation fails
	ErrOrderCreationFailed = errors.New("failed to create order")

	// ErrOrderUpdateFailed is returned when order update fails
	ErrOrderUpdateFailed = errors.New("failed to update order")

	// ErrInvalidStatusTransition is returned when an invalid status transition is attempted
	ErrInvalidStatusTransition = errors.New("invalid status transition")

	// ErrDishNotAvailable is returned when a dish is not available
	ErrDishNotAvailable = errors.New("dish not available")

	// ErrOrderAlreadyExists is returned when trying to create an order that already exists
	ErrOrderAlreadyExists = errors.New("order already exists")

	// ErrTableNotAvailable is returned when a table is not available for orders
	ErrTableNotAvailable = errors.New("table not available")
)
