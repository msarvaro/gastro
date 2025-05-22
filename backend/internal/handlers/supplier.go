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

type SupplierHandler struct {
	db *database.DB
}

func NewSupplierHandler(db *database.DB) *SupplierHandler {
	return &SupplierHandler{db: db}
}

func (h *SupplierHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	suppliers, err := h.db.GetAllSuppliers(businessID)
	if err != nil {
		http.Error(w, "Failed to get suppliers", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"suppliers": suppliers})
}

func (h *SupplierHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	supplier, err := h.db.GetSupplierByID(id, businessID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(supplier)
}

func (h *SupplierHandler) Create(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	var supplier models.Supplier
	if err := json.NewDecoder(r.Body).Decode(&supplier); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err := h.db.CreateSupplier(&supplier, businessID); err != nil {
		http.Error(w, "Failed to create", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(supplier)
}

func (h *SupplierHandler) Update(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var supplier models.Supplier
	if err := json.NewDecoder(r.Body).Decode(&supplier); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	supplier.ID = id
	if err := h.db.UpdateSupplier(&supplier, businessID); err != nil {
		log.Printf("Error Update - updating supplier: %v", err)
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(supplier)
}

func (h *SupplierHandler) Delete(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Business ID not found")
		return
	}
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.db.DeleteSupplier(id, businessID); err != nil {
		log.Printf("Error Delete - deleting supplier: %v", err)
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
