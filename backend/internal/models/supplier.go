package models

import "time"

type Supplier struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Categories []string  `json:"categories"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	Address    string    `json:"address"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
