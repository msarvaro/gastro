package handlers

import (
	"encoding/json"
	"net/http"
	"restaurant-management/internal/database"
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
	orders, err := h.db.GetOrderHistory()
	if err != nil {
		http.Error(w, "Failed to get order history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
