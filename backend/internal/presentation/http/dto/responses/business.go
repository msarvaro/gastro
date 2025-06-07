package responses

import "time"

// BusinessResponse represents a business entity
type BusinessResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Website     string    `json:"website,omitempty"`
	Logo        string    `json:"logo,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BusinessListResponse represents a list of businesses
type BusinessListResponse struct {
	Businesses []*BusinessResponse `json:"businesses"`
}

// BusinessUsersResponse represents a list of users associated with a business
type BusinessUsersResponse struct {
	Users []*UserResponse `json:"users"`
}

// UserResponse represents a user entity
type UserResponse struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	BusinessID   *int       `json:"business_id,omitempty"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
