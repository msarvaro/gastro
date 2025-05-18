package database

import (
	"log"
	"restaurant-management/internal/models"
	"time"
)

// GetWaiterProfile возвращает полную информацию профиля официанта
func (db *DB) GetWaiterProfile(waiterID int) (*models.WaiterProfileResponse, error) {
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
	currentShift, upcomingShifts, err := db.GetCurrentAndUpcomingShifts(waiterID)
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
	assignedTables, err := db.GetTablesAssignedToWaiter(waiterID)
	if err != nil {
		log.Printf("Warning: Could not get assigned tables for waiter %d: %v", waiterID, err)
		// Продолжаем выполнение, это некритичная ошибка
	} else {
		profile.AssignedTables = assignedTables
	}

	// 4. Получаем статистику по активным заказам
	orderStats, err := db.GetWaiterOrderStats(waiterID)
	if err != nil {
		log.Printf("Warning: Could not get order stats for waiter %d: %v", waiterID, err)
		// По умолчанию инициализируем пустую статистику
		orderStats = models.OrderStatusCounts{}
	}
	profile.OrderStats = orderStats

	// 5. Получаем показатели эффективности
	performanceData, err := db.GetWaiterPerformanceMetrics(waiterID)
	if err != nil {
		log.Printf("Warning: Could not get performance metrics for waiter %d: %v", waiterID, err)
		// По умолчанию инициализируем пустую статистику
		performanceData = models.PerformanceMetrics{}
	}
	profile.PerformanceData = performanceData

	return profile, nil
}

// GetWaiterShifts возвращает текущую и предстоящие смены официанта
func (db *DB) GetWaiterShifts(waiterID int) (*models.ShiftInfo, []models.ShiftInfo, error) {
	// TODO: Реализовать получение смен из БД
	// В рамках текущего задания используем тестовые данные

	// Текущее время для определения активной смены
	now := time.Now()

	// Мок для текущей смены (если сейчас рабочее время)
	var currentShift *models.ShiftInfo

	// Проверяем рабочее время (с 9:00 до 22:00)
	if now.Hour() >= 9 && now.Hour() < 22 {
		shiftStart := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
		shiftEnd := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, now.Location())

		currentShift = &models.ShiftInfo{
			ID:        1,
			StartTime: shiftStart,
			EndTime:   shiftEnd,
			IsActive:  true,
		}
	}

	// Мок для предстоящих смен
	tomorrow := now.Add(24 * time.Hour)
	dayAfterTomorrow := now.Add(48 * time.Hour)

	upcomingShifts := []models.ShiftInfo{
		{
			ID:        2,
			StartTime: time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 9, 0, 0, 0, now.Location()),
			EndTime:   time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 22, 0, 0, 0, now.Location()),
			IsActive:  false,
		},
		{
			ID:        3,
			StartTime: time.Date(dayAfterTomorrow.Year(), dayAfterTomorrow.Month(), dayAfterTomorrow.Day(), 9, 0, 0, 0, now.Location()),
			EndTime:   time.Date(dayAfterTomorrow.Year(), dayAfterTomorrow.Month(), dayAfterTomorrow.Day(), 22, 0, 0, 0, now.Location()),
			IsActive:  false,
		},
	}

	return currentShift, upcomingShifts, nil
}

// GetTablesAssignedToWaiter возвращает столы, назначенные официанту
func (db *DB) GetTablesAssignedToWaiter(waiterID int) ([]models.Table, error) {
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
func (db *DB) GetWaiterOrderStats(waiterID int) (models.OrderStatusCounts, error) {
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
func (db *DB) GetWaiterPerformanceMetrics(waiterID int) (models.PerformanceMetrics, error) {
	// TODO: Реализовать получение метрик эффективности из БД
	// В рамках текущего задания используем тестовые данные

	metrics := models.PerformanceMetrics{
		TablesServed:       24,
		OrdersCompleted:    36,
		AverageServiceTime: 42.5, // в минутах
	}

	return metrics, nil
}
