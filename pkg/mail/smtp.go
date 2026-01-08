package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/pixelvide/laravel-go/pkg/config"
)

// SMTPMailer implements Mailer using net/smtp
type SMTPMailer struct {
	cfg config.MailConfig
}

// NewSMTPMailer creates a new SMTPMailer
func NewSMTPMailer(cfg config.MailConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

// Send sends the given message using SMTP
func (m *SMTPMailer) Send(ctx context.Context, msg *Message) error {
	addr := fmt.Sprintf("%s:%s", m.cfg.Host, m.cfg.Port)

	// Set default From address if not provided
	if msg.From == "" {
		if m.cfg.FromAddress != "" {
			msg.From = m.cfg.FromAddress
			if m.cfg.FromName != "" {
				msg.From = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.FromAddress)
			}
		}
	}

	// Build the email body
	body, err := buildEmailBody(msg)
	if err != nil {
		return fmt.Errorf("failed to build email body: %w", err)
	}

	// Determine authentication
	var auth smtp.Auth
	if m.cfg.Username != "" && m.cfg.Password != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	fromAddr, err := parseEmailAddress(msg.From)
	if err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}

	recipients := getAllRecipients(msg)

	// Handle Implicit TLS (usually port 465)
	if m.cfg.Encryption == "ssl" || m.cfg.Port == "465" {
		return m.sendWithImplicitTLS(addr, auth, fromAddr, recipients, []byte(body))
	}

	// Handle STARTTLS or Unencrypted (usually port 587 or 25)
	// smtp.SendMail handles STARTTLS automatically if the server supports it
	return smtp.SendMail(addr, auth, fromAddr, recipients, []byte(body))
}

func (m *SMTPMailer) sendWithImplicitTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         m.cfg.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to dial TLS: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer func() {
		_ = client.Quit()
	}()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, t := range to {
		if err = client.Rcpt(t); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", t, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

func getAllRecipients(msg *Message) []string {
	recipients := make([]string, 0, len(msg.To)+len(msg.Cc)+len(msg.Bcc))
	recipients = append(recipients, msg.To...)
	recipients = append(recipients, msg.Cc...)
	recipients = append(recipients, msg.Bcc...)
	return recipients
}

func buildEmailBody(msg *Message) (string, error) {
	var headers []string

	// Sanitize headers
	sanitize := func(s string) string {
		return strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
	}

	headers = append(headers, fmt.Sprintf("From: %s", sanitize(msg.From)))
	headers = append(headers, fmt.Sprintf("To: %s", sanitize(strings.Join(msg.To, ", "))))
	if len(msg.Cc) > 0 {
		headers = append(headers, fmt.Sprintf("Cc: %s", sanitize(strings.Join(msg.Cc, ", "))))
	}
	headers = append(headers, fmt.Sprintf("Subject: %s", sanitize(msg.Subject)))

	contentType := "text/plain"
	if msg.ContentType != "" {
		contentType = msg.ContentType
	}
	headers = append(headers, fmt.Sprintf("Content-Type: %s; charset=UTF-8", sanitize(contentType)))

	return fmt.Sprintf("%s\r\n\r\n%s", strings.Join(headers, "\r\n"), msg.Body), nil
}

// parseEmailAddress extracts the address part using net/mail
func parseEmailAddress(input string) (string, error) {
	addr, err := mail.ParseAddress(input)
	if err != nil {
		// Fallback for simple cases or return error
		// If input is just "foo@bar.com" it works.
		// If input is "Name <foo@bar.com>" it works.
		return "", err
	}
	return addr.Address, nil
}
