package responses

// LoginResponse represents the response sent after successful authentication
type LoginResponse struct {
	Token      string `json:"token"`
	Role       string `json:"role"`
	Redirect   string `json:"redirect"`
	BusinessID int    `json:"business_id,omitempty"`
	UserID     int    `json:"user_id"`
	Name       string `json:"name,omitempty"`
	Username   string `json:"username"`
}

// GenericResponse represents a generic response with a message
type GenericResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
