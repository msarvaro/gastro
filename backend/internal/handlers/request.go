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

type RequestHandler struct {
	db *database.DB
}

func NewRequestHandler(db *database.DB) *RequestHandler {
	return &RequestHandler{db: db}
}

func (h *RequestHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	requests, err := h.db.GetAllRequests(businessID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get requests")
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"requests": requests})
}

func (h *RequestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	request, err := h.db.GetRequestByID(id, businessID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Request not found")
		return
	}
	respondWithJSON(w, http.StatusOK, request)
}

func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	var request models.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}
	if err := h.db.CreateRequest(&request, businessID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create request")
		return
	}
	respondWithJSON(w, http.StatusCreated, request)
}

func (h *RequestHandler) Update(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var request models.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid input")
		return
	}
	request.ID = id
	if err := h.db.UpdateRequest(&request, businessID); err != nil {
		log.Printf("Error Update - updating request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update request")
		return
	}
	respondWithJSON(w, http.StatusOK, request)
}

func (h *RequestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.db.DeleteRequest(id, businessID); err != nil {
		log.Printf("Error Delete - deleting request: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete request")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
