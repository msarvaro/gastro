package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/supplier"
	"restaurant-management/internal/middleware"
	"strconv"

	"github.com/gorilla/mux"
)

// SupplierController handles supplier-related operations
type SupplierController struct {
	supplierService supplier.Service
}

func NewSupplierController(supplierService supplier.Service) *SupplierController {
	return &SupplierController{supplierService: supplierService}
}

func (c *SupplierController) GetAll(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	suppliers, err := c.supplierService.GetAll(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting all suppliers: %v", err)
		http.Error(w, "Failed to fetch suppliers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"suppliers": suppliers})
}

func (c *SupplierController) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	supplier, err := c.supplierService.GetByID(r.Context(), id, businessID)
	if err != nil {
		log.Printf("Error getting supplier by ID %d: %v", id, err)
		http.Error(w, "Supplier not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(supplier)
}

func (c *SupplierController) Create(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var supplierReq supplier.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&supplierReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdSupplier, err := c.supplierService.Create(r.Context(), supplierReq, businessID)
	if err != nil {
		log.Printf("Error creating supplier: %v", err)
		http.Error(w, "Failed to create supplier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdSupplier)
}

func (c *SupplierController) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var supplierReq supplier.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&supplierReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updatedSupplier, err := c.supplierService.Update(r.Context(), id, supplierReq, businessID)
	if err != nil {
		log.Printf("Error updating supplier %d: %v", id, err)
		http.Error(w, "Failed to update supplier", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedSupplier)
}

func (c *SupplierController) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	if err := c.supplierService.Delete(r.Context(), id, businessID); err != nil {
		log.Printf("Error deleting supplier %d: %v", id, err)
		http.Error(w, "Failed to delete supplier", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
