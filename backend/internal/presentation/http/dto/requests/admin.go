package requests

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Username   string `json:"username" validate:"required,min=3,max=50"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=6"`
	Name       string `json:"name" validate:"required,max=100"`
	Role       string `json:"role" validate:"required,oneof=admin manager waiter kitchen"`
	BusinessID *int   `json:"business_id,omitempty"`
}

// UpdateUserRequest represents the request to update an existing user
type UpdateUserRequest struct {
	Username   *string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email      *string `json:"email,omitempty" validate:"omitempty,email"`
	Name       *string `json:"name,omitempty" validate:"omitempty,max=100"`
	Role       *string `json:"role,omitempty" validate:"omitempty,oneof=admin manager waiter kitchen"`
	Status     *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
	BusinessID *int    `json:"business_id,omitempty"`
}

// ListUsersRequest represents the request to list users with filters
type ListUsersRequest struct {
	BusinessID int    `json:"business_id,omitempty"`
	Role       string `json:"role,omitempty"`
	Status     string `json:"status,omitempty"`
	Page       int    `json:"page,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}
