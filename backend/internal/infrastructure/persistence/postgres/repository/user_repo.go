package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// userRepository implements the UserRepository interface using PostgreSQL
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	query := `
		SELECT id, username, email, password, name, role, status, business_id, 
		       last_active, created_at, updated_at
		FROM users 
		WHERE id = $1
	`

	var user entity.User
	var businessID sql.NullInt64
	var lastActive sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.Status,
		&businessID,
		&lastActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if businessID.Valid {
		bid := int(businessID.Int64)
		user.BusinessID = &bid
	}

	if lastActive.Valid {
		user.LastActiveAt = &lastActive.Time
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, username, email, password, name, role, status, business_id, 
		       last_active, created_at, updated_at
		FROM users 
		WHERE username = $1
	`

	var user entity.User
	var businessID sql.NullInt64
	var lastActive sql.NullTime

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.Status,
		&businessID,
		&lastActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if businessID.Valid {
		bid := int(businessID.Int64)
		user.BusinessID = &bid
	}

	if lastActive.Valid {
		user.LastActiveAt = &lastActive.Time
	}

	return &user, nil
}

// GetByBusinessID retrieves all users for a business
func (r *userRepository) GetByBusinessID(ctx context.Context, businessID int) ([]*entity.User, error) {
	query := `
		SELECT id, username, email, password, name, role, status, business_id, 
		       last_active, created_at, updated_at
		FROM users 
		WHERE business_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User

	for rows.Next() {
		var user entity.User
		var businessID sql.NullInt64
		var lastActive sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.Status,
			&businessID,
			&lastActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if businessID.Valid {
			bid := int(businessID.Int64)
			user.BusinessID = &bid
		}

		if lastActive.Valid {
			user.LastActiveAt = &lastActive.Time
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetByRole retrieves users by role for a business
func (r *userRepository) GetByRole(ctx context.Context, businessID int, role string) ([]*entity.User, error) {
	query := `
		SELECT id, username, email, password, name, role, status, business_id, 
		       last_active, created_at, updated_at
		FROM users 
		WHERE business_id = $1 AND role = $2
	`

	rows, err := r.db.QueryContext(ctx, query, businessID, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User

	for rows.Next() {
		var user entity.User
		var businessID sql.NullInt64
		var lastActive sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.Status,
			&businessID,
			&lastActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if businessID.Valid {
			bid := int(businessID.Int64)
			user.BusinessID = &bid
		}

		if lastActive.Valid {
			user.LastActiveAt = &lastActive.Time
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Create adds a new user to the database
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO users (username, email, password, name, role, status, business_id, 
		                  last_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	var businessID sql.NullInt64
	if user.BusinessID != nil {
		businessID.Int64 = int64(*user.BusinessID)
		businessID.Valid = true
	}

	var lastActive sql.NullTime
	if user.LastActiveAt != nil {
		lastActive.Time = *user.LastActiveAt
		lastActive.Valid = true
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Name,
		user.Role,
		user.Status,
		businessID,
		lastActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	return err
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, name = $4, role = $5,
			status = $6, business_id = $7, last_active = $8, updated_at = $9
		WHERE id = $10
	`

	var businessID sql.NullInt64
	if user.BusinessID != nil {
		businessID.Int64 = int64(*user.BusinessID)
		businessID.Valid = true
	}

	var lastActive sql.NullTime
	if user.LastActiveAt != nil {
		lastActive.Time = *user.LastActiveAt
		lastActive.Valid = true
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Name,
		user.Role,
		user.Status,
		businessID,
		lastActive,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Delete removes a user
func (r *userRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateLastActiveAt updates the last active timestamp
func (r *userRepository) UpdateLastActiveAt(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET last_active = $1, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdatePassword updates a user's password
func (r *userRepository) UpdatePassword(ctx context.Context, userID int, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, hashedPassword, now, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetActiveWaiters retrieves all active waiters for a business
func (r *userRepository) GetActiveWaiters(ctx context.Context, businessID int) ([]*entity.User, error) {
	query := `
		SELECT id, username, email, password, name, role, status, business_id, 
		       last_active, created_at, updated_at
		FROM users 
		WHERE business_id = $1 AND role = 'waiter' AND status = 'active'
	`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entity.User

	for rows.Next() {
		var user entity.User
		var businessID sql.NullInt64
		var lastActive sql.NullTime

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.Status,
			&businessID,
			&lastActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if businessID.Valid {
			bid := int(businessID.Int64)
			user.BusinessID = &bid
		}

		if lastActive.Valid {
			user.LastActiveAt = &lastActive.Time
		}

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
