package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
	"user-management/configs"
	"user-management/internal/models"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	*sql.DB
}

func NewDB(config *configs.Config) (*DB, error) {
	db, err := sql.Open("postgres", config.GetDBConnString())
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}

	// Используем временные переменные для сканирования NULL значений
	var createdAt, updatedAt sql.NullTime
	var lastActive sql.NullTime

	err := db.QueryRow(`
        SELECT id, username, email, password, role, status, 
               COALESCE(created_at, CURRENT_TIMESTAMP) as created_at,
               COALESCE(updated_at, CURRENT_TIMESTAMP) as updated_at,
               COALESCE(last_active, CURRENT_TIMESTAMP) as last_active
        FROM users 
        WHERE username = $1`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Status,
		&createdAt,
		&updatedAt,
		&lastActive,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with username: %s", username)
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Database error: %v", err)
		return nil, err
	}

	// Присваиваем значения из NullTime
	user.CreatedAt = createdAt.Time
	user.UpdatedAt = updatedAt.Time
	user.LastActive = lastActive.Time

	return user, nil
}

func (db *DB) GetUsers(page, limit int, status, role, search string) ([]models.User, int, error) {
	// Базовый запрос
	query := `
        SELECT id, username, email, role, status, last_active
        FROM users
        WHERE 1=1
    `
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}

	// Добавляем фильтры
	if status != "" {
		query += ` AND status = $` + strconv.Itoa(len(args)+1)
		countQuery += ` AND status = $` + strconv.Itoa(len(args)+1)
		args = append(args, status)
	}

	if role != "" {
		query += ` AND role = $` + strconv.Itoa(len(args)+1)
		countQuery += ` AND role = $` + strconv.Itoa(len(args)+1)
		args = append(args, role)
	}

	if search != "" {
		query += ` AND (username ILIKE $` + strconv.Itoa(len(args)+1) +
			` OR email ILIKE $` + strconv.Itoa(len(args)+1) + `)`
		countQuery += ` AND (username ILIKE $` + strconv.Itoa(len(args)+1) +
			` OR email ILIKE $` + strconv.Itoa(len(args)+1) + `)`
		args = append(args, "%"+search+"%")
	}

	// Получаем общее количество
	var total int
	err := db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Добавляем пагинацию
	query += ` ORDER BY id DESC LIMIT $` + strconv.Itoa(len(args)+1) +
		` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, limit, (page-1)*limit)

	// Выполняем запрос
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var lastActive sql.NullTime
		err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Status, &lastActive)
		if err != nil {
			return nil, 0, err
		}
		if lastActive.Valid {
			u.LastActive = lastActive.Time
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (db *DB) GetUserStats() (*models.UserStats, error) {
	stats := &models.UserStats{}

	err := db.QueryRow(`
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

func (db *DB) DeleteUser(id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = $1", id)
	return err
}

func (db *DB) UpdateUser(user *models.User) error {
	_, err := db.Exec(`
        UPDATE users 
        SET username = $1, email = $2, role = $3, status = $4, updated_at = $5
        WHERE id = $6`,
		user.Username, user.Email, user.Role, user.Status, time.Now(), user.ID,
	)
	return err
}

func (db *DB) CreateUser(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
        INSERT INTO users (username, email, password, role, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $6)`,
		user.Username, user.Email, string(hashedPassword), user.Role, user.Status, time.Now(),
	)
	return err
}

func (db *DB) GetStats() (map[string]int, error) {
	stats := map[string]int{}

	// Временные переменные для сканирования
	var total, active, inactive, new int

	// Общее количество пользователей
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Активные пользователи
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&active)
	if err != nil {
		return nil, err
	}
	stats["active"] = active

	// Неактивные пользователи
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'inactive'").Scan(&inactive)
	if err != nil {
		return nil, err
	}
	stats["inactive"] = inactive

	// Новые пользователи (за последние 24 часа)
	err = db.QueryRow(`
        SELECT COUNT(*) FROM users 
        WHERE created_at >= NOW() - INTERVAL '24 hours'
    `).Scan(&new)
	if err != nil {
		return nil, err
	}
	stats["new"] = new

	return stats, nil
}
