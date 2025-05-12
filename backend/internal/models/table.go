package models

import "time"

type Table struct {
	ID        int       `json:"id"`
	Number    int       `json:"number"`
	Seats     int       `json:"seats"`
	Status    string    `json:"status"` // free, occupied, reserved
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TableStatus struct {
	Total    int `json:"total"`
	Free     int `json:"free"`
	Occupied int `json:"occupied"`
	Reserved int `json:"reserved"`
}
