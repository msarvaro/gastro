package repository

import (
	"context"
	"database/sql"
	"restaurant-management/internal/domain/entity"
	"time"
)

type RequestRepository struct {
	db *sql.DB
}

func NewRequestRepository(db *sql.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

func (r *RequestRepository) GetByID(ctx context.Context, id int) (*entity.ServiceRequest, error) {
	query := `SELECT id, business_id, table_id, request_type, status, priority, requested_by,
	         assigned_to, notes, created_at, acknowledged_at, completed_at
	         FROM service_requests WHERE id = $1`

	req := &entity.ServiceRequest{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&req.ID, &req.BusinessID, &req.TableID, &req.RequestType, &req.Status,
		&req.Priority, &req.RequestedBy, &req.AssignedTo, &req.Notes,
		&req.CreatedAt, &req.AcknowledgedAt, &req.CompletedAt,
	)

	return req, err
}

func (r *RequestRepository) GetActiveByBusinessID(ctx context.Context, businessID int) ([]*entity.ServiceRequest, error) {
	query := `SELECT id, business_id, table_id, request_type, status, priority, requested_by,
	         assigned_to, notes, created_at, acknowledged_at, completed_at
	         FROM service_requests WHERE business_id = $1 AND status != 'completed'
	         ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*entity.ServiceRequest
	for rows.Next() {
		req := &entity.ServiceRequest{}
		err := rows.Scan(
			&req.ID, &req.BusinessID, &req.TableID, &req.RequestType, &req.Status,
			&req.Priority, &req.RequestedBy, &req.AssignedTo, &req.Notes,
			&req.CreatedAt, &req.AcknowledgedAt, &req.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (r *RequestRepository) GetByTableID(ctx context.Context, tableID int) ([]*entity.ServiceRequest, error) {
	query := `SELECT id, business_id, table_id, request_type, status, priority, requested_by,
	         assigned_to, notes, created_at, acknowledged_at, completed_at
	         FROM service_requests WHERE table_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*entity.ServiceRequest
	for rows.Next() {
		req := &entity.ServiceRequest{}
		err := rows.Scan(
			&req.ID, &req.BusinessID, &req.TableID, &req.RequestType, &req.Status,
			&req.Priority, &req.RequestedBy, &req.AssignedTo, &req.Notes,
			&req.CreatedAt, &req.AcknowledgedAt, &req.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (r *RequestRepository) GetByWaiterID(ctx context.Context, waiterID int) ([]*entity.ServiceRequest, error) {
	query := `SELECT id, business_id, table_id, request_type, status, priority, requested_by,
	         assigned_to, notes, created_at, acknowledged_at, completed_at
	         FROM service_requests WHERE assigned_to = $1 ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, waiterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*entity.ServiceRequest
	for rows.Next() {
		req := &entity.ServiceRequest{}
		err := rows.Scan(
			&req.ID, &req.BusinessID, &req.TableID, &req.RequestType, &req.Status,
			&req.Priority, &req.RequestedBy, &req.AssignedTo, &req.Notes,
			&req.CreatedAt, &req.AcknowledgedAt, &req.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (r *RequestRepository) Create(ctx context.Context, request *entity.ServiceRequest) error {
	query := `INSERT INTO service_requests (business_id, table_id, request_type, status, priority,
	         requested_by, assigned_to, notes, created_at)
	         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW()) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		request.BusinessID, request.TableID, request.RequestType, request.Status,
		request.Priority, request.RequestedBy, request.AssignedTo, request.Notes,
	).Scan(&request.ID)

	return err
}

func (r *RequestRepository) Update(ctx context.Context, request *entity.ServiceRequest) error {
	query := `UPDATE service_requests SET request_type = $1, status = $2, priority = $3,
	         requested_by = $4, assigned_to = $5, notes = $6, acknowledged_at = $7,
	         completed_at = $8 WHERE id = $9`

	_, err := r.db.ExecContext(ctx, query,
		request.RequestType, request.Status, request.Priority, request.RequestedBy,
		request.AssignedTo, request.Notes, request.AcknowledgedAt, request.CompletedAt,
		request.ID,
	)

	return err
}

func (r *RequestRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE service_requests SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *RequestRepository) AssignToWaiter(ctx context.Context, id int, waiterID int) error {
	query := `UPDATE service_requests SET assigned_to = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, waiterID, id)
	return err
}

func (r *RequestRepository) MarkCompleted(ctx context.Context, id int, completedAt time.Time) error {
	query := `UPDATE service_requests SET status = 'completed', completed_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, completedAt, id)
	return err
}
