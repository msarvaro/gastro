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
	"time"

	"github.com/gorilla/mux"
)

// Helper function to send JSON error responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// Helper function to send JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error during JSON marshalling"}`)) // Fallback
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

type WaiterHandler struct {
	db *database.DB
}

func NewWaiterHandler(db *database.DB) *WaiterHandler {
	return &WaiterHandler{db: db}
}

// GetTables returns all tables with their current status
func (h *WaiterHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	tables, err := h.db.GetAllTables()
	if err != nil {
		log.Printf("Error GetTables - fetching tables: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch tables")
		return
	}

	stats, err := h.db.GetTableStats()
	if err != nil {
		log.Printf("Error GetTables - fetching table stats: %v", err)
		// Continue with empty stats if fetching stats fails, but log it
		stats = &models.TableStats{}
	}

	response := struct {
		Tables []models.Table     `json:"tables"`
		Stats  *models.TableStats `json:"stats"`
	}{
		Tables: tables,
		Stats:  stats,
	}
	respondWithJSON(w, http.StatusOK, response)
}

// GetOrders returns all active orders with items
func (h *WaiterHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.db.GetActiveOrdersWithItems()
	if err != nil {
		log.Printf("Error GetOrders - fetching active orders: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch active orders")
		return
	}

	stats, err := h.db.GetOrderStatus()
	if err != nil {
		log.Printf("Error GetOrders - fetching order stats: %v", err)
		stats = &models.OrderStats{} // Default to empty stats on error
	}

	response := struct {
		Orders []models.Order     `json:"orders"`
		Stats  *models.OrderStats `json:"stats"`
	}{
		Orders: orders,
		Stats:  stats,
	}
	respondWithJSON(w, http.StatusOK, response)
}

// GetOrderHistory returns completed and cancelled orders
func (h *WaiterHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	orders, err := h.db.GetOrderHistoryWithItems()
	if err != nil {
		log.Printf("Error GetOrderHistory - fetching order history: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch order history")
		return
	}

	// Assuming stats are not strictly necessary for history or can be defaulted
	stats := &models.OrderStats{}

	response := struct {
		Orders []models.Order     `json:"orders"`
		Stats  *models.OrderStats `json:"stats"` // Kept for consistency if needed later
	}{
		Orders: orders,
		Stats:  stats,
	}
	respondWithJSON(w, http.StatusOK, response)
}

// CreateOrder creates a new order
func (h *WaiterHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error CreateOrder - decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if req.TableID < 0 { // Ensure TableID is provided
		respondWithError(w, http.StatusBadRequest, "Table ID is required")
		return
	}

	if len(req.Items) == 0 {
		respondWithError(w, http.StatusBadRequest, "Order must contain at least one item")
		return
	}

	waiterID32, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: Waiter ID not found in token")
		return
	}
	waiterID := int(waiterID32)

	table, err := h.db.GetTableByID(req.TableID)
	if err != nil {
		log.Printf("Error CreateOrder - fetching table %d: %v", req.TableID, err)
		respondWithError(w, http.StatusInternalServerError, "Error fetching table details")
		return
	}
	if table == nil {
		respondWithError(w, http.StatusNotFound, fmt.Sprintf("Table with ID %d not found", req.TableID))
		return
	}
	originalTableStatus := table.Status

	order := models.Order{
		TableID:     req.TableID,
		WaiterID:    waiterID,
		Status:      models.OrderStatusNew,
		Comment:     req.Comment,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Items:       make([]models.OrderItem, 0, len(req.Items)),
		TotalAmount: 0,
	}

	for _, itemInput := range req.Items {
		if itemInput.DishID == 0 { // Basic validation for DishID
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Dish ID is required for item: %v", itemInput))
			return
		}
		if itemInput.Quantity <= 0 { // Basic validation for Quantity
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Quantity must be positive for dish ID %d", itemInput.DishID))
			return
		}
		dish, err := h.db.GetDishByID(itemInput.DishID)
		if err != nil {
			log.Printf("Error CreateOrder - fetching dish %d: %v", itemInput.DishID, err)
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get dish details for ID %d", itemInput.DishID))
			return
		}
		if dish == nil {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Dish with ID %d not found", itemInput.DishID))
			return
		}
		if !dish.IsAvailable {
			respondWithError(w, http.StatusConflict, fmt.Sprintf("Dish '%s' (ID: %d) is currently not available", dish.Name, dish.ID))
			return
		}

		orderItem := models.OrderItem{
			DishID:   itemInput.DishID,
			Name:     dish.Name,
			Quantity: itemInput.Quantity,
			Price:    dish.Price, // Price at the time of order
			Total:    float64(itemInput.Quantity) * dish.Price,
			Notes:    itemInput.Notes,
		}
		order.Items = append(order.Items, orderItem)
		order.TotalAmount += orderItem.Total
	}

	createdOrder, err := h.db.CreateOrderAndItems(&order)
	if err != nil {
		log.Printf("Error CreateOrder - saving order: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create order in database")
		return
	}

	if originalTableStatus == models.TableStatusFree {
		if err := h.db.UpdateTableStatus(req.TableID, string(models.TableStatusOccupied)); err != nil {
			log.Printf("Warning: CreateOrder - failed to update table %d status to occupied: %v. Order created, but table status might be inconsistent.", req.TableID, err)
			// Not returning an error to the client here as the order was created.
			// This is a warning for server logs.
		}
	}

	respondWithJSON(w, http.StatusCreated, createdOrder)
}

