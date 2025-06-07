package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"restaurant-management/internal/domain/consts" // Added import
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// AuthService implements the auth service interface
type AuthService struct {
	userRepo repository.UserRepository
	// Could include a JWT manager/helper here
	jwtSecret    string
	tokenExpiry  time.Duration
	refreshToken map[string]string // Map of refresh tokens to user IDs (should use a proper store in production)
}

func CreateAuthService(userRepo repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		jwtSecret:    jwtSecret,
		tokenExpiry:  time.Hour * 24, // 24 hours
		refreshToken: make(map[string]string),
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(ctx context.Context, username, password string) (*entity.User, string, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Check if user is active
	if user.Status != consts.UserStatusActive { // Changed "active" to consts.UserStatusActive
		return nil, "", errors.New("account is inactive")
	}

	// Verify password
	if !verifyPassword(password, user.Password) {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	// Update last active time
	_ = s.userRepo.UpdateLastActiveAt(ctx, user.ID)

	return user, token, nil
}

// ValidateToken validates a JWT token and returns the user
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*entity.User, error) {
	// In a real implementation, would decode the JWT and validate it
	// Here's a simplified version:

	// This is where you'd verify the token signature and extract claims
	// For simplicity, let's assume we can extract a user ID from the token
	userID := 0 // placeholder for extracted user ID from token

	// Get the user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// Update last active time
	_ = s.userRepo.UpdateLastActiveAt(ctx, user.ID)

	return user, nil
}

// RefreshToken refreshes a JWT token
func (s *AuthService) RefreshToken(ctx context.Context, token string) (string, error) {
	// In a real implementation, would validate the refresh token and issue new tokens
	// Here's a simplified version:
	return "new-token", nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID int, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verify old password
	if !verifyPassword(oldPassword, user.Password) {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update password
	return s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
}

// ResetPassword resets a user's password and sends it via email
func (s *AuthService) ResetPassword(ctx context.Context, username string) error {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		// Don't reveal that the user doesn't exist to prevent user enumeration
		return nil
	}

	// Generate a secure random password
	newPassword, err := generateRandomPassword(12)
	if err != nil {
		return errors.New("failed to generate password")
	}

	// Hash the new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	// Update the password
	err = s.userRepo.UpdatePassword(ctx, user.ID, hashedPassword)
	if err != nil {
		return errors.New("failed to update password")
	}

	// In a real application, you would send the new password via email
	// For this example, we'll just return success
	return nil
}

// Helper functions

// generateToken generates a JWT token for a user
func (s *AuthService) generateToken(user *entity.User) (string, error) {
	expirationTime := time.Now().Add(s.tokenExpiry)

	// Create JWT claims with user information
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     expirationTime.Unix(),
	}

	// Add business_id if the user has one
	if user.BusinessID != nil {
		claims["business_id"] = *user.BusinessID
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// verifyPassword checks if a plain text password matches a hashed password
func verifyPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// hashPassword hashes a password
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// generateRandomPassword generates a secure random password
func generateRandomPassword(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
