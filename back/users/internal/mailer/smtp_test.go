package mailer

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestBuildTextMessageEncodesUnicodeSubject(t *testing.T) {
	message := string(buildTextMessage("sender@example.com", "user@example.com", "Аккаунт заблокирован", "Причина"))
	if !strings.Contains(message, "Subject: =?UTF-8?") || !strings.Contains(message, "Причина") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestSMTPMailerRejectsHeaderInjection(t *testing.T) {
	mailer := NewSMTPMailer(zerolog.Nop(), "smtp.example.com", "587", "", "", "sender@example.com", true, time.Second)
	if err := mailer.Send(context.Background(), "victim@example.com\r\nBcc: attacker@example.com", "subject", "body"); err == nil {
		t.Fatal("Send() error = nil")
	}
}

func TestSMTPMailerMissingConfigurationDoesNotPanic(t *testing.T) {
	var logs bytes.Buffer
	logger := zerolog.New(&logs)
	mailer := NewSMTPMailer(logger, "", "587", "", "", "", true, time.Second)

	if err := mailer.Send(context.Background(), "user@example.com", "subject", "body"); err == nil {
		t.Fatal("Send() error = nil, want disabled SMTP error")
	}
	if logOutput := logs.String(); !strings.Contains(logOutput, "SMTP is not configured, email delivery is disabled") ||
		!strings.Contains(logOutput, "SMTP_HOST") || !strings.Contains(logOutput, "SMTP_FROM") {
		t.Fatalf("startup log does not explain disabled SMTP: %s", logOutput)
	}
}
