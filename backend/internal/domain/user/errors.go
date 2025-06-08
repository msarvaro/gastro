package user

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserAlreadyExists is returned when trying to create a user that already exists
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidUserID is returned when an invalid user ID is provided
	ErrInvalidUserID = errors.New("invalid user ID")

	// ErrUserInactive is returned when trying to access an inactive user
	ErrUserInactive = errors.New("user is inactive")

	// ErrInvalidUserData is returned when user data validation fails
	ErrInvalidUserData = errors.New("invalid user data")

	// ErrUnauthorized is returned when a user is not authorized
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTokenGeneration is returned when token generation fails
	ErrTokenGeneration = errors.New("failed to generate token")
)
