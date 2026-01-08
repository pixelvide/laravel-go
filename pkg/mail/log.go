package mail

import (
	"context"
	"fmt"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/rs/zerolog/log"
)

// LogMailer implements Mailer by logging messages
type LogMailer struct {
	cfg config.MailConfig
}

// NewLogMailer creates a new LogMailer
func NewLogMailer(cfg config.MailConfig) *LogMailer {
	return &LogMailer{cfg: cfg}
}

// Send logs the message details
func (m *LogMailer) Send(ctx context.Context, msg *Message) error {
	// Set default From address if not provided
	if msg.From == "" {
		if m.cfg.FromAddress != "" {
			msg.From = m.cfg.FromAddress
			if m.cfg.FromName != "" {
				msg.From = fmt.Sprintf("%s <%s>", m.cfg.FromName, m.cfg.FromAddress)
			}
		}
	}

	logger := log.Ctx(ctx).With().
		Str("mailer", "log").
		Str("from", msg.From).
		Strs("to", msg.To).
		Str("subject", msg.Subject).
		Str("content_type", msg.ContentType).
		Logger()

	if len(msg.Cc) > 0 {
		logger = logger.With().Strs("cc", msg.Cc).Logger()
	}
	if len(msg.Bcc) > 0 {
		logger = logger.With().Strs("bcc", msg.Bcc).Logger()
	}

	logger.Info().Msg("Sending email")

	// Also log the body for debugging purposes, but maybe at debug level or just printed
	// Since this is a "log" mailer, the purpose is to see the email.
	logger.Info().Msgf("Body:\n%s", msg.Body)

	return nil
}
