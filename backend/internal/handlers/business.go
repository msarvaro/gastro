package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/database"
	"restaurant-management/internal/models"
	"strconv"

	"github.com/gorilla/mux"
)

// BusinessHandler handles business-related requests
type BusinessHandler struct {
	db *database.DB
}

// NewBusinessHandler creates a new BusinessHandler
func NewBusinessHandler(db *database.DB) *BusinessHandler {
	return &BusinessHandler{db: db}
}

// GetAllBusinesses retrieves all businesses
func (h *BusinessHandler) GetAllBusinesses(w http.ResponseWriter, r *http.Request) {
	businesses, err := h.db.GetAllBusinesses()
	if err != nil {
		log.Printf("Error getting all businesses: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch businesses")
		return
	}

	stats, err := h.db.GetBusinessStats()
	if err != nil {
		log.Printf("Error getting business stats: %v", err)
		// Continue with empty stats if fetching stats fails
		stats = &models.BusinessStats{}
	}

	response := struct {
		Businesses []models.Business     `json:"businesses"`
		Stats      *models.BusinessStats `json:"stats"`
	}{
		Businesses: businesses,
		Stats:      stats,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetBusinessByID retrieves a specific business
func (h *BusinessHandler) GetBusinessByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing business ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business ID format")
		return
	}

	business, err := h.db.GetBusinessByID(id)
	if err != nil {
		log.Printf("Error getting business %d: %v", id, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch business")
		return
	}

	respondWithJSON(w, http.StatusOK, business)
}

// CreateBusiness creates a new business
func (h *BusinessHandler) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	var business models.Business
	if err := json.NewDecoder(r.Body).Decode(&business); err != nil {
		log.Printf("Error decoding business: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Set default status if not provided
	if business.Status == "" {
		business.Status = "active"
	}

	if err := h.db.CreateBusiness(&business); err != nil {
		log.Printf("Error creating business: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create business")
		return
	}

	respondWithJSON(w, http.StatusCreated, business)
}

// UpdateBusiness updates an existing business
func (h *BusinessHandler) UpdateBusiness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing business ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business ID format")
		return
	}

	// Check if business exists
	existingBusiness, err := h.db.GetBusinessByID(id)
	if err != nil {
		log.Printf("Error checking for existing business %d: %v", id, err)
		respondWithError(w, http.StatusNotFound, "Business not found")
		return
	}

	// Decode request body
	var updatedBusiness models.Business
	if err := json.NewDecoder(r.Body).Decode(&updatedBusiness); err != nil {
		log.Printf("Error decoding business update: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Set the ID from path parameter
	updatedBusiness.ID = id

	// Preserve created_at timestamp
	updatedBusiness.CreatedAt = existingBusiness.CreatedAt

	if err := h.db.UpdateBusiness(&updatedBusiness); err != nil {
		log.Printf("Error updating business %d: %v", id, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update business")
		return
	}

	respondWithJSON(w, http.StatusOK, updatedBusiness)
}

// DeleteBusiness deletes a business
func (h *BusinessHandler) DeleteBusiness(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing business ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business ID format")
		return
	}

	// Check if business exists
	_, err = h.db.GetBusinessByID(id)
	if err != nil {
		log.Printf("Error checking for existing business %d: %v", id, err)
		respondWithError(w, http.StatusNotFound, "Business not found")
		return
	}

	if err := h.db.DeleteBusiness(id); err != nil {
		log.Printf("Error deleting business %d: %v", id, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete business")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Business deleted successfully"})
}

// SetBusinessCookie sets a cookie with the business ID
func (h *BusinessHandler) SetBusinessCookie(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing business ID")
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid business ID format")
		return
	}

	// Check if business exists
	business, err := h.db.GetBusinessByID(id)
	if err != nil {
		log.Printf("Error checking for existing business %d: %v", id, err)
		respondWithError(w, http.StatusNotFound, "Business not found")
		return
	}

	// Set cookie with business ID
	http.SetCookie(w, &http.Cookie{
		Name:     "business_id",
		Value:    idStr,
		Path:     "/",
		MaxAge:   86400 * 30, // 30 days
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Business selected successfully",
		"business": business,
	})
}
