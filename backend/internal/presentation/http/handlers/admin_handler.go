package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/dto/responses"
	"restaurant-management/internal/presentation/http/middleware"

	"github.com/gorilla/mux"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	userService     services.UserService
	businessService services.BusinessService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userService services.UserService, businessService services.BusinessService) *AdminHandler {
	return &AdminHandler{
		userService:     userService,
		businessService: businessService,
	}
}

// GetUsers retrieves all users for the current business
func (h *AdminHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	users, err := h.userService.ListUsers(r.Context(), businessID, nil)
	if err != nil {
		log.Printf("GetUsers: Error retrieving users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userResponses := make([]*responses.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = mapUserToResponse(user)
	}

	log.Printf("GetUsers: Retrieved %d users", len(users))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponses)
}

// GetStats retrieves administrative statistics
func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	// For now, return basic stats. This could be expanded to use a dedicated admin service
	stats := responses.AdminStatsResponse{
		TotalUsers:      0,
		ActiveUsers:     0,
		UsersByRole:     make(map[string]int),
		UsersByStatus:   make(map[string]int),
		UsersByBusiness: make(map[string]int),
		SystemHealth: responses.SystemHealthResponse{
			Status: "healthy",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// DeleteUser deletes a user by ID
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		log.Printf("DeleteUser: Error deleting user: %v", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateUser updates an existing user
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current user from context for authorization
	currentUserRole, _ := middleware.GetUserRoleFromContext(r.Context())
	businessID, _ := middleware.GetBusinessIDFromContext(r.Context())

	// Get existing user
	existingUser, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Apply updates to existing user
	updatedUser := *existingUser

	if req.Username != nil {
		updatedUser.Username = *req.Username
	}
	if req.Email != nil {
		updatedUser.Email = *req.Email
	}
	if req.Name != nil {
		updatedUser.Name = *req.Name
	}
	if req.Role != nil {
		updatedUser.Role = *req.Role
	}
	if req.Status != nil {
		updatedUser.Status = *req.Status
	}

	// Apply business rules similar to the old handler
	if req.BusinessID != nil {
		// Only admins can change business_id
		if currentUserRole != "admin" {
			log.Printf("UpdateUser: Non-admin user (role %s) attempted to change business_id", currentUserRole)
			http.Error(w, "Only admins can change business_id", http.StatusForbidden)
			return
		}
		updatedUser.BusinessID = req.BusinessID
	} else if currentUserRole == "manager" {
		// For managers, enforce their own business_id if none provided
		updatedUser.BusinessID = &businessID
	}

	log.Printf("UpdateUser: Performing update on user ID %d", userID)
	err = h.userService.UpdateUser(r.Context(), &updatedUser)
	if err != nil {
		log.Printf("UpdateUser: Failed to update user: %v", err)
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateUser creates a new user
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req requests.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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
	if req.BusinessID != nil {
		// Allow admins to override the business ID
		if userRole == "admin" {
			businessID = *req.BusinessID
			log.Printf("CreateUser: Admin overriding with business ID %d from request", businessID)
		} else {
			// Non-admins can only create users in their own business
			log.Printf("CreateUser: Non-admin attempted to override business ID, using context business ID %d instead", businessID)
		}
	}

	log.Printf("CreateUser: Using business ID %d for new user", businessID)

	// Create user entity
	user := &entity.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password, // Will be hashed by the service
		Name:     req.Name,
		Role:     req.Role,
		Status:   "active", // Default status
	}

	if businessID > 0 {
		user.BusinessID = &businessID
	}

	_, err := h.userService.CreateUser(r.Context(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetUserByID retrieves a specific user by ID
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

	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("GetUserByID: Error retrieving user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if user belongs to the current business (for non-admin users)
	if user.BusinessID == nil || *user.BusinessID != businessID {
		currentUserRole, _ := middleware.GetUserRoleFromContext(r.Context())
		if currentUserRole != "admin" {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
	}

	userResponse := mapUserToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResponse)
}
