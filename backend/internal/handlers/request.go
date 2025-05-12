package handlers

import (
	"encoding/json"
	"net/http"
	"restaurant-management/internal/database"
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
	requests, err := h.db.GetAllRequests()
	if err != nil {
		http.Error(w, "Failed to get requests", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"requests": requests})
}

func (h *RequestHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	request, err := h.db.GetRequestByID(id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(request)
}

func (h *RequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request models.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err := h.db.CreateRequest(&request); err != nil {
		http.Error(w, "Failed to create", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(request)
}

func (h *RequestHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var request models.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	request.ID = id
	if err := h.db.UpdateRequest(&request); err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(request)
}

func (h *RequestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.db.DeleteRequest(id); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
