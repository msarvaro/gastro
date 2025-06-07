package responses

import "time"

// StaffMemberResponse represents a staff member in API responses
type StaffMemberResponse struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	BusinessID   *int       `json:"business_id,omitempty"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// StaffListResponse represents a list of staff members
type StaffListResponse struct {
	Staff    []StaffMemberResponse `json:"staff"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// DailyReportResponse represents a daily business report
type DailyReportResponse struct {
	Date              time.Time                  `json:"date"`
	TotalOrders       int                        `json:"total_orders"`
	TotalRevenue      float64                    `json:"total_revenue"`
	AverageOrderValue float64                    `json:"average_order_value"`
	TotalCustomers    int                        `json:"total_customers"`
	PeakHour          string                     `json:"peak_hour"`
	TopDishes         []TopDishResponse          `json:"top_dishes"`
	StaffPerformance  []StaffPerformanceResponse `json:"staff_performance"`
	TableTurnover     float64                    `json:"table_turnover"`
}

// TopDishResponse represents a top-selling dish
type TopDishResponse struct {
	DishID   int     `json:"dish_id"`
	DishName string  `json:"dish_name"`
	Quantity int     `json:"quantity"`
	Revenue  float64 `json:"revenue"`
}

// StaffPerformanceResponse represents staff performance metrics
type StaffPerformanceResponse struct {
	UserID          int     `json:"user_id"`
	Username        string  `json:"username"`
	Name            string  `json:"name"`
	Role            string  `json:"role"`
	OrdersServed    int     `json:"orders_served"`
	Revenue         float64 `json:"revenue"`
	CustomerRating  float64 `json:"customer_rating"`
	EfficiencyScore float64 `json:"efficiency_score"`
}

// RevenueReportResponse represents a revenue report
type RevenueReportResponse struct {
	StartDate         time.Time                 `json:"start_date"`
	EndDate           time.Time                 `json:"end_date"`
	TotalRevenue      float64                   `json:"total_revenue"`
	DailyBreakdown    []DailyRevenueResponse    `json:"daily_breakdown"`
	CategoryBreakdown []CategoryRevenueResponse `json:"category_breakdown"`
	Growth            float64                   `json:"growth_percentage"`
}

// DailyRevenueResponse represents daily revenue data
type DailyRevenueResponse struct {
	Date    time.Time `json:"date"`
	Revenue float64   `json:"revenue"`
	Orders  int       `json:"orders"`
}

// CategoryRevenueResponse represents revenue by category
type CategoryRevenueResponse struct {
	Category   string  `json:"category"`
	Revenue    float64 `json:"revenue"`
	Percentage float64 `json:"percentage"`
}

// BusinessStatisticsResponse represents overall business statistics
type BusinessStatisticsResponse struct {
	TotalStaff      int     `json:"total_staff"`
	ActiveStaff     int     `json:"active_staff"`
	TotalTables     int     `json:"total_tables"`
	OccupiedTables  int     `json:"occupied_tables"`
	TodayOrders     int     `json:"today_orders"`
	TodayRevenue    float64 `json:"today_revenue"`
	MonthlyRevenue  float64 `json:"monthly_revenue"`
	AverageRating   float64 `json:"average_rating"`
	PendingRequests int     `json:"pending_requests"`
}
