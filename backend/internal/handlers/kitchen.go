package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"restaurant-management/internal/database"
	"restaurant-management/internal/middleware"
	"restaurant-management/internal/models"
	"strconv"

	"github.com/gorilla/mux"
)

type KitchenHandler struct {
	db *database.DB
}

func NewKitchenHandler(db *database.DB) *KitchenHandler {
	return &KitchenHandler{db: db}
}

// Helper function to send JSON error responses (duplicate, but keeping for now)
func (h *KitchenHandler) respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Helper function to send JSON responses (duplicate, but keeping for now)
func (h *KitchenHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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

// GetKitchenOrders returns all orders with status 'preparing'
func (h *KitchenHandler) GetKitchenOrders(w http.ResponseWriter, r *http.Request) {
	// Extract business_id from context
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		businessID = 0 // Fallback to 0 if not available
	}

	orders, err := h.db.GetOrdersByStatus("preparing", businessID)
	if err != nil {
		log.Printf("Error GetKitchenOrders - fetching orders: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to fetch kitchen orders")
		return
	}

	// You might want to add kitchen-specific stats here later if needed

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{"orders": orders})
}

// UpdateOrderStatusByCook allows cook to set order status to 'ready'
func (h *KitchenHandler) UpdateOrderStatusByCook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr, ok := vars["id"]
	if !ok {
		h.respondWithError(w, http.StatusBadRequest, "Missing order ID")
		return
	}
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid order ID format")
		return
	}

	// Extract business_id from context
	businessID, ok := middleware.GetBusinessIDFromContext(r.Context())
	if !ok {
		businessID = 0 // Fallback to 0 if not available
	}

	// Verify the current status is 'preparing' before allowing update to 'ready'
	order, err := h.db.GetOrderByID(orderID, businessID)
	if err != nil {
		log.Printf("Error UpdateOrderStatusByCook - fetching order %d: %v", orderID, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get order details")
		return
	}
	if order == nil {
		h.respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}
	// Only allow updating from 'preparing' to 'ready' by cook
	if order.Status != models.OrderStatusPreparing {
		log.Printf("Attempted to update order %d status from %s to ready by cook. Denied.", orderID, order.Status)
		h.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Order status is '%s', cannot set to 'ready'", order.Status))
		return
	}

	// Update the status to 'ready'
	// Reuse the waiter's UpdateOrder logic for status update
	// Note: This might need refinement if cook-specific status update logic is different,
	// but for now, it reuses the core database update and time setting logic.

	// Manually update the order object and call the database method
	order.Status = models.OrderStatusReady
	// The UpdateOrder method in DB should handle setting UpdatedAt and CompletedAt/CancelledAt based on status

	if err := h.db.UpdateOrder(order); err != nil {
		log.Printf("Error UpdateOrderStatusByCook - updating order %d status: %v", orderID, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	h.respondWithJSON(w, http.StatusOK, order)
}

// GetKitchenHistory returns completed and cancelled orders for kitchen history (status 'ready' in this context)
func (h *KitchenHandler) GetKitchenHistory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		h.respondWithError(w, http.StatusBadRequest, "business_id not found in context")
		return
	}

	orders, err := h.db.GetOrderHistoryWithItems(businessID)
	if err != nil {
		log.Printf("Error GetKitchenHistory - fetching history: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to fetch kitchen history")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{"orders": orders})
}

// GetInventory returns all inventory items
func (h *KitchenHandler) GetInventory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		h.respondWithError(w, http.StatusBadRequest, "business_id not found in context")
		return
	}

	items, err := h.db.GetAllInventory(businessID)
	if err != nil {
		log.Printf("Error GetInventory - fetching inventory: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to fetch inventory")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

// UpdateInventory updates the quantity of an inventory item
func (h *KitchenHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemIDStr, ok := vars["id"]
	if !ok {
		h.respondWithError(w, http.StatusBadRequest, "Missing inventory item ID")
		return
	}
	itemID, err := strconv.Atoi(itemIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid inventory item ID format")
		return
	}

	var item models.Inventory
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		log.Printf("Error UpdateInventory - decoding request for item %d: %v", itemID, err)
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Basic validation
	if item.Quantity < 0 {
		h.respondWithError(w, http.StatusBadRequest, "Quantity cannot be negative")
		return
	}

	// Set the ID from the URL parameter
	item.ID = itemID

	// Get context for logging user ID
	ctx := r.Context()

	if err := h.db.UpdateInventory(ctx, &item); err != nil {
		log.Printf("Error UpdateInventory - updating inventory item %d: %v", itemID, err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update inventory item")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Inventory updated successfully"})
}

// GetOrderHistory returns completed and cancelled orders for kitchen history (status 'ready' in this context)
func (h *KitchenHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		h.respondWithError(w, http.StatusBadRequest, "business_id not found in context")
		return
	}

	orders, err := h.db.GetOrderHistoryWithItems(businessID)
	if err != nil {
		log.Printf("Error GetOrderHistory - fetching history: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to fetch order history")
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{"orders": orders})
}
