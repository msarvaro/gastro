package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"restaurant-management/internal/database"
	"restaurant-management/internal/models"

	"github.com/gorilla/mux"
)

// GetTables returns all tables with their current status
func GetTables(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, number, seats, status, created_at, updated_at FROM tables ORDER BY number")
	if err != nil {
		http.Error(w, "Failed to fetch tables", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tables []models.Table
	for rows.Next() {
		var t models.Table
		err := rows.Scan(&t.ID, &t.Number, &t.Seats, &t.Status, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan table", http.StatusInternalServerError)
			return
		}
		tables = append(tables, t)
	}

	var stats struct {
		Total     int     `json:"total"`
		Free      int     `json:"free"`
		Occupied  int     `json:"occupied"`
		Reserved  int     `json:"reserved"`
		Occupancy float64 `json:"occupancy_percentage"`
	}
	for _, table := range tables {
		stats.Total++
		switch table.Status {
		case models.TableStatusFree:
			stats.Free++
		case models.TableStatusOccupied:
			stats.Occupied++
		case models.TableStatusReserved:
			stats.Reserved++
		}
	}
	if stats.Total > 0 {
		stats.Occupancy = float64(stats.Occupied+stats.Reserved) / float64(stats.Total) * 100
	}

	response := struct {
		Tables []models.Table `json:"tables"`
		Stats  interface{}    `json:"stats"`
	}{
		Tables: tables,
		Stats:  stats,
	}
	json.NewEncoder(w).Encode(response)
}

// GetOrders returns all active orders with items
func GetOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, table_id, waiter_id, status, total_amount, comment, created_at, updated_at FROM orders WHERE status NOT IN ('completed', 'cancelled') ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.TotalAmount, &o.Comment, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan order", http.StatusInternalServerError)
			return
		}

		// Fetch order items
		itemRows, err := database.DB.Query("SELECT id, order_id, dish_id, quantity, price, notes FROM order_items WHERE order_id = $1", o.ID)
		if err != nil {
			http.Error(w, "Failed to fetch order items", http.StatusInternalServerError)
			return
		}
		var items []models.OrderItem
		for itemRows.Next() {
			var item models.OrderItem
			err := itemRows.Scan(&item.ID, &item.OrderID, &item.DishID, &item.Quantity, &item.Price, &item.Notes)
			if err != nil {
				http.Error(w, "Failed to scan order item", http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}
		itemRows.Close()
		// Optionally: o.Items = items (if you add []OrderItem to Order struct)
		orders = append(orders, o)
	}

	var stats struct {
		Active     int `json:"active"`
		New        int `json:"new"`
		InProgress int `json:"in_progress"`
		Ready      int `json:"ready"`
	}
	for _, order := range orders {
		switch order.Status {
		case models.OrderStatusNew:
			stats.New++
		case models.OrderStatusInProgress:
			stats.InProgress++
		case models.OrderStatusReady:
			stats.Ready++
		}
		stats.Active++
	}

	response := struct {
		Orders []models.Order `json:"orders"`
		Stats  interface{}    `json:"stats"`
	}{
		Orders: orders,
		Stats:  stats,
	}
	json.NewEncoder(w).Encode(response)
}

// GetOrderHistory returns completed and cancelled orders
func GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, table_id, waiter_id, status, total_amount, comment, created_at, updated_at, completed_at FROM orders WHERE status IN ('completed', 'cancelled') ORDER BY completed_at DESC")
	if err != nil {
		http.Error(w, "Failed to fetch order history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		err := rows.Scan(&o.ID, &o.TableID, &o.WaiterID, &o.Status, &o.TotalAmount, &o.Comment, &o.CreatedAt, &o.UpdatedAt, &o.CompletedAt)
		if err != nil {
			http.Error(w, "Failed to scan order", http.StatusInternalServerError)
			return
		}
		orders = append(orders, o)
	}

	var stats struct {
		Completed   int     `json:"completed"`
		Cancelled   int     `json:"cancelled"`
		TotalAmount float64 `json:"total_amount"`
	}
	for _, order := range orders {
		if order.Status == models.OrderStatusCompleted {
			stats.Completed++
			stats.TotalAmount += order.TotalAmount
		} else {
			stats.Cancelled++
		}
	}

	response := struct {
		Orders []models.Order `json:"orders"`
		Stats  interface{}    `json:"stats"`
	}{
		Orders: orders,
		Stats:  stats,
	}
	json.NewEncoder(w).Encode(response)
}

// CreateOrder creates a new order
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	err = tx.QueryRow("INSERT INTO orders (table_id, waiter_id, status, total_amount, comment, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id", order.TableID, order.WaiterID, order.Status, order.TotalAmount, order.Comment).Scan(&order.ID)
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	for _, item := range order.Items {
		_, err = tx.Exec("INSERT INTO order_items (order_id, dish_id, quantity, price, notes) VALUES ($1, $2, $3, $4, $5)", order.ID, item.DishID, item.Quantity, item.Price, item.Notes)
		if err != nil {
			http.Error(w, "Failed to create order item", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(order)
}

// UpdateOrderStatus updates the status of an order
func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var status struct {
		Status models.OrderStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := database.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2", status.Status, orderID)
	if err != nil {
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	if status.Status == models.OrderStatusCompleted || status.Status == models.OrderStatusCancelled {
		// Optionally, free the table if needed
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
