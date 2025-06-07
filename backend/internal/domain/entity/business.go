package entity

import "time"

// Business represents a restaurant or other business
type Business struct {
	ID          int
	Name        string
	Description string
	Address     string
	Phone       string
	Email       string
	Website     string
	Logo        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsActive checks if the business is active
func (b *Business) IsActive() bool {
	return b.Status == "active"
}

// SetStatus updates the business status
func (b *Business) SetStatus(status string) {
	b.Status = status
	b.UpdatedAt = time.Now()
}

func (b *Business) IsOpen(currentTime time.Time) bool {
	// Implementation would check current time against opening/closing times
	return b.IsActive()
}
