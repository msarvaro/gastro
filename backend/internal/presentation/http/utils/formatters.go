package utils

import (
	"fmt"
	"time"

	"restaurant-management/internal/domain/consts"
	"restaurant-management/internal/presentation/http/dto/responses"
)

// FormatMoney formats an amount in cents to a localized money string
func FormatMoney(amountCents int) responses.MoneyResponse {
	amount := float64(amountCents) / 100.0

	return responses.MoneyResponse{
		Amount:    amountCents,
		Formatted: fmt.Sprintf("%.2f KZT", amount),
		Display:   fmt.Sprintf("%.0f KZT", amount), // No decimals for display
	}
}

// FormatDate formats a time to multiple formats for frontend consumption
func FormatDate(t time.Time) responses.FormattedDateResponse {
	if t.IsZero() {
		return responses.FormattedDateResponse{}
	}

	return responses.FormattedDateResponse{
		Raw:      t,
		Display:  formatDisplayDate(t),
		Date:     t.Format("2006-01-02"),
		Time:     t.Format("15:04"),
		Relative: formatRelativeTime(t),
		ISO:      t.Format(time.RFC3339),
	}
}

// FormatStatus provides status with translation and styling
func FormatStatus(status string, statusType string) responses.StatusResponse {
	var displayText, class, color string

	switch statusType {
	case "order":
		displayText = translateOrderStatus(status)
		class = fmt.Sprintf("status-%s", status)
		color = getOrderStatusColor(status)
	case "table":
		displayText = translateTableStatus(status)
		class = fmt.Sprintf("status-%s", status)
		color = getTableStatusColor(status)
	case "user":
		displayText = translateUserStatus(status)
		class = fmt.Sprintf("status-%s", status)
		color = getUserStatusColor(status)
	case "shift":
		displayText = translateShiftStatus(status)
		class = fmt.Sprintf("status-%s", status)
		color = getShiftStatusColor(status)
	default:
		displayText = status
		class = "status-default"
		color = "#6b7280"
	}

	return responses.StatusResponse{
		Value:       status,
		DisplayText: displayText,
		Class:       class,
		Color:       color,
	}
}

// CalculateStats calculates statistics with change indicators
func CalculateStats(current, previous int) responses.StatsResponse {
	var percentage float64
	var change float64
	var changeClass string

	if previous > 0 {
		percentage = float64(current) / float64(previous) * 100
		change = float64(current-previous) / float64(previous) * 100
	}

	if change > 0 {
		changeClass = "positive"
	} else if change < 0 {
		changeClass = "negative"
	} else {
		changeClass = "neutral"
	}

	return responses.StatsResponse{
		Total:       current,
		Active:      current, // This should be calculated based on context
		Inactive:    0,       // This should be calculated based on context
		Percentage:  percentage,
		Change:      change,
		ChangeClass: changeClass,
	}
}

// FormatDuration formats duration in a human-readable way
func FormatDuration(start, end time.Time) string {
	if start.IsZero() {
		return ""
	}

	var duration time.Duration
	if end.IsZero() {
		duration = time.Since(start)
	} else {
		duration = end.Sub(start)
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dч %dм", hours, minutes)
	}
	return fmt.Sprintf("%dм", minutes)
}

// CreateAvailableActions creates action buttons based on entity state and user permissions
func CreateAvailableActions(entityType string, entityStatus string, userRole string) []responses.ActionResponse {
	var actions []responses.ActionResponse

	switch entityType {
	case "order":
		actions = getOrderActions(entityStatus, userRole)
	case "table":
		actions = getTableActions(entityStatus, userRole)
	case "user":
		actions = getUserActions(entityStatus, userRole)
	}

	return actions
}

// TranslateRole translates role names to display text
func TranslateRole(role string) string {
	translations := map[string]string{
		consts.RoleAdmin:   "Администратор",
		consts.RoleManager: "Менеджер",
		consts.RoleWaiter:  "Официант",
		consts.RoleKitchen: "Повар",
		consts.RoleCashier: "Кассир",
	}

	if translation, exists := translations[role]; exists {
		return translation
	}
	return role
}

// Private helper functions

func formatDisplayDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Hour:
		return fmt.Sprintf("%d минут назад", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d часов назад", int(diff.Hours()))
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d дней назад", days)
	default:
		return t.Format("02.01.2006 15:04")
	}
}

func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "только что"
	case diff < time.Hour:
		return fmt.Sprintf("%d мин назад", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%d ч назад", int(diff.Hours()))
	default:
		return t.Format("02.01")
	}
}

func translateOrderStatus(status string) string {
	translations := map[string]string{
		consts.OrderStatusPending:   "Ожидает",
		consts.OrderStatusConfirmed: "Подтвержден",
		consts.OrderStatusPreparing: "Готовится",
		consts.OrderStatusReady:     "Готов",
		consts.OrderStatusDelivered: "Доставлен",
		consts.OrderStatusPaid:      "Оплачен",
		consts.OrderStatusCanceled:  "Отменен",
	}

	if translation, exists := translations[status]; exists {
		return translation
	}
	return status
}

