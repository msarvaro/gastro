package notification

import "context"

// Service defines the interface for notification business logic
type Service interface {
	// CreateNotification creates a new notification and queues it for sending
	CreateNotification(ctx context.Context, businessID int, req CreateNotificationRequest) (*Notification, error)

	// SendNotification sends a notification via email
	SendNotification(ctx context.Context, notification *Notification) error

	// ProcessPendingNotifications processes and sends pending notifications
	ProcessPendingNotifications(ctx context.Context) error

	// GetRecentNotifications retrieves recent notifications for dashboard
	GetRecentNotifications(ctx context.Context, businessID int) ([]Notification, error)

	// GetNotificationStats retrieves notification statistics
	GetNotificationStats(ctx context.Context, businessID int) (*NotificationStats, error)

	// SendLowInventoryAlert sends an alert when inventory is low
	SendLowInventoryAlert(ctx context.Context, businessID int, itemName string, currentStock, minStock float64, unit string) error

	// SendNewHiringAlert sends an alert for new hiring applications
	SendNewHiringAlert(ctx context.Context, businessID int, applicantName, position string, experience string, location string) error

	// SendWeeklyReport sends a weekly business report
	SendWeeklyReport(ctx context.Context, businessID int, reportData interface{}) error

	// SendDailyReport sends a daily business report
	SendDailyReport(ctx context.Context, businessID int, reportData interface{}) error
}
