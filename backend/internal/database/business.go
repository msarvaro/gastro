package database

import (
	"database/sql"
	"log"
	"restaurant-management/internal/models"
	"time"
)

// CreateBusiness creates a new business in the database
func (db *DB) CreateBusiness(business *models.Business) error {
	query := `
		INSERT INTO businesses (name, description, address, phone, email, website, logo, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, created_at, updated_at`

	now := time.Now()

	// Create NullString values for nullable fields
	description := sql.NullString{String: business.Description, Valid: business.Description != ""}
	address := sql.NullString{String: business.Address, Valid: business.Address != ""}
	phone := sql.NullString{String: business.Phone, Valid: business.Phone != ""}
	email := sql.NullString{String: business.Email, Valid: business.Email != ""}
	website := sql.NullString{String: business.Website, Valid: business.Website != ""}
	logo := sql.NullString{String: business.Logo, Valid: business.Logo != ""}

	err := db.QueryRow(
		query,
		business.Name,
		description,
		address,
		phone,
		email,
		website,
		logo,
		business.Status,
		now,
	).Scan(&business.ID, &business.CreatedAt, &business.UpdatedAt)

	if err != nil {
		log.Printf("Error creating business: %v", err)
		return err
	}

	return nil
}

// GetBusinessByID retrieves a business by its ID
func (db *DB) GetBusinessByID(id int) (*models.Business, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, logo, status, created_at, updated_at
		FROM businesses
		WHERE id = $1`

	business := &models.Business{}

	// Use NullString for potentially NULL string columns
	var description, address, phone, email, website, logo sql.NullString

	err := db.QueryRow(query, id).Scan(
		&business.ID,
		&business.Name,
		&description,
		&address,
		&phone,
		&email,
		&website,
		&logo,
		&business.Status,
		&business.CreatedAt,
		&business.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error fetching business by ID %d: %v", id, err)
		return nil, err
	}

	// Set values from NullString to string, using empty string for NULL values
	if description.Valid {
		business.Description = description.String
	}
	if address.Valid {
		business.Address = address.String
	}
	if phone.Valid {
		business.Phone = phone.String
	}
	if email.Valid {
		business.Email = email.String
	}
	if website.Valid {
		business.Website = website.String
	}
	if logo.Valid {
		business.Logo = logo.String
	}

	return business, nil
}

// GetAllBusinesses retrieves all businesses from the database
func (db *DB) GetAllBusinesses() ([]models.Business, error) {
	query := `
		SELECT id, name, description, address, phone, email, website, logo, status, created_at, updated_at
		FROM businesses
		ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error fetching all businesses: %v", err)
		return nil, err
	}
	defer rows.Close()

	var businesses []models.Business

	for rows.Next() {
		var business models.Business

		// Use NullString for potentially NULL string columns
		var description, address, phone, email, website, logo sql.NullString

		err := rows.Scan(
			&business.ID,
			&business.Name,
			&description,
			&address,
			&phone,
			&email,
			&website,
			&logo,
			&business.Status,
			&business.CreatedAt,
			&business.UpdatedAt,
		)

		if err != nil {
			log.Printf("Error scanning business row: %v", err)
			return nil, err
		}

		// Set values from NullString to string, using empty string for NULL values
		if description.Valid {
			business.Description = description.String
		}
		if address.Valid {
			business.Address = address.String
		}
		if phone.Valid {
			business.Phone = phone.String
		}
		if email.Valid {
			business.Email = email.String
		}
		if website.Valid {
			business.Website = website.String
		}
		if logo.Valid {
			business.Logo = logo.String
		}

		businesses = append(businesses, business)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating business rows: %v", err)
		return nil, err
	}

	return businesses, nil
}

// UpdateBusiness updates an existing business
func (db *DB) UpdateBusiness(business *models.Business) error {
	query := `
		UPDATE businesses
		SET name = $1, description = $2, address = $3, phone = $4, 
		    email = $5, website = $6, logo = $7, status = $8, updated_at = $9
		WHERE id = $10`

	now := time.Now()

	// Create NullString values for nullable fields
	description := sql.NullString{String: business.Description, Valid: business.Description != ""}
	address := sql.NullString{String: business.Address, Valid: business.Address != ""}
	phone := sql.NullString{String: business.Phone, Valid: business.Phone != ""}
	email := sql.NullString{String: business.Email, Valid: business.Email != ""}
	website := sql.NullString{String: business.Website, Valid: business.Website != ""}
	logo := sql.NullString{String: business.Logo, Valid: business.Logo != ""}

	_, err := db.Exec(
		query,
		business.Name,
		description,
		address,
		phone,
		email,
		website,
		logo,
		business.Status,
		now,
		business.ID,
	)

	if err != nil {
		log.Printf("Error updating business ID %d: %v", business.ID, err)
		return err
	}

	business.UpdatedAt = now

	return nil
}

// DeleteBusiness deletes a business by ID
func (db *DB) DeleteBusiness(id int) error {
	query := `DELETE FROM businesses WHERE id = $1`

	_, err := db.Exec(query, id)

	if err != nil {
		log.Printf("Error deleting business ID %d: %v", id, err)
		return err
	}

	return nil
}

// GetBusinessStats retrieves business statistics
func (db *DB) GetBusinessStats() (*models.BusinessStats, error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive,
			COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as new
		FROM businesses`

	stats := &models.BusinessStats{}

	err := db.QueryRow(query).Scan(
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
