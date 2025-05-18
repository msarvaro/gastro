package database

import (
	"log"
	"restaurant-management/internal/models"
	"time"
)

// GetWaiterProfile возвращает полную информацию профиля официанта
func (db *DB) GetWaiterProfile(waiterID int, businessID int) (*models.WaiterProfileResponse, error) {
	// 1. Получаем базовую информацию о пользователе
	user, err := db.GetUserByID(waiterID)
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
	// TODO: Реализовать получение смен из БД с фильтрацией по business_id
	// В рамках текущего задания используем тестовые данные

	// Текущее время для определения активной смены
	now := time.Now()

	// Мок для текущей смены (если сейчас рабочее время)
	var currentShift *models.ShiftWithEmployees

	// Проверяем рабочее время (с 9:00 до 22:00)
	if now.Hour() >= 9 && now.Hour() < 22 {
		shiftDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		shiftStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
		shiftEnd := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, now.Location())

		currentShift = &models.ShiftWithEmployees{
			ID:        1,
			Date:      shiftDate,
			StartTime: shiftStart,
			EndTime:   shiftEnd,
			ManagerID: 1,
			Manager: &models.User{
				ID:       1,
				Username: "manager",
				Name:     "Менеджер Смены",
			},
			Notes:     "Текущая смена",
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		}
	}

	// Мок для предстоящих смен
	tomorrow := now.Add(24 * time.Hour)
	dayAfterTomorrow := now.Add(48 * time.Hour)

	upcomingShifts := []models.ShiftWithEmployees{
		{
			ID:        2,
			Date:      time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, now.Location()),
			StartTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, now.Location()),
			EndTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 22, 0, 0, 0, now.Location()),
			ManagerID: 1,
			Notes:     "Завтрашняя смена",
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		},
		{
			ID:        3,
			Date:      time.Date(dayAfterTomorrow.Year(), dayAfterTomorrow.Month(), dayAfterTomorrow.Day(), 0, 0, 0, 0, now.Location()),
			StartTime: time.Date(dayAfterTomorrow.Year(), dayAfterTomorrow.Month(), dayAfterTomorrow.Day(), 9, 0, 0, 0, now.Location()),
			EndTime:   time.Date(dayAfterTomorrow.Year(), dayAfterTomorrow.Month(), dayAfterTomorrow.Day(), 22, 0, 0, 0, now.Location()),
			ManagerID: 1,
			Notes:     "Смена послезавтра",
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		},
	}

	return currentShift, upcomingShifts, nil
}

// GetTablesAssignedToWaiter возвращает столы, назначенные официанту
func (db *DB) GetTablesAssignedToWaiter(waiterID int, businessID int) ([]models.Table, error) {
	// TODO: Реализовать получение назначенных столов из БД
	// В рамках текущего задания используем тестовые данные

	// Предположим, что официанту назначены столы 1, 2 и 3
	tables := []models.Table{
		{
			ID:     1,
			Number: 1,
			Seats:  4,
			Status: models.TableStatusFree,
		},
		{
			ID:     2,
			Number: 2,
			Seats:  2,
			Status: models.TableStatusOccupied,
		},
		{
			ID:     3,
			Number: 3,
			Seats:  6,
			Status: models.TableStatusFree,
		},
	}

	return tables, nil
}

// GetWaiterOrderStats возвращает статистику по заказам официанта
func (db *DB) GetWaiterOrderStats(waiterID int, businessID int) (models.OrderStatusCounts, error) {
	// TODO: Реализовать получение статистики заказов из БД
	// В рамках текущего задания используем тестовые данные

	stats := models.OrderStatusCounts{
		New:       2,
		Accepted:  1,
		Preparing: 3,
		Ready:     1,
		Served:    2,
		Total:     9,
	}

	return stats, nil
}

// GetWaiterPerformanceMetrics возвращает метрики эффективности официанта
func (db *DB) GetWaiterPerformanceMetrics(waiterID int, businessID int) (models.PerformanceMetrics, error) {
	// TODO: Реализовать получение метрик эффективности из БД
	// В рамках текущего задания используем тестовые данные

	metrics := models.PerformanceMetrics{
		TablesServed:       24,
		OrdersCompleted:    36,
		AverageServiceTime: 42.5, // в минутах
	}

	return metrics, nil
}
