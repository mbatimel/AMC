// Package mailer contains the SMTP delivery adapter shared by admin flows.
package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/smtp"
	"net/textproto"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/admin/pkg/models"
)

type SMTPMailer struct {
	host       string
	port       string
	username   string
	password   string
	from       string
	tlsEnabled bool
	timeout    time.Duration
}

type lineWriter struct {
	w      io.Writer
	column int
	width  int
}

func (w *lineWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		remaining := w.width - w.column
		if remaining > len(p) {
			remaining = len(p)
		}
		n, err := w.w.Write(p[:remaining])
		written += n
		w.column += n
		if err != nil {
			return written, err
		}
		p = p[n:]
		if w.column == w.width {
			if _, err = io.WriteString(w.w, "\r\n"); err != nil {
				return written, err
			}
			w.column = 0
		}
	}
	return written, nil
}

func NewSMTPMailer(logger zerolog.Logger, host, port, username, password, from string, tlsEnabled bool, timeout time.Duration) *SMTPMailer {
	mailer := &SMTPMailer{
		host: host, port: port, username: username, password: password,
		from: from, tlsEnabled: tlsEnabled, timeout: timeout,
	}
	missing := make([]string, 0, 2)
	if strings.TrimSpace(host) == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if strings.TrimSpace(from) == "" {
		missing = append(missing, "SMTP_FROM")
	}
	if len(missing) > 0 {
		logger.Warn().Str("component", "smtp").Strs("missing", missing).Msg("SMTP is not configured, email delivery is disabled")
	}
	return mailer
}

func (m *SMTPMailer) Send(ctx context.Context, to string, subject string, body string) error {
	if err := validateHeaders(m.from, to, subject); err != nil {
		return err
	}
	return m.send(ctx, to, buildTextMessage(m.from, to, subject, body))
}

func (m *SMTPMailer) SendWithAttachments(ctx context.Context, to string, subject string, body string, attachments []models.ImageFile) error {
	if err := validateHeaders(m.from, to, subject); err != nil {
		return err
	}
	message, err := buildMultipartMessage(m.from, to, subject, body, attachments)
	if err != nil {
		return err
	}
	return m.send(ctx, to, message)
}

func (m *SMTPMailer) send(ctx context.Context, to string, message []byte) error {
	if strings.TrimSpace(m.host) == "" || strings.TrimSpace(m.from) == "" {
		return fmt.Errorf("SMTP is not configured")
	}
	if m.timeout <= 0 {
		return fmt.Errorf("SMTP timeout must be positive")
	}

	addr := net.JoinHostPort(m.host, m.port)
	dialer := &net.Dialer{Timeout: m.timeout}
	deadline := time.Now().Add(m.timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}

	var conn net.Conn
	var err error
	tlsConfig := &tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12}
	if m.tlsEnabled && m.port == "465" {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("dial SMTP: %w", err)
	}
	defer conn.Close()
	if err = conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("set SMTP deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return fmt.Errorf("create SMTP client: %w", err)
	}
	defer client.Close()
	if m.tlsEnabled && m.port != "465" {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("SMTP server does not support STARTTLS")
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("start SMTP TLS: %w", err)
		}
	}
	if m.username != "" {
		if err = client.Auth(smtp.PlainAuth("", m.username, m.password, m.host)); err != nil {
			return fmt.Errorf("authenticate SMTP: %w", err)
		}
	}
	if err = client.Mail(m.from); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message: %w", err)
	}
	if _, err = w.Write(message); err != nil {
		_ = w.Close()
		return fmt.Errorf("write SMTP message: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("finish SMTP message: %w", err)
	}
	if err = client.Quit(); err != nil {
		return fmt.Errorf("quit SMTP session: %w", err)
	}
	return nil
}

func containsHeaderInjection(value string) bool {
	return strings.ContainsAny(value, "\r\n\x00")
}

func validateHeaders(values ...string) error {
	for _, value := range values {
		if containsHeaderInjection(value) {
			return fmt.Errorf("refusing to send email: header value contains CR, LF or NUL")
		}
	}
	return nil
}

func writeCommonHeaders(w io.Writer, from, to, subject string) {
	fmt.Fprintf(w, "From: %s\r\n", from)
	fmt.Fprintf(w, "To: %s\r\n", to)
	fmt.Fprintf(w, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", subject))
	fmt.Fprint(w, "MIME-Version: 1.0\r\n")
}

func buildTextMessage(from string, to string, subject string, body string) []byte {
	var b bytes.Buffer
	writeCommonHeaders(&b, from, to, subject)
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	b.WriteString(body)
	return b.Bytes()
}

func buildMultipartMessage(from, to, subject, body string, attachments []models.ImageFile) ([]byte, error) {
	var b bytes.Buffer
	writeCommonHeaders(&b, from, to, subject)
	multipartWriter := multipart.NewWriter(&b)
	fmt.Fprintf(&b, "Content-Type: multipart/mixed; boundary=%q\r\n\r\n", multipartWriter.Boundary())

	textPart, err := multipartWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type": []string{"text/plain; charset=UTF-8"},
	})
	if err != nil {
		return nil, fmt.Errorf("create email text part: %w", err)
	}
	if _, err = io.WriteString(textPart, body); err != nil {
		return nil, fmt.Errorf("write email text part: %w", err)
	}

	for _, attachment := range attachments {
		name := filepath.Base(attachment.FileName)
		if name == "." || name == "" || containsHeaderInjection(name) {
			return nil, fmt.Errorf("invalid attachment name")
		}
		part, createErr := multipartWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              []string{http.DetectContentType(attachment.Content)},
			"Content-Disposition":       []string{mime.FormatMediaType("attachment", map[string]string{"filename": name})},
			"Content-Transfer-Encoding": []string{"base64"},
		})
		if createErr != nil {
			return nil, fmt.Errorf("create email attachment part: %w", createErr)
		}
		encoder := base64.NewEncoder(base64.StdEncoding, &lineWriter{w: part, width: 76})
		if _, err = encoder.Write(attachment.Content); err != nil {
			return nil, fmt.Errorf("encode email attachment: %w", err)
		}
		if err = encoder.Close(); err != nil {
			return nil, fmt.Errorf("finish email attachment: %w", err)
		}
	}
	if err = multipartWriter.Close(); err != nil {
		return nil, fmt.Errorf("finish multipart email: %w", err)
	}
	return b.Bytes(), nil
}
