package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// businessRepository implements the BusinessRepository interface using PostgreSQL
type businessRepository struct {
	db *sql.DB
}

// NewBusinessRepository creates a new business repository
func NewBusinessRepository(db *sql.DB) repository.BusinessRepository {
	return &businessRepository{
		db: db,
	}
}

// GetByID retrieves a business by ID
func (r *businessRepository) GetByID(ctx context.Context, id int) (*entity.Business, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, logo, status, created_at, updated_at
		FROM businesses 
		WHERE id = $1
	`

	var business entity.Business

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&business.ID,
		&business.Name,
		&business.Description,
		&business.Address,
		&business.Phone,
		&business.Email,
		&business.Website,
		&business.Logo,
		&business.Status,
		&business.CreatedAt,
		&business.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &business, nil
}

// GetAll retrieves all businesses
func (r *businessRepository) GetAll(ctx context.Context) ([]*entity.Business, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, logo, status, created_at, updated_at
		FROM businesses
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []*entity.Business

	for rows.Next() {
		var business entity.Business

		err := rows.Scan(
			&business.ID,
			&business.Name,
			&business.Description,
			&business.Address,
			&business.Phone,
			&business.Email,
			&business.Website,
			&business.Logo,
			&business.Status,
			&business.CreatedAt,
			&business.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		businesses = append(businesses, &business)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return businesses, nil
}

// Create adds a new business
func (r *businessRepository) Create(ctx context.Context, business *entity.Business) error {
	query := `
		INSERT INTO businesses (name, description, address, phone, email, website, logo, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	// Set defaults if not provided
	if business.Status == "" {
		business.Status = "active"
	}

	now := time.Now()
	if business.CreatedAt.IsZero() {
		business.CreatedAt = now
	}
	if business.UpdatedAt.IsZero() {
		business.UpdatedAt = now
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		business.Name,
		business.Description,
		business.Address,
		business.Phone,
		business.Email,
		business.Website,
		business.Logo,
		business.Status,
		business.CreatedAt,
		business.UpdatedAt,
	).Scan(&business.ID)

	return err
}

// Update updates an existing business
func (r *businessRepository) Update(ctx context.Context, business *entity.Business) error {
	query := `
		UPDATE businesses
		SET name = $1, description = $2, address = $3, phone = $4, email = $5, 
		    website = $6, logo = $7, status = $8, updated_at = $9
		WHERE id = $10
	`

	business.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		business.Name,
		business.Description,
		business.Address,
		business.Phone,
		business.Email,
		business.Website,
		business.Logo,
		business.Status,
		business.UpdatedAt,
		business.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("business not found")
	}

	return nil
}

// Delete removes a business
func (r *businessRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM businesses WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("business not found")
	}

	return nil
}

// GetByUserID retrieves businesses by user ID
func (r *businessRepository) GetByUserID(ctx context.Context, userID int) ([]*entity.Business, error) {
	query := `
		SELECT b.id, b.name, b.description, b.address, b.phone, b.email, 
		       b.website, b.logo, b.status, b.created_at, b.updated_at
		FROM businesses b
		JOIN users u ON b.id = u.business_id
		WHERE u.id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []*entity.Business

	for rows.Next() {
		var business entity.Business

		err := rows.Scan(
			&business.ID,
			&business.Name,
			&business.Description,
			&business.Address,
			&business.Phone,
			&business.Email,
			&business.Website,
			&business.Logo,
			&business.Status,
			&business.CreatedAt,
			&business.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		businesses = append(businesses, &business)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return businesses, nil
}

// GetUserBusinesses is an alias for GetByUserID
func (r *businessRepository) GetUserBusinesses(ctx context.Context, userID int) ([]*entity.Business, error) {
	return r.GetByUserID(ctx, userID)
}

// UpdateStatus updates a business's status
func (r *businessRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE businesses
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("business not found")
	}

	return nil
}

// GetByStatus retrieves businesses by status
func (r *businessRepository) GetByStatus(ctx context.Context, status string) ([]*entity.Business, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, logo, status, created_at, updated_at
		FROM businesses
		WHERE status = $1
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var businesses []*entity.Business

	for rows.Next() {
		var business entity.Business

		err := rows.Scan(
			&business.ID,
			&business.Name,
			&business.Description,
			&business.Address,
			&business.Phone,
			&business.Email,
			&business.Website,
			&business.Logo,
			&business.Status,
			&business.CreatedAt,
			&business.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		businesses = append(businesses, &business)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return businesses, nil
}
