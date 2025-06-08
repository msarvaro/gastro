package business

import "errors"

var (
	// ErrBusinessNotFound is returned when a business is not found
	ErrBusinessNotFound = errors.New("business not found")

	// ErrInvalidBusinessID is returned when an invalid business ID is provided
	ErrInvalidBusinessID = errors.New("invalid business ID")

	// ErrInvalidBusinessData is returned when business data validation fails
	ErrInvalidBusinessData = errors.New("invalid business data")

	// ErrBusinessAlreadyExists is returned when trying to create a business that already exists
	ErrBusinessAlreadyExists = errors.New("business already exists")

	// ErrBusinessUpdateFailed is returned when business update fails
	ErrBusinessUpdateFailed = errors.New("failed to update business")

	// ErrUnauthorizedBusiness is returned when user is not authorized for the business
	ErrUnauthorizedBusiness = errors.New("unauthorized access to business")
)
