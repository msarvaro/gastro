package shift

import "time"

// Shift represents a work shift entity
type Shift struct {
	ID         int       `json:"id"`
	Date       time.Time `json:"date"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	ManagerID  int       `json:"manager_id"`
	BusinessID int       `json:"business_id"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// User represents a simplified user entity for shifts
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// ShiftWithEmployees represents a shift with associated employees and manager
type ShiftWithEmployees struct {
	ID         int       `json:"id"`
	Date       time.Time `json:"date"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	ManagerID  int       `json:"manager_id"`
	BusinessID int       `json:"business_id"`
	Manager    *User     `json:"manager,omitempty"`
	Notes      string    `json:"notes"`
	Employees  []User    `json:"employees"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ShiftEmployee represents the assignment of an employee to a shift
type ShiftEmployee struct {
	ID         int       `json:"id"`
	ShiftID    int       `json:"shift_id"`
	BusinessID int       `json:"business_id"`
	EmployeeID int       `json:"employee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateShiftRequest represents data for creating a shift
type CreateShiftRequest struct {
	Date        string `json:"date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	ManagerID   string `json:"manager_id"`
	Notes       string `json:"notes"`
	EmployeeIDs []int  `json:"employee_ids"`
}

// UpdateShiftRequest represents data for updating a shift
type UpdateShiftRequest struct {
	Date        string `json:"date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	ManagerID   string `json:"manager_id"`
	Notes       string `json:"notes"`
	EmployeeIDs []int  `json:"employee_ids"`
}

// ShiftResponse represents a shift response
type ShiftResponse struct {
	ID        int       `json:"id"`
	Date      string    `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Manager   User      `json:"manager"`
	Notes     string    `json:"notes"`
	Employees []User    `json:"employees"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
