package database

import (
	"database/sql"
	"log"
	"restaurant-management/internal/models"
	"time"
)

// GetAllShifts возвращает все смены с ограниченными данными
func (db *DB) GetAllShifts(page, limit int) ([]models.ShiftWithEmployees, int, error) {
	// Базовый запрос для получения смен
	query := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, 
			s.manager_id, s.notes, s.created_at, s.updated_at,
			COUNT(se.employee_id) as employee_count
		FROM shifts s
		LEFT JOIN shift_employees se ON s.id = se.shift_id
		GROUP BY s.id
		ORDER BY s.date DESC, s.start_time DESC
	`

	// Запрос для получения общего количества
	countQuery := "SELECT COUNT(*) FROM shifts"

	// Получаем общее количество
	var total int
	err := db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		log.Printf("Error counting shifts: %v", err)
		return nil, 0, err
	}

	// Добавляем пагинацию
	if limit > 0 {
		query += " LIMIT $1 OFFSET $2"
	}

	var rows *sql.Rows
	var queryErr error
	if limit > 0 {
		rows, queryErr = db.Query(query, limit, (page-1)*limit)
	} else {
		rows, queryErr = db.Query(query)
	}

	if queryErr != nil {
		log.Printf("Error querying shifts: %v", queryErr)
		return nil, 0, queryErr
	}
	defer rows.Close()

	shifts := []models.ShiftWithEmployees{}
	for rows.Next() {
		var shift models.ShiftWithEmployees
		var employeeCount int

		err := rows.Scan(
			&shift.ID, &shift.Date, &shift.StartTime, &shift.EndTime,
			&shift.ManagerID, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt,
			&employeeCount,
		)
		if err != nil {
			log.Printf("Error scanning shift row: %v", err)
			continue
		}

		// Получаем данные менеджера
		manager, err := db.GetUserByID(shift.ManagerID)
		if err != nil {
			log.Printf("Warning: Could not get manager details for shift %d: %v", shift.ID, err)
		} else {
			shift.Manager = manager
		}

		// Получаем сотрудников для этой смены
		shift.Employees, err = db.GetShiftEmployees(shift.ID)
		if err != nil {
			log.Printf("Warning: Could not get employees for shift %d: %v", shift.ID, err)
		}

		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating shift rows: %v", err)
		return nil, 0, err
	}

	return shifts, total, nil
}

// GetShiftByID возвращает информацию о конкретной смене
func (db *DB) GetShiftByID(shiftID int) (*models.ShiftWithEmployees, error) {
	var shift models.ShiftWithEmployees

	// Получаем основную информацию о смене
	err := db.QueryRow(`
		SELECT id, date, start_time, end_time, manager_id, notes, created_at, updated_at 
		FROM shifts 
		WHERE id = $1`,
		shiftID,
	).Scan(
		&shift.ID, &shift.Date, &shift.StartTime, &shift.EndTime,
		&shift.ManagerID, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Смена не найдена
		}
		log.Printf("Error getting shift by ID %d: %v", shiftID, err)
		return nil, err
	}

	// Получаем данные менеджера
	manager, err := db.GetUserByID(shift.ManagerID)
	if err != nil {
		log.Printf("Warning: Could not get manager details for shift %d: %v", shift.ID, err)
	} else {
		shift.Manager = manager
	}

	// Получаем сотрудников для этой смены
	shift.Employees, err = db.GetShiftEmployees(shift.ID)
	if err != nil {
		log.Printf("Warning: Could not get employees for shift %d: %v", shift.ID, err)
	}

	return &shift, nil
}

// GetShiftEmployees возвращает список сотрудников, назначенных на смену
func (db *DB) GetShiftEmployees(shiftID int) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.name, u.email, u.role, u.status
		FROM users u
		JOIN shift_employees se ON u.id = se.employee_id
		WHERE se.shift_id = $1
		ORDER BY u.role, u.name, u.username
	`

	rows, err := db.Query(query, shiftID)
	if err != nil {
		log.Printf("Error querying shift employees: %v", err)
		return nil, err
	}
	defer rows.Close()

	var employees []models.User
	for rows.Next() {
		var employee models.User
		var name sql.NullString // Для обработки NULL в поле name

		err := rows.Scan(
			&employee.ID, &employee.Username, &name,
			&employee.Email, &employee.Role, &employee.Status,
		)
		if err != nil {
			log.Printf("Error scanning employee row: %v", err)
			continue
		}

		// Преобразуем NullString в строку
		if name.Valid {
			employee.Name = name.String
		}

		employees = append(employees, employee)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating employee rows: %v", err)
		return nil, err
	}

	return employees, nil
}

