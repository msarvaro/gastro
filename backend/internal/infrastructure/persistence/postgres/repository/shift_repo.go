package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"restaurant-management/internal/domain/entity"
	"restaurant-management/internal/domain/interfaces/repository"
)

// shiftRepository implements the ShiftRepository interface using PostgreSQL
type shiftRepository struct {
	db *sql.DB
}

// NewShiftRepository creates a new shift repository
func NewShiftRepository(db *sql.DB) repository.ShiftRepository {
	return &shiftRepository{
		db: db,
	}
}

// GetByID retrieves a shift by ID
func (r *shiftRepository) GetByID(ctx context.Context, id int) (*entity.Shift, error) {
	query := `
		SELECT id, date, start_time, end_time, manager_id, notes, business_id, created_at, updated_at
		FROM shifts 
		WHERE id = $1
	`

	var shift entity.Shift

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&shift.ID,
		&shift.Date,
		&shift.StartTime,
		&shift.EndTime,
		&shift.ManagerID,
		&shift.Notes,
		&shift.BusinessID,
		&shift.CreatedAt,
		&shift.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Load employees assigned to this shift
	employeesQuery := `
		SELECT u.id, u.username, u.email, u.password, u.name, u.role, u.status, 
		       u.business_id, u.last_active, u.created_at, u.updated_at
		FROM users u
		JOIN shift_employees se ON u.id = se.employee_id
		WHERE se.shift_id = $1
	`

	rows, err := r.db.QueryContext(ctx, employeesQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*entity.User

	for rows.Next() {
		var employee entity.User
		var businessID sql.NullInt64
		var lastActive sql.NullTime

		err := rows.Scan(
			&employee.ID,
			&employee.Username,
			&employee.Email,
			&employee.Password,
			&employee.Name,
			&employee.Role,
			&employee.Status,
			&businessID,
			&lastActive,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if businessID.Valid {
			bid := int(businessID.Int64)
			employee.BusinessID = &bid
		}

		if lastActive.Valid {
			employee.LastActiveAt = &lastActive.Time
		}

		employees = append(employees, &employee)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	shift.Employees = employees

	return &shift, nil
}

// GetByBusinessID retrieves all shifts for a business
func (r *shiftRepository) GetByBusinessID(ctx context.Context, businessID int) ([]*entity.Shift, error) {
	query := `
		SELECT id, date, start_time, end_time, manager_id, notes, business_id, created_at, updated_at
		FROM shifts 
		WHERE business_id = $1
		ORDER BY date DESC, start_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*entity.Shift

	for rows.Next() {
		var shift entity.Shift

		err := rows.Scan(
			&shift.ID,
			&shift.Date,
			&shift.StartTime,
			&shift.EndTime,
			&shift.ManagerID,
			&shift.Notes,
			&shift.BusinessID,
			&shift.CreatedAt,
			&shift.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		shifts = append(shifts, &shift)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// Create adds a new shift
func (r *shiftRepository) Create(ctx context.Context, shift *entity.Shift) error {
	query := `
		INSERT INTO shifts (date, start_time, end_time, manager_id, notes, business_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	now := time.Now()
	if shift.CreatedAt.IsZero() {
		shift.CreatedAt = now
	}
	if shift.UpdatedAt.IsZero() {
		shift.UpdatedAt = now
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		shift.Date,
		shift.StartTime,
		shift.EndTime,
		shift.ManagerID,
		shift.Notes,
		shift.BusinessID,
		shift.CreatedAt,
		shift.UpdatedAt,
	).Scan(&shift.ID)

	if err != nil {
		return err
	}

	// If employees are specified, assign them to the shift
	if len(shift.Employees) > 0 {
		for _, employee := range shift.Employees {
			err = r.AssignEmployeeToShift(ctx, shift.ID, employee.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Update updates an existing shift
func (r *shiftRepository) Update(ctx context.Context, shift *entity.Shift) error {
	query := `
		UPDATE shifts
		SET date = $1, start_time = $2, end_time = $3, manager_id = $4, notes = $5, updated_at = $6
		WHERE id = $7 AND business_id = $8
	`

	shift.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		shift.Date,
		shift.StartTime,
		shift.EndTime,
		shift.ManagerID,
		shift.Notes,
		shift.UpdatedAt,
		shift.ID,
		shift.BusinessID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("shift not found")
	}

	return nil
}

// Delete removes a shift
func (r *shiftRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM shifts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("shift not found")
	}

	return nil
}

// GetShiftsByDate retrieves shifts for a specific date
func (r *shiftRepository) GetShiftsByDate(ctx context.Context, businessID int, date time.Time) ([]*entity.Shift, error) {
	query := `
		SELECT id, date, start_time, end_time, manager_id, notes, business_id, created_at, updated_at
		FROM shifts 
		WHERE business_id = $1 AND date = $2
		ORDER BY start_time
	`

	// Format date to match SQL date format (YYYY-MM-DD)
	formattedDate := date.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, query, businessID, formattedDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*entity.Shift

	for rows.Next() {
		var shift entity.Shift

		err := rows.Scan(
			&shift.ID,
			&shift.Date,
			&shift.StartTime,
			&shift.EndTime,
			&shift.ManagerID,
			&shift.Notes,
			&shift.BusinessID,
			&shift.CreatedAt,
			&shift.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		shifts = append(shifts, &shift)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// GetShiftsByDateRange retrieves shifts within a date range
func (r *shiftRepository) GetShiftsByDateRange(ctx context.Context, businessID int, startDate, endDate time.Time) ([]*entity.Shift, error) {
	query := `
		SELECT id, date, start_time, end_time, manager_id, notes, business_id, created_at, updated_at
		FROM shifts 
		WHERE business_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date, start_time
	`

	// Format dates to match SQL date format (YYYY-MM-DD)
	formattedStartDate := startDate.Format("2006-01-02")
	formattedEndDate := endDate.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, query, businessID, formattedStartDate, formattedEndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*entity.Shift

	for rows.Next() {
		var shift entity.Shift

		err := rows.Scan(
			&shift.ID,
			&shift.Date,
			&shift.StartTime,
			&shift.EndTime,
			&shift.ManagerID,
			&shift.Notes,
			&shift.BusinessID,
			&shift.CreatedAt,
			&shift.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		shifts = append(shifts, &shift)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// GetShiftsByEmployeeID retrieves shifts for a specific employee
func (r *shiftRepository) GetShiftsByEmployeeID(ctx context.Context, employeeID int) ([]*entity.Shift, error) {
	query := `
		SELECT s.id, s.date, s.start_time, s.end_time, s.manager_id, s.notes, s.business_id, s.created_at, s.updated_at
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		ORDER BY s.date DESC, s.start_time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*entity.Shift

	for rows.Next() {
		var shift entity.Shift

		err := rows.Scan(
			&shift.ID,
			&shift.Date,
			&shift.StartTime,
			&shift.EndTime,
			&shift.ManagerID,
			&shift.Notes,
			&shift.BusinessID,
			&shift.CreatedAt,
			&shift.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		shifts = append(shifts, &shift)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// GetCurrentShiftForEmployee retrieves the current active shift for an employee
func (r *shiftRepository) GetCurrentShiftForEmployee(ctx context.Context, employeeID int) (*entity.Shift, error) {
	// Get current date and time
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	query := `
		SELECT s.id, s.date, s.start_time, s.end_time, s.manager_id, s.notes, s.business_id, s.created_at, s.updated_at
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		  AND s.date = $2
		  AND s.start_time <= $3
		  AND s.end_time >= $3
		LIMIT 1
	`

	var shift entity.Shift

	err := r.db.QueryRowContext(ctx, query, employeeID, currentDate, currentTime).Scan(
		&shift.ID,
		&shift.Date,
		&shift.StartTime,
		&shift.EndTime,
		&shift.ManagerID,
		&shift.Notes,
		&shift.BusinessID,
		&shift.CreatedAt,
		&shift.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &shift, nil
}

// GetUpcomingShiftsForEmployee retrieves future shifts for an employee
func (r *shiftRepository) GetUpcomingShiftsForEmployee(ctx context.Context, employeeID int) ([]*entity.Shift, error) {
	// Get current date and time
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	query := `
		SELECT s.id, s.date, s.start_time, s.end_time, s.manager_id, s.notes, s.business_id, s.created_at, s.updated_at
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		  AND (
			  (s.date = $2 AND s.start_time > $3) OR
			  s.date > $2
		  )
		ORDER BY s.date, s.start_time
		LIMIT 5
	`

	rows, err := r.db.QueryContext(ctx, query, employeeID, currentDate, currentTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shifts []*entity.Shift

	for rows.Next() {
		var shift entity.Shift

		err := rows.Scan(
			&shift.ID,
			&shift.Date,
			&shift.StartTime,
			&shift.EndTime,
			&shift.ManagerID,
			&shift.Notes,
			&shift.BusinessID,
			&shift.CreatedAt,
			&shift.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		shifts = append(shifts, &shift)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return shifts, nil
}

// AssignEmployeeToShift assigns an employee to a shift
func (r *shiftRepository) AssignEmployeeToShift(ctx context.Context, shiftID, employeeID int) error {
	// First check if the shift exists
	shiftQuery := `SELECT id, business_id FROM shifts WHERE id = $1`
	var shift entity.Shift
	err := r.db.QueryRowContext(ctx, shiftQuery, shiftID).Scan(&shift.ID, &shift.BusinessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("shift not found")
		}
		return err
	}

	// Then check if the employee exists and belongs to the same business
	employeeQuery := `SELECT id, business_id FROM users WHERE id = $1`
	var employee entity.User
	var businessID sql.NullInt64

	err = r.db.QueryRowContext(ctx, employeeQuery, employeeID).Scan(&employee.ID, &businessID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("employee not found")
		}
		return err
	}

	if !businessID.Valid || int(businessID.Int64) != shift.BusinessID {
		return errors.New("employee does not belong to the same business as the shift")
	}

	// Check if the assignment already exists
	checkQuery := `SELECT id FROM shift_employees WHERE shift_id = $1 AND employee_id = $2`
	var assignmentID int
	err = r.db.QueryRowContext(ctx, checkQuery, shiftID, employeeID).Scan(&assignmentID)
	if err == nil {
		// Assignment already exists
		return nil
	} else if err != sql.ErrNoRows {
		return err
	}

	// Insert the assignment
	insertQuery := `
		INSERT INTO shift_employees (shift_id, employee_id, business_id, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err = r.db.ExecContext(ctx, insertQuery, shiftID, employeeID, shift.BusinessID, time.Now())
	return err
}

// RemoveEmployeeFromShift removes an employee from a shift
func (r *shiftRepository) RemoveEmployeeFromShift(ctx context.Context, shiftID, employeeID int) error {
	query := `DELETE FROM shift_employees WHERE shift_id = $1 AND employee_id = $2`

	result, err := r.db.ExecContext(ctx, query, shiftID, employeeID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("shift assignment not found")
	}

	return nil
}

// GetEmployeesByShiftID retrieves all employees assigned to a shift
func (r *shiftRepository) GetEmployeesByShiftID(ctx context.Context, shiftID int) ([]*entity.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password, u.name, u.role, u.status, 
		       u.business_id, u.last_active, u.created_at, u.updated_at
		FROM users u
		JOIN shift_employees se ON u.id = se.employee_id
		WHERE se.shift_id = $1
		ORDER BY u.name
	`

	rows, err := r.db.QueryContext(ctx, query, shiftID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*entity.User

	for rows.Next() {
		var employee entity.User
		var businessID sql.NullInt64
		var lastActive sql.NullTime

		err := rows.Scan(
			&employee.ID,
			&employee.Username,
			&employee.Email,
			&employee.Password,
			&employee.Name,
			&employee.Role,
			&employee.Status,
			&businessID,
			&lastActive,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		if businessID.Valid {
			bid := int(businessID.Int64)
			employee.BusinessID = &bid
		}

		if lastActive.Valid {
			employee.LastActiveAt = &lastActive.Time
		}

		employees = append(employees, &employee)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}
