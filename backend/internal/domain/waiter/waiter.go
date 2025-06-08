package waiter

import (
	"time"
)

// WaiterProfile represents a waiter's complete profile information
type WaiterProfile struct {
	ID                  int                `json:"id"`
	Username            string             `json:"username"`
	Name                string             `json:"name"`
	Email               string             `json:"email"`
	CurrentShift        *ShiftInfo         `json:"current_shift,omitempty"`
	CurrentShiftManager string             `json:"current_shift_manager,omitempty"`
	UpcomingShifts      []ShiftInfo        `json:"upcoming_shifts"`
	AssignedTables      []Table            `json:"assigned_tables"`
	OrderStats          OrderStatusCounts  `json:"order_stats"`
	PerformanceData     PerformanceMetrics `json:"performance_data"`
}

// ShiftInfo contains information about a waiter's shift
type ShiftInfo struct {
	ID        int       `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsActive  bool      `json:"is_active"`
}

// OrderStatusCounts contains order counts by status
type OrderStatusCounts struct {
	New       int `json:"new"`
	Accepted  int `json:"accepted"`
	Preparing int `json:"preparing"`
	Ready     int `json:"ready"`
	Served    int `json:"served"`
	Total     int `json:"total"`
}

// PerformanceMetrics contains waiter performance metrics
type PerformanceMetrics struct {
	TablesServed       int     `json:"tables_served"`
	OrdersCompleted    int     `json:"orders_completed"`
	AverageServiceTime float64 `json:"average_service_time"` // in minutes
}

// Table represents a table entity for waiter operations
type Table struct {
	ID           int         `json:"id"`
	Number       int         `json:"number"`
	Seats        int         `json:"seats"`
	Status       string      `json:"status"`
	Orders       []OrderInfo `json:"orders,omitempty"`
	ReservedAt   *time.Time  `json:"reserved_at,omitempty"`
	OccupiedAt   *time.Time  `json:"occupied_at,omitempty"`
	CurrentOrder *int        `json:"current_order,omitempty"`
}

// OrderInfo represents simplified order information
type OrderInfo struct {
	ID      int       `json:"id"`
	Time    time.Time `json:"time"`
	Status  string    `json:"status"`
	Comment *string   `json:"comment,omitempty"`
}

// ShiftWithEmployees represents a shift with employees for waiter domain
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

// User represents a user entity for waiter operations
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}
