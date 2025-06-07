package requests

import "time"

// CreateStaffMemberRequest represents a request to create a new staff member
type CreateStaffMemberRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Role     string `json:"role" validate:"required,oneof=manager waiter kitchen cashier"`
}

// UpdateStaffMemberRequest represents a request to update a staff member
type UpdateStaffMemberRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Role     *string `json:"role,omitempty" validate:"omitempty,oneof=manager waiter kitchen cashier"`
	Status   *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
}

// ReportRequest represents a request for various reports
type ReportRequest struct {
	Type      string     `json:"type" validate:"required,oneof=daily revenue staff_performance"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Date      *time.Time `json:"date,omitempty"`
	Period    *string    `json:"period,omitempty" validate:"omitempty,oneof=daily weekly monthly"`
}

// UpdateBusinessHoursRequest represents a request to update business hours
type UpdateBusinessHoursRequest struct {
	OpenTime  string `json:"open_time" validate:"required"`
	CloseTime string `json:"close_time" validate:"required"`
}
