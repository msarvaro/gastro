package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/middleware"

	"github.com/gorilla/mux"
)

// KitchenHandler handles kitchen-related HTTP requests
type KitchenHandler struct {
	kitchenService   services.KitchenService
	orderService     services.OrderService
	inventoryService services.InventoryService
}

// NewKitchenHandler creates a new kitchen handler
func NewKitchenHandler(kitchenService services.KitchenService, orderService services.OrderService, inventoryService services.InventoryService) *KitchenHandler {
	return &KitchenHandler{
		kitchenService:   kitchenService,
		orderService:     orderService,
		inventoryService: inventoryService,
	}
}

// GetKitchenOrders returns all orders with status 'preparing'
func (h *KitchenHandler) GetKitchenOrders(w http.ResponseWriter, r *http.Request) {
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		businessID = 0 // Fallback to 0 if not available
	}

	orders, err := h.kitchenService.GetPendingOrders(r.Context(), businessID)
	if err != nil {
		log.Printf("Error GetKitchenOrders - fetching orders: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch kitchen orders")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"orders": orders})
}

// UpdateOrderStatusByCook allows cook to set order status to 'ready'
func (h *KitchenHandler) UpdateOrderStatusByCook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing order ID")
		return
	}
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	var req requests.KitchenUpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get order and verify it's in preparing status
	order, err := h.orderService.GetOrderByID(r.Context(), orderID)
	if err != nil {
		log.Printf("Error UpdateOrderStatusByCook - fetching order %d: %v", orderID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get order details")
		return
	}
	if order == nil {
		respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Only allow updating from 'preparing' to 'ready' by cook
	if order.Status != "preparing" {
		log.Printf("Attempted to update order %d status from %s to %s by cook. Denied.", orderID, order.Status, req.Status)
		respondWithError(w, http.StatusBadRequest, "Order status cannot be updated from current state")
		return
	}

	// Update the status
	err = h.orderService.UpdateOrderStatus(r.Context(), orderID, req.Status)
	if err != nil {
		log.Printf("Error UpdateOrderStatusByCook - updating order %d status: %v", orderID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	// Get updated order for response
	updatedOrder, _ := h.orderService.GetOrderByID(r.Context(), orderID)
	respondWithJSON(w, http.StatusOK, updatedOrder)
}

// GetKitchenHistory returns completed and cancelled orders for kitchen history
func (h *KitchenHandler) GetKitchenHistory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		respondWithError(w, http.StatusBadRequest, "business_id not found in context")
		return
	}

	// Get completed/cancelled orders
	orders, err := h.orderService.GetActiveOrders(r.Context(), businessID)
	if err != nil {
		log.Printf("Error GetKitchenHistory - fetching history: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch kitchen history")
		return
	}

	// Filter for completed orders (this should be done in service layer in real implementation)
	completedOrders := make([]*entity.Order, 0)
	for _, order := range orders {
		if order.Status == "completed" || order.Status == "cancelled" {
			completedOrders = append(completedOrders, order)
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"orders": completedOrders})
}

// GetInventory returns all inventory items
func (h *KitchenHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		respondWithError(w, http.StatusBadRequest, "business_id not found in context")
		return
	}

	items, err := h.inventoryService.GetInventoryItems(r.Context(), businessID)
	if err != nil {
		log.Printf("Error GetInventory - fetching inventory: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch inventory")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// UpdateInventory updates the quantity of an inventory item
func (h *KitchenHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemIDStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing inventory item ID")
		return
	}
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid inventory item ID format")
		return
	}

	var req requests.UpdateInventoryFromKitchenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error UpdateInventory - decoding request for item %d: %v", itemID, err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Basic validation
	if req.Quantity < 0 {
		respondWithError(w, http.StatusBadRequest, "Quantity cannot be negative")
		return
	}

	// Update stock using the inventory service
	err = h.inventoryService.UpdateStock(r.Context(), itemID, req.Quantity, "Kitchen update")
	if err != nil {
		log.Printf("Error UpdateInventory - updating inventory item %d: %v", itemID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update inventory item")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Inventory updated successfully"})
}

// Helper functions (these would ideally be in a shared package)
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error during JSON marshalling"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
