package notification

// EmailMessage represents an email message
type EmailMessage struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// EmailService defines the interface for email operations
type EmailService interface {
	// SendEmail sends an email message
	SendEmail(message EmailMessage) error

	// ValidateEmail validates if an email address is valid
	ValidateEmail(email string) bool
}
