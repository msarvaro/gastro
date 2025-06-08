package postgres

import (
	"context"
	"database/sql"
	"log"
	"restaurant-management/internal/domain/business"
	"time"
)

type BusinessRepository struct {
	db *DB
}

func NewBusinessRepository(db *DB) business.Repository {
	return &BusinessRepository{db: db}
}

// CreateBusiness creates a new business in the database
func (r *BusinessRepository) CreateBusiness(ctx context.Context, b *business.Business) error {
	query := `
		INSERT INTO businesses (name, description, address, phone, email, website, logo, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, created_at, updated_at`

	now := time.Now()

	// Create NullString values for nullable fields
	description := sql.NullString{String: b.Description, Valid: b.Description != ""}
	address := sql.NullString{String: b.Address, Valid: b.Address != ""}
	phone := sql.NullString{String: b.Phone, Valid: b.Phone != ""}
	email := sql.NullString{String: b.Email, Valid: b.Email != ""}
	website := sql.NullString{String: b.Website, Valid: b.Website != ""}
	logo := sql.NullString{String: b.Logo, Valid: b.Logo != ""}

	err := r.db.QueryRowContext(ctx, query,
		b.Name,
		description,
		address,
		phone,
		email,
		website,
		logo,
		b.Status,
		now,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)

	if err != nil {
		log.Printf("Error creating business: %v", err)
		return err
	}

	return nil
}

// GetBusinessByID retrieves a business by its ID
func (r *BusinessRepository) GetBusinessByID(ctx context.Context, id int) (*business.Business, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, logo, status, created_at, updated_at
		FROM businesses
		WHERE id = $1`

	b := &business.Business{}

	// Use NullString for potentially NULL string columns
	var description, address, phone, email, website, logo sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID,
		&b.Name,
		&description,
		&address,
		&phone,
		&email,
		&website,
		&logo,
		&b.Status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error fetching business by ID %d: %v", id, err)
		return nil, err
	}

	// Set values from NullString to string, using empty string for NULL values
	if description.Valid {
		b.Description = description.String
	}
	if address.Valid {
		b.Address = address.String
	}
	if phone.Valid {
		b.Phone = phone.String
	}
	if email.Valid {
		b.Email = email.String
	}
	if website.Valid {
		b.Website = website.String
	}
	if logo.Valid {
		b.Logo = logo.String
	}

	return b, nil
}

// GetAllBusinesses retrieves all businesses from the database
func (r *BusinessRepository) GetAllBusinesses(ctx context.Context) ([]business.Business, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, logo, status, created_at, updated_at
		FROM businesses
		ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error fetching all businesses: %v", err)
		return nil, err
	}
	defer rows.Close()

	var businesses []business.Business

	for rows.Next() {
		var b business.Business

		// Use NullString for potentially NULL string columns
		var description, address, phone, email, website, logo sql.NullString

		err := rows.Scan(
			&b.ID,
			&b.Name,
			&description,
			&address,
			&phone,
			&email,
			&website,
			&logo,
			&b.Status,
			&b.CreatedAt,
			&b.UpdatedAt,
		)

		if err != nil {
			log.Printf("Error scanning business row: %v", err)
			return nil, err
		}

		// Set values from NullString to string, using empty string for NULL values
		if description.Valid {
			b.Description = description.String
		}
		if address.Valid {
			b.Address = address.String
		}
		if phone.Valid {
			b.Phone = phone.String
		}
		if email.Valid {
			b.Email = email.String
		}
		if website.Valid {
			b.Website = website.String
		}
		if logo.Valid {
			b.Logo = logo.String
		}

		businesses = append(businesses, b)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating business rows: %v", err)
		return nil, err
	}

	return businesses, nil
}

// UpdateBusiness updates an existing business
func (r *BusinessRepository) UpdateBusiness(ctx context.Context, b *business.Business) error {
	query := `
		UPDATE businesses
		SET name = $1, description = $2, address = $3, phone = $4, 
		    email = $5, website = $6, logo = $7, status = $8, updated_at = $9
		WHERE id = $10`

	now := time.Now()

	// Create NullString values for nullable fields
	description := sql.NullString{String: b.Description, Valid: b.Description != ""}
	address := sql.NullString{String: b.Address, Valid: b.Address != ""}
	phone := sql.NullString{String: b.Phone, Valid: b.Phone != ""}
	email := sql.NullString{String: b.Email, Valid: b.Email != ""}
	website := sql.NullString{String: b.Website, Valid: b.Website != ""}
	logo := sql.NullString{String: b.Logo, Valid: b.Logo != ""}

	_, err := r.db.ExecContext(ctx, query,
		b.Name,
		description,
		address,
		phone,
		email,
		website,
		logo,
		b.Status,
		now,
		b.ID,
	)

	if err != nil {
		log.Printf("Error updating business ID %d: %v", b.ID, err)
		return err
	}

	b.UpdatedAt = now

	return nil
}

// DeleteBusiness deletes a business by ID
func (r *BusinessRepository) DeleteBusiness(ctx context.Context, id int) error {
	query := `DELETE FROM businesses WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		log.Printf("Error deleting business ID %d: %v", id, err)
		return err
	}

	return nil
}

// GetBusinessStats retrieves business statistics
func (r *BusinessRepository) GetBusinessStats(ctx context.Context) (*business.BusinessStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as new
		FROM businesses`

	stats := &business.BusinessStats{}

	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.Total,
		&stats.Active,
		&stats.Inactive,
		&stats.New,
	)

	if err != nil {
		log.Printf("Error fetching business stats: %v", err)
		return nil, err
	}

	return stats, nil
}
