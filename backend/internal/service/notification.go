package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"restaurant-management/internal/domain/notification"
	"restaurant-management/internal/domain/user"
)

type NotificationService struct {
	repo         notification.Repository
	emailService notification.EmailService
	userService  user.Service
}

func NewNotificationService(repo notification.Repository, emailService notification.EmailService, userService user.Service) notification.Service {
	return &NotificationService{
		repo:         repo,
		emailService: emailService,
		userService:  userService,
	}
}

func (s *NotificationService) CreateNotification(ctx context.Context, businessID int, req notification.CreateNotificationRequest) (*notification.Notification, error) {
	// Validate recipients
	for _, email := range req.Recipients {
		if !s.emailService.ValidateEmail(email) {
			return nil, fmt.Errorf("invalid email address: %s", email)
		}
	}

	n := &notification.Notification{
		BusinessID: businessID,
		Type:       req.Type,
		Subject:    req.Subject,
		Body:       req.Body,
		Recipients: req.Recipients,
		Status:     notification.NotificationStatusPending,
	}

	err := s.repo.Create(ctx, n)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (s *NotificationService) SendNotification(ctx context.Context, n *notification.Notification) error {
	message := notification.EmailMessage{
		To:      n.Recipients,
		Subject: n.Subject,
		Body:    n.Body,
		IsHTML:  true,
	}

	err := s.emailService.SendEmail(message)
	sentAt := time.Now()

	if err != nil {
		errorMsg := err.Error()
		updateErr := s.repo.UpdateStatus(ctx, n.ID, notification.NotificationStatusFailed, nil, &errorMsg)
		if updateErr != nil {
			log.Printf("Failed to update notification status: %v", updateErr)
		}
		return err
	}

	err = s.repo.UpdateStatus(ctx, n.ID, notification.NotificationStatusSent, &sentAt, nil)
	if err != nil {
		log.Printf("Failed to update notification status to sent: %v", err)
	}

	return nil
}

func (s *NotificationService) ProcessPendingNotifications(ctx context.Context) error {
	notifications, err := s.repo.GetPendingNotifications(ctx, 50) // Process 50 at a time
	if err != nil {
		return err
	}

	for _, n := range notifications {
		if err := s.SendNotification(ctx, &n); err != nil {
			log.Printf("Failed to send notification %d: %v", n.ID, err)
		}
	}

	return nil
}

func (s *NotificationService) GetRecentNotifications(ctx context.Context, businessID int) ([]notification.Notification, error) {
	return s.repo.GetRecentNotifications(ctx, businessID, 10)
}

func (s *NotificationService) GetNotificationStats(ctx context.Context, businessID int) (*notification.NotificationStats, error) {
	return s.repo.GetStats(ctx, businessID)
}

// getManagerEmails gets all manager email addresses for a specific business
func (s *NotificationService) getManagerEmails(ctx context.Context, businessID int) ([]string, error) {
	users, err := s.userService.GetUsers(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for business %d: %w", businessID, err)
	}

	var managerEmails []string
	for _, user := range users {
		if user.Role == "manager" && user.Email != "" {
			managerEmails = append(managerEmails, user.Email)
		}
	}

	if len(managerEmails) == 0 {
		return nil, fmt.Errorf("no managers found for business %d", businessID)
	}

	return managerEmails, nil
}

func (s *NotificationService) SendLowInventoryAlert(ctx context.Context, businessID int, itemName string, currentStock, minStock float64, unit string) error {
	subject := fmt.Sprintf("⚠️ Низкий запас: %s", itemName)
	body := s.generateLowInventoryHTML(itemName, currentStock, minStock, unit)

	// Get actual manager emails for this business
	recipients, err := s.getManagerEmails(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to get manager emails: %w", err)
	}

	req := notification.CreateNotificationRequest{
		Type:       notification.NotificationTypeLowInventory,
		Subject:    subject,
		Body:       body,
		Recipients: recipients,
	}

	_, err = s.CreateNotification(ctx, businessID, req)
	return err
}

func (s *NotificationService) SendNewHiringAlert(ctx context.Context, businessID int, applicantName, position, experience, location string) error {
	subject := fmt.Sprintf("📋 Новая заявка на найм: %s", position)
	body := s.generateNewHiringHTML(applicantName, position, experience, location)

	// Get actual manager emails for this business (they handle hiring)
	recipients, err := s.getManagerEmails(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to get manager emails: %w", err)
	}

	req := notification.CreateNotificationRequest{
		Type:       notification.NotificationTypeNewHiring,
		Subject:    subject,
		Body:       body,
		Recipients: recipients,
	}

	_, err = s.CreateNotification(ctx, businessID, req)
	return err
}

func (s *NotificationService) SendWeeklyReport(ctx context.Context, businessID int, reportData interface{}) error {
	subject := "📊 Еженедельный отчет готов"
	body := s.generateWeeklyReportHTML(reportData)

	// Get actual manager emails for this business
	recipients, err := s.getManagerEmails(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to get manager emails: %w", err)
	}

	req := notification.CreateNotificationRequest{
		Type:       notification.NotificationTypeWeeklyReport,
		Subject:    subject,
		Body:       body,
		Recipients: recipients,
	}

	_, err = s.CreateNotification(ctx, businessID, req)
	return err
}

func (s *NotificationService) SendDailyReport(ctx context.Context, businessID int, reportData interface{}) error {
	subject := "📈 Ежедневный отчет готов"
	body := s.generateDailyReportHTML(reportData)

	// Get actual manager emails for this business
	recipients, err := s.getManagerEmails(ctx, businessID)
	if err != nil {
		return fmt.Errorf("failed to get manager emails: %w", err)
	}

	req := notification.CreateNotificationRequest{
		Type:       notification.NotificationTypeDailyReport,
		Subject:    subject,
		Body:       body,
		Recipients: recipients,
	}

	_, err = s.CreateNotification(ctx, businessID, req)
	return err
}

// HTML template generators
func (s *NotificationService) generateLowInventoryHTML(itemName string, currentStock, minStock float64, unit string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Низкий запас товара</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4;">
    <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h2 style="color: #d32f2f; margin-bottom: 20px;">⚠️ Предупреждение о низком запасе</h2>
        <p>Здравствуйте!</p>
        <p>Уведомляем вас о том, что запас товара <strong>%s</strong> достиг критически низкого уровня.</p>
        <div style="background-color: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 4px; margin: 20px 0;">
            <p style="margin: 0;"><strong>Текущий остаток:</strong> %.2f %s</p>
            <p style="margin: 5px 0 0 0;"><strong>Минимальный уровень:</strong> %.2f %s</p>
        </div>
        <p>Рекомендуется пополнить запас как можно скорее.</p>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">
            Это автоматическое уведомление из системы управления рестораном.
        </p>
    </div>
</body>
</html>
`, itemName, currentStock, unit, minStock, unit)
}

func (s *NotificationService) generateNewHiringHTML(applicantName, position, experience, location string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Новая заявка на найм</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4;">
    <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h2 style="color: #1976d2; margin-bottom: 20px;">📋 Новая заявка на найм</h2>
        <p>Поступила новая заявка на трудоустройство:</p>
        <div style="background-color: #e3f2fd; border: 1px solid #bbdefb; padding: 15px; border-radius: 4px; margin: 20px 0;">
            <p style="margin: 0;"><strong>Кандидат:</strong> %s</p>
            <p style="margin: 5px 0;"><strong>Должность:</strong> %s</p>
            <p style="margin: 5px 0;"><strong>Опыт работы:</strong> %s</p>
            <p style="margin: 5px 0 0 0;"><strong>Локация:</strong> %s</p>
        </div>
        <p>Пожалуйста, ознакомьтесь с заявкой в системе управления.</p>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">
            Это автоматическое уведомление из системы управления рестораном.
        </p>
    </div>
</body>
</html>
`, applicantName, position, experience, location)
}

