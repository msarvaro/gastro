package postgres

import (
	"context"
	"log"
	"restaurant-management/internal/domain/supplier"

	"github.com/lib/pq"
)

type SupplierRepository struct {
	db *DB
}

func NewSupplierRepository(db *DB) supplier.Repository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) GetAll(ctx context.Context, businessID int) ([]supplier.Supplier, error) {
	query := `SELECT id, name, categories, phone, email, address, status, created_at, updated_at FROM suppliers WHERE business_id = $1`
	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		log.Printf("Error querying suppliers: %v", err)
		return nil, err
	}
	defer rows.Close()

	var suppliers []supplier.Supplier
	for rows.Next() {
		var s supplier.Supplier
		var categories []string
		if err := rows.Scan(&s.ID, &s.Name, pq.Array(&categories), &s.Phone, &s.Email, &s.Address, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			log.Printf("Error scanning supplier row: %v", err)
			return nil, err
		}
		s.Categories = categories
		s.BusinessID = businessID
		suppliers = append(suppliers, s)
	}
	return suppliers, nil
}

func (r *SupplierRepository) GetByID(ctx context.Context, id int, businessID int) (*supplier.Supplier, error) {
	query := `SELECT id, name, categories, phone, email, address, status, created_at, updated_at FROM suppliers WHERE id = $1 AND business_id = $2`
	var s supplier.Supplier
	var categories []string
	err := r.db.QueryRowContext(ctx, query, id, businessID).Scan(&s.ID, &s.Name, pq.Array(&categories), &s.Phone, &s.Email, &s.Address, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		log.Printf("Error getting supplier by ID %d: %v", id, err)
		return nil, err
	}
	s.Categories = categories
	s.BusinessID = businessID
	return &s, nil
}

func (r *SupplierRepository) Create(ctx context.Context, supplierReq supplier.CreateSupplierRequest, businessID int) (*supplier.Supplier, error) {
	query := `
		INSERT INTO suppliers (name, categories, phone, email, address, status, business_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	s := &supplier.Supplier{
		Name:       supplierReq.Name,
		Categories: supplierReq.Categories,
		Phone:      supplierReq.Phone,
		Email:      supplierReq.Email,
		Address:    supplierReq.Address,
		Status:     supplierReq.Status,
		BusinessID: businessID,
	}

	// Set default status if not provided
	if s.Status == "" {
		s.Status = "active"
	}

	err := r.db.QueryRowContext(ctx, query,
		s.Name,
		pq.Array(s.Categories),
		s.Phone,
		s.Email,
		s.Address,
		s.Status,
		businessID,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)

	if err != nil {
		log.Printf("Error creating supplier: %v", err)
		return nil, err
	}

	return s, nil
}

func (r *SupplierRepository) Update(ctx context.Context, id int, supplierReq supplier.UpdateSupplierRequest, businessID int) (*supplier.Supplier, error) {
	// First get the existing supplier
	existing, err := r.GetByID(ctx, id, businessID)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if supplierReq.Name != "" {
		existing.Name = supplierReq.Name
	}
	if len(supplierReq.Categories) > 0 {
		existing.Categories = supplierReq.Categories
	}
	if supplierReq.Phone != "" {
		existing.Phone = supplierReq.Phone
	}
	if supplierReq.Email != "" {
		existing.Email = supplierReq.Email
	}
	if supplierReq.Address != "" {
		existing.Address = supplierReq.Address
	}
	if supplierReq.Status != "" {
		existing.Status = supplierReq.Status
	}

	query := `
		UPDATE suppliers
		SET name = $1, categories = $2, phone = $3, email = $4, address = $5, status = $6, updated_at = NOW()
		WHERE id = $7 AND business_id = $8
		RETURNING updated_at`

	err = r.db.QueryRowContext(ctx, query,
		existing.Name,
		pq.Array(existing.Categories),
		existing.Phone,
		existing.Email,
		existing.Address,
		existing.Status,
		id,
		businessID,
	).Scan(&existing.UpdatedAt)

	if err != nil {
		log.Printf("Error updating supplier %d: %v", id, err)
		return nil, err
	}

	return existing, nil
}

func (r *SupplierRepository) Delete(ctx context.Context, id int, businessID int) error {
	query := `DELETE FROM suppliers WHERE id = $1 AND business_id = $2`
	result, err := r.db.ExecContext(ctx, query, id, businessID)
	if err != nil {
		log.Printf("Error deleting supplier %d: %v", id, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No supplier found with ID %d for business %d", id, businessID)
		return err
	}

	return nil
}
