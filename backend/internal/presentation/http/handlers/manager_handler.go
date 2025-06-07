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

type ManagerHandler struct {
	managerService services.ManagerService
}

func NewManagerHandler(managerService services.ManagerService) *ManagerHandler {
	return &ManagerHandler{
		managerService: managerService,
	}
}

// GetStaffList gets all staff members for a business
func (h *ManagerHandler) GetStaffList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	staff, err := h.managerService.GetStaffList(ctx, businessID)
	if err != nil {
		http.Error(w, "Failed to get staff list", http.StatusInternalServerError)
		return
	}

	staffResponses := make([]responses.StaffMemberResponse, len(staff))
	for i, member := range staff {
		staffResponses[i] = h.staffToResponse(member)
	}

	response := responses.StaffListResponse{
		Staff:    staffResponses,
		Total:    len(staffResponses),
		Page:     1,
		PageSize: len(staffResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateStaffMember creates a new staff member
func (h *ManagerHandler) CreateStaffMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	var req requests.CreateStaffMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user := &entity.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   req.Password, // Should be hashed by the service
		Name:       req.Name,
		Role:       req.Role,
		Status:     "active",
		BusinessID: &businessID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := h.managerService.CreateStaffMember(ctx, user); err != nil {
		http.Error(w, "Failed to create staff member", http.StatusInternalServerError)
		return
	}

	response := h.staffToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateStaffMember updates an existing staff member
func (h *ManagerHandler) UpdateStaffMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateStaffMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing staff member (this would need to be implemented in manager service)
	// For now, create a user entity with the updates
	user := &entity.User{
		ID:        userID,
		UpdatedAt: time.Now(),
	}

	// Update fields if provided
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := h.managerService.UpdateStaffMember(ctx, user); err != nil {
		http.Error(w, "Failed to update staff member", http.StatusInternalServerError)
		return
	}

	response := h.staffToResponse(user)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeactivateStaffMember deactivates a staff member
func (h *ManagerHandler) DeactivateStaffMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.managerService.DeactivateStaffMember(ctx, userID); err != nil {
		http.Error(w, "Failed to deactivate staff member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetDailyReport gets a daily business report
func (h *ManagerHandler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	// Parse date from query parameter
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	} else {
		date = time.Now()
	}

	report, err := h.managerService.GetDailyReport(ctx, businessID, date)
	if err != nil {
		http.Error(w, "Failed to get daily report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetRevenueReport gets a revenue report for a date range
func (h *ManagerHandler) GetRevenueReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date and end_date are required", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	report, err := h.managerService.GetRevenueReport(ctx, businessID, startDate, endDate)
	if err != nil {
		http.Error(w, "Failed to get revenue report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// GetStaffPerformanceReport gets staff performance report
func (h *ManagerHandler) GetStaffPerformanceReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "monthly" // default
	}

	report, err := h.managerService.GetStaffPerformanceReport(ctx, businessID, period)
	if err != nil {
		http.Error(w, "Failed to get staff performance report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// UpdateBusinessHours updates business operating hours
func (h *ManagerHandler) UpdateBusinessHours(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	var req requests.UpdateBusinessHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.managerService.UpdateBusinessHours(ctx, businessID, req.OpenTime, req.CloseTime); err != nil {
		http.Error(w, "Failed to update business hours", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBusinessStatistics gets overall business statistics
func (h *ManagerHandler) GetBusinessStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	stats, err := h.managerService.GetBusinessStatistics(ctx, businessID)
	if err != nil {
		http.Error(w, "Failed to get business statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetOrderHistory gets order history (existing method from old handler)
func (h *ManagerHandler) GetOrderHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	// This is a simplified implementation - would need proper order service integration
	report, err := h.managerService.GetDailyReport(ctx, businessID, time.Now())
	if err != nil {
		http.Error(w, "Failed to get order history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

// Helper methods
func (h *ManagerHandler) staffToResponse(user *entity.User) responses.StaffMemberResponse {
	return responses.StaffMemberResponse{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Name:         user.Name,
		Role:         user.Role,
		Status:       user.Status,
		BusinessID:   user.BusinessID,
		LastActiveAt: user.LastActiveAt,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
