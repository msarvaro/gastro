package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"restaurant-management/internal/domain/notification"
)

type NotificationRepository struct {
	db *DB
}

func NewNotificationRepository(db *DB) notification.Repository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	recipientsJSON, err := json.Marshal(n.Recipients)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO notifications (business_id, type, subject, body, recipients, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	return r.db.QueryRow(
		query,
		n.BusinessID,
		n.Type,
		n.Subject,
		n.Body,
		recipientsJSON,
		n.Status,
		now,
		now,
	).Scan(&n.ID)
}

func (r *NotificationRepository) GetByID(ctx context.Context, id int) (*notification.Notification, error) {
	query := `
		SELECT id, business_id, type, subject, body, recipients, status, sent_at, error_msg, created_at, updated_at
		FROM notifications
		WHERE id = $1
	`

	var n notification.Notification
	var recipientsJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&n.ID,
		&n.BusinessID,
		&n.Type,
		&n.Subject,
		&n.Body,
		&recipientsJSON,
		&n.Status,
		&n.SentAt,
		&n.ErrorMsg,
		&n.CreatedAt,
		&n.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(recipientsJSON, &n.Recipients); err != nil {
		return nil, err
	}

	return &n, nil
}

func (r *NotificationRepository) GetByBusinessID(ctx context.Context, businessID int, limit, offset int) ([]notification.Notification, error) {
	query := `
		SELECT id, business_id, type, subject, body, recipients, status, sent_at, error_msg, created_at, updated_at
		FROM notifications
		WHERE business_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, businessID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []notification.Notification

	for rows.Next() {
		var n notification.Notification
		var recipientsJSON []byte

		err := rows.Scan(
			&n.ID,
			&n.BusinessID,
			&n.Type,
			&n.Subject,
			&n.Body,
			&recipientsJSON,
			&n.Status,
			&n.SentAt,
			&n.ErrorMsg,
			&n.CreatedAt,
			&n.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(recipientsJSON, &n.Recipients); err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (r *NotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]notification.Notification, error) {
	query := `
		SELECT id, business_id, type, subject, body, recipients, status, sent_at, error_msg, created_at, updated_at
		FROM notifications
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(query, notification.NotificationStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []notification.Notification

	for rows.Next() {
		var n notification.Notification
		var recipientsJSON []byte

		err := rows.Scan(
			&n.ID,
			&n.BusinessID,
			&n.Type,
			&n.Subject,
			&n.Body,
			&recipientsJSON,
			&n.Status,
			&n.SentAt,
			&n.ErrorMsg,
			&n.CreatedAt,
			&n.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(recipientsJSON, &n.Recipients); err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id int, status notification.NotificationStatus, sentAt *time.Time, errorMsg *string) error {
	query := `
		UPDATE notifications
		SET status = $1, sent_at = $2, error_msg = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(query, status, sentAt, errorMsg, time.Now(), id)
	return err
}

func (r *NotificationRepository) GetStats(ctx context.Context, businessID int) (*notification.NotificationStats, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN status = 'sent' THEN 1 END) as total_sent,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as total_pending,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as total_failed,
			COUNT(CASE WHEN status = 'sent' AND DATE(sent_at) = CURRENT_DATE THEN 1 END) as sent_today,
			COUNT(CASE WHEN status = 'sent' AND sent_at >= DATE_TRUNC('week', CURRENT_DATE) THEN 1 END) as sent_this_week,
			COUNT(CASE WHEN status = 'sent' AND sent_at >= DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as sent_this_month
		FROM notifications
		WHERE business_id = $1
	`

	var stats notification.NotificationStats
	err := r.db.QueryRow(query, businessID).Scan(
		&stats.TotalSent,
		&stats.TotalPending,
		&stats.TotalFailed,
		&stats.SentToday,
		&stats.SentThisWeek,
		&stats.SentThisMonth,
	)

	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *NotificationRepository) GetRecentNotifications(ctx context.Context, businessID int, limit int) ([]notification.Notification, error) {
	query := `
		SELECT id, business_id, type, subject, body, recipients, status, sent_at, error_msg, created_at, updated_at
		FROM notifications
		WHERE business_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, businessID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []notification.Notification

	for rows.Next() {
		var n notification.Notification
		var recipientsJSON []byte

		err := rows.Scan(
			&n.ID,
			&n.BusinessID,
			&n.Type,
			&n.Subject,
			&n.Body,
			&recipientsJSON,
			&n.Status,
			&n.SentAt,
			&n.ErrorMsg,
			&n.CreatedAt,
			&n.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(recipientsJSON, &n.Recipients); err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	return notifications, rows.Err()
}
