package notification

import (
	"time"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeLowInventory NotificationType = "low_inventory"
	NotificationTypeNewHiring    NotificationType = "new_hiring"
	NotificationTypeWeeklyReport NotificationType = "weekly_report"
	NotificationTypeDailyReport  NotificationType = "daily_report"
	NotificationTypeOrderUpdate  NotificationType = "order_update"
	NotificationTypeShiftAlert   NotificationType = "shift_alert"
	NotificationTypeSystemAlert  NotificationType = "system_alert"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// Notification represents an email notification
type Notification struct {
	ID         int                `json:"id" db:"id"`
	BusinessID int                `json:"business_id" db:"business_id"`
	Type       NotificationType   `json:"type" db:"type"`
	Subject    string             `json:"subject" db:"subject"`
	Body       string             `json:"body" db:"body"`
	Recipients []string           `json:"recipients" db:"recipients"`
	Status     NotificationStatus `json:"status" db:"status"`
	SentAt     *time.Time         `json:"sent_at,omitempty" db:"sent_at"`
	ErrorMsg   *string            `json:"error_msg,omitempty" db:"error_msg"`
	CreatedAt  time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" db:"updated_at"`
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	Type       NotificationType `json:"type"`
	Subject    string           `json:"subject"`
	Body       string           `json:"body"`
	Recipients []string         `json:"recipients"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalSent     int `json:"total_sent"`
	TotalPending  int `json:"total_pending"`
	TotalFailed   int `json:"total_failed"`
	SentToday     int `json:"sent_today"`
	SentThisWeek  int `json:"sent_this_week"`
	SentThisMonth int `json:"sent_this_month"`
}
