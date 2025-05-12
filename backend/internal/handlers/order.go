package handlers

import (
	"encoding/json"
	"net/http"
	"restaurant-management/internal/database"
	"restaurant-management/internal/models"
	"strconv"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	db *database.DB
}

func NewOrderHandler(db *database.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	orders, err := h.db.GetAllOrders()
	if err != nil {
		http.Error(w, "Failed to get orders", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	order, err := h.db.GetOrderByID(id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	if err := h.db.CreateOrder(&order); err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var order models.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	order.ID = id
	if err := h.db.UpdateOrder(&order); err != nil {
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.db.GetOrderStatus()
	if err != nil {
		http.Error(w, "Failed to get order status", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(status)
}

func (h *OrderHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	orders, err := h.db.GetOrderHistory()
	if err != nil {
		http.Error(w, "Failed to get order history", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(orders)
}
