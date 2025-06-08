package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/user"
	"strconv"
)

type AuthController struct {
	userService user.Service
}

func NewAuthController(userService user.Service) *AuthController {
	return &AuthController{
		userService: userService,
	}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding login request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := c.userService.Login(r.Context(), req)
	if err != nil {
		switch err {
		case user.ErrInvalidCredentials:
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		case user.ErrUserInactive:
			http.Error(w, "User is inactive", http.StatusForbidden)
		case user.ErrTokenGeneration:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		default:
			log.Printf("Login error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Set the auth token in a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    response.Token,
		Path:     "/",
		MaxAge:   86400, // 24 hours in seconds
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	// Set business ID cookie if present
	if response.BusinessID > 0 {
		http.SetCookie(w, &http.Cookie{
			Name:     "business_id",
			Value:    strconv.Itoa(response.BusinessID),
			Path:     "/",
			MaxAge:   86400, // 24 hours in seconds
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})
		log.Printf("Login controller: Business ID cookie set to: %d", response.BusinessID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
