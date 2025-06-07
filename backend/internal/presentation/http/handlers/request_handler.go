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

type RequestHandler struct {
	requestService services.RequestService
}

func NewRequestHandler(requestService services.RequestService) *RequestHandler {
	return &RequestHandler{
		requestService: requestService,
	}
}

// GetRequests gets all service requests for a business
func (h *RequestHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	serviceRequests, err := h.requestService.GetActiveRequestsByBusinessID(ctx, businessID)
	if err != nil {
		http.Error(w, "Failed to get requests", http.StatusInternalServerError)
		return
	}

	requestResponses := make([]responses.ServiceRequestResponse, len(serviceRequests))
	for i, req := range serviceRequests {
		requestResponses[i] = h.requestToResponse(req)
	}

	response := responses.ServiceRequestsListResponse{
		Requests: requestResponses,
		Total:    len(requestResponses),
		Page:     1,
		PageSize: len(requestResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRequestByID gets a specific service request by ID
func (h *RequestHandler) GetRequestByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	serviceRequest, err := h.requestService.GetRequestByID(ctx, requestID)
	if err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	response := h.requestToResponse(serviceRequest)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateRequest creates a new service request
func (h *RequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	var req requests.CreateServiceRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	serviceRequest := &entity.ServiceRequest{
		BusinessID:  businessID,
		TableID:     req.TableID,
		RequestType: req.RequestType,
		Status:      "pending",
		Priority:    req.Priority,
		RequestedBy: req.RequestedBy,
		Notes:       req.Notes,
		CreatedAt:   time.Now(),
	}

	if err := h.requestService.CreateRequest(ctx, serviceRequest); err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	response := h.requestToResponse(serviceRequest)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateRequest updates an existing service request
func (h *RequestHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateServiceRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	serviceRequest, err := h.requestService.GetRequestByID(ctx, requestID)
	if err != nil {
		http.Error(w, "Request not found", http.StatusNotFound)
		return
	}

	// Update fields if provided
	if req.Status != nil {
		serviceRequest.Status = *req.Status

		// Set timestamps based on status
		now := time.Now()
		switch *req.Status {
		case "acknowledged":
			serviceRequest.AcknowledgedAt = &now
		case "completed":
			serviceRequest.CompletedAt = &now
		}
	}
	if req.Priority != nil {
		serviceRequest.Priority = *req.Priority
	}
	if req.AssignedTo != nil {
		serviceRequest.AssignedTo = req.AssignedTo
	}
	if req.Notes != nil {
		serviceRequest.Notes = *req.Notes
	}

	if err := h.requestService.UpdateRequest(ctx, serviceRequest); err != nil {
		http.Error(w, "Failed to update request", http.StatusInternalServerError)
		return
	}

	response := h.requestToResponse(serviceRequest)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateRequestStatus updates the status of a service request
func (h *RequestHandler) UpdateRequestStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateRequestStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.requestService.UpdateRequestStatus(ctx, requestID, req.Status); err != nil {
		http.Error(w, "Failed to update request status", http.StatusInternalServerError)
		return
	}

	response := responses.RequestStatusResponse{
		ID:        requestID,
		Status:    req.Status,
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AssignRequest assigns a service request to a waiter
func (h *RequestHandler) AssignRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	var req requests.AssignRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.requestService.AssignRequestToWaiter(ctx, requestID, req.WaiterID); err != nil {
		http.Error(w, "Failed to assign request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AcknowledgeRequest acknowledges a service request
func (h *RequestHandler) AcknowledgeRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	if err := h.requestService.AcknowledgeRequest(ctx, requestID); err != nil {
		http.Error(w, "Failed to acknowledge request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CompleteRequest marks a service request as completed
func (h *RequestHandler) CompleteRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	requestID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	if err := h.requestService.CompleteRequest(ctx, requestID); err != nil {
		http.Error(w, "Failed to complete request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetRequestTypes returns available request types
func (h *RequestHandler) GetRequestTypes(w http.ResponseWriter, r *http.Request) {
	types := []responses.RequestTypeInfo{
		{
			Value:       "call_waiter",
			Label:       "Call Waiter",
			Description: "Customer needs assistance from waiter",
			Priority:    "medium",
		},
		{
			Value:       "bill",
			Label:       "Request Bill",
			Description: "Customer wants to pay the bill",
			Priority:    "high",
		},
		{
			Value:       "water",
			Label:       "Water",
			Description: "Customer needs water",
			Priority:    "low",
		},
		{
			Value:       "napkins",
			Label:       "Napkins",
			Description: "Customer needs napkins",
			Priority:    "low",
		},
		{
			Value:       "help",
			Label:       "General Help",
			Description: "Customer needs general assistance",
			Priority:    "medium",
		},
		{
			Value:       "complaint",
			Label:       "Complaint",
			Description: "Customer has a complaint",
			Priority:    "urgent",
		},
	}

	response := responses.RequestTypesResponse{
		Types: types,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods
func (h *RequestHandler) requestToResponse(req *entity.ServiceRequest) responses.ServiceRequestResponse {
	var responseTime *string
	if req.AcknowledgedAt != nil {
		duration := req.AcknowledgedAt.Sub(req.CreatedAt)
		durationStr := duration.String()
		responseTime = &durationStr
	}

	return responses.ServiceRequestResponse{
		ID:             req.ID,
		BusinessID:     req.BusinessID,
		TableID:        req.TableID,
		TableNumber:    req.TableID, // Assuming table ID matches table number for now
		RequestType:    req.RequestType,
		Status:         req.Status,
		Priority:       req.Priority,
		RequestedBy:    req.RequestedBy,
		AssignedTo:     req.AssignedTo,
		AssignedToName: nil, // This would need to be populated from user service
		Notes:          req.Notes,
		CreatedAt:      req.CreatedAt,
		AcknowledgedAt: req.AcknowledgedAt,
		CompletedAt:    req.CompletedAt,
		ResponseTime:   responseTime,
	}
}
