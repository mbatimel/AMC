package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

type fakeMailer struct {
	to      string
	subject string
	body    string
	sendErr error
	calls   int
}

func (f *fakeMailer) Send(_ context.Context, to string, subject string, body string) error {
	f.calls++
	f.to, f.subject, f.body = to, subject, body
	return f.sendErr
}

func TestRequestPasswordReset_KnownActiveUser_CreatesTokenAndSendsEmail(t *testing.T) {
	userID := uuid.New()
	var createdUserID uuid.UUID
	var createdHash string
	var invalidatedFor uuid.UUID
	storage := &stubStorage{
		getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
			return postgres.User{ID: userID, Email: "user@example.com", IsActive: true}, nil
		},
		invalidateUserPasswordResetTokensFn: func(_ context.Context, uid uuid.UUID) error {
			invalidatedFor = uid
			return nil
		},
		createPasswordResetTokenFn: func(_ context.Context, uid uuid.UUID, hash string, _ time.Time) error {
			createdUserID, createdHash = uid, hash
			return nil
		},
	}
	mail := &fakeMailer{}
	svc := &service{logger: zerolog.Nop(), storage: storage, mailer: mail, publicFrontBaseURL: "https://app.volint.ru", resetTokenTTL: 30 * time.Minute}

	sent, err := svc.RequestPasswordReset(context.Background(), " User@Example.com ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Fatal("expected emailSent=true")
	}
	if createdUserID != userID || createdHash == "" {
		t.Fatalf("expected token created for user %v, got userID=%v hash=%q", userID, createdUserID, createdHash)
	}
	if invalidatedFor != userID {
		t.Fatalf("expected old tokens invalidated for %v, got %v", userID, invalidatedFor)
	}
	if mail.calls != 1 || mail.to != "user@example.com" {
		t.Fatalf("expected email sent to user@example.com once, got calls=%d to=%q", mail.calls, mail.to)
	}
	if !strings.Contains(mail.body, "https://app.volint.ru/reset-password?token=") {
		t.Fatalf("expected reset link in email body, got: %s", mail.body)
	}
}

func TestRequestPasswordReset_UnknownEmail_NeutralSuccessNoEmail(t *testing.T) {
	storage := &stubStorage{getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
		return postgres.User{}, postgres.ErrUserNotFound
	}}
	mail := &fakeMailer{}
	svc := &service{logger: zerolog.Nop(), storage: storage, mailer: mail}

	sent, err := svc.RequestPasswordReset(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Fatal("expected emailSent=true even for an unknown email — must not leak whether the account exists")
	}
	if mail.calls != 0 {
		t.Fatal("must not send an email for an unknown account")
	}
}

func TestRequestPasswordReset_BlockedUser_NeutralSuccessNoEmail(t *testing.T) {
	storage := &stubStorage{getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
		return postgres.User{ID: uuid.New(), Email: "blocked@example.com", IsActive: false}, nil
	}}
	mail := &fakeMailer{}
	svc := &service{logger: zerolog.Nop(), storage: storage, mailer: mail}

	sent, err := svc.RequestPasswordReset(context.Background(), "blocked@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Fatal("expected emailSent=true for a blocked account too — same neutral response")
	}
	if mail.calls != 0 {
		t.Fatal("must not send an email for a blocked account")
	}
}

func TestRequestPasswordReset_InvalidEmail_ValidationError(t *testing.T) {
	svc := &service{logger: zerolog.Nop(), storage: &stubStorage{}}

	_, err := svc.RequestPasswordReset(context.Background(), "not-an-email")
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRequestPasswordReset_MailerFails_ReturnsEmailSentFalse(t *testing.T) {
	storage := &stubStorage{getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
		return postgres.User{ID: uuid.New(), Email: "user@example.com", IsActive: true}, nil
	}}
	mail := &fakeMailer{sendErr: errors.New("smtp down")}
	svc := &service{logger: zerolog.Nop(), storage: storage, mailer: mail, resetTokenTTL: time.Minute}

	sent, err := svc.RequestPasswordReset(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("mailer failure must not fail the request, got: %v", err)
	}
	if sent {
		t.Fatal("expected emailSent=false when mailer fails")
	}
}

