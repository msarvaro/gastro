package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"restaurant-management/internal/domain/shift"
	"time"
)

type ShiftRepository struct {
	db *DB
}

func NewShiftRepository(db *DB) shift.Repository {
	return &ShiftRepository{db: db}
}

// Helper method to get shift by ID without employees
func (r *ShiftRepository) getShiftByID(ctx context.Context, shiftID int, businessID int) (*shift.Shift, error) {
	query := `
		SELECT id, date, start_time, end_time, manager_id, business_id, notes, created_at, updated_at
		FROM shifts
		WHERE id = $1 AND business_id = $2`

	var s shift.Shift
	err := r.db.QueryRowContext(ctx, query, shiftID, businessID).Scan(
		&s.ID,
		&s.Date,
		&s.StartTime,
		&s.EndTime,
		&s.ManagerID,
		&s.BusinessID,
		&s.Notes,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shift with ID %d not found", shiftID)
	}
	if err != nil {
		log.Printf("Error scanning shift by ID %d: %v", shiftID, err)
		return nil, err
	}
	return &s, nil
}

func (r *ShiftRepository) GetAllShifts(ctx context.Context, page, limit int, businessID int) ([]shift.ShiftWithEmployees, int, error) {
	offset := (page - 1) * limit

	// First get the total count
	countQuery := `SELECT COUNT(*) FROM shifts WHERE business_id = $1`
	var totalCount int
	err := r.db.QueryRowContext(ctx, countQuery, businessID).Scan(&totalCount)
	if err != nil {
		log.Printf("Error getting shifts count: %v", err)
		return nil, 0, err
	}

	// Then get the paginated shifts with manager information
	shiftsQuery := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, s.manager_id, s.business_id, s.notes, s.created_at, s.updated_at,
			m.username as manager_username, m.name as manager_name, m.email as manager_email, m.role as manager_role, m.status as manager_status
		FROM shifts s
		LEFT JOIN users m ON s.manager_id = m.id
		WHERE s.business_id = $1
		ORDER BY s.date DESC, s.start_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, shiftsQuery, businessID, limit, offset)
	if err != nil {
		log.Printf("Error querying shifts: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var shifts []shift.ShiftWithEmployees
	for rows.Next() {
		var swe shift.ShiftWithEmployees
		var manager shift.User

		err := rows.Scan(
			&swe.ID, &swe.Date, &swe.StartTime, &swe.EndTime, &swe.ManagerID, &swe.BusinessID, &swe.Notes, &swe.CreatedAt, &swe.UpdatedAt,
			&manager.Username, &manager.Name, &manager.Email, &manager.Role, &manager.Status,
		)
		if err != nil {
			log.Printf("Error scanning shift: %v", err)
			return nil, 0, err
		}
		manager.ID = swe.ManagerID
		swe.Manager = &manager

		// Get employees for this shift
		employees, err := r.GetShiftEmployees(ctx, swe.ID, businessID)
		if err != nil {
			log.Printf("Error getting employees for shift %d: %v", swe.ID, err)
			// Continue with empty employees instead of failing
		}
		swe.Employees = employees

		shifts = append(shifts, swe)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating shift rows: %v", err)
		return nil, 0, err
	}

	return shifts, totalCount, nil
}

func (r *ShiftRepository) GetShiftByID(ctx context.Context, shiftID int, businessID int) (*shift.ShiftWithEmployees, error) {
	// Get the basic shift information with manager details
	query := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, s.manager_id, s.business_id, s.notes, s.created_at, s.updated_at,
			m.username as manager_username, m.name as manager_name, m.email as manager_email, m.role as manager_role, m.status as manager_status
		FROM shifts s
		LEFT JOIN users m ON s.manager_id = m.id
		WHERE s.id = $1 AND s.business_id = $2`

	var swe shift.ShiftWithEmployees
	var manager shift.User

	err := r.db.QueryRowContext(ctx, query, shiftID, businessID).Scan(
		&swe.ID, &swe.Date, &swe.StartTime, &swe.EndTime, &swe.ManagerID, &swe.BusinessID, &swe.Notes, &swe.CreatedAt, &swe.UpdatedAt,
		&manager.Username, &manager.Name, &manager.Email, &manager.Role, &manager.Status,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shift with ID %d not found", shiftID)
	}
	if err != nil {
		log.Printf("Error scanning shift by ID %d: %v", shiftID, err)
		return nil, err
	}

	manager.ID = swe.ManagerID
	swe.Manager = &manager

	// Get employees for this shift
	employees, err := r.GetShiftEmployees(ctx, shiftID, businessID)
	if err != nil {
		log.Printf("Error getting employees for shift %d: %v", shiftID, err)
		return nil, err
	}
	swe.Employees = employees

	return &swe, nil
}

func (r *ShiftRepository) GetShiftEmployees(ctx context.Context, shiftID int, businessID int) ([]shift.User, error) {
	query := `
		SELECT u.id, u.username, u.name, u.email, u.role, u.status
		FROM shift_employees se
		JOIN users u ON se.employee_id = u.id
		WHERE se.shift_id = $1 AND se.business_id = $2
		ORDER BY u.name`

	rows, err := r.db.QueryContext(ctx, query, shiftID, businessID)
	if err != nil {
		log.Printf("Error querying shift employees: %v", err)
		return nil, err
	}
	defer rows.Close()

	var employees []shift.User
	for rows.Next() {
		var user shift.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Name,
			&user.Email,
			&user.Role,
			&user.Status,
		)
		if err != nil {
			log.Printf("Error scanning employee: %v", err)
			return nil, err
		}
		employees = append(employees, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating employee rows: %v", err)
		return nil, err
	}

	return employees, nil
}

func (r *ShiftRepository) CreateShift(ctx context.Context, s *shift.Shift, employeeIDs []int, businessID int) (*shift.Shift, error) {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for creating shift: %v", err)
		return nil, err
	}

	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	// Insert the shift
	shiftSQL := `
		INSERT INTO shifts (date, start_time, end_time, manager_id, business_id, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err = tx.QueryRowContext(ctx, shiftSQL,
		s.Date, s.StartTime, s.EndTime, s.ManagerID, businessID, s.Notes, s.CreatedAt, s.UpdatedAt,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		tx.Rollback()
		log.Printf("Error inserting shift: %v", err)
		return nil, err
	}

	s.BusinessID = businessID

	// Insert shift employees
	employeeSQL := `INSERT INTO shift_employees (shift_id, employee_id, business_id) VALUES ($1, $2, $3)`
	for _, employeeID := range employeeIDs {
		_, err = tx.ExecContext(ctx, employeeSQL, s.ID, employeeID, businessID)
		if err != nil {
			tx.Rollback()
			log.Printf("Error inserting shift employee %d: %v", employeeID, err)
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction for creating shift: %v", err)
		return nil, err
	}

	return s, nil
}

func (r *ShiftRepository) UpdateShift(ctx context.Context, s *shift.Shift, employeeIDs []int, businessID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for updating shift: %v", err)
		return err
	}

	s.UpdatedAt = time.Now()

	// Update the shift
	shiftSQL := `
		UPDATE shifts 
		SET date = $1, start_time = $2, end_time = $3, manager_id = $4, notes = $5, updated_at = $6
		WHERE id = $7 AND business_id = $8`

	result, err := tx.ExecContext(ctx, shiftSQL,
		s.Date, s.StartTime, s.EndTime, s.ManagerID, s.Notes, s.UpdatedAt, s.ID, businessID,
	)
	if err != nil {
		tx.Rollback()
		log.Printf("Error updating shift %d: %v", s.ID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		log.Printf("Error getting affected rows for shift update %d: %v", s.ID, err)
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("shift with ID %d not found", s.ID)
	}

	// Delete existing shift employees
	_, err = tx.ExecContext(ctx, "DELETE FROM shift_employees WHERE shift_id = $1", s.ID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error deleting existing shift employees for shift %d: %v", s.ID, err)
		return err
	}

	// Insert new shift employees
	employeeSQL := `INSERT INTO shift_employees (shift_id, employee_id, business_id) VALUES ($1, $2, $3)`
	for _, employeeID := range employeeIDs {
		_, err = tx.ExecContext(ctx, employeeSQL, s.ID, employeeID, businessID)
		if err != nil {
			tx.Rollback()
			log.Printf("Error inserting shift employee %d: %v", employeeID, err)
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction for updating shift: %v", err)
		return err
	}

	return nil
}

func (r *ShiftRepository) DeleteShift(ctx context.Context, shiftID int, businessID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		log.Printf("Error starting transaction for deleting shift: %v", err)
		return err
	}

	// Delete shift employees first (foreign key constraint)
	_, err = tx.ExecContext(ctx, "DELETE FROM shift_employees WHERE shift_id = $1", shiftID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error deleting shift employees for shift %d: %v", shiftID, err)
		return err
	}

	// Delete the shift
	result, err := tx.ExecContext(ctx, "DELETE FROM shifts WHERE id = $1 AND business_id = $2", shiftID, businessID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error deleting shift %d: %v", shiftID, err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		log.Printf("Error getting affected rows for shift deletion %d: %v", shiftID, err)
		return err
	}

	if rowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("shift with ID %d not found", shiftID)
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction for deleting shift: %v", err)
		return err
	}

	return nil
}

func (r *ShiftRepository) GetEmployeeShifts(ctx context.Context, employeeID int, businessID int) ([]shift.ShiftWithEmployees, error) {
	// Get all shifts for a specific employee
	query := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, s.manager_id, s.business_id, s.notes, s.created_at, s.updated_at,
			m.username as manager_username, m.name as manager_name, m.email as manager_email, m.role as manager_role, m.status as manager_status
		FROM shifts s
		LEFT JOIN users m ON s.manager_id = m.id
		INNER JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1 AND s.business_id = $2
		ORDER BY s.date ASC, s.start_time ASC`

	rows, err := r.db.QueryContext(ctx, query, employeeID, businessID)
	if err != nil {
		log.Printf("Error querying employee shifts: %v", err)
		return nil, err
	}
	defer rows.Close()

	var shifts []shift.ShiftWithEmployees
	for rows.Next() {
		var swe shift.ShiftWithEmployees
		var manager shift.User

		err := rows.Scan(
			&swe.ID, &swe.Date, &swe.StartTime, &swe.EndTime, &swe.ManagerID, &swe.BusinessID, &swe.Notes, &swe.CreatedAt, &swe.UpdatedAt,
			&manager.Username, &manager.Name, &manager.Email, &manager.Role, &manager.Status,
		)
		if err != nil {
			log.Printf("Error scanning employee shift: %v", err)
			return nil, err
		}
		manager.ID = swe.ManagerID
		swe.Manager = &manager

		// Get employees for this shift
		employees, err := r.GetShiftEmployees(ctx, swe.ID, businessID)
		if err != nil {
			log.Printf("Error getting employees for shift %d: %v", swe.ID, err)
			// Continue with empty employees instead of failing
		}
		swe.Employees = employees

		shifts = append(shifts, swe)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating employee shift rows: %v", err)
		return nil, err
	}

	return shifts, nil
}

func (r *ShiftRepository) GetCurrentAndUpcomingShifts(ctx context.Context, employeeID int, businessID int) (*shift.ShiftWithEmployees, []shift.ShiftWithEmployees, error) {
	// Get current shift (today) and upcoming shifts for a specific employee
	currentQuery := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, s.manager_id, s.business_id, s.notes, s.created_at, s.updated_at,
			m.username as manager_username, m.name as manager_name, m.email as manager_email, m.role as manager_role, m.status as manager_status
		FROM shifts s
		LEFT JOIN users m ON s.manager_id = m.id
		INNER JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1 AND s.business_id = $2 AND s.date = CURRENT_DATE
		ORDER BY s.start_time ASC
		LIMIT 1`

	var currentShift *shift.ShiftWithEmployees
	var manager shift.User
	var swe shift.ShiftWithEmployees

	err := r.db.QueryRowContext(ctx, currentQuery, employeeID, businessID).Scan(
		&swe.ID, &swe.Date, &swe.StartTime, &swe.EndTime, &swe.ManagerID, &swe.BusinessID, &swe.Notes, &swe.CreatedAt, &swe.UpdatedAt,
		&manager.Username, &manager.Name, &manager.Email, &manager.Role, &manager.Status,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error querying current shift: %v", err)
		return nil, nil, err
	}

	if err != sql.ErrNoRows {
		manager.ID = swe.ManagerID
		swe.Manager = &manager

		// Get employees for current shift
		employees, err := r.GetShiftEmployees(ctx, swe.ID, businessID)
		if err != nil {
			log.Printf("Error getting employees for current shift %d: %v", swe.ID, err)
		} else {
			swe.Employees = employees
		}
		currentShift = &swe
	}

	// Get upcoming shifts
	upcomingQuery := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, s.manager_id, s.business_id, s.notes, s.created_at, s.updated_at,
			m.username as manager_username, m.name as manager_name, m.email as manager_email, m.role as manager_role, m.status as manager_status
		FROM shifts s
		LEFT JOIN users m ON s.manager_id = m.id
		INNER JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1 AND s.business_id = $2 AND s.date > CURRENT_DATE
		ORDER BY s.date ASC, s.start_time ASC`

	rows, err := r.db.QueryContext(ctx, upcomingQuery, employeeID, businessID)
	if err != nil {
		log.Printf("Error querying upcoming shifts: %v", err)
		return currentShift, nil, err
	}
	defer rows.Close()

	var upcomingShifts []shift.ShiftWithEmployees
	for rows.Next() {
		var uswe shift.ShiftWithEmployees
		var umanager shift.User

		err := rows.Scan(
			&uswe.ID, &uswe.Date, &uswe.StartTime, &uswe.EndTime, &uswe.ManagerID, &uswe.BusinessID, &uswe.Notes, &uswe.CreatedAt, &uswe.UpdatedAt,
			&umanager.Username, &umanager.Name, &umanager.Email, &umanager.Role, &umanager.Status,
		)
		if err != nil {
			log.Printf("Error scanning upcoming shift: %v", err)
			return currentShift, nil, err
		}
		umanager.ID = uswe.ManagerID
		uswe.Manager = &umanager

		// Get employees for this shift
		employees, err := r.GetShiftEmployees(ctx, uswe.ID, businessID)
		if err != nil {
			log.Printf("Error getting employees for upcoming shift %d: %v", uswe.ID, err)
			// Continue with empty employees instead of failing
		}
		uswe.Employees = employees

		upcomingShifts = append(upcomingShifts, uswe)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating upcoming shift rows: %v", err)
		return currentShift, nil, err
	}

	return currentShift, upcomingShifts, nil
}
