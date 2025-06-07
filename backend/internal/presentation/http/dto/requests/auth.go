package requests

// LoginRequest represents the data needed to authenticate a user
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ChangePasswordRequest represents the data needed to change a user's password
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ResetPasswordRequest represents the data needed to reset a user's password
type ResetPasswordRequest struct {
	Username string `json:"username"`
}
