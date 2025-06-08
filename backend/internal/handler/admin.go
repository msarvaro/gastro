package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/user"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

type AdminController struct {
	userService user.Service
}

func NewAdminController(userService user.Service) *AdminController {
	return &AdminController{
		userService: userService,
	}
}

func (c *AdminController) GetUsers(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	users, err := c.userService.GetUsers(r.Context(), businessID)
	if err != nil {
		log.Printf("GetUsers: Error getting users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("GetUsers: Retrieved %d users", len(users))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (c *AdminController) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := c.userService.GetUserStats(r.Context())
	if err != nil {
		http.Error(w, "Failed to get stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (c *AdminController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = c.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		log.Printf("DeleteUser: Error deleting user: %v", err)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *AdminController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var userUpdate user.User
	if err := json.NewDecoder(r.Body).Decode(&userUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userUpdate.ID = userID

	// Extract business ID and user role from context to apply same business rules as in CreateUser
	businessID, _ := middleware.GetBusinessIDFromContext(r.Context())
	userRole, _ := middleware.GetUserRoleFromContext(r.Context())

	// If this is a business_id update, enforce business context rules
	if userUpdate.BusinessID > 0 {
		// Only admins can change business_id
		if userRole != "admin" {
			log.Printf("UpdateUser: Non-admin user (role %s) attempted to change business_id", userRole)
			http.Error(w, "Only admins can change business_id", http.StatusForbidden)
			return
		}
	} else if userRole == "manager" || userRole == "admin" {
		// For managers, enforce their own business_id if none provided
		userUpdate.BusinessID = businessID
	}

	log.Printf("UpdateUser: Performing selective update on user ID %d", userID)
	err = c.userService.UpdateUser(r.Context(), &userUpdate)
	if err != nil {
		log.Printf("UpdateUser: Failed to update user: %v", err)
		switch err {
		case user.ErrInvalidUserID:
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
		case user.ErrInvalidUserData:
			http.Error(w, "Invalid user data", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *AdminController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser user.User
	if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
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
	if newUser.BusinessID > 0 {
		// Allow admins to override the business ID
		if userRole == "admin" {
			businessID = newUser.BusinessID
			log.Printf("CreateUser: Admin overriding with business ID %d from request", businessID)
		} else {
			// Non-admins can only create users in their own business
			log.Printf("CreateUser: Non-admin attempted to override business ID, using context business ID %d instead", businessID)
		}
	}

	log.Printf("CreateUser: Using business ID %d for new user", businessID)

	if err := c.userService.CreateUser(r.Context(), &newUser, businessID); err != nil {
		switch err {
		case user.ErrUserAlreadyExists:
			http.Error(w, "User already exists", http.StatusConflict)
		case user.ErrInvalidUserData:
			http.Error(w, "Invalid user data", http.StatusBadRequest)
		default:
			log.Printf("CreateUser: Error creating user: %v", err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *AdminController) GetUserByID(w http.ResponseWriter, r *http.Request) {
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

	userResult, err := c.userService.GetUserByID(r.Context(), userID, businessID)
	if err != nil {
		log.Printf("GetUserByID: Error getting user: %v", err)
		switch err {
		case user.ErrUserNotFound:
			http.Error(w, "User not found", http.StatusNotFound)
		case user.ErrInvalidUserID:
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userResult)
}
