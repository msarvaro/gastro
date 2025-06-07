package entity

import "time"

// Shift represents a work shift for employees
type Shift struct {
	ID         int
	Date       time.Time
	StartTime  time.Time
	EndTime    time.Time
	ManagerID  int
	Notes      string
	BusinessID int
	CreatedAt  time.Time
	UpdatedAt  time.Time
	BreakStart *time.Time // Optional break start time
	BreakEnd   *time.Time // Optional break end time
	Employees  []*User    // Employees assigned to this shift
}

// ShiftEmployee represents the relationship between shifts and employees
type ShiftEmployee struct {
	ID         int
	ShiftID    int
	EmployeeID int
	BusinessID int
	CreatedAt  time.Time
}

// IsActive checks if the shift is currently active
func (s *Shift) IsActive() bool {
	now := time.Now()

	// Convert current date and shift date to same day for time comparison
	currentDate := now.Format("2006-01-02")
	shiftDate := s.Date.Format("2006-01-02")

	// If it's not the same day, the shift is not active
	if currentDate != shiftDate {
		return false
	}

	// Create full datetime by combining date with start/end times
	startTime := combineDateTime(s.Date, s.StartTime)
	endTime := combineDateTime(s.Date, s.EndTime)

	// Handle overnight shifts
	if endTime.Before(startTime) {
		endTime = endTime.Add(24 * time.Hour)
	}

	return now.After(startTime) && now.Before(endTime)
}

// IsUpcoming checks if the shift is in the future
func (s *Shift) IsUpcoming() bool {
	now := time.Now()
	startTime := combineDateTime(s.Date, s.StartTime)
	return now.Before(startTime)
}

// Helper function to combine date and time parts
func combineDateTime(date, timeComponent time.Time) time.Time {
	year, month, day := date.Date()
	hour, min, sec := timeComponent.Clock()
	return time.Date(year, month, day, hour, min, sec, 0, time.Local)
}

type ShiftSummary struct {
	ShiftID      int
	TotalOrders  int
	TotalRevenue float64
	TotalTips    float64
	Performance  float64 // Performance score 0-100
}

// Business methods
func (s *Shift) Duration() time.Duration {
	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}

func (s *Shift) IsOnBreak() bool {
	if s.BreakStart == nil {
		return false
	}
	return s.BreakEnd == nil
}
