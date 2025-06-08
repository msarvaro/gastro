package user

import (
	"context"
	"time"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GoogleLoginRequest represents a Google OAuth login request
type GoogleLoginRequest struct {
	GoogleToken string `json:"google_token"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token      string `json:"token"`
	Role       string `json:"role"`
	Redirect   string `json:"redirect"`
	BusinessID int    `json:"business_id,omitempty"`
}

// WaiterProfile represents detailed waiter profile information
type WaiterProfile struct {
	ID                  int                `json:"id"`
	Username            string             `json:"username"`
	Name                string             `json:"name"`
	Email               string             `json:"email"`
	CurrentShift        *ShiftInfo         `json:"current_shift,omitempty"`
	CurrentShiftManager string             `json:"current_shift_manager,omitempty"`
	UpcomingShifts      []ShiftInfo        `json:"upcoming_shifts,omitempty"`
	AssignedTables      []AssignedTable    `json:"assigned_tables,omitempty"`
	OrderStats          OrderStatusCounts  `json:"order_stats"`
	PerformanceData     PerformanceMetrics `json:"performance_data"`
}

// ShiftInfo contains information about a shift
type ShiftInfo struct {
	ID        int       `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsActive  bool      `json:"is_active"`
}

// AssignedTable represents a table assigned to a waiter
type AssignedTable struct {
	ID         int        `json:"id"`
	Number     int        `json:"number"`
	Seats      int        `json:"seats"`
	Status     string     `json:"status"`
	ReservedAt *time.Time `json:"reserved_at,omitempty"`
	OccupiedAt *time.Time `json:"occupied_at,omitempty"`
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

// Service defines the user service interface
type Service interface {
	// Login authenticates a user and returns a token
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)

	// GoogleLogin authenticates a user via Google OAuth and returns a token
	GoogleLogin(ctx context.Context, req GoogleLoginRequest) (*LoginResponse, error)

	// GetUserByGoogleEmail retrieves a user by Google email
	GetUserByGoogleEmail(ctx context.Context, googleEmail string) (*User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id int, businessID int) (*User, error)

	// GetUsers retrieves users by business ID
	GetUsers(ctx context.Context, businessID int) ([]User, error)

	// CreateUser creates a new user with validation
	CreateUser(ctx context.Context, user *User, businessID int) error

	// UpdateUser updates an existing user with validation
	UpdateUser(ctx context.Context, user *User) error

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id int) error

	// GetUserStats retrieves user statistics
	GetUserStats(ctx context.Context) (*UserStats, error)

	// GetStats retrieves general user statistics
	GetStats(ctx context.Context) (map[string]int, error)
}
