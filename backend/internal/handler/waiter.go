package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/order"
	"restaurant-management/internal/domain/table"
	"restaurant-management/internal/domain/user"
	"restaurant-management/internal/domain/waiter"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

type WaiterController struct {
	waiterService waiter.Service
	orderService  order.Service
	tableService  table.Service
	userService   user.Service
}

func NewWaiterController(orderService order.Service, tableService table.Service, userService user.Service, waiterService waiter.Service) *WaiterController {
	return &WaiterController{
		waiterService: waiterService,
		orderService:  orderService,
		tableService:  tableService,
		userService:   userService,
	}
}

func (c *WaiterController) GetTables(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	tables, err := c.tableService.GetTables(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting tables: %v", err)
		http.Error(w, "Failed to fetch tables", http.StatusInternalServerError)
		return
	}

	stats, err := c.tableService.GetTableStats(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting table stats: %v", err)
		// Continue with empty stats if fetching stats fails
		stats = &table.TableStats{}
	}

	response := struct {
		Tables []table.Table     `json:"tables"`
		Stats  *table.TableStats `json:"stats"`
	}{
		Tables: tables,
		Stats:  stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *WaiterController) UpdateTableStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tableID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid table ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var statusUpdate table.TableStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.tableService.UpdateTableStatus(r.Context(), tableID, statusUpdate, businessID); err != nil {
		log.Printf("Error updating table status: %v", err)

		// Handle specific error types with appropriate user-friendly messages
		switch err {
		case table.ErrTableHasActiveOrders:
			http.Error(w, "Стол имеет активные заказы", http.StatusBadRequest)
		case table.ErrTableNotFound:
			http.Error(w, "Table not found", http.StatusNotFound)
		case table.ErrInvalidTableData:
			http.Error(w, "Invalid table data", http.StatusBadRequest)
		default:
			http.Error(w, "Failed to update table status", http.StatusInternalServerError)
		}
		return
	}

	// Fetch the updated table to return in the response
	updatedTable, err := c.tableService.GetTableByID(r.Context(), tableID)
	if err != nil {
		log.Printf("Warning: Failed to fetch updated table %d: %v", tableID, err)
		// Still return success even if we can't fetch the updated table
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Table status updated successfully"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTable)
}

func (c *WaiterController) GetActiveOrders(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	orders, err := c.orderService.GetActiveOrders(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting active orders: %v", err)
		http.Error(w, "Failed to fetch active orders", http.StatusInternalServerError)
		return
	}

	stats, err := c.orderService.GetOrderStats(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting order stats: %v", err)
		stats = &order.OrderStats{} // Default to empty stats on error
	}

	response := struct {
		Orders []order.Order     `json:"orders"`
		Stats  *order.OrderStats `json:"stats"`
	}{
		Orders: orders,
		Stats:  stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *WaiterController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var orderRequest order.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&orderRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get waiter ID from context
	userID, exists := middleware.GetUserIDFromContext(r.Context())
	if !exists {
		http.Error(w, "user_id not found in context", http.StatusBadRequest)
		return
	}

	createdOrder, err := c.orderService.CreateOrder(r.Context(), orderRequest, userID, businessID)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdOrder)
}

func (c *WaiterController) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
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

	if err := c.orderService.UpdateOrderStatus(r.Context(), orderID, statusUpdate, businessID); err != nil {
		log.Printf("Error updating order status: %v", err)
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Order status updated successfully"})
}

func (c *WaiterController) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	history, err := c.orderService.GetOrderHistory(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting order history: %v", err)
		http.Error(w, "Failed to fetch order history", http.StatusInternalServerError)
		return
	}
	stats, err := c.orderService.GetOrderStats(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting order stats: %v", err)
		stats = &order.OrderStats{}
	}

	response := struct {
		Orders []order.Order     `json:"orders"`
		Stats  *order.OrderStats `json:"stats"`
	}{
		Orders: history,
		Stats:  stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *WaiterController) GetProfile(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	userID, exists := middleware.GetUserIDFromContext(r.Context())
	if !exists {
		http.Error(w, "user_id not found in context", http.StatusBadRequest)
		return
	}

	profile, err := c.waiterService.GetWaiterProfile(r.Context(), int(userID), businessID)
	if err != nil {
		log.Printf("Error getting waiter profile for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch profile information", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
