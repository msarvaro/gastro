package requests

// CreateServiceRequestRequest represents a request to create a new service request
type CreateServiceRequestRequest struct {
	TableID     int    `json:"table_id" validate:"required"`
	RequestType string `json:"request_type" validate:"required,oneof=call_waiter bill water napkins help complaint"`
	Priority    string `json:"priority" validate:"required,oneof=low medium high urgent"`
	RequestedBy string `json:"requested_by"`
	Notes       string `json:"notes"`
}

// UpdateServiceRequestRequest represents a request to update an existing service request
type UpdateServiceRequestRequest struct {
	Status     *string `json:"status,omitempty" validate:"omitempty,oneof=pending acknowledged completed"`
	Priority   *string `json:"priority,omitempty" validate:"omitempty,oneof=low medium high urgent"`
	AssignedTo *int    `json:"assigned_to,omitempty"`
	Notes      *string `json:"notes,omitempty"`
}

// AssignRequestRequest represents a request to assign a service request to a waiter
type AssignRequestRequest struct {
	WaiterID int `json:"waiter_id" validate:"required"`
}

// UpdateRequestStatusRequest represents a request to update request status
type UpdateRequestStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending acknowledged completed"`
}
