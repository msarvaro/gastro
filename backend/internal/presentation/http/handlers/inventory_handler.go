package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/dto/responses"
	"restaurant-management/internal/presentation/http/middleware"

	"github.com/gorilla/mux"
)

// InventoryHandler handles inventory-related HTTP requests
type InventoryHandler struct {
	inventoryService services.InventoryService
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(inventoryService services.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

// GetAll retrieves all inventory items for the current business
func (h *InventoryHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	items, err := h.inventoryService.GetInventoryItems(r.Context(), businessID)
	if err != nil {
		http.Error(w, "Failed to get inventory", http.StatusInternalServerError)
		return
	}

	itemResponses := make([]*responses.InventoryItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = mapInventoryItemToResponse(item)
	}

	response := responses.InventoryListResponse{
		Items: itemResponses,
		Total: len(itemResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetByID retrieves a specific inventory item by ID
func (h *InventoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid inventory item ID", http.StatusBadRequest)
		return
	}

	// Note: In the entity approach, we would need to implement GetByID in the service
	// For now, we'll get all items and filter by ID
	items, err := h.inventoryService.GetInventoryItems(r.Context(), businessID)
	if err != nil {
		http.Error(w, "Failed to get inventory item", http.StatusInternalServerError)
		return
	}

	var item *entity.InventoryItem
	for _, i := range items {
		if i.ID == id {
			item = i
			break
		}
	}

	if item == nil {
		http.Error(w, "Inventory item not found", http.StatusNotFound)
		return
	}

	itemResponse := mapInventoryItemToResponse(item)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itemResponse)
}

// Create creates a new inventory item
func (h *InventoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	var req requests.CreateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Create inventory item entity
	item := &entity.InventoryItem{
		BusinessID:      businessID,
		Name:            req.Name,
		SKU:             req.SKU,
		Category:        req.Category,
		Unit:            req.Unit,
		CurrentStock:    req.CurrentStock,
		MinimumStock:    req.MinimumStock,
		MaximumStock:    req.MaximumStock,
		ReorderPoint:    req.ReorderPoint,
		Cost:            req.Cost,
		SupplierID:      req.SupplierID,
		ExpiryDate:      req.ExpiryDate,
		StorageLocation: req.StorageLocation,
		IsActive:        true, // Default to active
	}

	err := h.inventoryService.AddInventoryItem(r.Context(), item)
	if err != nil {
		log.Printf("Error Create - creating inventory item: %v", err)
		http.Error(w, "Failed to create", http.StatusInternalServerError)
		return
	}

	itemResponse := mapInventoryItemToResponse(item)
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itemResponse)
}

// Update updates an existing inventory item
func (h *InventoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Business ID not found", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid inventory item ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateInventoryItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Get existing item first
	items, err := h.inventoryService.GetInventoryItems(r.Context(), businessID)
	if err != nil {
		http.Error(w, "Failed to get inventory item", http.StatusInternalServerError)
		return
	}

	var existingItem *entity.InventoryItem
	for _, i := range items {
		if i.ID == id {
			existingItem = i
			break
		}
	}

	if existingItem == nil {
		http.Error(w, "Inventory item not found", http.StatusNotFound)
		return
	}

	// Apply updates
	updatedItem := *existingItem
	updatedItem.BusinessID = businessID

	if req.Name != nil {
		updatedItem.Name = *req.Name
	}
	if req.SKU != nil {
		updatedItem.SKU = *req.SKU
	}
	if req.Category != nil {
		updatedItem.Category = *req.Category
	}
	if req.Unit != nil {
		updatedItem.Unit = *req.Unit
	}
	if req.CurrentStock != nil {
		updatedItem.CurrentStock = *req.CurrentStock
	}
	if req.MinimumStock != nil {
		updatedItem.MinimumStock = *req.MinimumStock
	}
	if req.MaximumStock != nil {
		updatedItem.MaximumStock = *req.MaximumStock
	}
	if req.ReorderPoint != nil {
		updatedItem.ReorderPoint = *req.ReorderPoint
	}
	if req.Cost != nil {
		updatedItem.Cost = *req.Cost
	}
	if req.SupplierID != nil {
		updatedItem.SupplierID = req.SupplierID
	}
	if req.ExpiryDate != nil {
		updatedItem.ExpiryDate = req.ExpiryDate
	}
	if req.StorageLocation != nil {
		updatedItem.StorageLocation = *req.StorageLocation
	}
	if req.IsActive != nil {
		updatedItem.IsActive = *req.IsActive
	}

	err = h.inventoryService.UpdateInventoryItem(r.Context(), &updatedItem)
	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	itemResponse := mapInventoryItemToResponse(&updatedItem)
	log.Printf("Inventory item updated: %v", updatedItem)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itemResponse)
}

// Delete deletes an inventory item
func (h *InventoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid inventory item ID", http.StatusBadRequest)
		return
	}

	err = h.inventoryService.RemoveInventoryItem(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to map inventory entity to response DTO
func mapInventoryItemToResponse(item *entity.InventoryItem) *responses.InventoryItemResponse {
	response := &responses.InventoryItemResponse{
		ID:              item.ID,
		BusinessID:      item.BusinessID,
		Name:            item.Name,
		SKU:             item.SKU,
		Category:        item.Category,
		Unit:            item.Unit,
		CurrentStock:    item.CurrentStock,
		MinimumStock:    item.MinimumStock,
		MaximumStock:    item.MaximumStock,
		ReorderPoint:    item.ReorderPoint,
		Cost:            item.Cost,
		SupplierID:      item.SupplierID,
		ExpiryDate:      item.ExpiryDate,
		StorageLocation: item.StorageLocation,
		IsActive:        item.IsActive,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
		// Computed fields from business logic
		NeedsReorder: item.NeedsReorder(),
		IsExpired:    item.IsExpired(),
		IsLowStock:   item.IsLowStock(),
	}

	// Determine stock status
	if item.IsLowStock() {
		response.StockStatus = "critical"
	} else if item.CurrentStock <= item.MinimumStock*1.5 {
		response.StockStatus = "low"
	} else if item.CurrentStock >= item.MaximumStock*0.8 {
		response.StockStatus = "high"
	} else {
		response.StockStatus = "normal"
	}

	return response
}
