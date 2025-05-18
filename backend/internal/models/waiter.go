package models

import "time"

// WaiterProfileResponse представляет данные профиля официанта
type WaiterProfileResponse struct {
	ID                  int                `json:"id"`
	Username            string             `json:"username"`
	Name                string             `json:"name"`
	Email               string             `json:"email"`
	CurrentShift        *ShiftInfo         `json:"current_shift,omitempty"`
	CurrentShiftManager string             `json:"current_shift_manager,omitempty"`
	UpcomingShifts      []ShiftInfo        `json:"upcoming_shifts,omitempty"`
	AssignedTables      []Table            `json:"assigned_tables,omitempty"`
	OrderStats          OrderStatusCounts  `json:"order_stats"`
	PerformanceData     PerformanceMetrics `json:"performance_data"`
}

// ShiftInfo содержит информацию о смене официанта
type ShiftInfo struct {
	ID        int       `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	IsActive  bool      `json:"is_active"`
}

// OrderStatusCounts содержит количество заказов по статусам
type OrderStatusCounts struct {
	New       int `json:"new"`
	Accepted  int `json:"accepted"`
	Preparing int `json:"preparing"`
	Ready     int `json:"ready"`
	Served    int `json:"served"`
	Total     int `json:"total"`
}

// PerformanceMetrics содержит метрики эффективности официанта
type PerformanceMetrics struct {
	TablesServed       int     `json:"tables_served"`
	OrdersCompleted    int     `json:"orders_completed"`
	AverageServiceTime float64 `json:"average_service_time"` // в минутах
}
