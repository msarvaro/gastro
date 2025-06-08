package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/order"
	"restaurant-management/internal/middleware"
)

type ManagerController struct {
	orderService order.Service
}

func NewManagerController(orderService order.Service) *ManagerController {
	return &ManagerController{
		orderService: orderService,
	}
}

func (c *ManagerController) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// For now just return a basic dashboard response
	dashboard := map[string]string{
		"message": "Manager dashboard",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

func (c *ManagerController) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	orders, err := c.orderService.GetOrderHistory(r.Context(), businessID)
	if err != nil {
		log.Printf("Error retrieving order history: %v", err)
		if orders == nil {
			orders = []order.Order{} // Return empty array instead of null
		}
	}

	if orders == nil {
		orders = []order.Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}
