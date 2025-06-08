package shift

import "errors"

var (
	// ErrShiftNotFound is returned when a shift is not found
	ErrShiftNotFound = errors.New("shift not found")

	// ErrInvalidShiftID is returned when an invalid shift ID is provided
	ErrInvalidShiftID = errors.New("invalid shift ID")

	// ErrInvalidShiftData is returned when shift data validation fails
	ErrInvalidShiftData = errors.New("invalid shift data")

	// ErrShiftAlreadyExists is returned when trying to create a shift that already exists
	ErrShiftAlreadyExists = errors.New("shift already exists")

	// ErrShiftOverlap is returned when shifts overlap in time
	ErrShiftOverlap = errors.New("shift overlaps with existing shift")

	// ErrEmployeeNotFound is returned when an employee is not found
	ErrEmployeeNotFound = errors.New("employee not found")

	// ErrManagerNotFound is returned when a manager is not found
	ErrManagerNotFound = errors.New("manager not found")

	// ErrInvalidTimeRange is returned when start time is after end time
	ErrInvalidTimeRange = errors.New("invalid time range: start time must be before end time")
)
