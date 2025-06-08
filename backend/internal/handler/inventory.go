package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/inventory"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

type InventoryController struct {
	inventoryService inventory.Service
}

func NewInventoryController(inventoryService inventory.Service) *InventoryController {
	return &InventoryController{
		inventoryService: inventoryService,
	}
}

func (c *InventoryController) GetAll(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	items, err := c.inventoryService.GetAllInventory(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting all inventory: %v", err)
		http.Error(w, "Failed to fetch inventory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"items": items})
}

func (c *InventoryController) GetByID(w http.ResponseWriter, r *http.Request) {
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

	item, err := c.inventoryService.GetInventoryByID(r.Context(), id, businessID)
	if err != nil {
		log.Printf("Error getting inventory item by ID %d: %v", id, err)
		switch err {
		case inventory.ErrInventoryItemNotFound:
			http.Error(w, "Inventory item not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to fetch inventory item", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (c *InventoryController) Create(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var item inventory.Inventory
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.inventoryService.CreateInventory(r.Context(), &item, businessID); err != nil {
		log.Printf("Error creating inventory item: %v", err)
		switch err {
		case inventory.ErrInvalidInventoryData:
			http.Error(w, "Invalid inventory data", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to create inventory item", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (c *InventoryController) Update(w http.ResponseWriter, r *http.Request) {
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

	var item inventory.Inventory
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	item.ID = id

	if err := c.inventoryService.UpdateInventory(r.Context(), &item, businessID); err != nil {
		log.Printf("Error updating inventory item %d: %v", id, err)
		switch err {
		case inventory.ErrInventoryItemNotFound:
			http.Error(w, "Inventory item not found", http.StatusNotFound)
		case inventory.ErrInvalidInventoryData:
			http.Error(w, "Invalid inventory data", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update inventory item", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (c *InventoryController) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := c.inventoryService.DeleteInventory(r.Context(), id, businessID); err != nil {
		log.Printf("Error deleting inventory item %d: %v", id, err)
		switch err {
		case inventory.ErrInventoryItemNotFound:
			http.Error(w, "Inventory item not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to delete inventory item", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
