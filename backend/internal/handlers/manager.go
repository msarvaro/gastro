package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"restaurant-management/internal/database"
	"restaurant-management/internal/models"
)

type ManagerHandler struct {
	db *database.DB
}

func NewManagerHandler(db *database.DB) *ManagerHandler {
	return &ManagerHandler{db: db}
}

func (h *ManagerHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Для проверки авторизации достаточно вернуть успешный статус
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"role":   "manager",
	})
}

func (h *ManagerHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	orders, err := h.db.GetOrderHistoryWithItems()
	if err != nil {
		log.Printf("DEBUG: GetOrderHistoryWithItems returned error: %v (type: %T)", err, err)
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode([]models.Order{})
			return
		}
		log.Printf("Error retrieving order history (not sql.ErrNoRows): %v", err)
		http.Error(w, "Failed to get order history", http.StatusInternalServerError)
		return
	}

	if orders == nil {
		orders = []models.Order{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}
