package postgres

import (
	"context"
	"database/sql"
	"log"
	"restaurant-management/internal/domain/waiter"
	"time"
)

type WaiterRepository struct {
	db *DB
}

func NewWaiterRepository(db *DB) waiter.Repository {
	return &WaiterRepository{db: db}
}

func (r *WaiterRepository) GetWaiterProfile(ctx context.Context, waiterID int, businessID int) (*waiter.WaiterProfile, error) {
	// First get the user repository to get basic user info
	userRepo := NewUserRepository(r.db)
	user, err := userRepo.GetUserByID(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Error getting user for waiter profile: %v", err)
		return nil, err
	}

	profile := &waiter.WaiterProfile{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		Email:    user.Email,
	}

	// Get current and upcoming shifts
	currentShift, upcomingShifts, err := r.GetWaiterCurrentAndUpcomingShifts(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get shifts for waiter %d: %v", waiterID, err)
	} else {
		if currentShift != nil {
			shiftInfo := &waiter.ShiftInfo{
				ID:        currentShift.ID,
				StartTime: currentShift.StartTime,
				EndTime:   currentShift.EndTime,
				IsActive:  true,
			}
			profile.CurrentShift = shiftInfo

			if currentShift.Manager != nil {
				managerName := currentShift.Manager.Name
				if managerName == "" {
					managerName = currentShift.Manager.Username
				}
				profile.CurrentShiftManager = managerName
			}
		}

		if len(upcomingShifts) > 0 {
			for _, shift := range upcomingShifts {
				shiftInfo := waiter.ShiftInfo{
					ID:        shift.ID,
					StartTime: shift.StartTime,
					EndTime:   shift.EndTime,
					IsActive:  false,
				}
				profile.UpcomingShifts = append(profile.UpcomingShifts, shiftInfo)
			}
		}
	}

	// Get assigned tables
	assignedTables, err := r.GetTablesAssignedToWaiter(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get assigned tables for waiter %d: %v", waiterID, err)
	} else {
		profile.AssignedTables = assignedTables
	}

	// Get order stats
	orderStats, err := r.GetWaiterOrderStats(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get order stats for waiter %d: %v", waiterID, err)
		orderStats = waiter.OrderStatusCounts{}
	}
	profile.OrderStats = orderStats

	// Get performance metrics
	performanceData, err := r.GetWaiterPerformanceMetrics(ctx, waiterID, businessID)
	if err != nil {
		log.Printf("Warning: Could not get performance metrics for waiter %d: %v", waiterID, err)
		performanceData = waiter.PerformanceMetrics{}
	}
	profile.PerformanceData = performanceData

	return profile, nil
}

