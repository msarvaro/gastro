package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/dto/responses"
	"strconv"
	"time"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	authService services.AuthService
	userService services.UserService
	jwtKey      string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService services.AuthService, userService services.UserService, jwtKey string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		jwtKey:      jwtKey,
	}
}

// Login authenticates a user and returns a JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login handler: Processing login request")

	var req requests.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Login handler: Error decoding request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Login handler: Login attempt for username: %s", req.Username)

	// Authenticate user
	user, token, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		log.Printf("Login handler: Authentication failed: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Set cookies
	expirationTime := time.Now().Add(24 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Set business ID cookie if available
	if user.BusinessID != nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "business_id",
			Value:    strconv.Itoa(*user.BusinessID),
			Expires:  expirationTime,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		log.Printf("Login handler: Business ID cookie set to: %d", *user.BusinessID)
	}

	// Determine the redirect path based on role
	redirectPath := "/"

	// Only admins without a business should see business selection
	if user.Role == "admin" && user.BusinessID == nil {
		// Redirect admin to business selection if no business is associated
		redirectPath = "/select-business"
		log.Printf("Login handler: Admin user without business, redirecting to: %s", redirectPath)
	} else {
		// For all other roles, or if business ID is already set, redirect to role-specific page
		switch user.Role {
		case "manager":
			redirectPath = "/manager"
		case "waiter":
			redirectPath = "/waiter"
		case "cook":
			redirectPath = "/kitchen"
		}
	}

	// Prepare response
	businessID := 0
	if user.BusinessID != nil {
		businessID = *user.BusinessID
	}

	response := responses.LoginResponse{
		Token:      token,
		Role:       user.Role,
		Redirect:   redirectPath,
		BusinessID: businessID,
		UserID:     user.ID,
		Name:       user.Name,
		Username:   user.Username,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangePassword changes a user's password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req requests.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := h.getUserIDFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Change password
	err := h.authService.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Printf("Change password failed: %v", err)
		http.Error(w, "Failed to change password: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := responses.GenericResponse{
		Success: true,
		Message: "Password changed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ResetPassword initiates a password reset process
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req requests.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Reset password
	err := h.authService.ResetPassword(r.Context(), req.Username)
	if err != nil {
		log.Printf("Reset password failed: %v", err)
		// Don't expose error details to avoid user enumeration
		http.Error(w, "Password reset request processed", http.StatusOK)
		return
	}

	response := responses.GenericResponse{
		Success: true,
		Message: "Password reset request processed. Check your email for instructions.",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper method to get user ID from context
func (h *AuthHandler) getUserIDFromContext(r *http.Request) (int, bool) {
	userIDVal := r.Context().Value("user_id")
	if userIDVal == nil {
		return 0, false
	}

	// Handle different types
	switch v := userIDVal.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		id, err := strconv.Atoi(v)
		if err == nil {
			return id, true
		}
	}

	return 0, false
}
