package user

import "context"

// Repository defines the interface for user data operations
type Repository interface {
	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id int, businessID int) (*User, error)

	// GetUsers retrieves users by business ID
	GetUsers(ctx context.Context, businessID int) ([]User, error)

	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *User, businessID int) error

	// UpdateUser updates an existing user
	UpdateUser(ctx context.Context, user *User) error

	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id int) error

	// GetUserStats retrieves user statistics
	GetUserStats(ctx context.Context) (*UserStats, error)

	// GetUserBusinessID retrieves the business ID associated with a user
	GetUserBusinessID(ctx context.Context, userID int) (int, error)

	// GetStats retrieves general user statistics
	GetStats(ctx context.Context) (map[string]int, error)

	// GetByGoogleEmail finds a user by Google email
	GetByGoogleEmail(ctx context.Context, googleEmail string) (*User, error)
}
