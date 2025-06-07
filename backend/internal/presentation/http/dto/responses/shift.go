package responses

import "time"

// ShiftDetailResponse represents a detailed shift in API responses
type ShiftDetailResponse struct {
	ID         int                 `json:"id"`
	Date       time.Time           `json:"date"`
	StartTime  time.Time           `json:"start_time"`
	EndTime    time.Time           `json:"end_time"`
	ManagerID  int                 `json:"manager_id"`
	Manager    ShiftUserResponse   `json:"manager"`
	Notes      string              `json:"notes"`
	BusinessID int                 `json:"business_id"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
	Employees  []ShiftUserResponse `json:"employees"`
	IsActive   bool                `json:"is_active"`
	Duration   string              `json:"duration"` // Duration as string
}

// ShiftUserResponse represents a user in shift context
type ShiftUserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Email    string `json:"email"`
}

// ShiftsListResponse represents a list of shifts
type ShiftsListResponse struct {
	Shifts   []ShiftDetailResponse `json:"shifts"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// ShiftSummaryResponse represents a shift summary with statistics
type ShiftSummaryResponse struct {
	Shift        ShiftDetailResponse `json:"shift"`
	TotalOrders  int                 `json:"total_orders"`
	TotalRevenue float64             `json:"total_revenue"`
	TotalTips    float64             `json:"total_tips"`
	Performance  float64             `json:"performance"`
}

// EmployeeShiftDetailResponse represents an employee's shifts
type EmployeeShiftDetailResponse struct {
	UserID         int                   `json:"user_id"`
	Username       string                `json:"username"`
	Name           string                `json:"name"`
	CurrentShift   *ShiftDetailResponse  `json:"current_shift,omitempty"`
	UpcomingShifts []ShiftDetailResponse `json:"upcoming_shifts"`
	TotalHours     float64               `json:"total_hours_this_week"`
}

// ShiftAssignmentResponse represents the result of assigning/removing employees
type ShiftAssignmentResponse struct {
	ShiftID   int                 `json:"shift_id"`
	Employees []ShiftUserResponse `json:"employees"`
	UpdatedAt time.Time           `json:"updated_at"`
}
