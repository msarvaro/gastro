package requests

// CreateBusinessRequest represents the data needed to create a new business
type CreateBusinessRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Website     string `json:"website,omitempty"`
	Logo        string `json:"logo,omitempty"`
}

// UpdateBusinessRequest represents the data needed to update a business
type UpdateBusinessRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Address     string `json:"address,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	Website     string `json:"website,omitempty"`
	Logo        string `json:"logo,omitempty"`
	Status      string `json:"status,omitempty"`
}

// AddUserToBusinessRequest represents the data needed to add a user to a business
type AddUserToBusinessRequest struct {
	UserID int `json:"user_id"`
}

// BusinessSelectionRequest represents the data for selecting a business
type BusinessSelectionRequest struct {
	BusinessID int `json:"business_id"`
}
