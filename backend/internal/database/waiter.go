package database

import (
	"database/sql"
	"log"
	"restaurant-management/internal/models"
	"time"
)

// GetWaiterProfile возвращает полную информацию профиля официанта
func (db *DB) GetWaiterProfile(waiterID int, businessID int) (*models.WaiterProfileResponse, error) {
	// 1. Получаем базовую информацию о пользователе
	user, err := db.GetUserByID(waiterID, businessID)
	if err != nil {
		log.Printf("Error getting user for waiter profile: %v", err)
		return nil, err
	}

	profile := &models.WaiterProfileResponse{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		Email:    user.Email,
	}

	// 2. Получаем информацию о текущей и предстоящих сменах из реальной БД
	// Filter shifts by businessID
	currentShift, upcomingShifts, err := db.GetWaiterCurrentAndUpcomingShifts(waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get shifts for waiter %d: %v", waiterID, err)
		// Продолжаем выполнение, это некритичная ошибка
	} else {
		if currentShift != nil {
			shiftInfo := &models.ShiftInfo{
				ID:        currentShift.ID,
				StartTime: currentShift.StartTime,
				EndTime:   currentShift.EndTime,
				IsActive:  true,
			}
			profile.CurrentShift = shiftInfo

			// Если есть менеджер, добавляем его имя
			if currentShift.Manager != nil {
				managerName := currentShift.Manager.Name
				if managerName == "" {
					managerName = currentShift.Manager.Username
				}
				profile.CurrentShiftManager = managerName
			}
		}

		// Преобразуем предстоящие смены
		if len(upcomingShifts) > 0 {
			for _, shift := range upcomingShifts {
				shiftInfo := models.ShiftInfo{
					ID:        shift.ID,
					StartTime: shift.StartTime,
					EndTime:   shift.EndTime,
					IsActive:  false,
				}
				profile.UpcomingShifts = append(profile.UpcomingShifts, shiftInfo)
			}
		}
	}

	// 3. Получаем назначенные столы
	// Filter tables by businessID
	assignedTables, err := db.GetTablesAssignedToWaiter(waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get assigned tables for waiter %d: %v", waiterID, err)
		// Продолжаем выполнение, это некритичная ошибка
	} else {
		profile.AssignedTables = assignedTables
	}

	// 4. Получаем статистику по активным заказам
	// Filter orders by businessID
	orderStats, err := db.GetWaiterOrderStats(waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get order stats for waiter %d: %v", waiterID, err)
		// По умолчанию инициализируем пустую статистику
		orderStats = models.OrderStatusCounts{}
	}
	profile.OrderStats = orderStats

	// 5. Получаем показатели эффективности
	// Filter performance metrics by businessID
	performanceData, err := db.GetWaiterPerformanceMetrics(waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get performance metrics for waiter %d: %v", waiterID, err)
		// По умолчанию инициализируем пустую статистику
		performanceData = models.PerformanceMetrics{}
	}
	profile.PerformanceData = performanceData

	return profile, nil
}

