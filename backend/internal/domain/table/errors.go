package table

import "errors"

var (
	// ErrTableNotFound is returned when a table is not found
	ErrTableNotFound = errors.New("table not found")

	// ErrInvalidTableData is returned when table data validation fails
	ErrInvalidTableData = errors.New("invalid table data")

	// ErrTableHasActiveOrders is returned when trying to free a table that has active orders
	ErrTableHasActiveOrders = errors.New("table has active orders")

	// ErrInvalidTableStatus is returned when an invalid table status is provided
	ErrInvalidTableStatus = errors.New("invalid table status")

	// ErrTableAlreadyOccupied is returned when trying to occupy an already occupied table
	ErrTableAlreadyOccupied = errors.New("table already occupied")

	// ErrInvalidTableID is returned when an invalid table ID is provided
	ErrInvalidTableID = errors.New("invalid table ID")

	// ErrTableAlreadyReserved is returned when trying to reserve an already reserved table
	ErrTableAlreadyReserved = errors.New("table is already reserved")

	// ErrTableUpdateFailed is returned when table update fails
	ErrTableUpdateFailed = errors.New("failed to update table")
)
