// back/admin/internal/mailer/smtp.go
package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/rs/zerolog"
)

// SMTPMailer sends plain-text notification emails. There is no email
// infrastructure elsewhere in the project, so this is a small, self-contained
// net/smtp sender rather than a shared package. If the host is not
// configured, Send logs and no-ops instead of failing the caller — email
// delivery must never block the admin action that triggered it (e.g.
// rejecting a signup request still deletes the account even if SMTP is down
// or unconfigured in this environment).
type SMTPMailer struct {
	host     string
	port     string
	username string
	password string
	from     string
	logger   zerolog.Logger
}

func NewSMTPMailer(logger zerolog.Logger, host, port, username, password, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, username: username, password: password, from: from, logger: logger}
}

func (m *SMTPMailer) Send(_ context.Context, to string, subject string, body string) error {
	if m.host == "" || m.from == "" {
		m.logger.Warn().Str("to", to).Msg("SMTP is not configured, skipping email")
		return nil
	}
	// Defense in depth: callers are expected to validate `to` before it gets
	// here (see internal/service.normalizeSignupEmail), but a header/subject
	// containing CR/LF/NUL must never reach net/smtp regardless.
	if containsHeaderInjection(m.from) || containsHeaderInjection(to) || containsHeaderInjection(subject) {
		return fmt.Errorf("refusing to send email: header value contains CR, LF or NUL")
	}

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	if err := smtp.SendMail(addr, auth, m.from, []string{to}, buildMessage(m.from, to, subject, body)); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}

func containsHeaderInjection(value string) bool {
	return strings.ContainsAny(value, "\r\n\x00")
}

func buildMessage(from string, to string, subject string, body string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	b.WriteString(body)
	return []byte(b.String())
}
