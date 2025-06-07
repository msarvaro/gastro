package responses

import "time"

// ServiceRequestResponse represents a service request in API responses
type ServiceRequestResponse struct {
	ID             int        `json:"id"`
	BusinessID     int        `json:"business_id"`
	TableID        int        `json:"table_id"`
	TableNumber    int        `json:"table_number"`
	RequestType    string     `json:"request_type"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	RequestedBy    string     `json:"requested_by"`
	AssignedTo     *int       `json:"assigned_to,omitempty"`
	AssignedToName *string    `json:"assigned_to_name,omitempty"`
	Notes          string     `json:"notes"`
	CreatedAt      time.Time  `json:"created_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	ResponseTime   *string    `json:"response_time,omitempty"` // Duration as string
}

// ServiceRequestsListResponse represents a list of service requests
type ServiceRequestsListResponse struct {
	Requests []ServiceRequestResponse `json:"requests"`
	Total    int                      `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

// RequestTypesResponse represents available request types
type RequestTypesResponse struct {
	Types []RequestTypeInfo `json:"types"`
}

// RequestTypeInfo represents information about a request type
type RequestTypeInfo struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Priority    string `json:"default_priority"`
}

// RequestStatusResponse represents the status update response
type RequestStatusResponse struct {
	ID        int       `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}
