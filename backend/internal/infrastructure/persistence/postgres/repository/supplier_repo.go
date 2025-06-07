package repository

import (
	"context"
	"database/sql"
	"restaurant-management/internal/domain/entity"
)

type SupplierRepository struct {
	db *sql.DB
}

func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// Supplier operations
func (r *SupplierRepository) GetByID(ctx context.Context, id int) (*entity.Supplier, error) {
	query := `SELECT id, business_id, name, contact_person, email, phone, address, tax_id,
	         payment_terms, rating, is_active, created_at, updated_at
	         FROM suppliers WHERE id = $1`

	supplier := &entity.Supplier{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&supplier.ID, &supplier.BusinessID, &supplier.Name, &supplier.ContactPerson,
		&supplier.Email, &supplier.Phone, &supplier.Address, &supplier.TaxID,
		&supplier.PaymentTerms, &supplier.Rating, &supplier.IsActive,
		&supplier.CreatedAt, &supplier.UpdatedAt,
	)

	return supplier, err
}

func (r *SupplierRepository) GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Supplier, error) {
	query := `SELECT id, business_id, name, contact_person, email, phone, address, tax_id,
	         payment_terms, rating, is_active, created_at, updated_at
	         FROM suppliers WHERE business_id = $1 ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []*entity.Supplier
	for rows.Next() {
		supplier := &entity.Supplier{}
		err := rows.Scan(
			&supplier.ID, &supplier.BusinessID, &supplier.Name, &supplier.ContactPerson,
			&supplier.Email, &supplier.Phone, &supplier.Address, &supplier.TaxID,
			&supplier.PaymentTerms, &supplier.Rating, &supplier.IsActive,
			&supplier.CreatedAt, &supplier.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, nil
}

func (r *SupplierRepository) GetActive(ctx context.Context, businessID int) ([]*entity.Supplier, error) {
	query := `SELECT id, business_id, name, contact_person, email, phone, address, tax_id,
	         payment_terms, rating, is_active, created_at, updated_at
	         FROM suppliers WHERE business_id = $1 AND is_active = true ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var suppliers []*entity.Supplier
	for rows.Next() {
		supplier := &entity.Supplier{}
		err := rows.Scan(
			&supplier.ID, &supplier.BusinessID, &supplier.Name, &supplier.ContactPerson,
			&supplier.Email, &supplier.Phone, &supplier.Address, &supplier.TaxID,
			&supplier.PaymentTerms, &supplier.Rating, &supplier.IsActive,
			&supplier.CreatedAt, &supplier.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		suppliers = append(suppliers, supplier)
	}

	return suppliers, nil
}

func (r *SupplierRepository) Create(ctx context.Context, supplier *entity.Supplier) error {
	query := `INSERT INTO suppliers (business_id, name, contact_person, email, phone, address,
	         tax_id, payment_terms, rating, is_active, created_at, updated_at)
	         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		supplier.BusinessID, supplier.Name, supplier.ContactPerson, supplier.Email,
		supplier.Phone, supplier.Address, supplier.TaxID, supplier.PaymentTerms,
		supplier.Rating, supplier.IsActive,
	).Scan(&supplier.ID)

	return err
}

func (r *SupplierRepository) Update(ctx context.Context, supplier *entity.Supplier) error {
	query := `UPDATE suppliers SET name = $1, contact_person = $2, email = $3, phone = $4,
	         address = $5, tax_id = $6, payment_terms = $7, rating = $8, is_active = $9,
	         updated_at = NOW() WHERE id = $10`

	_, err := r.db.ExecContext(ctx, query,
		supplier.Name, supplier.ContactPerson, supplier.Email, supplier.Phone,
		supplier.Address, supplier.TaxID, supplier.PaymentTerms, supplier.Rating,
		supplier.IsActive, supplier.ID,
	)

	return err
}

func (r *SupplierRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM suppliers WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Purchase order operations
func (r *SupplierRepository) GetPurchaseOrderByID(ctx context.Context, id int) (*entity.PurchaseOrder, error) {
	query := `SELECT id, business_id, supplier_id, order_number, status, total_amount,
	         expected_delivery, actual_delivery, created_by, approved_by, notes,
	         created_at, updated_at FROM purchase_orders WHERE id = $1`

	po := &entity.PurchaseOrder{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&po.ID, &po.BusinessID, &po.SupplierID, &po.OrderNumber, &po.Status,
		&po.TotalAmount, &po.ExpectedDelivery, &po.ActualDelivery,
		&po.CreatedBy, &po.ApprovedBy, &po.Notes,
		&po.CreatedAt, &po.UpdatedAt,
	)

	return po, err
}

func (r *SupplierRepository) GetPurchaseOrdersBySupplierID(ctx context.Context, supplierID int) ([]*entity.PurchaseOrder, error) {
	query := `SELECT id, business_id, supplier_id, order_number, status, total_amount,
	         expected_delivery, actual_delivery, created_by, approved_by, notes,
	         created_at, updated_at FROM purchase_orders WHERE supplier_id = $1
	         ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.PurchaseOrder
	for rows.Next() {
		po := &entity.PurchaseOrder{}
		err := rows.Scan(
			&po.ID, &po.BusinessID, &po.SupplierID, &po.OrderNumber, &po.Status,
			&po.TotalAmount, &po.ExpectedDelivery, &po.ActualDelivery,
			&po.CreatedBy, &po.ApprovedBy, &po.Notes,
			&po.CreatedAt, &po.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, po)
	}

	return orders, nil
}

func (r *SupplierRepository) GetPendingPurchaseOrders(ctx context.Context, businessID int) ([]*entity.PurchaseOrder, error) {
	query := `SELECT id, business_id, supplier_id, order_number, status, total_amount,
	         expected_delivery, actual_delivery, created_by, approved_by, notes,
	         created_at, updated_at FROM purchase_orders 
	         WHERE business_id = $1 AND status IN ('draft', 'sent', 'confirmed')
	         ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.PurchaseOrder
	for rows.Next() {
		po := &entity.PurchaseOrder{}
		err := rows.Scan(
			&po.ID, &po.BusinessID, &po.SupplierID, &po.OrderNumber, &po.Status,
			&po.TotalAmount, &po.ExpectedDelivery, &po.ActualDelivery,
			&po.CreatedBy, &po.ApprovedBy, &po.Notes,
			&po.CreatedAt, &po.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, po)
	}

	return orders, nil
}

func (r *SupplierRepository) CreatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error {
	query := `INSERT INTO purchase_orders (business_id, supplier_id, order_number, status,
	         total_amount, expected_delivery, created_by, notes, created_at, updated_at)
	         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		order.BusinessID, order.SupplierID, order.OrderNumber, order.Status,
		order.TotalAmount, order.ExpectedDelivery, order.CreatedBy, order.Notes,
	).Scan(&order.ID)

	return err
}

func (r *SupplierRepository) UpdatePurchaseOrder(ctx context.Context, order *entity.PurchaseOrder) error {
	query := `UPDATE purchase_orders SET status = $1, total_amount = $2, expected_delivery = $3,
	         actual_delivery = $4, approved_by = $5, notes = $6, updated_at = NOW()
	         WHERE id = $7`

	_, err := r.db.ExecContext(ctx, query,
		order.Status, order.TotalAmount, order.ExpectedDelivery, order.ActualDelivery,
		order.ApprovedBy, order.Notes, order.ID,
	)

	return err
}

func (r *SupplierRepository) UpdatePurchaseOrderStatus(ctx context.Context, id int, status string) error {
	query := `UPDATE purchase_orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