// UpdateOrderStatus updates the status of an order
func (h *WaiterHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
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

	var reqBody models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error UpdateOrderStatus - decoding request for order %d: %v", orderID, err)
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if reqBody.Status == "" { // Validate that status is not empty
		respondWithError(w, http.StatusBadRequest, "Status is required")
		return
	}

	order, err := h.db.GetOrderByID(orderID)
	if err != nil {
		log.Printf("Error UpdateOrderStatus - fetching order %d: %v", orderID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get order details")
		return
	}
	if order == nil {
		respondWithError(w, http.StatusNotFound, "Order not found")
		return
	}

	// Update order fields
	order.Status = reqBody.Status
	order.UpdatedAt = time.Now()

	nilTime := (*time.Time)(nil) // Helper for setting nullable time fields to null

	switch order.Status {
	case models.OrderStatusCompleted:
		now := time.Now()
		order.CompletedAt = &now
		order.CancelledAt = nilTime
	case models.OrderStatusCancelled:
		now := time.Now()
		order.CancelledAt = &now
		order.CompletedAt = nilTime
	default:
		// For other statuses, ensure these are null if previously set
		order.CompletedAt = nilTime
		order.CancelledAt = nilTime
	}

	if err := h.db.UpdateOrder(order); err != nil {
		log.Printf("Error UpdateOrderStatus - updating order %d: %v", orderID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update order status")
		return
	}

	// If order is completed or cancelled, check if the table should be freed
	if order.Status == models.OrderStatusCompleted || order.Status == models.OrderStatusCancelled {
		isLast, errCheck := h.db.IsLastActiveOrderForTable(order.TableID, order.ID)
		if errCheck != nil {
			log.Printf("Warning: UpdateOrderStatus - failed to check if order %d was last active for table %d: %v. Order status updated, but table status may need manual check.", order.ID, order.TableID, errCheck)
		} else if isLast {
			if err := h.db.UpdateTableStatus(order.TableID, string(models.TableStatusFree)); err != nil {
				log.Printf("Warning: UpdateOrderStatus - failed to update table %d status to free: %v. Order status updated, but table status might be inconsistent.", order.TableID, err)
			}
		}
	}
	respondWithJSON(w, http.StatusOK, order) // Return the updated order
}

// GetProfile returns the waiter's profile information (placeholder)
func (h *WaiterHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID32, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: User ID not found")
		return
	}
	userID := int(userID32)

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Printf("Error GetProfile - fetching user %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch profile information")
		return
	}
	if user == nil {
		respondWithError(w, http.StatusNotFound, "User profile not found")
		return
	}
	// For security, create a specific profile response model if not all user fields should be returned
	respondWithJSON(w, http.StatusOK, user)
}
