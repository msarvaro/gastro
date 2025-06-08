package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/business"
	"strconv"

	"github.com/gorilla/mux"
)

// BusinessController handles business-related requests
type BusinessController struct {
	businessService business.Service
}

// NewBusinessController creates a new BusinessController
func NewBusinessController(businessService business.Service) *BusinessController {
	return &BusinessController{
		businessService: businessService,
	}
}

// GetAllBusinesses retrieves all businesses
func (c *BusinessController) GetAllBusinesses(w http.ResponseWriter, r *http.Request) {
	businesses, stats, err := c.businessService.GetAllBusinesses(r.Context())
	if err != nil {
		log.Printf("Error getting all businesses: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch businesses")
		return
	}

	response := struct {
		Businesses []business.Business     `json:"businesses"`
		Stats      *business.BusinessStats `json:"stats"`
	}{
		Businesses: businesses,
		Stats:      stats,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetBusinessByID retrieves a specific business
func (c *BusinessController) GetBusinessByID(w http.ResponseWriter, r *http.Request) {
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

	businessResult, err := c.businessService.GetBusinessByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting business %d: %v", id, err)
		switch err {
		case business.ErrBusinessNotFound:
			respondWithError(w, http.StatusNotFound, "Business not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch business")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, businessResult)
}

// CreateBusiness creates a new business
func (c *BusinessController) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	var newBusiness business.Business
	if err := json.NewDecoder(r.Body).Decode(&newBusiness); err != nil {
		log.Printf("Error decoding business: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if err := c.businessService.CreateBusiness(r.Context(), &newBusiness); err != nil {
		log.Printf("Error creating business: %v", err)
		switch err {
		case business.ErrInvalidBusinessData:
			respondWithError(w, http.StatusBadRequest, "Invalid business data")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to create business")
		}
		return
	}

	respondWithJSON(w, http.StatusCreated, newBusiness)
}

// UpdateBusiness updates an existing business
func (c *BusinessController) UpdateBusiness(w http.ResponseWriter, r *http.Request) {
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

	var updatedBusiness business.Business
	if err := json.NewDecoder(r.Body).Decode(&updatedBusiness); err != nil {
		log.Printf("Error decoding business update: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Set the ID from path parameter
	updatedBusiness.ID = id

	if err := c.businessService.UpdateBusiness(r.Context(), &updatedBusiness); err != nil {
		log.Printf("Error updating business %d: %v", id, err)
		switch err {
		case business.ErrBusinessNotFound:
			respondWithError(w, http.StatusNotFound, "Business not found")
		case business.ErrInvalidBusinessData:
			respondWithError(w, http.StatusBadRequest, "Invalid business data")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to update business")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, updatedBusiness)
}

// DeleteBusiness deletes a business
func (c *BusinessController) DeleteBusiness(w http.ResponseWriter, r *http.Request) {
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

	if err := c.businessService.DeleteBusiness(r.Context(), id); err != nil {
		log.Printf("Error deleting business %d: %v", id, err)
		switch err {
		case business.ErrBusinessNotFound:
			respondWithError(w, http.StatusNotFound, "Business not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to delete business")
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Business deleted successfully"})
}

// SetBusinessCookie sets a cookie with the business ID
func (c *BusinessController) SetBusinessCookie(w http.ResponseWriter, r *http.Request) {
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
	businessResult, err := c.businessService.GetBusinessByID(r.Context(), id)
	if err != nil {
		log.Printf("Error checking for existing business %d: %v", id, err)
		switch err {
		case business.ErrBusinessNotFound:
			respondWithError(w, http.StatusNotFound, "Business not found")
		default:
			respondWithError(w, http.StatusInternalServerError, "Failed to verify business")
		}
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
		"business": businessResult,
	})
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