func (r *WaiterRepository) GetWaiterCurrentAndUpcomingShifts(ctx context.Context, waiterID int, businessID int) (*waiter.ShiftWithEmployees, []waiter.ShiftWithEmployees, error) {
	now := time.Now().UTC().Add(5 * time.Hour)
	currentDate := now.Format("2006-01-02")
	currentTime := now.Format("15:04:05")

	// Query for current shift
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

	var currentShift *waiter.ShiftWithEmployees
	var shift waiter.ShiftWithEmployees

	err := r.db.QueryRowContext(ctx, currentShiftQuery, waiterID, businessID, currentDate, currentTime).Scan(
		&shift.ID, &shift.Date, &shift.StartTime, &shift.EndTime,
		&shift.ManagerID, &shift.Notes, &shift.CreatedAt, &shift.UpdatedAt, &shift.BusinessID,
	)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Error getting current shift for waiter %d: %v", waiterID, err)
		}
	} else {
		// Get manager data
		userRepo := NewUserRepository(r.db)
		manager, err := userRepo.GetUserByID(ctx, shift.ManagerID, businessID)
		if err == nil {
			shift.Manager = &waiter.User{
				ID:       manager.ID,
				Username: manager.Username,
				Name:     manager.Name,
				Email:    manager.Email,
				Role:     manager.Role,
				Status:   manager.Status,
			}
		} else {
			log.Printf("Error getting manager for shift %d: %v", shift.ID, err)
		}

		// Get shift employees
		shiftRepo := NewShiftRepository(r.db)
		shiftEmployees, err := shiftRepo.GetShiftEmployees(ctx, shift.ID, businessID)
		if err == nil {
			for _, emp := range shiftEmployees {
				shift.Employees = append(shift.Employees, waiter.User{
					ID:       emp.ID,
					Username: emp.Username,
					Name:     emp.Name,
					Email:    emp.Email,
					Role:     emp.Role,
					Status:   emp.Status,
				})
			}
		}
		if err != nil {
			log.Printf("Warning: Could not get employees for shift %d: %v", shift.ID, err)
		}

		currentShift = &shift
	}

	// Query for upcoming shifts
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

	rows, err := r.db.QueryContext(ctx, upcomingShiftsQuery, waiterID, businessID, currentDate, currentTime)
	if err != nil {
		log.Printf("Error querying upcoming shifts for waiter %d: %v", waiterID, err)
		return currentShift, nil, err
	}
	defer rows.Close()

	upcomingShifts := []waiter.ShiftWithEmployees{}
	userRepo := NewUserRepository(r.db)
	shiftRepo := NewShiftRepository(r.db)

	for rows.Next() {
		var upcoming waiter.ShiftWithEmployees

		err := rows.Scan(
			&upcoming.ID, &upcoming.Date, &upcoming.StartTime, &upcoming.EndTime,
			&upcoming.ManagerID, &upcoming.Notes, &upcoming.CreatedAt, &upcoming.UpdatedAt, &upcoming.BusinessID,
		)

		if err != nil {
			log.Printf("Error scanning upcoming shift: %v", err)
			continue
		}

		// Get manager data
		manager, err := userRepo.GetUserByID(ctx, upcoming.ManagerID, businessID)
		if err == nil {
			upcoming.Manager = &waiter.User{
				ID:       manager.ID,
				Username: manager.Username,
				Name:     manager.Name,
				Email:    manager.Email,
				Role:     manager.Role,
				Status:   manager.Status,
			}
		} else {
			log.Printf("Error getting manager for upcoming shift %d: %v", upcoming.ID, err)
		}

		// Get shift employees
		shiftEmployees, err := shiftRepo.GetShiftEmployees(ctx, upcoming.ID, businessID)
		if err == nil {
			for _, emp := range shiftEmployees {
				upcoming.Employees = append(upcoming.Employees, waiter.User{
					ID:       emp.ID,
					Username: emp.Username,
					Name:     emp.Name,
					Email:    emp.Email,
					Role:     emp.Role,
					Status:   emp.Status,
				})
			}
		}
		if err != nil {
			log.Printf("Warning: Could not get employees for upcoming shift %d: %v", upcoming.ID, err)
		}

		upcomingShifts = append(upcomingShifts, upcoming)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating upcoming shifts for waiter %d: %v", waiterID, err)
		return currentShift, nil, err
	}

	return currentShift, upcomingShifts, nil
}