func translateTableStatus(status string) string {
	translations := map[string]string{
		consts.TableStatusAvailable:   "Свободен",
		consts.TableStatusOccupied:    "Занят",
		consts.TableStatusReserved:    "Забронирован",
		consts.TableStatusMaintenance: "На обслуживании",
	}

	if translation, exists := translations[status]; exists {
		return translation
	}
	return status
}

func translateUserStatus(status string) string {
	translations := map[string]string{
		"active":   "Активен",
		"inactive": "Неактивен",
		"blocked":  "Заблокирован",
	}

	if translation, exists := translations[status]; exists {
		return translation
	}
	return status
}

func translateShiftStatus(status string) string {
	translations := map[string]string{
		"active":    "Активная",
		"scheduled": "Запланирована",
		"completed": "Завершена",
		"cancelled": "Отменена",
	}

	if translation, exists := translations[status]; exists {
		return translation
	}
	return status
}

func getOrderStatusColor(status string) string {
	colors := map[string]string{
		consts.OrderStatusPending:   "#2196F3",
		consts.OrderStatusConfirmed: "#9C27B0",
		consts.OrderStatusPreparing: "#FF9800",
		consts.OrderStatusReady:     "#4CAF50",
		consts.OrderStatusDelivered: "#607D8B",
		consts.OrderStatusPaid:      "#4CAF50",
		consts.OrderStatusCanceled:  "#F44336",
	}

	if color, exists := colors[status]; exists {
		return color
	}
	return "#6b7280"
}

func getTableStatusColor(status string) string {
	colors := map[string]string{
		consts.TableStatusAvailable:   "#4CAF50",
		consts.TableStatusOccupied:    "#F44336",
		consts.TableStatusReserved:    "#FF9800",
		consts.TableStatusMaintenance: "#9E9E9E",
	}

	if color, exists := colors[status]; exists {
		return color
	}
	return "#6b7280"
}

func getUserStatusColor(status string) string {
	colors := map[string]string{
		"active":   "#4CAF50",
		"inactive": "#F44336",
		"blocked":  "#9E9E9E",
	}

	if color, exists := colors[status]; exists {
		return color
	}
	return "#6b7280"
}

func getShiftStatusColor(status string) string {
	colors := map[string]string{
		"active":    "#4CAF50",
		"scheduled": "#2196F3",
		"completed": "#9E9E9E",
		"cancelled": "#F44336",
	}

	if color, exists := colors[status]; exists {
		return color
	}
	return "#6b7280"
}

func getOrderActions(status string, userRole string) []responses.ActionResponse {
	var actions []responses.ActionResponse

	if userRole == consts.RoleWaiter {
		switch status {
		case consts.OrderStatusPending:
			actions = append(actions, responses.ActionResponse{
				ID:      "accept",
				Label:   "Принять",
				Icon:    "check",
				Variant: "primary",
			})
			actions = append(actions, responses.ActionResponse{
				ID:      "cancel",
				Label:   "Отменить",
				Icon:    "x",
				Variant: "danger",
			})
		case consts.OrderStatusReady:
			actions = append(actions, responses.ActionResponse{
				ID:      "deliver",
				Label:   "Доставить",
				Icon:    "truck",
				Variant: "primary",
			})
		}
	}

	return actions
}

func getTableActions(status string, userRole string) []responses.ActionResponse {
	var actions []responses.ActionResponse

	if userRole == consts.RoleWaiter {
		switch status {
		case consts.TableStatusAvailable:
			actions = append(actions, responses.ActionResponse{
				ID:      "occupy",
				Label:   "Занять",
				Icon:    "user",
				Variant: "primary",
			})
			actions = append(actions, responses.ActionResponse{
				ID:      "reserve",
				Label:   "Забронировать",
				Icon:    "calendar",
				Variant: "secondary",
			})
		case consts.TableStatusOccupied:
			actions = append(actions, responses.ActionResponse{
				ID:      "free",
				Label:   "Освободить",
				Icon:    "check",
				Variant: "primary",
			})
		}
	}

	return actions
}

func getUserActions(status string, userRole string) []responses.ActionResponse {
	var actions []responses.ActionResponse

	if userRole == consts.RoleManager || userRole == consts.RoleAdmin {
		actions = append(actions, responses.ActionResponse{
			ID:      "edit",
			Label:   "Редактировать",
			Icon:    "edit",
			Variant: "secondary",
		})

		if status == "active" {
			actions = append(actions, responses.ActionResponse{
				ID:      "deactivate",
				Label:   "Деактивировать",
				Icon:    "x",
				Variant: "danger",
			})
		} else {
			actions = append(actions, responses.ActionResponse{
				ID:      "activate",
				Label:   "Активировать",
				Icon:    "check",
				Variant: "primary",
			})
		}
	}

	return actions
}
