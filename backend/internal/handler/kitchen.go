package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/inventory"
	"restaurant-management/internal/domain/order"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

type KitchenController struct {
	orderService     order.Service
	inventoryService inventory.Service
}

func NewKitchenController(orderService order.Service, inventoryService inventory.Service) *KitchenController {
	return &KitchenController{
		orderService:     orderService,
		inventoryService: inventoryService,
	}
}

func (c *KitchenController) GetKitchenOrders(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	orders, err := c.orderService.GetKitchenOrders(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting kitchen orders: %v", err)
		http.Error(w, "Failed to fetch kitchen orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"orders": orders})
}

func (c *KitchenController) UpdateOrderStatusByCook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var statusUpdate order.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.orderService.UpdateOrderStatusByCook(r.Context(), orderID, statusUpdate, businessID); err != nil {
		log.Printf("Error updating order status by cook: %v", err)
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	// Fetch the updated order to return in the response (like the old handler)
	updatedOrder, err := c.orderService.GetOrderByID(r.Context(), orderID, businessID)
	if err != nil {
		log.Printf("Warning: Failed to fetch updated order %d: %v", orderID, err)
		// Still return success even if we can't fetch the updated order
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Order status updated successfully"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedOrder)
}

func (c *KitchenController) GetKitchenHistory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	orders, err := c.orderService.GetOrderHistory(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting kitchen history: %v", err)
		http.Error(w, "Failed to fetch kitchen history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"orders": orders})
}

func (c *KitchenController) GetInventory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	inventory, err := c.inventoryService.GetAllInventory(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting inventory for kitchen: %v", err)
		http.Error(w, "Failed to fetch inventory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"items": inventory})
}

func (c *KitchenController) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid inventory ID", http.StatusBadRequest)
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
		http.Error(w, "Failed to update inventory item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}
