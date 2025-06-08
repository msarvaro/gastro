package business

import "time"

// Business represents a business entity
type Business struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Address     string    `json:"address,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	Email       string    `json:"email,omitempty"`
	Website     string    `json:"website,omitempty"`
	Logo        string    `json:"logo,omitempty"`
	Status      string    `json:"status"` // active, inactive, suspended
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BusinessStats represents business statistics
type BusinessStats struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Inactive int `json:"inactive"`
	New      int `json:"new"` // created in the last 7 days
}
