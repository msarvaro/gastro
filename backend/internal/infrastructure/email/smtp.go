package email

import (
	"fmt"
	"net/smtp"
	"regexp"
	"restaurant-management/configs"
	"restaurant-management/internal/domain/notification"
	"strings"
)

type SMTPService struct {
	config *configs.SMTPConfig
}

func NewSMTPService(config *configs.SMTPConfig) notification.EmailService {
	return &SMTPService{config: config}
}

func (s *SMTPService) SendEmail(message notification.EmailMessage) error {
	// Setup authentication
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	// Prepare the email headers and body
	to := strings.Join(message.To, ",")

	var contentType string
	if message.IsHTML {
		contentType = "text/html; charset=UTF-8"
	} else {
		contentType = "text/plain; charset=UTF-8"
	}

	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: %s\r\n"+
			"\r\n"+
			"%s\r\n",
		s.config.From,
		to,
		message.Subject,
		contentType,
		message.Body,
	)

	// Send the email
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, s.config.From, message.To, []byte(msg))
}

func (s *SMTPService) ValidateEmail(email string) bool {
	// Simple email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
