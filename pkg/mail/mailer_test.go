package mail

import (
	"context"
	"bytes"
	"testing"
	"strings"

	"github.com/pixelvide/laravel-go/pkg/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	tests := []struct {
		name      string
		config    config.MailConfig
		wantType  interface{}
		expectErr bool
	}{
		{
			name: "smtp",
			config: config.MailConfig{
				Mailer: "smtp",
			},
			wantType:  &SMTPMailer{},
			expectErr: false,
		},
		{
			name: "log",
			config: config.MailConfig{
				Mailer: "log",
			},
			wantType:  &LogMailer{},
			expectErr: false,
		},
		{
			name: "invalid",
			config: config.MailConfig{
				Mailer: "invalid",
			},
			wantType:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMailer(tt.config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, tt.wantType, got)
			}
		})
	}
}

func TestLogMailer_Send(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	ctx := logger.WithContext(context.Background())

	cfg := config.MailConfig{
		Mailer:      "log",
		FromAddress: "test@example.com",
		FromName:    "Test Sender",
	}
	mailer := NewLogMailer(cfg)

	msg := &Message{
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Body:    "Test Body",
	}

	err := mailer.Send(ctx, msg)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Sending email")
	assert.Contains(t, output, "Test Sender <test@example.com>")
	assert.Contains(t, output, "recipient@example.com")
	assert.Contains(t, output, "Test Subject")
	assert.Contains(t, output, "Test Body")
}

func TestSMTPHelper_ParseEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"test@example.com", "test@example.com", false},
		{"Name <test@example.com>", "test@example.com", false},
		{"<test@example.com>", "test@example.com", false},
		{"Invalid <test@example.com", "", true}, // net/mail is strict
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseEmailAddress(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestSMTPHelper_BuildBody(t *testing.T) {
	msg := &Message{
		From:        "sender@example.com",
		To:          []string{"to@example.com"},
		Subject:     "Test",
		Body:        "Body",
		ContentType: "text/html",
	}

	body, err := buildEmailBody(msg)
	assert.NoError(t, err)
	assert.Contains(t, body, "From: sender@example.com")
	assert.Contains(t, body, "To: to@example.com")
	assert.Contains(t, body, "Subject: Test")
	assert.Contains(t, body, "Content-Type: text/html")
	assert.True(t, strings.HasSuffix(body, "\r\n\r\nBody"))
}

func TestSMTPHelper_BuildBody_Sanitization(t *testing.T) {
	msg := &Message{
		From:        "sender@example.com",
		To:          []string{"to@example.com"},
		Subject:     "Test\r\nInjected: Header",
		Body:        "Body",
	}

	body, err := buildEmailBody(msg)
	assert.NoError(t, err)
	assert.Contains(t, body, "Subject: TestInjected: Header")
	assert.NotContains(t, body, "Subject: Test\r\n")
}

func TestSMTPHelper_Recipients(t *testing.T) {
	msg := &Message{
		To:  []string{"to1@example.com", "to2@example.com"},
		Cc:  []string{"cc1@example.com"},
		Bcc: []string{"bcc1@example.com"},
	}

	recipients := getAllRecipients(msg)
	assert.Len(t, recipients, 4)
	assert.Contains(t, recipients, "to1@example.com")
	assert.Contains(t, recipients, "cc1@example.com")
	assert.Contains(t, recipients, "bcc1@example.com")
}
