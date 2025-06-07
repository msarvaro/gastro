package entity

import "time"

type ServiceRequest struct {
	ID             int
	BusinessID     int
	TableID        int
	RequestType    string // e.g., "call_waiter", "bill", "water", "napkins"
	Status         string // pending, acknowledged, completed
	Priority       string // low, medium, high, urgent
	RequestedBy    string // Could be customer name or table number
	AssignedTo     *int   // Waiter ID
	Notes          string
	CreatedAt      time.Time
	AcknowledgedAt *time.Time
	CompletedAt    *time.Time
}

// Business methods
func (r *ServiceRequest) IsActive() bool {
	return r.Status != "completed"
}

func (r *ServiceRequest) ResponseTime() *time.Duration {
	if r.AcknowledgedAt == nil {
		return nil
	}
	duration := r.AcknowledgedAt.Sub(r.CreatedAt)
	return &duration
}
