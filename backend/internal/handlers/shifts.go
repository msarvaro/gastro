package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"restaurant-management/internal/database"
	"restaurant-management/internal/middleware"
	"restaurant-management/internal/models"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type ShiftHandler struct {
	db *database.DB
}

func NewShiftHandler(db *database.DB) *ShiftHandler {
	return &ShiftHandler{db: db}
}

// GetAllShifts возвращает список всех смен
func (h *ShiftHandler) GetAllShifts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in token")
		return
	}

	// Получаем информацию о пользователе для проверки роли
	user, err := h.db.GetUserByID(int(userID))
	if err != nil {
		log.Printf("Error GetAllShifts - fetching user %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}

	// Проверяем роль пользователя
	if user.Role != "manager" && user.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "Access denied: Requires manager or admin role")
		return
	}

	// Получаем параметры пагинации из запроса
	page := 1
	limit := 10

	pageParam := r.URL.Query().Get("page")
	if pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	limitParam := r.URL.Query().Get("limit")
	if limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	// Получаем все смены
	shifts, total, err := h.db.GetAllShifts(page, limit)
	if err != nil {
		log.Printf("Error GetAllShifts - fetching shifts: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch shifts")
		return
	}

	// Формируем ответ
	response := struct {
		Shifts []models.ShiftWithEmployees `json:"shifts"`
		Total  int                         `json:"total"`
		Page   int                         `json:"page"`
		Limit  int                         `json:"limit"`
	}{
		Shifts: shifts,
		Total:  total,
		Page:   page,
		Limit:  limit,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetShiftByID возвращает информацию о конкретной смене
func (h *ShiftHandler) GetShiftByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in token")
		return
	}

	// Получаем информацию о пользователе для проверки роли
	user, err := h.db.GetUserByID(int(userID))
	if err != nil {
		log.Printf("Error GetShiftByID - fetching user %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}

	// Проверяем роль пользователя
	if user.Role != "manager" && user.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "Access denied: Requires manager or admin role")
		return
	}

	// Получаем ID смены из параметров URL
	vars := mux.Vars(r)
	shiftIDStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing shift ID")
		return
	}

	// Преобразуем ID в число
	shiftID, err := strconv.Atoi(shiftIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid shift ID")
		return
	}

	// Получаем информацию о смене
	shift, err := h.db.GetShiftByID(shiftID)
	if err != nil {
		log.Printf("Error GetShiftByID - fetching shift %d: %v", shiftID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch shift details")
		return
	}

	if shift == nil {
		respondWithError(w, http.StatusNotFound, "Shift not found")
		return
	}

	// Возвращаем данные о смене
	respondWithJSON(w, http.StatusOK, shift)
}

// CreateShift создает новую смену
func (h *ShiftHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in token")
		return
	}

	// Получаем информацию о пользователе для проверки роли
	user, err := h.db.GetUserByID(int(userID))
	if err != nil {
		log.Printf("Error CreateShift - fetching user %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}

	// Проверяем роль пользователя
	if user.Role != "manager" && user.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "Access denied: Requires manager or admin role")
		return
	}

	// Декодируем данные запроса
	var req models.CreateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error CreateShift - decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Валидация данных
	if req.StartTime == "" || req.EndTime == "" || req.ManagerID == 0 {
		respondWithError(w, http.StatusBadRequest, "Missing required fields: start_time, end_time, manager_id")
		return
	}

	// Парсим дату и время
	var shiftDate time.Time
	var startTime, endTime time.Time
	var parseErr error

	// Парсим дату (если не указана, используем сегодняшнюю)
	if req.Date == "" {
		shiftDate = time.Now()
	} else {
		shiftDate, parseErr = time.Parse("2006-01-02", req.Date)
		if parseErr != nil {
			log.Printf("Error CreateShift - parsing date %s: %v", req.Date, parseErr)
			respondWithError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
			return
		}
	}

	// Парсим время начала и конца
	startTime, parseErr = time.Parse("15:04", req.StartTime)
	if parseErr != nil {
		log.Printf("Error CreateShift - parsing start time %s: %v", req.StartTime, parseErr)
		respondWithError(w, http.StatusBadRequest, "Invalid start time format, use HH:MM")
		return
	}

	endTime, parseErr = time.Parse("15:04", req.EndTime)
	if parseErr != nil {
		log.Printf("Error CreateShift - parsing end time %s: %v", req.EndTime, parseErr)
		respondWithError(w, http.StatusBadRequest, "Invalid end time format, use HH:MM")
		return
	}

	// Создаем объект смены
	shift := &models.Shift{
		Date:      shiftDate,
		StartTime: time.Date(shiftDate.Year(), shiftDate.Month(), shiftDate.Day(), startTime.Hour(), startTime.Minute(), 0, 0, time.Local),
		EndTime:   time.Date(shiftDate.Year(), shiftDate.Month(), shiftDate.Day(), endTime.Hour(), endTime.Minute(), 0, 0, time.Local),
		ManagerID: req.ManagerID,
		Notes:     req.Notes,
	}

	// Создаем смену в БД
	createdShift, err := h.db.CreateShift(shift, req.EmployeeIDs)
	if err != nil {
		log.Printf("Error CreateShift - saving shift: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to create shift")
		return
	}

	// Возвращаем созданную смену
	respondWithJSON(w, http.StatusCreated, createdShift)
}

