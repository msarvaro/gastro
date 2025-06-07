package services

import (
	"context"
	"restaurant-management/internal/domain/consts" // Added import
	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
	"time"
)

// ShiftService implements shift service operations
type ShiftService struct {
	shiftRepo repository.ShiftRepository
	userRepo  repository.UserRepository
}

func createShiftService(shiftRepo repository.ShiftRepository, userRepo repository.UserRepository) *ShiftService {
	return &ShiftService{
		shiftRepo: shiftRepo,
		userRepo:  userRepo,
	}
}

// GetShiftByID retrieves a shift by ID
func (s *ShiftService) GetShiftByID(ctx context.Context, id int) (*entity.Shift, error) {
	shift, err := s.shiftRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load employees for the shift
	employees, err := s.shiftRepo.GetEmployeesByShiftID(ctx, id)
	if err == nil {
		shift.Employees = employees
	}

	return shift, nil
}

// GetShiftsByBusinessID retrieves all shifts for a business
func (s *ShiftService) GetShiftsByBusinessID(ctx context.Context, businessID int) ([]*entity.Shift, error) {
	shifts, err := s.shiftRepo.GetByBusinessID(ctx, businessID)
	if err != nil {
		return nil, err
	}

	// Load employees for each shift
	for _, shift := range shifts {
		employees, err := s.shiftRepo.GetEmployeesByShiftID(ctx, shift.ID)
		if err == nil {
			shift.Employees = employees
		}
	}

	return shifts, nil
}

// GetShiftsByDate retrieves all shifts for a specific date
func (s *ShiftService) GetShiftsByDate(ctx context.Context, businessID int, date time.Time) ([]*entity.Shift, error) {
	shifts, err := s.shiftRepo.GetShiftsByDate(ctx, businessID, date)
	if err != nil {
		return nil, err
	}

	// Load employees for each shift
	for _, shift := range shifts {
		employees, err := s.shiftRepo.GetEmployeesByShiftID(ctx, shift.ID)
		if err == nil {
			shift.Employees = employees
		}
	}

	return shifts, nil
}

// GetShiftsByEmployeeID retrieves all shifts for a specific employee
func (s *ShiftService) GetShiftsByEmployeeID(ctx context.Context, employeeID int) ([]*entity.Shift, error) {
	return s.shiftRepo.GetShiftsByEmployeeID(ctx, employeeID)
}

// GetCurrentShiftForEmployee retrieves the current active shift for an employee
func (s *ShiftService) GetCurrentShiftForEmployee(ctx context.Context, employeeID int) (*entity.Shift, error) {
	return s.shiftRepo.GetCurrentShiftForEmployee(ctx, employeeID)
}

// GetUpcomingShiftsForEmployee retrieves upcoming shifts for an employee
func (s *ShiftService) GetUpcomingShiftsForEmployee(ctx context.Context, employeeID int) ([]*entity.Shift, error) {
	return s.shiftRepo.GetUpcomingShiftsForEmployee(ctx, employeeID)
}

// CreateShift creates a new shift
func (s *ShiftService) CreateShift(ctx context.Context, shift *entity.Shift) error {
	// Set creation time
	shift.CreatedAt = time.Now()
	shift.UpdatedAt = time.Now()

	// Verify manager exists and has appropriate role
	manager, err := s.userRepo.GetByID(ctx, shift.ManagerID)
	if err != nil {
		return err
	}

	if manager.Role != consts.RoleManager && manager.Role != consts.RoleAdmin { // Changed
		return nil // Allow for now, could add validation
	}

	err = s.shiftRepo.Create(ctx, shift)
	if err != nil {
		return err
	}

	// Assign employees to the shift
	for _, employee := range shift.Employees {
		err = s.shiftRepo.AssignEmployeeToShift(ctx, shift.ID, employee.ID)
		if err != nil {
			// Log error but continue with other employees
			continue
		}
	}

	return nil
}

// UpdateShift updates an existing shift
func (s *ShiftService) UpdateShift(ctx context.Context, shift *entity.Shift) error {
	// Set update time
	shift.UpdatedAt = time.Now()

	// Verify shift exists
	_, err := s.shiftRepo.GetByID(ctx, shift.ID)
	if err != nil {
		return err
	}

	return s.shiftRepo.Update(ctx, shift)
}

// DeleteShift deletes a shift
func (s *ShiftService) DeleteShift(ctx context.Context, id int) error {
	return s.shiftRepo.Delete(ctx, id)
}

// AssignEmployeeToShift assigns an employee to a shift
func (s *ShiftService) AssignEmployeeToShift(ctx context.Context, shiftID, employeeID int) error {
	// Verify employee exists
	_, err := s.userRepo.GetByID(ctx, employeeID)
	if err != nil {
		return err
	}

	// Verify shift exists
	_, err = s.shiftRepo.GetByID(ctx, shiftID)
	if err != nil {
		return err
	}

	return s.shiftRepo.AssignEmployeeToShift(ctx, shiftID, employeeID)
}

// RemoveEmployeeFromShift removes an employee from a shift
func (s *ShiftService) RemoveEmployeeFromShift(ctx context.Context, shiftID, employeeID int) error {
	return s.shiftRepo.RemoveEmployeeFromShift(ctx, shiftID, employeeID)
}

// GetEmployeesByShiftID retrieves all employees assigned to a shift
func (s *ShiftService) GetEmployeesByShiftID(ctx context.Context, shiftID int) ([]*entity.User, error) {
	return s.shiftRepo.GetEmployeesByShiftID(ctx, shiftID)
}
