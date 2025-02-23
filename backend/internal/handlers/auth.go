package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"user-management/internal/database"

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
	Token string `json:"token"`
	Role  string `json:"role"`
}

func NewAuthHandler(db *database.DB, jwtKey string) *AuthHandler {
	return &AuthHandler{
		db:     db,
		jwtKey: jwtKey,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt - Username: %s, Password length: %d", req.Username, len(req.Password))

	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("GetUserByUsername error: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Found user - ID: %d, Username: %s, Role: %s", user.ID, user.Username, user.Role)
	log.Printf("Stored password hash: %s", user.Password)
	log.Printf("Comparing password: %s with hash", req.Password)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Password comparison failed: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Password comparison successful")

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
		log.Printf("Error signing token: %v", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	log.Printf("Generated token successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: tokenString,
		Role:  user.Role,
	})
}
