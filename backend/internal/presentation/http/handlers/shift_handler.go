package handlers

import (
	"context"
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

type ShiftHandler struct {
	shiftService services.ShiftService
}

func NewShiftHandler(shiftService services.ShiftService) *ShiftHandler {
	return &ShiftHandler{
		shiftService: shiftService,
	}
}

// GetShifts gets all shifts for a business
func (h *ShiftHandler) GetShifts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	shifts, err := h.shiftService.GetShiftsByBusinessID(ctx, businessID)
	if err != nil {
		http.Error(w, "Failed to get shifts", http.StatusInternalServerError)
		return
	}

	shiftResponses := make([]responses.ShiftDetailResponse, len(shifts))
	for i, shift := range shifts {
		shiftResponses[i] = h.shiftToDetailResponse(shift)
	}

	response := responses.ShiftsListResponse{
		Shifts:   shiftResponses,
		Total:    len(shiftResponses),
		Page:     1,
		PageSize: len(shiftResponses),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetShiftByID gets a specific shift by ID
func (h *ShiftHandler) GetShiftByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	shift, err := h.shiftService.GetShiftByID(ctx, shiftID)
	if err != nil {
		http.Error(w, "Shift not found", http.StatusNotFound)
		return
	}

	response := h.shiftToDetailResponse(shift)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateShift creates a new shift
func (h *ShiftHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	businessID := ctx.Value("business_id").(int)

	var req requests.CreateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shift := &entity.Shift{
		Date:       req.Date,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		ManagerID:  req.ManagerID,
		Notes:      req.Notes,
		BusinessID: businessID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := h.shiftService.CreateShift(ctx, shift); err != nil {
		http.Error(w, "Failed to create shift", http.StatusInternalServerError)
		return
	}

	// Assign employees if provided
	for _, employeeID := range req.Employees {
		if err := h.shiftService.AssignEmployeeToShift(ctx, shift.ID, employeeID); err != nil {
			// Log error but don't fail the entire request
			continue
		}
	}

	response := h.shiftToDetailResponse(shift)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateShift updates an existing shift
func (h *ShiftHandler) UpdateShift(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	var req requests.UpdateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shift, err := h.shiftService.GetShiftByID(ctx, shiftID)
	if err != nil {
		http.Error(w, "Shift not found", http.StatusNotFound)
		return
	}

	// Update fields if provided
	if req.Date != nil {
		shift.Date = *req.Date
	}
	if req.StartTime != nil {
		shift.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		shift.EndTime = *req.EndTime
	}
	if req.ManagerID != nil {
		shift.ManagerID = *req.ManagerID
	}
	if req.Notes != nil {
		shift.Notes = *req.Notes
	}

	shift.UpdatedAt = time.Now()

	if err := h.shiftService.UpdateShift(ctx, shift); err != nil {
		http.Error(w, "Failed to update shift", http.StatusInternalServerError)
		return
	}

	// Update employee assignments if provided
	if req.Employees != nil {
		// Get current employees
		currentEmployees, _ := h.shiftService.GetEmployeesByShiftID(ctx, shiftID)

		// Remove all current employees
		for _, emp := range currentEmployees {
			h.shiftService.RemoveEmployeeFromShift(ctx, shiftID, emp.ID)
		}

		// Add new employees
		for _, employeeID := range req.Employees {
			h.shiftService.AssignEmployeeToShift(ctx, shiftID, employeeID)
		}
	}

	response := h.shiftToDetailResponse(shift)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteShift deletes a shift
func (h *ShiftHandler) DeleteShift(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	if err := h.shiftService.DeleteShift(ctx, shiftID); err != nil {
		http.Error(w, "Failed to delete shift", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AssignEmployee assigns an employee to a shift
func (h *ShiftHandler) AssignEmployee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	var req requests.AssignEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.shiftService.AssignEmployeeToShift(ctx, shiftID, req.EmployeeID); err != nil {
		http.Error(w, "Failed to assign employee", http.StatusInternalServerError)
		return
	}

	// Return updated employee list
	employees, err := h.shiftService.GetEmployeesByShiftID(ctx, shiftID)
	if err != nil {
		http.Error(w, "Failed to get updated employee list", http.StatusInternalServerError)
		return
	}

	response := responses.ShiftAssignmentResponse{
		ShiftID:   shiftID,
		Employees: h.usersToShiftUserResponses(employees),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RemoveEmployee removes an employee from a shift
func (h *ShiftHandler) RemoveEmployee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)

	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	var req requests.RemoveEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.shiftService.RemoveEmployeeFromShift(ctx, shiftID, req.EmployeeID); err != nil {
		http.Error(w, "Failed to remove employee", http.StatusInternalServerError)
		return
	}

	// Return updated employee list
	employees, err := h.shiftService.GetEmployeesByShiftID(ctx, shiftID)
	if err != nil {
		http.Error(w, "Failed to get updated employee list", http.StatusInternalServerError)
		return
	}

	response := responses.ShiftAssignmentResponse{
		ShiftID:   shiftID,
		Employees: h.usersToShiftUserResponses(employees),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEmployeeShifts gets shifts for an employee
func (h *ShiftHandler) GetEmployeeShifts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value("user_id").(int)

	// Get current shift
	currentShift, _ := h.shiftService.GetCurrentShiftForEmployee(ctx, userID)

	// Get upcoming shifts
	upcomingShifts, err := h.shiftService.GetUpcomingShiftsForEmployee(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to get employee shifts", http.StatusInternalServerError)
		return
	}

	var currentShiftResponse *responses.ShiftDetailResponse
	if currentShift != nil {
		resp := h.shiftToDetailResponse(currentShift)
		currentShiftResponse = &resp
	}

	upcomingShiftResponses := make([]responses.ShiftDetailResponse, len(upcomingShifts))
	for i, shift := range upcomingShifts {
		upcomingShiftResponses[i] = h.shiftToDetailResponse(shift)
	}

	response := responses.EmployeeShiftDetailResponse{
		UserID:         userID,
		Username:       "", // Would need to get from user service
		Name:           "", // Would need to get from user service
		CurrentShift:   currentShiftResponse,
		UpcomingShifts: upcomingShiftResponses,
		TotalHours:     0.0, // Would need to calculate
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods
func (h *ShiftHandler) shiftToDetailResponse(shift *entity.Shift) responses.ShiftDetailResponse {
	employees, _ := h.shiftService.GetEmployeesByShiftID(context.Background(), shift.ID)

	duration := shift.EndTime.Sub(shift.StartTime).String()

	return responses.ShiftDetailResponse{
		ID:         shift.ID,
		Date:       shift.Date,
		StartTime:  shift.StartTime,
		EndTime:    shift.EndTime,
		ManagerID:  shift.ManagerID,
		Manager:    responses.ShiftUserResponse{}, // Would need manager details
		Notes:      shift.Notes,
		BusinessID: shift.BusinessID,
		CreatedAt:  shift.CreatedAt,
		UpdatedAt:  shift.UpdatedAt,
		Employees:  h.usersToShiftUserResponses(employees),
		IsActive:   shift.IsActive(),
		Duration:   duration,
	}
}

func (h *ShiftHandler) usersToShiftUserResponses(users []*entity.User) []responses.ShiftUserResponse {
	result := make([]responses.ShiftUserResponse, len(users))
	for i, user := range users {
		result[i] = responses.ShiftUserResponse{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Role:     user.Role,
			Email:    user.Email,
		}
	}
	return result
}
