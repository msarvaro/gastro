package notification

import (
	"context"
	"time"
)

// Repository defines the interface for notification data access
type Repository interface {
	// Create creates a new notification
	Create(ctx context.Context, notification *Notification) error

	// GetByID retrieves a notification by ID
	GetByID(ctx context.Context, id int) (*Notification, error)

	// GetByBusinessID retrieves notifications for a specific business
	GetByBusinessID(ctx context.Context, businessID int, limit, offset int) ([]Notification, error)

	// GetPendingNotifications retrieves pending notifications for sending
	GetPendingNotifications(ctx context.Context, limit int) ([]Notification, error)

	// UpdateStatus updates the status of a notification
	UpdateStatus(ctx context.Context, id int, status NotificationStatus, sentAt *time.Time, errorMsg *string) error

	// GetStats retrieves notification statistics for a business
	GetStats(ctx context.Context, businessID int) (*NotificationStats, error)

	// GetRecentNotifications retrieves recent notifications for dashboard display
	GetRecentNotifications(ctx context.Context, businessID int, limit int) ([]Notification, error)
}
