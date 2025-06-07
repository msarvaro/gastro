package responses

// TableResponse represents a table entity with enhanced formatting
type TableResponse struct {
	ID               int                    `json:"id"`
	Number           int                    `json:"number"`
	Seats            int                    `json:"seats"`
	Status           StatusResponse         `json:"status"`
	AssignedWaiter   *string                `json:"assigned_waiter"`
	Orders           []*OrderShortResponse  `json:"orders"`
	LastActivity     *FormattedDateResponse `json:"last_activity"`
	OccupiedDuration *string                `json:"occupied_duration"` // "2ч 15м"
	AvailableActions []ActionResponse       `json:"available_actions"`
	Revenue          *MoneyResponse         `json:"revenue"` // Today's revenue for this table
}

// TablesListResponse represents a list of tables with backend calculations
type TablesListResponse struct {
	Tables      []*TableResponse       `json:"tables"`
	Stats       TablesStatsResponse    `json:"stats"`
	Filters     []FilterOptionResponse `json:"filters"`
	Actions     []ActionResponse       `json:"actions"`
	LastUpdated FormattedDateResponse  `json:"last_updated"`
}

// TablesStatsResponse represents statistics about tables with calculations
type TablesStatsResponse struct {
	Total          int           `json:"total"`
	Free           int           `json:"free"`
	Occupied       int           `json:"occupied"`
	Reserved       int           `json:"reserved"`
	OccupancyRate  float64       `json:"occupancy_rate"`
	AverageRevenue MoneyResponse `json:"average_revenue"`
	TurnoverRate   string        `json:"turnover_rate"` // "2.3 оборота/день"
}

// OrderShortResponse represents a short order information for table display
type OrderShortResponse struct {
	ID     int                   `json:"id"`
	Status StatusResponse        `json:"status"`
	Time   FormattedDateResponse `json:"time"`
}

// OrderItemResponse represents an item in an order with enhanced formatting
type OrderItemResponse struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Quantity   int            `json:"quantity"`
	Price      MoneyResponse  `json:"price"`
	TotalPrice MoneyResponse  `json:"total_price"`
	Notes      string         `json:"notes"`
	Status     StatusResponse `json:"status"`
}

// OrderResponse represents an order entity with enhanced formatting
type OrderResponse struct {
	ID               int                    `json:"id"`
	TableNumber      int                    `json:"table_number"`
	Status           StatusResponse         `json:"status"`
	TotalAmount      MoneyResponse          `json:"total_amount"`
	Comment          string                 `json:"comment"`
	Items            []*OrderItemResponse   `json:"items"`
	CreatedAt        FormattedDateResponse  `json:"created_at"`
	UpdatedAt        FormattedDateResponse  `json:"updated_at"`
	Duration         string                 `json:"duration"` // "15 минут назад"
	EstimatedReady   *FormattedDateResponse `json:"estimated_ready"`
	AvailableActions []ActionResponse       `json:"available_actions"`
	Priority         string                 `json:"priority"` // "high", "normal", "low"
}

// OrdersListResponse represents a list of orders with backend calculations
type OrdersListResponse struct {
	Orders      []*OrderResponse       `json:"orders"`
	Stats       OrdersStatsResponse    `json:"stats"`
	Filters     []FilterOptionResponse `json:"filters"`
	Actions     []ActionResponse       `json:"actions"`
	LastUpdated FormattedDateResponse  `json:"last_updated"`
}

// OrdersStatsResponse represents statistics about orders with calculations
type OrdersStatsResponse struct {
	TotalActiveOrders int           `json:"total_active_orders"`
	New               int           `json:"new"`
	Accepted          int           `json:"accepted"`
	Preparing         int           `json:"preparing"`
	Ready             int           `json:"ready"`
	Served            int           `json:"served"`
	TotalRevenue      MoneyResponse `json:"total_revenue"`
	AverageOrderValue MoneyResponse `json:"average_order_value"`
	PendingTime       string        `json:"pending_time"` // "Средне ожидание: 12 мин"
}

