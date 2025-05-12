package models

import "time"

type Request struct {
	ID          int        `json:"id"`
	Branch      string     `json:"branch"`
	SupplierID  int        `json:"supplier_id"`
	Items       []string   `json:"items"`
	Priority    string     `json:"priority"`
	Comment     string     `json:"comment"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
