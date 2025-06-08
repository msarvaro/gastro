package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"restaurant-management/internal/domain/user"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) user.Repository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	u := &user.User{}

	// Используем временные переменные для сканирования NULL значений
	var createdAt, updatedAt sql.NullTime
	var lastActive sql.NullTime

	err := r.db.QueryRowContext(ctx, `
        SELECT id, username, email, password, role, status, 
               COALESCE(created_at, CURRENT_TIMESTAMP) as created_at,
               COALESCE(updated_at, CURRENT_TIMESTAMP) as updated_at,
               COALESCE(last_active, CURRENT_TIMESTAMP) as last_active
        FROM users 
        WHERE username = $1`,
		username,
	).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.Role,
		&u.Status,
		&createdAt,
		&updatedAt,
		&lastActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No user found with username: %s", username)
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Database error: %v", err)
		return nil, err
	}

	// Присваиваем значения из NullTime
	u.CreatedAt = createdAt.Time
	u.UpdatedAt = updatedAt.Time
	u.LastActive = lastActive.Time

	return u, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int, businessID int) (*user.User, error) {
	u := &user.User{}
	var name sql.NullString

	// First get the requesting user's role
	var requestingUserRole string
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM users WHERE id = $1`, id).Scan(&requestingUserRole)
	if err != nil {
		return nil, err
	}

	// Modify query based on user role
	var query string
	var args []interface{}

	if requestingUserRole == "admin" {
		// Admin can view any user
		query = `
			SELECT id, username, name, email, role, status, business_id, created_at, updated_at
			FROM users
			WHERE id = $1`
		args = []interface{}{id}
	} else {
		// Regular users can only view users from their business
		query = `
			SELECT id, username, name, email, role, status, business_id, created_at, updated_at
			FROM users
			WHERE id = $1 AND business_id = $2`
		args = []interface{}{id, businessID}
	}

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&u.ID, &u.Username, &name, &u.Email, &u.Role, &u.Status, &u.BusinessID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Convert NULL name to empty string if needed
	if name.Valid {
		u.Name = name.String
	} else {
		u.Name = ""
	}

	log.Printf("GetUserByID called with id=%d, businessID=%d, role=%s", id, businessID, requestingUserRole)
	return u, nil
}

func (r *UserRepository) GetUsers(ctx context.Context, businessID int) ([]user.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, name, email, role, status, business_id, created_at, updated_at
		FROM users
		WHERE business_id = $1
		ORDER BY created_at DESC`, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		var name, status sql.NullString
		err := rows.Scan(&u.ID, &u.Username, &name, &u.Email, &u.Role, &status, &u.BusinessID, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if name.Valid {
			u.Name = name.String
		} else {
			u.Name = ""
		}
		if status.Valid {
			u.Status = status.String
		} else {
			u.Status = ""
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) GetUserStats(ctx context.Context) (*user.UserStats, error) {
	stats := &user.UserStats{}

	err := r.db.QueryRowContext(ctx, `
        SELECT 
            COUNT(*) as total,
            COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
            COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive,
            COUNT(CASE WHEN role = 'admin' THEN 1 END) as admins,
            COUNT(CASE WHEN created_at >= NOW() - INTERVAL '7 days' THEN 1 END) as new
        FROM users
    `).Scan(&stats.Total, &stats.Active, &stats.Inactive, &stats.Admins, &stats.New)

	return stats, err
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

func (r *UserRepository) UpdateUser(ctx context.Context, u *user.User) error {
	existingUser, err := r.GetUserByID(ctx, u.ID, u.BusinessID)
	if err != nil {
		log.Printf("UpdateUser: Error retrieving existing user data: %v", err)
		return err
	}

	if u.Username != "" {
		existingUser.Username = u.Username
	}
	if u.Name != "" {
		existingUser.Name = u.Name
	}
	if u.Email != "" {
		existingUser.Email = u.Email
	}
	if u.Role != "" {
		existingUser.Role = u.Role
	}
	if u.Status != "" {
		existingUser.Status = u.Status
	}

	_, err = r.db.ExecContext(ctx, `
        UPDATE users 
        SET username = $1, name = $2, email = $3, role = $4, status = $5, updated_at = $6
        WHERE id = $7`,
		existingUser.Username, existingUser.Name, existingUser.Email,
		existingUser.Role, existingUser.Status, time.Now(), u.ID,
	)

	if err != nil {
		log.Printf("UpdateUser: Error updating user ID %d: %v", u.ID, err)
	}

	return err
}

func (r *UserRepository) CreateUser(ctx context.Context, u *user.User, businessID int) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Проверка на пустое имя - используем username если имя не задано
	if u.Name == "" {
		u.Name = u.Username
		log.Printf("CreateUser DB: Имя было пустым, используем username: %q", u.Name)
	}

	// Отладка: перед выполнением запроса
	log.Printf("CreateUser DB: Saving user - Username: %q, Name: %q, Email: %q, Role: %q, Status: %q, BusinessID: %d",
		u.Username, u.Name, u.Email, u.Role, u.Status, businessID)

	_, err = r.db.ExecContext(ctx, `
        INSERT INTO users (username, name, email, password, role, status, business_id, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)`,
		u.Username, u.Name, u.Email, string(hashedPassword), u.Role, u.Status, businessID, time.Now(),
	)

	if err != nil {
		log.Printf("CreateUser DB: Error creating user: %v", err)
	} else {
		log.Printf("CreateUser DB: User created successfully with business ID: %d", businessID)
	}

	return err
}

// GetUserBusinessID retrieves the business ID associated with a user
func (r *UserRepository) GetUserBusinessID(ctx context.Context, userID int) (int, error) {
	var businessID int

	// First check the users table for a direct association
	err := r.db.QueryRowContext(ctx, `
        SELECT business_id FROM users WHERE id = $1
    `, userID).Scan(&businessID)

	if err == nil && businessID > 0 {
		log.Printf("Found business ID %d for user %d in users table", businessID, userID)
		return businessID, nil
	}

	// If not found in users table, check user_businesses table if it exists
	// This is for systems with many-to-many user-business relationships
	var exists bool
	err = r.db.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_name = 'user_businesses'
        )
    `).Scan(&exists)

	if err != nil {
		log.Printf("Error checking if user_businesses table exists: %v", err)
		return 0, err
	}

	if exists {
		// Check for the user's primary business in user_businesses table
		err = r.db.QueryRowContext(ctx, `
            SELECT business_id FROM user_businesses 
            WHERE user_id = $1 AND is_primary = true
        `, userID).Scan(&businessID)

		if err == nil && businessID > 0 {
			log.Printf("Found primary business ID %d for user %d in user_businesses table", businessID, userID)
			return businessID, nil
		}

		// If no primary business found, get the first business associated with the user
		err = r.db.QueryRowContext(ctx, `
            SELECT business_id FROM user_businesses 
            WHERE user_id = $1 
            LIMIT 1
        `, userID).Scan(&businessID)

		if err == nil && businessID > 0 {
			log.Printf("Found business ID %d for user %d in user_businesses table", businessID, userID)
			return businessID, nil
		}
	}

	// If no business found or there was an error
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error retrieving business ID for user %d: %v", userID, err)
		return 0, err
	}

	log.Printf("No business found for user %d", userID)
	return 0, nil
}

// GetStats retrieves general user statistics
func (r *UserRepository) GetStats(ctx context.Context) (map[string]int, error) {
	stats := map[string]int{}

	// Временные переменные для сканирования
	var total, active, inactive, new int

	// Общее количество пользователей
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Активные пользователи
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&active)
	if err != nil {
		return nil, err
	}
	stats["active"] = active

	// Неактивные пользователи
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE status = 'inactive'").Scan(&inactive)
	if err != nil {
		return nil, err
	}
	stats["inactive"] = inactive

	// Новые пользователи (за последние 24 часа)
	err = r.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM users 
        WHERE created_at >= NOW() - INTERVAL '24 hours'
    `).Scan(&new)
	if err != nil {
		return nil, err
	}
	stats["new"] = new

	return stats, nil
}
