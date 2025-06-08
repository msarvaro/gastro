package inventory

import "errors"

var (
	// ErrInventoryItemNotFound is returned when an inventory item is not found
	ErrInventoryItemNotFound = errors.New("inventory item not found")

	// ErrInvalidInventoryData is returned when inventory data validation fails
	ErrInvalidInventoryData = errors.New("invalid inventory data")

	// ErrInventoryItemAlreadyExists is returned when trying to create an inventory item that already exists
	ErrInventoryItemAlreadyExists = errors.New("inventory item already exists")

	// ErrInvalidInventoryID is returned when an invalid inventory ID is provided
	ErrInvalidInventoryID = errors.New("invalid inventory ID")

	// ErrInventoryUpdateFailed is returned when inventory update fails
	ErrInventoryUpdateFailed = errors.New("failed to update inventory")

	// ErrLowStock is returned when inventory stock is low
	ErrLowStock = errors.New("inventory stock is low")
)
