package mail

import "context"

// Message represents an email message
type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	ContentType string // e.g., "text/plain", "text/html"
}

// Mailer is the interface for sending emails
type Mailer interface {
	// Send sends the given message
	Send(ctx context.Context, msg *Message) error
}
