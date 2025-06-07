package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/services"
	"restaurant-management/internal/presentation/http/dto/requests"
	"restaurant-management/internal/presentation/http/dto/responses"

	"github.com/gorilla/mux"
)

type SupplierHandler struct {
	supplierService services.SupplierService
}

func NewSupplierHandler(supplierService services.SupplierService) *SupplierHandler {
	return &SupplierHandler{
		supplierService: supplierService,
	}
}

// GetSuppliers gets all suppliers for a business
func (h *SupplierHandler) GetSuppliers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	suppliers, err := h.supplierService.GetSuppliersByBusinessID(ctx, businessID)
	if err != nil {
		http.Error(w, "Failed to get suppliers", http.StatusInternalServerError)
		return
	}

	supplierResponses := make([]responses.SupplierResponse, len(suppliers))
	for i, supplier := range suppliers {
		supplierResponses[i] = h.supplierToResponse(supplier)
	}

	response := responses.SuppliersListResponse{
		Suppliers: supplierResponses,
		Total:     len(supplierResponses),
		Page:      1,
		PageSize:  len(supplierResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSupplierByID gets a specific supplier by ID
func (h *SupplierHandler) GetSupplierByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	supplierID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}

	supplier, err := h.supplierService.GetSupplierByID(ctx, supplierID)
	if err != nil {
		http.Error(w, "Supplier not found", http.StatusNotFound)
		return
	}

	response := h.supplierToResponse(supplier)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateSupplier creates a new supplier
func (h *SupplierHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	var req requests.CreateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	supplier := &entity.Supplier{
		BusinessID:    businessID,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		TaxID:         req.TaxID,
		PaymentTerms:  req.PaymentTerms,
		Rating:        0.0,
		IsActive:      true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.supplierService.CreateSupplier(ctx, supplier); err != nil {
		http.Error(w, "Failed to create supplier", http.StatusInternalServerError)
		return
	}

	response := h.supplierToResponse(supplier)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateSupplier updates an existing supplier
func (h *SupplierHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	supplierID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	supplier, err := h.supplierService.GetSupplierByID(ctx, supplierID)
	if err != nil {
		http.Error(w, "Supplier not found", http.StatusNotFound)
		return
	}

	// Update fields if provided
	if req.Name != nil {
		supplier.Name = *req.Name
	}
	if req.ContactPerson != nil {
		supplier.ContactPerson = *req.ContactPerson
	}
	if req.Email != nil {
		supplier.Email = *req.Email
	}
	if req.Phone != nil {
		supplier.Phone = *req.Phone
	}
	if req.Address != nil {
		supplier.Address = *req.Address
	}
	if req.TaxID != nil {
		supplier.TaxID = *req.TaxID
	}
	if req.PaymentTerms != nil {
		supplier.PaymentTerms = *req.PaymentTerms
	}
	if req.Rating != nil {
		supplier.Rating = *req.Rating
	}
	if req.IsActive != nil {
		supplier.IsActive = *req.IsActive
	}

	supplier.UpdatedAt = time.Now()

	if err := h.supplierService.UpdateSupplier(ctx, supplier); err != nil {
		http.Error(w, "Failed to update supplier", http.StatusInternalServerError)
		return
	}

	response := h.supplierToResponse(supplier)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteSupplier deletes a supplier
func (h *SupplierHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	supplierID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}

	if err := h.supplierService.DeleteSupplier(ctx, supplierID); err != nil {
		http.Error(w, "Failed to delete supplier", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetPurchaseOrders gets purchase orders for a supplier
func (h *SupplierHandler) GetPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	supplierID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid supplier ID", http.StatusBadRequest)
		return
	}

	orders, err := h.supplierService.GetPurchaseOrdersBySupplier(ctx, supplierID)
	if err != nil {
		http.Error(w, "Failed to get purchase orders", http.StatusInternalServerError)
		return
	}

	orderResponses := make([]responses.PurchaseOrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = h.purchaseOrderToResponse(order)
	}

	response := responses.PurchaseOrdersListResponse{
		Orders:   orderResponses,
		Total:    len(orderResponses),
		Page:     1,
		PageSize: len(orderResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreatePurchaseOrder creates a new purchase order
func (h *SupplierHandler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	var req requests.CreatePurchaseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Calculate total amount from items
	var totalAmount float64
	for _, item := range req.Items {
		totalAmount += item.Quantity * item.UnitPrice
	}

	order := &entity.PurchaseOrder{
		BusinessID:       businessID,
		SupplierID:       req.SupplierID,
		OrderNumber:      req.OrderNumber,
		Status:           "draft",
		TotalAmount:      totalAmount,
		ExpectedDelivery: req.ExpectedDelivery,
		CreatedBy:        ctx.Value("user_id").(int),
		Notes:            req.Notes,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := h.supplierService.CreatePurchaseOrder(ctx, order); err != nil {
		http.Error(w, "Failed to create purchase order", http.StatusInternalServerError)
		return
	}

	response := h.purchaseOrderToResponse(order)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Helper methods
func (h *SupplierHandler) supplierToResponse(supplier *entity.Supplier) responses.SupplierResponse {
	return responses.SupplierResponse{
		ID:            supplier.ID,
		BusinessID:    supplier.BusinessID,
		Name:          supplier.Name,
		ContactPerson: supplier.ContactPerson,
		Email:         supplier.Email,
		Phone:         supplier.Phone,
		Address:       supplier.Address,
		TaxID:         supplier.TaxID,
		PaymentTerms:  supplier.PaymentTerms,
		Rating:        supplier.Rating,
		IsActive:      supplier.IsActive,
		CreatedAt:     supplier.CreatedAt,
		UpdatedAt:     supplier.UpdatedAt,
	}
}

func (h *SupplierHandler) purchaseOrderToResponse(order *entity.PurchaseOrder) responses.PurchaseOrderResponse {
	items := make([]responses.PurchaseOrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = responses.PurchaseOrderItemResponse{
			ID:              item.ID,
			PurchaseOrderID: item.PurchaseOrderID,
			InventoryItemID: item.InventoryItemID,
			ItemName:        "", // This would need to be populated from inventory
			Quantity:        item.Quantity,
			UnitPrice:       item.UnitPrice,
			TotalPrice:      item.TotalPrice,
			ReceivedQty:     item.ReceivedQty,
			Notes:           item.Notes,
		}
	}

	return responses.PurchaseOrderResponse{
		ID:               order.ID,
		BusinessID:       order.BusinessID,
		SupplierID:       order.SupplierID,
		SupplierName:     "", // This would need to be populated from supplier
		OrderNumber:      order.OrderNumber,
		Status:           order.Status,
		TotalAmount:      order.TotalAmount,
		ExpectedDelivery: order.ExpectedDelivery,
		ActualDelivery:   order.ActualDelivery,
		Items:            items,
		CreatedBy:        order.CreatedBy,
		CreatedByName:    "", // This would need to be populated from user
		ApprovedBy:       order.ApprovedBy,
		ApprovedByName:   nil, // This would need to be populated from user
		Notes:            order.Notes,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	}
}
