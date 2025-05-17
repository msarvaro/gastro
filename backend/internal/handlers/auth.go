package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/database"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db     *database.DB
	jwtKey string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Role     string `json:"role"`
	Redirect string `json:"redirect"`
}

func NewAuthHandler(db *database.DB, jwtKey string) *AuthHandler {
	return &AuthHandler{
		db:     db,
		jwtKey: jwtKey,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login handler: Processing login request")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Login handler: Error decoding request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Login handler: Login attempt for username: %s", req.Username)

	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("Login handler: GetUserByUsername error: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Login handler: User found - ID: %d, Role: %s", user.ID, user.Role)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Login handler: Password comparison failed: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Login handler: Password verified successfully")

	expirationTime := time.Now().Add(24 * time.Hour)
	if req.Remember {
		expirationTime = time.Now().Add(30 * 24 * time.Hour)
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.jwtKey))
	if err != nil {
		log.Printf("Login handler: Error signing token: %v", err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	log.Printf("Login handler: Token generated successfully")

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    tokenString,
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	redirectPath := "/"
	switch user.Role {
	case "admin":
		redirectPath = "/admin"
	case "manager":
		redirectPath = "/manager"
	case "waiter":
		redirectPath = "/waiter"
	case "cook":
		redirectPath = "/kitchen"
	}

	log.Printf("Login handler: Sending successful response with redirect to: %s", redirectPath)

	response := LoginResponse{
		Token:    tokenString,
		Role:     user.Role,
		Redirect: redirectPath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
