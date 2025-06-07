package responses

import "time"

// UserListResponse represents a paginated list of users
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// AdminStatsResponse represents administrative statistics
type AdminStatsResponse struct {
	TotalUsers      int                    `json:"total_users"`
	ActiveUsers     int                    `json:"active_users"`
	UsersByRole     map[string]int         `json:"users_by_role"`
	UsersByStatus   map[string]int         `json:"users_by_status"`
	UsersByBusiness map[string]int         `json:"users_by_business"`
	RecentActivity  []UserActivityResponse `json:"recent_activity"`
	SystemHealth    SystemHealthResponse   `json:"system_health"`
}

// UserActivityResponse represents user activity information
type UserActivityResponse struct {
	UserID    int       `json:"user_id"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// SystemHealthResponse represents system health metrics
type SystemHealthResponse struct {
	Status         string             `json:"status"`
	DatabaseHealth DatabaseHealthInfo `json:"database_health"`
	ServerUptime   string             `json:"server_uptime"`
	MemoryUsage    MemoryUsageInfo    `json:"memory_usage"`
	ActiveSessions int                `json:"active_sessions"`
}

// DatabaseHealthInfo represents database health information
type DatabaseHealthInfo struct {
	Status            string `json:"status"`
	ConnectionsTotal  int    `json:"connections_total"`
	ConnectionsActive int    `json:"connections_active"`
	ResponseTime      string `json:"response_time"`
}

// MemoryUsageInfo represents memory usage information
type MemoryUsageInfo struct {
	Used    string  `json:"used"`
	Total   string  `json:"total"`
	Percent float64 `json:"percent"`
}