func TestConfirmPasswordReset_ValidToken_UpdatesPasswordAndInvalidatesTokens(t *testing.T) {
	userID := uuid.New()
	var updatedUserID uuid.UUID
	var invalidatedFor uuid.UUID
	storage := &stubStorage{
		getPasswordResetTokenFn: func(context.Context, string) (postgres.PasswordResetToken, error) {
			return postgres.PasswordResetToken{UserID: userID, ExpiresAt: time.Now().Add(time.Minute)}, nil
		},
		updateUserPasswordFn: func(_ context.Context, uid uuid.UUID, _ string) error {
			updatedUserID = uid
			return nil
		},
		invalidateUserPasswordResetTokensFn: func(_ context.Context, uid uuid.UUID) error {
			invalidatedFor = uid
			return nil
		},
	}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	err := svc.ConfirmPasswordReset(context.Background(), "raw-token-value", "new-strong-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedUserID != userID {
		t.Fatalf("expected password updated for %v, got %v", userID, updatedUserID)
	}
	if invalidatedFor != userID {
		t.Fatalf("expected tokens invalidated for %v, got %v", userID, invalidatedFor)
	}
}

func TestConfirmPasswordReset_UnknownToken_TokenInvalid(t *testing.T) {
	storage := &stubStorage{getPasswordResetTokenFn: func(context.Context, string) (postgres.PasswordResetToken, error) {
		return postgres.PasswordResetToken{}, postgres.ErrTokenNotFound
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	err := svc.ConfirmPasswordReset(context.Background(), "bogus-token", "new-strong-password")
	if !customErrors.Is(err, customErrors.TokenInvalidError()) {
		t.Fatalf("expected token invalid error, got: %v", err)
	}
}

func TestConfirmPasswordReset_UsedToken_TokenInvalid(t *testing.T) {
	storage := &stubStorage{getPasswordResetTokenFn: func(context.Context, string) (postgres.PasswordResetToken, error) {
		return postgres.PasswordResetToken{
			UserID: uuid.New(), ExpiresAt: time.Now().Add(time.Minute),
			UsedAt: sql.NullTime{Time: time.Now(), Valid: true},
		}, nil
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	err := svc.ConfirmPasswordReset(context.Background(), "used-token", "new-strong-password")
	if !customErrors.Is(err, customErrors.TokenInvalidError()) {
		t.Fatalf("expected token invalid error for an already-used token, got: %v", err)
	}
}

func TestConfirmPasswordReset_ExpiredToken_TokenExpired(t *testing.T) {
	storage := &stubStorage{getPasswordResetTokenFn: func(context.Context, string) (postgres.PasswordResetToken, error) {
		return postgres.PasswordResetToken{UserID: uuid.New(), ExpiresAt: time.Now().Add(-time.Minute)}, nil
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	err := svc.ConfirmPasswordReset(context.Background(), "expired-token", "new-strong-password")
	if !customErrors.Is(err, customErrors.TokenExpiredError()) {
		t.Fatalf("expected token expired error, got: %v", err)
	}
}

func TestConfirmPasswordReset_WeakPassword_ValidationError(t *testing.T) {
	storage := &stubStorage{getPasswordResetTokenFn: func(context.Context, string) (postgres.PasswordResetToken, error) {
		return postgres.PasswordResetToken{UserID: uuid.New(), ExpiresAt: time.Now().Add(time.Minute)}, nil
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	err := svc.ConfirmPasswordReset(context.Background(), "some-token", "short")
	if err == nil {
		t.Fatal("expected validation error for a password shorter than the minimum")
	}
}

func TestConfirmPasswordReset_EmptyToken_ValidationError(t *testing.T) {
	svc := &service{logger: zerolog.Nop(), storage: &stubStorage{}}

	err := svc.ConfirmPasswordReset(context.Background(), "", "new-strong-password")
	if err == nil {
		t.Fatal("expected validation error for an empty token")
	}
}
