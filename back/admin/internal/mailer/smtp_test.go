package mailer

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mbatimel/AMC/admin/pkg/models"
	"github.com/rs/zerolog"
)

func TestBuildTextMessageEncodesUnicodeSubject(t *testing.T) {
	message := string(buildTextMessage("sender@example.com", "user@example.com", "Заявка отклонена", "Причина"))
	if !strings.Contains(message, "Subject: =?UTF-8?") || !strings.Contains(message, "Причина") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestBuildMultipartMessageContainsAttachments(t *testing.T) {
	message, err := buildMultipartMessage(
		"sender@example.com", "order@voint.ru", "Новая заявка", "Текст заявки",
		[]models.ImageFile{{FileName: "document.pdf", Content: []byte("%PDF-1.4\n")}},
	)
	if err != nil {
		t.Fatalf("buildMultipartMessage() error = %v", err)
	}
	text := string(message)
	for _, expected := range []string{"multipart/mixed", "document.pdf", "JVBERi0xLjQK", "Текст заявки"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("message does not contain %q: %s", expected, text)
		}
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
