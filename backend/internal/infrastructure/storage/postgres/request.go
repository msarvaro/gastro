package postgres

import (
	"context"
	"log"
	"restaurant-management/internal/domain/request"
	"time"

	"github.com/lib/pq"
)

type RequestRepository struct {
	db *DB
}

func NewRequestRepository(db *DB) request.Repository {
	return &RequestRepository{db: db}
}

func (r *RequestRepository) GetAll(ctx context.Context, businessID int) ([]request.Request, error) {
	query := `SELECT id, supplier_id, items, priority, comment, status, created_at, completed_at FROM requests WHERE business_id = $1`
	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		log.Printf("Error querying requests: %v", err)
		return nil, err
	}
	defer rows.Close()

	var requests []request.Request
	for rows.Next() {
		var req request.Request
		var items []string
		if err := rows.Scan(&req.ID, &req.SupplierID, pq.Array(&items), &req.Priority, &req.Comment, &req.Status, &req.CreatedAt, &req.CompletedAt); err != nil {
			log.Printf("Error scanning request row: %v", err)
			return nil, err
		}
		req.Items = items
		req.BusinessID = businessID
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *RequestRepository) GetByID(ctx context.Context, id int, businessID int) (*request.Request, error) {
	query := `SELECT id, supplier_id, items, priority, comment, status, created_at, completed_at FROM requests WHERE id = $1 AND business_id = $2`
	var req request.Request
	var items []string
	err := r.db.QueryRowContext(ctx, query, id, businessID).Scan(&req.ID, &req.SupplierID, pq.Array(&items), &req.Priority, &req.Comment, &req.Status, &req.CreatedAt, &req.CompletedAt)
	if err != nil {
		log.Printf("Error getting request by ID %d: %v", id, err)
		return nil, err
	}
	req.Items = items
	req.BusinessID = businessID
	return &req, nil
}

func (r *RequestRepository) Create(ctx context.Context, requestReq request.CreateRequestRequest, businessID int) (*request.Request, error) {
	query := `
		INSERT INTO requests (supplier_id, items, priority, comment, status, business_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING id, created_at`

	req := &request.Request{
		SupplierID: requestReq.SupplierID,
		Items:      requestReq.Items,
		Priority:   requestReq.Priority,
		Comment:    requestReq.Comment,
		Status:     requestReq.Status,
		BusinessID: businessID,
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = "pending"
	}

	err := r.db.QueryRowContext(ctx, query,
		req.SupplierID,
		pq.Array(req.Items),
		req.Priority,
		req.Comment,
		req.Status,
		businessID,
	).Scan(&req.ID, &req.CreatedAt)

	if err != nil {
		log.Printf("Error creating request: %v", err)
		return nil, err
	}

	return req, nil
}

func (r *RequestRepository) Update(ctx context.Context, id int, requestReq request.UpdateRequestRequest, businessID int) (*request.Request, error) {
	// First get the existing request
	existing, err := r.GetByID(ctx, id, businessID)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if requestReq.SupplierID != 0 {
		existing.SupplierID = requestReq.SupplierID
	}
	if len(requestReq.Items) > 0 {
		existing.Items = requestReq.Items
	}
	if requestReq.Priority != "" {
		existing.Priority = requestReq.Priority
	}
	if requestReq.Comment != "" {
		existing.Comment = requestReq.Comment
	}
	if requestReq.Status != "" {
		existing.Status = requestReq.Status
		// Set completed_at if status changes to completed
		if requestReq.Status == "completed" && existing.CompletedAt == nil {
			now := time.Now()
			existing.CompletedAt = &now
		}
	}

	query := `
		UPDATE requests
		SET supplier_id = $1, items = $2, priority = $3, comment = $4, status = $5
		WHERE id = $6 AND business_id = $7`

	_, err = r.db.ExecContext(ctx, query,
		existing.SupplierID,
		pq.Array(existing.Items),
		existing.Priority,
		existing.Comment,
		existing.Status,
		id,
		businessID,
	)

	if err != nil {
		log.Printf("Error updating request %d: %v", id, err)
		return nil, err
	}

	return existing, nil
}

func (r *RequestRepository) Delete(ctx context.Context, id int, businessID int) error {
	query := `DELETE FROM requests WHERE id = $1 AND business_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, businessID)
	if err != nil {
		log.Printf("Error deleting request %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No request found with ID %d for business %d", id, businessID)
		return err
	}

	return nil
}