// OrderHistoryResponse represents an order in history with enhanced formatting
type OrderHistoryResponse struct {
	ID                   int                    `json:"id"`
	TableNumber          int                    `json:"table_number"`
	Status               StatusResponse         `json:"status"`
	TotalAmount          MoneyResponse          `json:"total_amount"`
	Items                []*OrderItemResponse   `json:"items"`
	CreatedAt            FormattedDateResponse  `json:"created_at"`
	CompletedAt          *FormattedDateResponse `json:"completed_at"`
	CancelledAt          *FormattedDateResponse `json:"cancelled_at"`
	Duration             string                 `json:"duration"`              // "Обслужен за 25 минут"
	CustomerSatisfaction *int                   `json:"customer_satisfaction"` // 1-5 rating if available
}

// OrderHistoryListResponse represents a list of order history with enhanced formatting
type OrderHistoryListResponse struct {
	Orders      []*OrderHistoryResponse `json:"orders"`
	Stats       HistoryStatsResponse    `json:"stats"`
	Filters     []FilterOptionResponse  `json:"filters"`
	Actions     []ActionResponse        `json:"actions"`
	LastUpdated FormattedDateResponse   `json:"last_updated"`
	Pagination  PaginationResponse      `json:"pagination"`
}

// HistoryStatsResponse represents statistics about order history with calculations
type HistoryStatsResponse struct {
	CompletedTotal       int           `json:"completed_total"`
	CancelledTotal       int           `json:"cancelled_total"`
	CompletedAmountTotal MoneyResponse `json:"completed_amount_total"`
	CancelledAmountTotal MoneyResponse `json:"cancelled_amount_total"`
	SuccessRate          float64       `json:"success_rate"`
	AverageServiceTime   string        `json:"average_service_time"`  // "23 минуты"
	TopPerformingPeriod  string        `json:"top_performing_period"` // "14:00-16:00"
}

// ShiftResponse represents a shift entity with enhanced formatting
type ShiftResponse struct {
	ID          int                  `json:"id"`
	Date        string               `json:"date"`
	StartTime   string               `json:"start_time"`
	EndTime     string               `json:"end_time"`
	Duration    string               `json:"duration"`
	Status      StatusResponse       `json:"status"`
	Manager     string               `json:"manager"`
	Performance *PerformanceResponse `json:"performance"`
}

// PerformanceResponse represents performance metrics with calculations
type PerformanceResponse struct {
	TablesServed    int           `json:"tables_served"`
	OrdersCompleted int           `json:"orders_completed"`
	Revenue         MoneyResponse `json:"revenue"`
	Tips            MoneyResponse `json:"tips"`
	EfficiencyScore float64       `json:"efficiency_score"` // 0-100
	CustomerRating  float64       `json:"customer_rating"`  // 0-5
	Ranking         string        `json:"ranking"`          // "Топ 10%"
}

// Menu-related responses for order creation
type MenuCategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
	Actions    []ActionResponse   `json:"actions"`
}

type MenuDishesResponse struct {
	Dishes     []DishResponse     `json:"dishes"`
	Category   *CategoryResponse  `json:"category"`
	Pagination PaginationResponse `json:"pagination"`
	Actions    []ActionResponse   `json:"actions"`
}

// Order creation response
type OrderCreatedResponse struct {
	Order          *OrderResponse        `json:"order"`
	Message        string                `json:"message"`
	EstimatedReady FormattedDateResponse `json:"estimated_ready"`
	NextActions    []ActionResponse      `json:"next_actions"`
	Notifications  []string              `json:"notifications"`
}

// WaiterProfileResponse represents waiter profile with enhanced formatting
type WaiterProfileResponse struct {
	ID               int                 `json:"id"`
	Name             string              `json:"name"`
	Username         string              `json:"username"`
	Email            string              `json:"email"`
	Role             string              `json:"role"`
	Status           StatusResponse      `json:"status"`
	AssignedTables   []*TableResponse    `json:"assigned_tables"`
	CurrentShift     *ShiftResponse      `json:"current_shift"`
	UpcomingShifts   []*ShiftResponse    `json:"upcoming_shifts"`
	OrderStats       OrdersStatsResponse `json:"order_stats"`
	PerformanceData  PerformanceResponse `json:"performance_data"`
	AvailableActions []ActionResponse    `json:"available_actions"`
}
