package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/table"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

type TableController struct {
	tableService table.Service
}

func NewTableController(tableService table.Service) *TableController {
	return &TableController{
		tableService: tableService,
	}
}

func (c *TableController) GetTableByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid table ID", http.StatusBadRequest)
		return
	}

	tableData, err := c.tableService.GetTableByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting table by ID %d: %v", id, err)
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tableData)
}

func (c *TableController) UpdateTableStatus(w http.ResponseWriter, r *http.Request) {
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

	var req table.TableStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := c.tableService.UpdateTableStatus(r.Context(), tableID, req, businessID); err != nil {
		log.Printf("Error updating table status: %v", err)
		http.Error(w, "Failed to update table status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Table status updated successfully"})
}

func (c *TableController) GetTableStats(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	stats, err := c.tableService.GetTableStats(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting table stats: %v", err)
		http.Error(w, "Failed to fetch table stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
