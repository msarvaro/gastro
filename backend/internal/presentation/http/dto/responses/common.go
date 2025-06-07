package responses

import "time"

// Common response structures to reduce frontend complexity

// StatusResponse provides translated status with styling info
type StatusResponse struct {
	Value       string `json:"value"`
	DisplayText string `json:"display_text"`
	Class       string `json:"class"`
	Color       string `json:"color"`
}

// FormattedDateResponse provides multiple date formats
type FormattedDateResponse struct {
	Raw      time.Time `json:"raw"`
	Display  string    `json:"display"`  // "15 января 2024, 14:30"
	Date     string    `json:"date"`     // "2024-01-15"
	Time     string    `json:"time"`     // "14:30"
	Relative string    `json:"relative"` // "2 часа назад"
	ISO      string    `json:"iso"`      // ISO format
}

// MoneyResponse provides formatted money values
type MoneyResponse struct {
	Amount    int    `json:"amount"`    // Raw amount in cents
	Formatted string `json:"formatted"` // "1,250.50 KZT"
	Display   string `json:"display"`   // "1 251 KZT" (localized)
}

// StatsResponse provides calculated statistics
type StatsResponse struct {
	Total       int     `json:"total"`
	Active      int     `json:"active"`
	Inactive    int     `json:"inactive"`
	Percentage  float64 `json:"percentage"`
	Change      float64 `json:"change"`
	ChangeClass string  `json:"change_class"` // "positive", "negative", "neutral"
}

// PaginationResponse provides pagination metadata
type PaginationResponse struct {
	CurrentPage  int  `json:"current_page"`
	TotalPages   int  `json:"total_pages"`
	TotalItems   int  `json:"total_items"`
	ItemsPerPage int  `json:"items_per_page"`
	HasNext      bool `json:"has_next"`
	HasPrev      bool `json:"has_prev"`
}

// ActionResponse provides available actions for entities
type ActionResponse struct {
	ID           string `json:"id"`
	Label        string `json:"label"`
	Icon         string `json:"icon"`
	Variant      string `json:"variant"` // "primary", "secondary", "danger"
	Disabled     bool   `json:"disabled"`
	RequiresAuth bool   `json:"requires_auth"`
}

// FilterOptionResponse provides filter options
type FilterOptionResponse struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Count    int    `json:"count"`
	Selected bool   `json:"selected"`
}

// ErrorDetailsResponse provides detailed error information
type ErrorDetailsResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Actions []ActionResponse       `json:"actions,omitempty"`
}

// SuccessResponse provides standardized success responses
type SuccessResponse struct {
	Message string           `json:"message"`
	Data    interface{}      `json:"data,omitempty"`
	Actions []ActionResponse `json:"actions,omitempty"`
}