// GetWaiterCurrentAndUpcomingShifts возвращает текущую и предстоящие смены официанта
func (db *DB) GetWaiterCurrentAndUpcomingShifts(waiterID int, businessID int) (*models.ShiftWithEmployees, []models.ShiftWithEmployees, error) {
	// Текущее время для определения активной смены
	now := time.Now()
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	// Запрос для получения текущей смены (сегодня и время между start_time и end_time)
	currentShiftQuery := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, 
			s.manager_id, s.notes, s.created_at, s.updated_at, s.business_id
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		AND s.business_id = $2
		AND s.date = $3::date
		AND $4::time BETWEEN s.start_time AND s.end_time
		LIMIT 1
	`

	// Получаем текущую смену
	var currentShift *models.ShiftWithEmployees
	var shift models.ShiftWithEmployees

	err := db.QueryRow(currentShiftQuery, waiterID, businessID, currentDate, currentTime).Scan(
		&shift.ID, &shift.Date, &shift.StartTime, &shift.EndTime,
		&shift.ManagerID, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt, &shift.BusinessID,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error getting current shift for waiter %d: %v", waiterID, err)
			// Продолжаем работу даже если текущая смена не найдена
		}
	} else {
		// Получаем данные менеджера
		manager, err := db.GetUserByID(shift.ManagerID, businessID)
		if err == nil {
			shift.Manager = manager
		} else {
			log.Printf("Error getting manager for shift %d: %v", shift.ID, err)
		}

		// Получаем сотрудников для этой смены
		shift.Employees, err = db.GetShiftEmployees(shift.ID, businessID)
		if err != nil {
			log.Printf("Warning: Could not get employees for shift %d: %v", shift.ID, err)
		}

		currentShift = &shift
	}

	// Запрос для получения предстоящих смен
	upcomingShiftsQuery := `
		SELECT 
			s.id, s.date, s.start_time, s.end_time, 
			s.manager_id, s.notes, s.created_at, s.updated_at, s.business_id
		FROM shifts s
		JOIN shift_employees se ON s.id = se.shift_id
		WHERE se.employee_id = $1
		AND s.business_id = $2
		AND (
			s.date > $3::date
			OR 
			(s.date = $3::date AND s.start_time > $4::time)
		)
		ORDER BY s.date ASC, s.start_time ASC
		LIMIT 5
	`

	// Получаем предстоящие смены
	rows, err := db.Query(upcomingShiftsQuery, waiterID, businessID, currentDate, currentTime)
	if err != nil {
		log.Printf("Error querying upcoming shifts for waiter %d: %v", waiterID, err)
		return currentShift, nil, err
	}
	defer rows.Close()

	upcomingShifts := []models.ShiftWithEmployees{}
	for rows.Next() {
		var upcoming models.ShiftWithEmployees

		err := rows.Scan(
			&upcoming.ID, &upcoming.Date, &upcoming.StartTime, &upcoming.EndTime,
			&upcoming.ManagerID, &upcoming.Notes, &upcoming.CreatedAt, &upcoming.UpdatedAt, &upcoming.BusinessID,
		)
		if err != nil {
			log.Printf("Error scanning upcoming shift row: %v", err)
			continue
		}

		// Получаем данные менеджера
		manager, err := db.GetUserByID(upcoming.ManagerID, businessID)
		if err == nil {
			upcoming.Manager = manager
		} else {
			log.Printf("Error getting manager for shift %d: %v", upcoming.ID, err)
		}

		// Получаем сотрудников для этой смены
		upcoming.Employees, err = db.GetShiftEmployees(upcoming.ID, businessID)
		if err != nil {
			log.Printf("Warning: Could not get employees for shift %d: %v", upcoming.ID, err)
		}

		upcomingShifts = append(upcomingShifts, upcoming)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating upcoming shift rows: %v", err)
		return currentShift, nil, err
	}

	return currentShift, upcomingShifts, nil
}

// GetTablesAssignedToWaiter возвращает столы, назначенные официанту
func (db *DB) GetTablesAssignedToWaiter(waiterID int, businessID int) ([]models.Table, error) {
	// Запрос для получения столов, назначенных официанту
	query := `
		SELECT 
			t.id, t.number, t.seats, t.status, t.reserved_at, t.occupied_at, t.business_id
		FROM tables t
		JOIN waiter_tables wt ON t.id = wt.table_id
		WHERE wt.waiter_id = $1
		AND t.business_id = $2
		ORDER BY t.number ASC
	`

	rows, err := db.Query(query, waiterID, businessID)
	if err != nil {
		log.Printf("Error querying tables assigned to waiter %d: %v", waiterID, err)
		return nil, err
	}
	defer rows.Close()

	tables := []models.Table{}
	for rows.Next() {
		var table models.Table
		var reservedAt, occupiedAt sql.NullTime

		var businessID int
		err := rows.Scan(
			&table.ID, &table.Number, &table.Seats, &table.Status,
			&reservedAt, &occupiedAt, &businessID,
		)
		if err != nil {
			log.Printf("Error scanning table row: %v", err)
			continue
		}

		// Преобразуем NullTime в *time.Time
		if reservedAt.Valid {
			table.ReservedAt = &reservedAt.Time
		}
		if occupiedAt.Valid {
			table.OccupiedAt = &occupiedAt.Time
		}

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating table rows: %v", err)
		return nil, err
	}

	// Если не удалось найти столы, возвращаем пустой слайс
	if len(tables) == 0 {
		return []models.Table{}, nil
	}

	return tables, nil
}

// GetWaiterOrderStats возвращает статистику по заказам официанта
func (db *DB) GetWaiterOrderStats(waiterID int, businessID int) (models.OrderStatusCounts, error) {
	stats := models.OrderStatusCounts{}

	// Запрос для получения количества заказов по статусам
	query := `
		SELECT 
			status, 
			COUNT(*) as count
		FROM orders
		WHERE waiter_id = $1 
		AND business_id = $2
		AND status NOT IN ('completed', 'cancelled')
		GROUP BY status
	`

	rows, err := db.Query(query, waiterID, businessID)
	if err != nil {
		log.Printf("Error querying order stats for waiter %d: %v", waiterID, err)
		return stats, err
	}
	defer rows.Close()

	// Вычисляем общее количество
	totalOrders := 0

	for rows.Next() {
		var status string
		var count int

		err := rows.Scan(&status, &count)
		if err != nil {
			log.Printf("Error scanning order stats row: %v", err)
			continue
		}

		// Увеличиваем счетчик для соответствующего статуса
		switch status {
		case "new":
			stats.New = count
		case "accepted":
			stats.Accepted = count
		case "preparing":
			stats.Preparing = count
		case "ready":
			stats.Ready = count
		case "served":
			stats.Served = count
		}

		totalOrders += count
	}

	stats.Total = totalOrders

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating order stats rows: %v", err)
		return stats, err
	}

	return stats, nil
}

// GetWaiterPerformanceMetrics возвращает метрики эффективности официанта
func (db *DB) GetWaiterPerformanceMetrics(waiterID int, businessID int) (models.PerformanceMetrics, error) {
	metrics := models.PerformanceMetrics{}

	// 1. Получаем количество обслуженных столов
	tablesServedQuery := `
		SELECT COUNT(DISTINCT table_id) 
		FROM orders 
		WHERE waiter_id = $1 
		AND business_id = $2
		AND status IN ('served', 'completed')
	`
	err := db.QueryRow(tablesServedQuery, waiterID, businessID).Scan(&metrics.TablesServed)
	if err != nil {
		log.Printf("Error getting tables served for waiter %d: %v", waiterID, err)
		// Продолжаем работу, но с дефолтным значением
	}

	// 2. Получаем количество выполненных заказов
	ordersCompletedQuery := `
		SELECT COUNT(*) 
		FROM orders 
		WHERE waiter_id = $1 
		AND business_id = $2
		AND status = 'completed'
	`
	err = db.QueryRow(ordersCompletedQuery, waiterID, businessID).Scan(&metrics.OrdersCompleted)
	if err != nil {
		log.Printf("Error getting orders completed for waiter %d: %v", waiterID, err)
		// Продолжаем работу, но с дефолтным значением
	}

	// 3. Получаем среднее время обслуживания (время между статусами 'new' и 'served')
	// Это приблизительная метрика, так как в нашей схеме мы не храним историю изменений статусов
	// Используем разницу между датой завершения и датой создания
	averageServiceTimeQuery := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (COALESCE(completed_at, updated_at) - created_at)) / 60), 0)
		FROM orders 
		WHERE waiter_id = $1 
		AND business_id = $2
		AND status = 'completed'
	`
	err = db.QueryRow(averageServiceTimeQuery, waiterID, businessID).Scan(&metrics.AverageServiceTime)
	if err != nil {
		log.Printf("Error getting average service time for waiter %d: %v", waiterID, err)
		// Устанавливаем значение по умолчанию
		metrics.AverageServiceTime = 0
	}

	return metrics, nil
}
