package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/domain/shift"
	"restaurant-management/internal/middleware"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type ShiftController struct {
	shiftService shift.Service
}

func NewShiftController(shiftService shift.Service) *ShiftController {
	return &ShiftController{
		shiftService: shiftService,
	}
}

func (c *ShiftController) GetAllShifts(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	shifts, totalCount, err := c.shiftService.GetAllShifts(r.Context(), page, limit, businessID)
	if err != nil {
		log.Printf("Error getting all shifts: %v", err)
		http.Error(w, "Failed to fetch shifts", http.StatusInternalServerError)
		return
	}

	response := struct {
		Shifts     []shift.ShiftWithEmployees `json:"shifts"`
		TotalCount int                        `json:"total_count"`
		Page       int                        `json:"page"`
		Limit      int                        `json:"limit"`
	}{
		Shifts:     shifts,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *ShiftController) GetShiftByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	shift, err := c.shiftService.GetShiftByID(r.Context(), shiftID, businessID)
	if err != nil {
		log.Printf("Error getting shift by ID %d: %v", shiftID, err)
		http.Error(w, "Shift not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shift)
}

func (c *ShiftController) CreateShift(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var req shift.CreateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shiftData, employeeIDs, err := c.parseShiftRequest(req)
	if err != nil {
		log.Printf("Error parsing shift request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	createdShift, err := c.shiftService.CreateShift(r.Context(), shiftData, employeeIDs, businessID)
	if err != nil {
		log.Printf("Error creating shift: %v", err)
		http.Error(w, "Failed to create shift", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdShift)
}

func (c *ShiftController) UpdateShift(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var req shift.UpdateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shiftData, employeeIDs, err := c.parseShiftRequest(req)
	if err != nil {
		log.Printf("Error parsing shift request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shiftData.ID = shiftID

	if err := c.shiftService.UpdateShift(r.Context(), shiftData, employeeIDs, businessID); err != nil {
		log.Printf("Error updating shift %d: %v", shiftID, err)
		http.Error(w, "Failed to update shift", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Shift updated successfully"})
}

func (c *ShiftController) DeleteShift(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shiftID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid shift ID", http.StatusBadRequest)
		return
	}

	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	if err := c.shiftService.DeleteShift(r.Context(), shiftID, businessID); err != nil {
		log.Printf("Error deleting shift %d: %v", shiftID, err)
		http.Error(w, "Failed to delete shift", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Shift deleted successfully"})
}

func (c *ShiftController) GetEmployeeShifts(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	shifts, err := c.shiftService.GetCurrentAndUpcomingShifts(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting employee shifts: %v", err)
		http.Error(w, "Failed to fetch employee shifts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shifts)
}

// Helper methods for parsing shift requests
func (c *ShiftController) parseShiftRequest(req interface{}) (*shift.Shift, []int, error) {
	switch r := req.(type) {
	case shift.CreateShiftRequest:
		return c.parseCreateShiftRequest(r)
	case shift.UpdateShiftRequest:
		return c.parseUpdateShiftRequest(r)
	default:
		return nil, nil, shift.ErrInvalidShiftData
	}
}

func (c *ShiftController) parseCreateShiftRequest(req shift.CreateShiftRequest) (*shift.Shift, []int, error) {
	managerID, err := c.parseManagerID(req.ManagerID)
	if err != nil {
		return nil, nil, err
	}

	if req.StartTime == "" || req.EndTime == "" {
		return nil, nil, shift.ErrInvalidShiftData
	}

	// Parse date (if empty, use today)
	var shiftDate time.Time
	if req.Date == "" {
		shiftDate = time.Now()
	} else {
		shiftDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, nil, shift.ErrInvalidTimeRange
		}
	}

	// Parse start and end times
	startTime, err := c.parseDateTime(req.Date, req.StartTime)
	if err != nil {
		return nil, nil, shift.ErrInvalidTimeRange
	}

	endTime, err := c.parseDateTime(req.Date, req.EndTime)
	if err != nil {
		return nil, nil, shift.ErrInvalidTimeRange
	}

	shiftData := &shift.Shift{
		Date:      shiftDate,
		StartTime: startTime,
		EndTime:   endTime,
		ManagerID: managerID,
		Notes:     req.Notes,
	}

	return shiftData, req.EmployeeIDs, nil
}

func (c *ShiftController) parseUpdateShiftRequest(req shift.UpdateShiftRequest) (*shift.Shift, []int, error) {
	managerID, err := c.parseManagerID(req.ManagerID)
	if err != nil {
		return nil, nil, err
	}

	// Parse date (if empty, use today)
	var shiftDate time.Time
	if req.Date == "" {
		shiftDate = time.Now()
	} else {
		shiftDate, err = time.Parse("2006-01-02", req.Date)
		if err != nil {
			return nil, nil, shift.ErrInvalidTimeRange
		}
	}

	// Parse start and end times
	startTime, err := c.parseDateTime(req.Date, req.StartTime)
	if err != nil {
		return nil, nil, shift.ErrInvalidTimeRange
	}

	endTime, err := c.parseDateTime(req.Date, req.EndTime)
	if err != nil {
		return nil, nil, shift.ErrInvalidTimeRange
	}

	shiftData := &shift.Shift{
		Date:      shiftDate,
		StartTime: startTime,
		EndTime:   endTime,
		ManagerID: managerID,
		Notes:     req.Notes,
	}

	return shiftData, req.EmployeeIDs, nil
}

func (c *ShiftController) parseDateTime(dateStr, timeStr string) (time.Time, error) {
	// Parse date (if empty, use today)
	var date time.Time
	var err error
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return time.Time{}, err
		}
	}

	// Parse time
	timePart, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}

	// Combine date and time
	combined := time.Date(
		date.Year(), date.Month(), date.Day(),
		timePart.Hour(), timePart.Minute(), 0, 0,
		date.Location(),
	)

	return combined, nil
}

func (c *ShiftController) parseManagerID(managerIDStr string) (int, error) {
	if strings.TrimSpace(managerIDStr) == "" {
		return 0, shift.ErrManagerNotFound
	}

	managerID, err := strconv.Atoi(managerIDStr)
	if err != nil || managerID <= 0 {
		return 0, shift.ErrManagerNotFound
	}

	return managerID, nil
}
