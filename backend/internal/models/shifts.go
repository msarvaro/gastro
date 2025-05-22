package models

import (
	"time"
)

// Shift представляет модель смены
type Shift struct {
	ID         int       `json:"id"`
	Date       time.Time `json:"date"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	ManagerID  int       `json:"manager_id"`
	BusinessID int       `json:"business_id"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ShiftWithEmployees представляет модель смены с назначенными сотрудниками
type ShiftWithEmployees struct {
	ID         int       `json:"id"`
	Date       time.Time `json:"date"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	ManagerID  int       `json:"manager_id"`
	BusinessID int       `json:"business_id"`
	Manager    *User     `json:"manager,omitempty"`
	Notes      string    `json:"notes"`
	Employees  []User    `json:"employees"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ShiftEmployee представляет модель связи смена-сотрудник
type ShiftEmployee struct {
	ID         int       `json:"id"`
	ShiftID    int       `json:"shift_id"`
	BusinessID int       `json:"business_id"`
	EmployeeID int       `json:"employee_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateShiftRequest представляет запрос на создание смены
type CreateShiftRequest struct {
	Date        string `json:"date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	ManagerID   string `json:"manager_id"`
	Notes       string `json:"notes"`
	EmployeeIDs []int  `json:"employee_ids"`
	// BusinessID is not required in request as it will be obtained from context
}

// UpdateShiftRequest представляет запрос на обновление смены
type UpdateShiftRequest struct {
	Date        string `json:"date"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	ManagerID   string `json:"manager_id"`
	Notes       string `json:"notes"`
	EmployeeIDs []int  `json:"employee_ids"`
}

// ShiftResponse представляет ответ API для смены
type ShiftResponse struct {
	ID        int       `json:"id"`
	Date      string    `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Manager   User      `json:"manager"`
	Notes     string    `json:"notes"`
	Employees []User    `json:"employees"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