func (r *WaiterRepository) GetTablesAssignedToWaiter(ctx context.Context, waiterID int, businessID int) ([]waiter.Table, error) {
	// Запрос для получения столов, назначенных официанту
	// ВАЖНО: Этот запрос должен обращаться к таблице waiter_tables или аналогичной для установления связи официант-стол
	query := `
		SELECT 
			t.id, t.number, t.seats, t.status, t.reserved_at, t.occupied_at
		FROM tables t
		JOIN waiter_tables wt ON t.id = wt.table_id
		WHERE wt.waiter_id = $1 AND t.business_id = $2
		ORDER BY t.number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, waiterID, businessID)
	if err != nil {
		log.Printf("Error querying tables assigned to waiter %d: %v", waiterID, err)
		return nil, err
	}
	defer rows.Close()

	var tables []waiter.Table
	for rows.Next() {
		var table waiter.Table
		err := rows.Scan(
			&table.ID,
			&table.Number,
			&table.Seats,
			&table.Status,
			&table.ReservedAt,
			&table.OccupiedAt,
		)
		if err != nil {
			log.Printf("Error scanning table assigned to waiter %d: %v", waiterID, err)
			continue
		}

		// Получаем активные заказы для этого стола
		orderQuery := `
			SELECT id, status, created_at, comment 
			FROM orders 
			WHERE table_id = $1 AND status NOT IN ('completed', 'cancelled') AND business_id = $2
			ORDER BY created_at ASC
		`

		orderRows, err := r.db.QueryContext(ctx, orderQuery, table.ID, businessID)
		if err != nil {
			log.Printf("Error querying orders for table %d: %v", table.ID, err)
			tables = append(tables, table)
			continue
		}

		var tableOrders []waiter.OrderInfo
		for orderRows.Next() {
			var orderInfo waiter.OrderInfo
			var comment sql.NullString
			if err := orderRows.Scan(&orderInfo.ID, &orderInfo.Status, &orderInfo.Time, &comment); err != nil {
				log.Printf("Error scanning order for table %d: %v", table.ID, err)
				continue
			}
			if comment.Valid {
				orderInfo.Comment = &comment.String
			}
			tableOrders = append(tableOrders, orderInfo)
		}
		orderRows.Close()

		table.Orders = tableOrders
		tables = append(tables, table)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating tables for waiter %d: %v", waiterID, err)
		return nil, err
	}

	return tables, nil
}

func (r *WaiterRepository) GetWaiterOrderStats(ctx context.Context, waiterID int, businessID int) (waiter.OrderStatusCounts, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN status = 'new' THEN 1 END) as new_orders,
			COUNT(CASE WHEN status = 'accepted' THEN 1 END) as accepted_orders,
			COUNT(CASE WHEN status = 'preparing' THEN 1 END) as preparing_orders,
			COUNT(CASE WHEN status = 'ready' THEN 1 END) as ready_orders,
			COUNT(CASE WHEN status = 'served' THEN 1 END) as served_orders,
			COUNT(*) as total_orders
		FROM orders
		WHERE waiter_id = $1 AND business_id = $2
		AND status NOT IN ('completed', 'cancelled')
		AND DATE(created_at) = CURRENT_DATE
	`

	var stats waiter.OrderStatusCounts
	err := r.db.QueryRowContext(ctx, query, waiterID, businessID).Scan(
		&stats.New,
		&stats.Accepted,
		&stats.Preparing,
		&stats.Ready,
		&stats.Served,
		&stats.Total,
	)

	if err != nil {
		log.Printf("Error getting order stats for waiter %d: %v", waiterID, err)
		return waiter.OrderStatusCounts{}, err
	}

	return stats, nil
}

func (r *WaiterRepository) GetWaiterPerformanceMetrics(ctx context.Context, waiterID int, businessID int) (waiter.PerformanceMetrics, error) {
	// Запрос для получения метрик производительности официанта за текущий день
	query := `
		SELECT 
			COUNT(DISTINCT table_id) as tables_served,
			COUNT(*) as orders_completed,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at))/60), 0) as avg_service_time
		FROM orders
		WHERE waiter_id = $1 AND business_id = $2
		AND status = 'completed'
		AND DATE(created_at) = CURRENT_DATE
	`

	var metrics waiter.PerformanceMetrics
	err := r.db.QueryRowContext(ctx, query, waiterID, businessID).Scan(
		&metrics.TablesServed,
		&metrics.OrdersCompleted,
		&metrics.AverageServiceTime,
	)

	if err != nil {
		log.Printf("Error getting performance metrics for waiter %d: %v", waiterID, err)
		return waiter.PerformanceMetrics{}, err
	}

	return metrics, nil
}
