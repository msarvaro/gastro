package service

import (
	"context"
	"log"
	"restaurant-management/internal/domain/user"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo   user.Repository
	jwtKey string
}

func NewUserService(repo user.Repository, jwtKey string) user.Service {
	return &UserService{
		repo:   repo,
		jwtKey: jwtKey,
	}
}

func (s *UserService) Login(ctx context.Context, req user.LoginRequest) (*user.LoginResponse, error) {
	if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, user.ErrInvalidCredentials
	}

	// Get user by username
	u, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		log.Printf("Login failed for username %s: user not found", req.Username)
		return nil, user.ErrInvalidCredentials
	}

	// Check if user is active
	if u.Status != "active" {
		log.Printf("Login failed for username %s: user is inactive", req.Username)
		return nil, user.ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		log.Printf("Login failed for username %s: invalid password", req.Username)
		return nil, user.ErrInvalidCredentials
	}

	// Get user's business ID
	businessID, err := s.repo.GetUserBusinessID(ctx, u.ID)
	if err != nil {
		log.Printf("Warning: Could not get business ID for user %s: %v", req.Username, err)
		// Continue without business ID (admin might not have a specific business)
		businessID = 0
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":     u.ID,
		"username":    u.Username,
		"role":        u.Role,
		"business_id": businessID,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtKey))
	if err != nil {
		log.Printf("Failed to generate token for user %s: %v", req.Username, err)
		return nil, user.ErrTokenGeneration
	}

	// Determine redirect path based on role
	var redirectPath string
	switch u.Role {
	case "admin":
		redirectPath = "/select-business"
	case "manager":
		redirectPath = "/manager"
	case "waiter":
		redirectPath = "/waiter"
	case "cook":
		redirectPath = "/kitchen"
	default:
		redirectPath = "/"
	}

	return &user.LoginResponse{
		Token:      tokenString,
		Role:       u.Role,
		Redirect:   redirectPath,
		BusinessID: businessID,
	}, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id int, businessID int) (*user.User, error) {
	if id <= 0 {
		return nil, user.ErrInvalidUserID
	}

	return s.repo.GetUserByID(ctx, id, businessID)
}

func (s *UserService) GetUsers(ctx context.Context, businessID int) ([]user.User, error) {
	if businessID <= 0 {
		return nil, user.ErrInvalidUserData
	}

	return s.repo.GetUsers(ctx, businessID)
}

func (s *UserService) CreateUser(ctx context.Context, u *user.User, businessID int) error {
	if businessID <= 0 {
		return user.ErrInvalidUserData
	}

	// Validation
	if strings.TrimSpace(u.Username) == "" {
		return user.ErrInvalidUserData
	}
	if strings.TrimSpace(u.Password) == "" {
		return user.ErrInvalidUserData
	}
	if strings.TrimSpace(u.Email) == "" {
		return user.ErrInvalidUserData
	}
	if strings.TrimSpace(u.Role) == "" {
		return user.ErrInvalidUserData
	}

	// Validate role
	validRoles := map[string]bool{
		"admin":   true,
		"manager": true,
		"waiter":  true,
		"cook":    true,
	}
	if !validRoles[u.Role] {
		return user.ErrInvalidUserData
	}

	// Set default status if not provided
	if u.Status == "" {
		u.Status = "active"
	}

	// Validate status
	if u.Status != "active" && u.Status != "inactive" {
		return user.ErrInvalidUserData
	}

	// Use username as name if name is empty
	if strings.TrimSpace(u.Name) == "" {
		u.Name = u.Username
	}

	// Check if user already exists
	existingUser, _ := s.repo.GetUserByUsername(ctx, u.Username)
	if existingUser != nil {
		return user.ErrUserAlreadyExists
	}

	return s.repo.CreateUser(ctx, u, businessID)
}

func (s *UserService) UpdateUser(ctx context.Context, u *user.User) error {
	if u.ID <= 0 {
		return user.ErrInvalidUserID
	}

	// Validation for non-empty fields
	if u.Username != "" && strings.TrimSpace(u.Username) == "" {
		return user.ErrInvalidUserData
	}
	if u.Email != "" && strings.TrimSpace(u.Email) == "" {
		return user.ErrInvalidUserData
	}
	if u.Role != "" {
		validRoles := map[string]bool{
			"admin":   true,
			"manager": true,
			"waiter":  true,
			"cook":    true,
		}
		if !validRoles[u.Role] {
			return user.ErrInvalidUserData
		}
	}
	if u.Status != "" && u.Status != "active" && u.Status != "inactive" {
		return user.ErrInvalidUserData
	}

	return s.repo.UpdateUser(ctx, u)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	if id <= 0 {
		return user.ErrInvalidUserID
	}

	return s.repo.DeleteUser(ctx, id)
}

func (s *UserService) GetUserStats(ctx context.Context) (*user.UserStats, error) {
	return s.repo.GetUserStats(ctx)
}

func (s *UserService) GetStats(ctx context.Context) (map[string]int, error) {
	return s.repo.GetStats(ctx)
}
