package mail

import (
	"fmt"

	"github.com/pixelvide/laravel-go/pkg/config"
)

// NewMailer creates a new Mailer based on the configuration
func NewMailer(cfg config.MailConfig) (Mailer, error) {
	switch cfg.Mailer {
	case "smtp":
		return NewSMTPMailer(cfg), nil
	case "log":
		return NewLogMailer(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported mailer: %s", cfg.Mailer)
	}
}
