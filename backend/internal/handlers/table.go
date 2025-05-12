package handlers

import (
	"encoding/json"
	"net/http"
	"restaurant-management/internal/database"
	"restaurant-management/internal/models"
	"strconv"

	"github.com/gorilla/mux"
)

type TableHandler struct {
	db *database.DB
}

func NewTableHandler(db *database.DB) *TableHandler {
	return &TableHandler{db: db}
}

func (h *TableHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tables, err := h.db.GetAllTables()
	if err != nil {
		http.Error(w, "Failed to get tables", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tables)
}

func (h *TableHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	table, err := h.db.GetTableByID(id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(table)
}

func (h *TableHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.db.GetTableStatus()
	if err != nil {
		http.Error(w, "Failed to get table status", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(status)
}

func (h *TableHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var table models.Table
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	table.ID = id
	if err := h.db.UpdateTableStatus(&table); err != nil {
		http.Error(w, "Failed to update table status", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(table)
}
