package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/dto/responses"
	"restaurant-management/internal/presentation/http/middleware"
	"strconv"
	"time"
)

// BusinessHandler handles business-related requests
type BusinessHandler struct {
	businessService services.BusinessService
	userService     services.UserService
}

// NewBusinessHandler creates a new business handler
func NewBusinessHandler(businessService services.BusinessService, userService services.UserService) *BusinessHandler {
	return &BusinessHandler{
		businessService: businessService,
		userService:     userService,
	}
}

// GetBusiness returns a specific business by ID
func (h *BusinessHandler) GetBusiness(w http.ResponseWriter, r *http.Request) {
	// Extract business ID from URL or context
	businessID, err := h.getBusinessIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}

	business, err := h.businessService.GetBusinessByID(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting business: %v", err)
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	response := mapBusinessToResponse(business)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListBusinesses returns all businesses
func (h *BusinessHandler) ListBusinesses(w http.ResponseWriter, r *http.Request) {
	businesses, err := h.businessService.GetAllBusinesses(r.Context())
	if err != nil {
		log.Printf("Error listing businesses: %v", err)
		http.Error(w, "Failed to retrieve businesses", http.StatusInternalServerError)
		return
	}

	var response responses.BusinessListResponse
	for _, business := range businesses {
		response.Businesses = append(response.Businesses, mapBusinessToResponse(business))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateBusiness creates a new business
func (h *BusinessHandler) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	var req requests.CreateBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	business := &entity.Business{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		Logo:        req.Logo,
		Status:      "active", // Default status for new businesses
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := h.businessService.CreateBusiness(r.Context(), business)
	if err != nil {
		log.Printf("Error creating business: %v", err)
		http.Error(w, "Failed to create business", http.StatusInternalServerError)
		return
	}

	// Get the created business to return as response
	createdBusiness, err := h.businessService.GetBusinessByID(r.Context(), business.ID)
	if err != nil {
		log.Printf("Error fetching created business: %v", err)
		http.Error(w, "Business created but couldn't fetch details", http.StatusInternalServerError)
		return
	}

	response := mapBusinessToResponse(createdBusiness)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateBusiness updates an existing business
func (h *BusinessHandler) UpdateBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get existing business
	existingBusiness, err := h.businessService.GetBusinessByID(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting business for update: %v", err)
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	// Update fields if they're provided
	if req.Name != "" {
		existingBusiness.Name = req.Name
	}
	if req.Description != "" {
		existingBusiness.Description = req.Description
	}
	if req.Address != "" {
		existingBusiness.Address = req.Address
	}
	if req.Phone != "" {
		existingBusiness.Phone = req.Phone
	}
	if req.Email != "" {
		existingBusiness.Email = req.Email
	}
	if req.Website != "" {
		existingBusiness.Website = req.Website
	}
	if req.Logo != "" {
		existingBusiness.Logo = req.Logo
	}
	if req.Status != "" {
		existingBusiness.Status = req.Status
	}
	existingBusiness.UpdatedAt = time.Now()

	err = h.businessService.UpdateBusiness(r.Context(), existingBusiness)
	if err != nil {
		log.Printf("Error updating business: %v", err)
		http.Error(w, "Failed to update business", http.StatusInternalServerError)
		return
	}

	// Get the updated business to return as response
	updatedBusiness, err := h.businessService.GetBusinessByID(r.Context(), businessID)
	if err != nil {
		log.Printf("Error fetching updated business: %v", err)
		http.Error(w, "Business updated but couldn't fetch details", http.StatusInternalServerError)
		return
	}

	response := mapBusinessToResponse(updatedBusiness)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteBusiness deletes a business
func (h *BusinessHandler) DeleteBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}

	err = h.businessService.DeleteBusiness(r.Context(), businessID)
	if err != nil {
		log.Printf("Error deleting business: %v", err)
		http.Error(w, "Failed to delete business", http.StatusInternalServerError)
		return
	}

	response := responses.GenericResponse{
		Success: true,
		Message: "Business deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetBusinessUsers returns all users associated with a business
func (h *BusinessHandler) GetBusinessUsers(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}

	users, err := h.businessService.GetBusinessUsers(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting business users: %v", err)
		http.Error(w, "Failed to retrieve business users", http.StatusInternalServerError)
		return
	}

	var response responses.BusinessUsersResponse
	for _, user := range users {
		response.Users = append(response.Users, mapUserToResponse(user))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AddUserToBusiness adds a user to a business
func (h *BusinessHandler) AddUserToBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}

	var req requests.AddUserToBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// First get the user
	user, err := h.userService.GetUserByID(r.Context(), req.UserID)
	if err != nil {
		log.Printf("Error getting user for adding to business: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Then add to business
	err = h.businessService.AddUserToBusiness(r.Context(), user, businessID)
	if err != nil {
		log.Printf("Error adding user to business: %v", err)
		http.Error(w, "Failed to add user to business", http.StatusInternalServerError)
		return
	}

	response := responses.GenericResponse{
		Success: true,
		Message: "User added to business successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveUserFromBusiness removes a user from a business
func (h *BusinessHandler) RemoveUserFromBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessIDFromRequest(r)
	if err != nil {
		http.Error(w, "Invalid business ID", http.StatusBadRequest)
		return
	}

	// Extract user ID from URL path parameter
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.businessService.RemoveUserFromBusiness(r.Context(), userID, businessID)
	if err != nil {
		log.Printf("Error removing user from business: %v", err)
		http.Error(w, "Failed to remove user from business", http.StatusInternalServerError)
		return
	}

	response := responses.GenericResponse{
		Success: true,
		Message: "User removed from business successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SelectBusiness selects a business for the current user session
func (h *BusinessHandler) SelectBusiness(w http.ResponseWriter, r *http.Request) {
	var req requests.BusinessSelectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify the business exists
	business, err := h.businessService.GetBusinessByID(r.Context(), req.BusinessID)
	if err != nil {
		log.Printf("Error getting business for selection: %v", err)
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	// Check if user has access to this business by checking business users
	users, err := h.businessService.GetBusinessUsers(r.Context(), req.BusinessID)
	if err != nil {
		log.Printf("Error getting business users: %v", err)
		http.Error(w, "Failed to check business access", http.StatusInternalServerError)
		return
	}

	hasAccess := false
	for _, user := range users {
		if user.ID == userID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		log.Printf("User %d does not have access to business %d", userID, req.BusinessID)
		http.Error(w, "You do not have access to this business", http.StatusForbidden)
		return
	}

	// Set business ID cookie
	expirationTime := time.Now().Add(24 * time.Hour)
	http.SetCookie(w, &http.Cookie{
		Name:     "business_id",
		Value:    strconv.Itoa(req.BusinessID),
		Expires:  expirationTime,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Get user's role for determining the redirect
	user, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting user for redirect: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Determine redirect path based on user's role
	redirectPath := "/"
	switch user.Role {
	case "admin":
		redirectPath = "/admin"
	case "manager":
		redirectPath = "/manager"
	case "waiter":
		redirectPath = "/waiter"
	case "kitchen":
		redirectPath = "/kitchen"
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "Business selected successfully",
		"redirect": redirectPath,
		"business": mapBusinessToResponse(business),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper method to get business ID from request (URL or context)
func (h *BusinessHandler) getBusinessIDFromRequest(r *http.Request) (int, error) {
	// First try from URL path parameter
	businessIDStr := r.URL.Query().Get("id")
	if businessIDStr != "" {
		return strconv.Atoi(businessIDStr)
	}

	// Then try from URL query parameter
	businessIDStr = r.URL.Query().Get("business_id")
	if businessIDStr != "" {
		return strconv.Atoi(businessIDStr)
	}

	// Finally try from context (set by middleware)
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if ok {
		return businessID, nil
	}

	return 0, errors.New("business ID not found")
}

// Helper function to map business entity to response DTO
func mapBusinessToResponse(business *entity.Business) *responses.BusinessResponse {
	return &responses.BusinessResponse{
		ID:          business.ID,
		Name:        business.Name,
		Description: business.Description,
		Address:     business.Address,
		Phone:       business.Phone,
		Email:       business.Email,
		Website:     business.Website,
		Logo:        business.Logo,
		Status:      business.Status,
		CreatedAt:   business.CreatedAt,
		UpdatedAt:   business.UpdatedAt,
	}
}

// Helper function to map user entity to response DTO
func mapUserToResponse(user *entity.User) *responses.UserResponse {
	return &responses.UserResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Name:         user.Name,
		Role:         user.Role,
		Status:       user.Status,
		BusinessID:   user.BusinessID,
		LastActiveAt: user.LastActiveAt,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
