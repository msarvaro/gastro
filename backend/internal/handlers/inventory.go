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

type InventoryHandler struct {
	db *database.DB
}

func NewInventoryHandler(db *database.DB) *InventoryHandler {
	return &InventoryHandler{db: db}
}

func (h *InventoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	items, err := h.db.GetAllInventory(businessID)
	if err != nil {
		http.Error(w, "Failed to get inventory", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
}

func (h *InventoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	item, err := h.db.GetInventoryByID(id, businessID)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(item)
}

func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	var item models.Inventory
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	item.BusinessID = businessID
	if err := h.db.CreateInventory(&item); err != nil {
		log.Printf("Error Create - creating inventory item: %v", err)
		http.Error(w, "Failed to create", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var item models.Inventory
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	item.ID = id
	item.BusinessID = businessID
	ctx := r.Context()
	if err := h.db.UpdateInventory(ctx, &item); err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}
	log.Printf("Inventory item updated: %v", item)
	json.NewEncoder(w).Encode(item)
}

func (h *InventoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.db.DeleteInventory(id, businessID); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
