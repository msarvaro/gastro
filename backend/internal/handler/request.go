package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/request"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

// RequestController handles request-related operations
type RequestController struct {
	requestService request.Service
}

func NewRequestController(requestService request.Service) *RequestController {
	return &RequestController{requestService: requestService}
}

func (c *RequestController) GetAll(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	requests, err := c.requestService.GetAll(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting all requests: %v", err)
		http.Error(w, "Failed to fetch requests", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"requests": requests})
}

func (c *RequestController) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	request, err := c.requestService.GetByID(r.Context(), id, businessID)
	if err != nil {
		log.Printf("Error getting request by ID %d: %v", id, err)
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(request)
}

func (c *RequestController) Create(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var requestReq request.CreateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&requestReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdRequest, err := c.requestService.Create(r.Context(), requestReq, businessID)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdRequest)
}

func (c *RequestController) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var requestReq request.UpdateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&requestReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedRequest, err := c.requestService.Update(r.Context(), id, requestReq, businessID)
	if err != nil {
		log.Printf("Error updating request %d: %v", id, err)
		http.Error(w, "Failed to update request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedRequest)
}

func (c *RequestController) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	if err := c.requestService.Delete(r.Context(), id, businessID); err != nil {
		log.Printf("Error deleting request %d: %v", id, err)
		http.Error(w, "Failed to delete request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
