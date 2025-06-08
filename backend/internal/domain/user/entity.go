package user

import "time"

// User represents a user entity
type User struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	BusinessID int       `json:"business_id"`
	LastActive time.Time `json:"last_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserStats represents user statistics
type UserStats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
	Admins   int `json:"admins"`
	New      int `json:"new"`
}