// UpdateShift обновляет информацию о смене
func (h *ShiftHandler) UpdateShift(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in token")
		return
	}

	// Получаем информацию о пользователе для проверки роли
	user, err := h.db.GetUserByID(int(userID))
	if err != nil {
		log.Printf("Error UpdateShift - fetching user %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}

	// Проверяем роль пользователя
	if user.Role != "manager" && user.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "Access denied: Requires manager or admin role")
		return
	}

	// Получаем ID смены из параметров URL
	vars := mux.Vars(r)
	shiftIDStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing shift ID")
		return
	}

	// Преобразуем ID в число
	shiftID, err := strconv.Atoi(shiftIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid shift ID")
		return
	}

	// Проверяем существование смены
	existingShift, err := h.db.GetShiftByID(shiftID)
	if err != nil {
		log.Printf("Error UpdateShift - fetching shift %d: %v", shiftID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch shift details")
		return
	}

	if existingShift == nil {
		respondWithError(w, http.StatusNotFound, "Shift not found")
		return
	}

	// Декодируем данные запроса
	var req models.UpdateShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error UpdateShift - decoding request: %v", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Парсим дату и время
	var shiftDate time.Time
	var startTime, endTime time.Time
	var parseErr error

	// Парсим дату (если не указана, используем существующую)
	if req.Date == "" {
		shiftDate = existingShift.Date
	} else {
		shiftDate, parseErr = time.Parse("2006-01-02", req.Date)
		if parseErr != nil {
			log.Printf("Error UpdateShift - parsing date %s: %v", req.Date, parseErr)
			respondWithError(w, http.StatusBadRequest, "Invalid date format, use YYYY-MM-DD")
			return
		}
	}

	// Парсим время начала (если не указано, используем существующее)
	if req.StartTime == "" {
		startTime = existingShift.StartTime
	} else {
		tempStartTime, parseErr := time.Parse("15:04", req.StartTime)
		if parseErr != nil {
			log.Printf("Error UpdateShift - parsing start time %s: %v", req.StartTime, parseErr)
			respondWithError(w, http.StatusBadRequest, "Invalid start time format, use HH:MM")
			return
		}
		startTime = time.Date(shiftDate.Year(), shiftDate.Month(), shiftDate.Day(), tempStartTime.Hour(), tempStartTime.Minute(), 0, 0, time.Local)
	}

	// Парсим время окончания (если не указано, используем существующее)
	if req.EndTime == "" {
		endTime = existingShift.EndTime
	} else {
		tempEndTime, parseErr := time.Parse("15:04", req.EndTime)
		if parseErr != nil {
			log.Printf("Error UpdateShift - parsing end time %s: %v", req.EndTime, parseErr)
			respondWithError(w, http.StatusBadRequest, "Invalid end time format, use HH:MM")
			return
		}
		endTime = time.Date(shiftDate.Year(), shiftDate.Month(), shiftDate.Day(), tempEndTime.Hour(), tempEndTime.Minute(), 0, 0, time.Local)
	}

	// Проверяем ID менеджера (если не указан, используем существующий)
	managerID := existingShift.ManagerID
	if req.ManagerID != 0 {
		managerID = req.ManagerID
	}

	// Создаем обновленный объект смены
	updatedShift := &models.Shift{
		ID:        shiftID,
		Date:      shiftDate,
		StartTime: startTime,
		EndTime:   endTime,
		ManagerID: managerID,
		Notes:     req.Notes,
	}

	// Если список сотрудников не указан, оставляем существующий
	employeeIDs := req.EmployeeIDs
	if employeeIDs == nil || len(employeeIDs) == 0 {
		employeeIDs = make([]int, len(existingShift.Employees))
		for i, employee := range existingShift.Employees {
			employeeIDs[i] = employee.ID
		}
	}

	// Обновляем смену в БД
	err = h.db.UpdateShift(updatedShift, employeeIDs)
	if err != nil {
		log.Printf("Error UpdateShift - updating shift %d: %v", shiftID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to update shift")
		return
	}

	// Получаем обновленную информацию о смене
	updatedShiftWithEmployees, err := h.db.GetShiftByID(shiftID)
	if err != nil {
		log.Printf("Error UpdateShift - fetching updated shift %d: %v", shiftID, err)
		respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
		return
	}

	// Возвращаем обновленную смену
	respondWithJSON(w, http.StatusOK, updatedShiftWithEmployees)
}

// DeleteShift удаляет смену
func (h *ShiftHandler) DeleteShift(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in token")
		return
	}

	// Получаем информацию о пользователе для проверки роли
	user, err := h.db.GetUserByID(int(userID))
	if err != nil {
		log.Printf("Error DeleteShift - fetching user %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}

	// Проверяем роль пользователя
	if user.Role != "manager" && user.Role != "admin" {
		respondWithError(w, http.StatusForbidden, "Access denied: Requires manager or admin role")
		return
	}

	// Получаем ID смены из параметров URL
	vars := mux.Vars(r)
	shiftIDStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing shift ID")
		return
	}

	// Преобразуем ID в число
	shiftID, err := strconv.Atoi(shiftIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid shift ID")
		return
	}

	// Проверяем существование смены
	existingShift, err := h.db.GetShiftByID(shiftID)
	if err != nil {
		log.Printf("Error DeleteShift - fetching shift %d: %v", shiftID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch shift details")
		return
	}

	if existingShift == nil {
		respondWithError(w, http.StatusNotFound, "Shift not found")
		return
	}

	// Удаляем смену
	err = h.db.DeleteShift(shiftID)
	if err != nil {
		log.Printf("Error DeleteShift - deleting shift %d: %v", shiftID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to delete shift")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// GetEmployeeShifts возвращает смены для конкретного сотрудника
func (h *ShiftHandler) GetEmployeeShifts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserIDFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized: User ID not found in token")
		return
	}

	// Получаем информацию о пользователе
	user, err := h.db.GetUserByID(int(userID))
	if err != nil {
		log.Printf("Error GetEmployeeShifts - fetching user %d: %v", userID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to verify user")
		return
	}

	// Определяем, для какого сотрудника запрашиваются смены
	targetEmployeeID := int(userID) // По умолчанию запрос для текущего пользователя

	// Если пользователь менеджер или админ, он может запросить смены для любого сотрудника
	if user.Role == "manager" || user.Role == "admin" {
		employeeIDParam := r.URL.Query().Get("employee_id")
		if employeeIDParam != "" {
			if empID, err := strconv.Atoi(employeeIDParam); err == nil && empID > 0 {
				targetEmployeeID = empID
			}
		}
	}

	// Получаем смены сотрудника
	shifts, err := h.db.GetEmployeeShifts(targetEmployeeID)
	if err != nil {
		log.Printf("Error GetEmployeeShifts - fetching shifts for employee %d: %v", targetEmployeeID, err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch employee shifts")
		return
	}

	// Формируем ответ
	response := struct {
		Shifts []models.ShiftWithEmployees `json:"shifts"`
		Total  int                         `json:"total"`
	}{
		Shifts: shifts,
		Total:  len(shifts),
	}

	respondWithJSON(w, http.StatusOK, response)
}
