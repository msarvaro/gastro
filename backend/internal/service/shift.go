package service

import (
	"context"
	"restaurant-management/internal/domain/shift"
	"strconv"
	"strings"
	"time"
)

type ShiftService struct {
	repo shift.Repository
}

func NewShiftService(repo shift.Repository) shift.Service {
	return &ShiftService{repo: repo}
}

func (s *ShiftService) GetAllShifts(ctx context.Context, page, limit int, businessID int) ([]shift.ShiftWithEmployees, int, error) {
	if businessID <= 0 {
		return nil, 0, shift.ErrInvalidShiftData
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	return s.repo.GetAllShifts(ctx, page, limit, businessID)
}

func (s *ShiftService) GetShiftByID(ctx context.Context, shiftID int, businessID int) (*shift.ShiftWithEmployees, error) {
	if shiftID <= 0 {
		return nil, shift.ErrInvalidShiftID
	}
	if businessID <= 0 {
		return nil, shift.ErrInvalidShiftData
	}

	return s.repo.GetShiftByID(ctx, shiftID, businessID)
}

func (s *ShiftService) GetShiftEmployees(ctx context.Context, shiftID int, businessID int) ([]shift.User, error) {
	if shiftID <= 0 {
		return nil, shift.ErrInvalidShiftID
	}
	if businessID <= 0 {
		return nil, shift.ErrInvalidShiftData
	}

	return s.repo.GetShiftEmployees(ctx, shiftID, businessID)
}

func (s *ShiftService) CreateShift(ctx context.Context, shiftData *shift.Shift, employeeIDs []int, businessID int) (*shift.Shift, error) {
	if businessID <= 0 {
		return nil, shift.ErrInvalidShiftData
	}

	// Validation
	if shiftData.Date.IsZero() {
		return nil, shift.ErrInvalidShiftData
	}
	if shiftData.StartTime.IsZero() || shiftData.EndTime.IsZero() {
		return nil, shift.ErrInvalidShiftData
	}
	if shiftData.StartTime.After(shiftData.EndTime) || shiftData.StartTime.Equal(shiftData.EndTime) {
		return nil, shift.ErrInvalidTimeRange
	}
	if shiftData.ManagerID <= 0 {
		return nil, shift.ErrManagerNotFound
	}
	if len(employeeIDs) == 0 {
		return nil, shift.ErrInvalidShiftData
	}

	// Validate that all employee IDs are positive
	for _, empID := range employeeIDs {
		if empID <= 0 {
			return nil, shift.ErrEmployeeNotFound
		}
	}

	// Check for overlapping shifts with the same employees
	if err := s.checkShiftOverlaps(ctx, shiftData, employeeIDs, businessID, 0); err != nil {
		return nil, err
	}

	return s.repo.CreateShift(ctx, shiftData, employeeIDs, businessID)
}

func (s *ShiftService) UpdateShift(ctx context.Context, shiftData *shift.Shift, employeeIDs []int, businessID int) error {
	if shiftData.ID <= 0 {
		return shift.ErrInvalidShiftID
	}
	if businessID <= 0 {
		return shift.ErrInvalidShiftData
	}

	// Validation
	if shiftData.Date.IsZero() {
		return shift.ErrInvalidShiftData
	}
	if shiftData.StartTime.IsZero() || shiftData.EndTime.IsZero() {
		return shift.ErrInvalidShiftData
	}
	if shiftData.StartTime.After(shiftData.EndTime) || shiftData.StartTime.Equal(shiftData.EndTime) {
		return shift.ErrInvalidTimeRange
	}
	if shiftData.ManagerID <= 0 {
		return shift.ErrManagerNotFound
	}
	if len(employeeIDs) == 0 {
		return shift.ErrInvalidShiftData
	}

	// Validate that all employee IDs are positive
	for _, empID := range employeeIDs {
		if empID <= 0 {
			return shift.ErrEmployeeNotFound
		}
	}

	// Verify shift exists
	_, err := s.repo.GetShiftByID(ctx, shiftData.ID, businessID)
	if err != nil {
		return shift.ErrShiftNotFound
	}

	// Check for overlapping shifts with the same employees (excluding current shift)
	if err := s.checkShiftOverlaps(ctx, shiftData, employeeIDs, businessID, shiftData.ID); err != nil {
		return err
	}

	return s.repo.UpdateShift(ctx, shiftData, employeeIDs, businessID)
}

func (s *ShiftService) DeleteShift(ctx context.Context, shiftID int, businessID int) error {
	if shiftID <= 0 {
		return shift.ErrInvalidShiftID
	}
	if businessID <= 0 {
		return shift.ErrInvalidShiftData
	}

	// Verify shift exists
	_, err := s.repo.GetShiftByID(ctx, shiftID, businessID)
	if err != nil {
		return shift.ErrShiftNotFound
	}

	return s.repo.DeleteShift(ctx, shiftID, businessID)
}

func (s *ShiftService) GetCurrentAndUpcomingShifts(ctx context.Context, businessID int) ([]shift.ShiftWithEmployees, error) {
	if businessID <= 0 {
		return nil, shift.ErrInvalidShiftData
	}

	// For this method, we'll return upcoming shifts for all employees
	// In a real implementation, you might want to filter by specific employee
	// For now, we'll get all upcoming shifts
	allShifts, _, err := s.repo.GetAllShifts(ctx, 1, 50, businessID)
	if err != nil {
		return nil, err
	}

	// Filter to current and upcoming shifts
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var currentAndUpcoming []shift.ShiftWithEmployees
	for _, shiftData := range allShifts {
		if shiftData.Date.Equal(today) || shiftData.Date.After(today) {
			currentAndUpcoming = append(currentAndUpcoming, shiftData)
		}
	}

	return currentAndUpcoming, nil
}

// Helper function to parse shift request data
func (s *ShiftService) ParseShiftRequest(req interface{}) (*shift.Shift, []int, error) {
	// This would be used by handlers to parse request data
	// For now, it's a placeholder for request parsing logic
	// In a real implementation, you'd parse the request struct
	return nil, nil, shift.ErrInvalidShiftData
}

// Helper function to parse date and time strings
func parseDateTime(dateStr, timeStr string) (time.Time, error) {
	// Parse date (YYYY-MM-DD format)
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, err
	}

	// Parse time (HH:MM format)
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

// Helper function to parse manager ID
func parseManagerID(managerIDStr string) (int, error) {
	if strings.TrimSpace(managerIDStr) == "" {
		return 0, shift.ErrManagerNotFound
	}

	managerID, err := strconv.Atoi(managerIDStr)
	if err != nil || managerID <= 0 {
		return 0, shift.ErrManagerNotFound
	}

	return managerID, nil
}

// checkShiftOverlaps checks if the given shift overlaps with existing shifts for the same employees
func (s *ShiftService) checkShiftOverlaps(ctx context.Context, shiftData *shift.Shift, employeeIDs []int, businessID int, excludeShiftID int) error {
	// Get all shifts for the same date
	allShifts, _, err := s.repo.GetAllShifts(ctx, 1, 1000, businessID) // Get a large number to check all
	if err != nil {
		return err
	}

	// Check for overlaps with each employee
	for _, employeeID := range employeeIDs {
		for _, existingShift := range allShifts {
			// Skip if it's the same shift (when updating)
			if existingShift.ID == excludeShiftID {
				continue
			}

			// Check if shift is on the same date
			if !existingShift.Date.Equal(shiftData.Date) {
				continue
			}

			// Check if the employee is assigned to this existing shift
			isEmployeeInShift := false
			for _, employee := range existingShift.Employees {
				if employee.ID == employeeID {
					isEmployeeInShift = true
					break
				}
			}

			if !isEmployeeInShift {
				continue
			}

			// Check for time overlap
			if s.shiftsOverlap(shiftData.StartTime, shiftData.EndTime, existingShift.StartTime, existingShift.EndTime) {
				return shift.ErrShiftOverlap
			}
		}
	}

	return nil
}

// shiftsOverlap checks if two time ranges overlap
func (s *ShiftService) shiftsOverlap(start1, end1, start2, end2 time.Time) bool {
	// Two shifts overlap if one starts before the other ends
	return start1.Before(end2) && start2.Before(end1)
}
