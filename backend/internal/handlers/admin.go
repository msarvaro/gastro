package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/database"
	"restaurant-management/internal/middleware"
	"restaurant-management/internal/models"
	"strconv"

	"github.com/gorilla/mux"
)

type AdminHandler struct {
	db *database.DB
}

func NewAdminHandler(db *database.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	users, err := h.db.GetUsers(businessID)
	if err != nil {
		log.Printf("GetUsers: Ошибка при получении пользователей: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("GetUsers: Получено %d пользователей", len(users))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetUserStats()
	if err != nil {
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.db.DeleteUser(userID)
	if err != nil {
		log.Printf("DeleteUser: Ошибка при удалении пользователя: %v", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.ID = userID

	// Extract business ID and user role from context to apply same business rules as in CreateUser
	businessID, _ := middleware.GetBusinessIDFromContext(r.Context())
	userRole, _ := middleware.GetUserRoleFromContext(r.Context())

	// If this is a business_id update, enforce business context rules
	if user.BusinessID > 0 {
		// Only admins can change business_id
		if userRole != "admin" {
			log.Printf("UpdateUser: Non-admin user (role %s) attempted to change business_id", userRole)
			http.Error(w, "Only admins can change business_id", http.StatusForbidden)
			return
		}
	} else if userRole == "manager" {
		// For managers, enforce their own business_id if none provided
		user.BusinessID = businessID
	}

	log.Printf("UpdateUser: Performing selective update on user ID %d", userID)
	err = h.db.UpdateUser(&user)
	if err != nil {
		log.Printf("UpdateUser: Failed to update user: %v", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract business ID from request context (added by BusinessMiddleware)
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		log.Printf("CreateUser: Could not parse business ID from context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Extract user role from context
	userRole, ok := middleware.GetUserRoleFromContext(r.Context())
	if !ok {
		log.Printf("CreateUser: Could not extract user role from context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("CreateUser: Request from user with role %s", userRole)

	// If the creator is a manager, they must have a business ID
	if userRole == "manager" && businessID == 0 {
		log.Printf("CreateUser: Manager attempted to create user without a business ID")
		http.Error(w, "Managers must specify a business ID", http.StatusBadRequest)
		return
	}

	// Check if user explicitly provided a business_id in the request
	if user.BusinessID > 0 {
		// Allow admins to override the business ID
		if userRole == "admin" {
			businessID = user.BusinessID
			log.Printf("CreateUser: Admin overriding with business ID %d from request", businessID)
		} else {
			// Non-admins can only create users in their own business
			log.Printf("CreateUser: Non-admin attempted to override business ID, using context business ID %d instead", businessID)
		}
	}

	log.Printf("CreateUser: Using business ID %d for new user", businessID)

	if err := h.db.CreateUser(&user, businessID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AdminHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByID(userID, businessID)
	if err != nil {
		log.Printf("GetUserByID: Ошибка при получении пользователя: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