// CreateShift создает новую смену и связывает ее с сотрудниками
func (db *DB) CreateShift(shift *models.Shift, employeeIDs []int) (*models.Shift, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	// Вставляем запись о смене
	err = tx.QueryRow(`
		INSERT INTO shifts (date, start_time, end_time, manager_id, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id, created_at, updated_at
	`, shift.Date, shift.StartTime, shift.EndTime, shift.ManagerID, shift.Notes, time.Now()).
		Scan(&shift.ID, &shift.CreatedAt, &shift.UpdatedAt)

	if err != nil {
		tx.Rollback()
		log.Printf("Error creating shift: %v", err)
		return nil, err
	}

	// Назначаем сотрудников на смену
	for _, employeeID := range employeeIDs {
		_, err := tx.Exec(`
			INSERT INTO shift_employees (shift_id, employee_id, created_at)
			VALUES ($1, $2, $3)
		`, shift.ID, employeeID, time.Now())

		if err != nil {
			tx.Rollback()
			log.Printf("Error assigning employee %d to shift %d: %v", employeeID, shift.ID, err)
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing shift transaction: %v", err)
		return nil, err
	}

	return shift, nil
}

// UpdateShift обновляет информацию о смене и перераспределяет сотрудников
func (db *DB) UpdateShift(shift *models.Shift, employeeIDs []int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Обновляем смену
	_, err = tx.Exec(`
		UPDATE shifts
		SET date = $1, start_time = $2, end_time = $3, manager_id = $4, notes = $5, updated_at = $6
		WHERE id = $7
	`, shift.Date, shift.StartTime, shift.EndTime, shift.ManagerID, shift.Notes, time.Now(), shift.ID)

	if err != nil {
		tx.Rollback()
		log.Printf("Error updating shift %d: %v", shift.ID, err)
		return err
	}

	// Удаляем все существующие связи с сотрудниками
	_, err = tx.Exec("DELETE FROM shift_employees WHERE shift_id = $1", shift.ID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error clearing employees for shift %d: %v", shift.ID, err)
		return err
	}

	// Добавляем новые связи с сотрудниками
	for _, employeeID := range employeeIDs {
		_, err := tx.Exec(`
			INSERT INTO shift_employees (shift_id, employee_id, created_at)
			VALUES ($1, $2, $3)
		`, shift.ID, employeeID, time.Now())

		if err != nil {
			tx.Rollback()
			log.Printf("Error reassigning employee %d to shift %d: %v", employeeID, shift.ID, err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Error committing shift update transaction: %v", err)
		return err
	}

	return nil
}

// DeleteShift удаляет смену и все связанные записи
func (db *DB) DeleteShift(shiftID int) error {
	// Каскадное удаление обеспечивается внешним ключом с ON DELETE CASCADE
	_, err := db.Exec("DELETE FROM shifts WHERE id = $1", shiftID)
	if err != nil {
		log.Printf("Error deleting shift %d: %v", shiftID, err)
		return err
	}
	return nil
}

// GetEmployeeShifts возвращает смены сотрудника
func (db *DB) GetEmployeeShifts(employeeID int) ([]models.ShiftWithEmployees, error) {
	query := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, 
			s.manager_id, s.notes, s.created_at, s.updated_at
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		ORDER BY s.date ASC, s.start_time ASC
	`

	rows, err := db.Query(query, employeeID)
	if err != nil {
		log.Printf("Error querying employee shifts: %v", err)
		return nil, err
	}
	defer rows.Close()

	shifts := []models.ShiftWithEmployees{}
	for rows.Next() {
		var shift models.ShiftWithEmployees

		err := rows.Scan(
			&shift.ID, &shift.Date, &shift.StartTime, &shift.EndTime,
			&shift.ManagerID, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning employee shift row: %v", err)
			continue
		}

		// Получаем данные менеджера
		manager, err := db.GetUserByID(shift.ManagerID)
		if err == nil {
			shift.Manager = manager
		}

		shifts = append(shifts, shift)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating employee shift rows: %v", err)
		return nil, err
	}

	return shifts, nil
}

// GetCurrentAndUpcomingShifts возвращает текущую смену и список предстоящих смен для сотрудника
func (db *DB) GetCurrentAndUpcomingShifts(employeeID int) (*models.ShiftWithEmployees, []models.ShiftWithEmployees, error) {
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	// Запрос для получения текущей смены (сегодня и время между start_time и end_time)
	currentShiftQuery := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, 
			s.manager_id, s.notes, s.created_at, s.updated_at
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		AND date_trunc('day', s.date) = date_trunc('day', $2::date)
		AND $3::time BETWEEN s.start_time AND s.end_time
		LIMIT 1
	`

	// Запрос для получения предстоящих смен
	upcomingShiftsQuery := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, 
			s.manager_id, s.notes, s.created_at, s.updated_at
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		AND (
			date_trunc('day', s.date) > date_trunc('day', $2::date)
			OR 
			(date_trunc('day', s.date) = date_trunc('day', $2::date) AND s.start_time > $3::time)
		)
		ORDER BY s.date ASC, s.start_time ASC
		LIMIT 5
	`

	// Получаем текущую смену
	var currentShift *models.ShiftWithEmployees
	row := db.QueryRow(currentShiftQuery, employeeID, currentDate, currentTime)

	// Временные переменные для сканирования
	var shift models.ShiftWithEmployees
	err := row.Scan(
		&shift.ID, &shift.Date, &shift.StartTime, &shift.EndTime,
		&shift.ManagerID, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error getting current shift for employee %d: %v", employeeID, err)
			return nil, nil, err
		}
	} else {
		// Получаем данные менеджера
		manager, err := db.GetUserByID(shift.ManagerID)
		if err == nil {
			shift.Manager = manager
		}
		currentShift = &shift
	}

	// Получаем предстоящие смены
	rows, err := db.Query(upcomingShiftsQuery, employeeID, currentDate, currentTime)
	if err != nil {
		log.Printf("Error querying upcoming shifts for employee %d: %v", employeeID, err)
		return currentShift, nil, err
	}
	defer rows.Close()

	upcomingShifts := []models.ShiftWithEmployees{}
	for rows.Next() {
		var upcoming models.ShiftWithEmployees

		err := rows.Scan(
			&upcoming.ID, &upcoming.Date, &upcoming.StartTime, &upcoming.EndTime,
			&upcoming.ManagerID, &upcoming.Notes, &upcoming.CreatedAt, &upcoming.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning upcoming shift row: %v", err)
			continue
		}

		// Получаем данные менеджера
		manager, err := db.GetUserByID(upcoming.ManagerID)
		if err == nil {
			upcoming.Manager = manager
		}

		upcomingShifts = append(upcomingShifts, upcoming)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating upcoming shift rows: %v", err)
		return currentShift, nil, err
	}

	return currentShift, upcomingShifts, nil
}
