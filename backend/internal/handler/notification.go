package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"restaurant-management/internal/domain/notification"
	"restaurant-management/internal/middleware"
)

type NotificationController struct {
	notificationService notification.Service
}

func NewNotificationController(notificationService notification.Service) *NotificationController {
	return &NotificationController{
		notificationService: notificationService,
	}
}

// GetRecentNotifications retrieves recent notifications for the dashboard
func (c *NotificationController) GetRecentNotifications(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	notifications, err := c.notificationService.GetRecentNotifications(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting recent notifications: %v", err)
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}

// GetNotificationStats retrieves notification statistics
func (c *NotificationController) GetNotificationStats(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	stats, err := c.notificationService.GetNotificationStats(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting notification stats: %v", err)
		http.Error(w, "Failed to fetch notification statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// CreateNotification creates a new notification
func (c *NotificationController) CreateNotification(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var req notification.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createdNotification, err := c.notificationService.CreateNotification(r.Context(), businessID, req)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdNotification)
}

// SendLowInventoryAlert sends a low inventory alert
func (c *NotificationController) SendLowInventoryAlert(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var req struct {
		ItemName     string  `json:"item_name"`
		CurrentStock float64 `json:"current_stock"`
		MinStock     float64 `json:"min_stock"`
		Unit         string  `json:"unit"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := c.notificationService.SendLowInventoryAlert(r.Context(), businessID, req.ItemName, req.CurrentStock, req.MinStock, req.Unit)
	if err != nil {
		log.Printf("Error sending low inventory alert: %v", err)
		http.Error(w, "Failed to send alert", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Low inventory alert sent successfully"})
}

// SendNewHiringAlert sends a new hiring alert
func (c *NotificationController) SendNewHiringAlert(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	var req struct {
		ApplicantName string `json:"applicant_name"`
		Position      string `json:"position"`
		Experience    string `json:"experience"`
		Location      string `json:"location"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := c.notificationService.SendNewHiringAlert(r.Context(), businessID, req.ApplicantName, req.Position, req.Experience, req.Location)
	if err != nil {
		log.Printf("Error sending new hiring alert: %v", err)
		http.Error(w, "Failed to send alert", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "New hiring alert sent successfully"})
}

// ProcessPendingNotifications manually processes pending notifications
func (c *NotificationController) ProcessPendingNotifications(w http.ResponseWriter, r *http.Request) {
	err := c.notificationService.ProcessPendingNotifications(r.Context())
	if err != nil {
		log.Printf("Error processing pending notifications: %v", err)
		http.Error(w, "Failed to process notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Pending notifications processed successfully"})
}

// GetNotificationsByBusiness retrieves paginated notifications for a business
func (c *NotificationController) GetNotificationsByBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, exists := middleware.GetBusinessIDFromContext(r.Context())
	if !exists {
		http.Error(w, "business_id not found in context", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters (not used in current implementation)
	// TODO: Implement pagination when adding GetByBusinessID method to service
	// limitStr := r.URL.Query().Get("limit")
	// offsetStr := r.URL.Query().Get("offset")

	// This would require adding the method to the repository
	// For now, we'll use GetRecentNotifications
	notifications, err := c.notificationService.GetRecentNotifications(r.Context(), businessID)
	if err != nil {
		log.Printf("Error getting notifications: %v", err)
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
