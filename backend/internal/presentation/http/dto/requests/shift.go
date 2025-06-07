package requests

import "time"

// CreateShiftRequest represents a request to create a new shift
type CreateShiftRequest struct {
	Date      time.Time `json:"date" validate:"required"`
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
	ManagerID int       `json:"manager_id" validate:"required"`
	Notes     string    `json:"notes"`
	Employees []int     `json:"employee_ids"`
}

// UpdateShiftRequest represents a request to update an existing shift
type UpdateShiftRequest struct {
	Date      *time.Time `json:"date,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	ManagerID *int       `json:"manager_id,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
	Employees []int      `json:"employee_ids,omitempty"`
}

// AssignEmployeeRequest represents a request to assign an employee to a shift
type AssignEmployeeRequest struct {
	EmployeeID int `json:"employee_id" validate:"required"`
}

// RemoveEmployeeRequest represents a request to remove an employee from a shift
type RemoveEmployeeRequest struct {
	EmployeeID int `json:"employee_id" validate:"required"`
}