func (s *NotificationService) generateWeeklyReportHTML(reportData interface{}) string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Еженедельный отчет</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4;">
    <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h2 style="color: #388e3c; margin-bottom: 20px;">📊 Еженедельный отчет</h2>
        <p>Еженедельный отчет по вашему ресторану готов!</p>
        <div style="background-color: #e8f5e8; border: 1px solid #c8e6c9; padding: 15px; border-radius: 4px; margin: 20px 0;">
            <p style="margin: 0;"><strong>Период:</strong> За прошедшую неделю</p>
            <p style="margin: 5px 0 0 0;"><strong>Основные метрики:</strong> выручка +15%, посещаемость +8%</p>
        </div>
        <p>Полный отчет доступен в системе управления рестораном.</p>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">
            Это автоматическое уведомление из системы управления рестораном.
        </p>
    </div>
</body>
</html>
`
}

func (s *NotificationService) generateDailyReportHTML(reportData interface{}) string {
	return `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Ежедневный отчет</title>
</head>
<body style="font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4;">
    <div style="max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h2 style="color: #1976d2; margin-bottom: 20px;">📈 Ежедневный отчет</h2>
        <p>Ежедневный отчет по вашему ресторану готов!</p>
        <div style="background-color: #e3f2fd; border: 1px solid #bbdefb; padding: 15px; border-radius: 4px; margin: 20px 0;">
            <p style="margin: 0;"><strong>Дата:</strong> Сегодня</p>
            <p style="margin: 5px 0 0 0;"><strong>Основные показатели:</strong> Готовы к просмотру</p>
        </div>
        <p>Подробный отчет доступен в системе управления рестораном.</p>
        <p style="color: #666; font-size: 14px; margin-top: 30px;">
            Это автоматическое уведомление из системы управления рестораном.
        </p>
    </div>
</body>
</html>
`
}
