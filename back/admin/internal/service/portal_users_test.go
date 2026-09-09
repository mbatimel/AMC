package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
)

func newPortalUsersService(storage *fakeStorage, mail *fakeMailer) *service {
	return NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, mail)
}

func TestInviteAdmin_CreatesUserAndSendsEmail(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	response, err := svc.InviteAdmin(context.Background(), uuid.New(), " NewAdmin@Example.com ", "  Иван Иванов  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Email != "newadmin@example.com" {
		t.Fatalf("expected normalized lowercase trimmed email, got %q", response.Email)
	}
	if len(response.Password) != passwordLength {
		t.Fatalf("expected generated password of length %d, got %q", passwordLength, response.Password)
	}
	if !response.EmailSent {
		t.Fatal("expected emailSent=true when mailer succeeds")
	}
	if mail.calls != 1 {
		t.Fatalf("expected mailer.Send called once, got %d", mail.calls)
	}
	if mail.to != "newadmin@example.com" {
		t.Fatalf("expected email sent to newadmin@example.com, got %q", mail.to)
	}
	if !strings.Contains(mail.body, response.Password) {
		t.Fatal("expected email body to contain the generated password")
	}
	if len(storage.inserted) != 1 || storage.inserted[0].action != "Приглашён администратор: newadmin@example.com" {
		t.Fatalf("expected audit log entry to be written, got %+v", storage.inserted)
	}
}

func TestInviteAdmin_NonAdminCaller_Forbidden(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: false}, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "new@example.com", "Имя")
	if err == nil {
		t.Fatal("expected error for non-admin caller")
	}
	if mail.calls != 0 {
		t.Fatal("mailer must not be called when caller is not admin")
	}
}

func TestInviteAdmin_EmailAlreadyTaken_Conflict(t *testing.T) {
	storage := &fakeStorage{createAdminUserFn: func(context.Context, string, string) (uuid.UUID, error) {
		return uuid.UUID{}, postgres.ErrEmailTaken
	}}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "taken@example.com", "Имя")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if mail.calls != 0 {
		t.Fatal("mailer must not be called when user creation fails")
	}
}

func TestInviteAdmin_InvalidEmail_BadRequest(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "not-an-email", "Имя")
	if err == nil {
		t.Fatal("expected validation error for invalid email")
	}
	if mail.calls != 0 {
		t.Fatal("mailer must not be called when email is invalid")
	}
}

func TestInviteAdmin_EmptyName_BadRequest(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "valid@example.com", "   ")
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}
}

func TestInviteAdmin_MailerFails_StillReturnsPasswordWithEmailSentFalse(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{sendErr: errors.New("smtp down")}
	svc := newPortalUsersService(storage, mail)

	response, err := svc.InviteAdmin(context.Background(), uuid.New(), "new@example.com", "Имя")
	if err != nil {
		t.Fatalf("mailer failure must not fail the request, got error: %v", err)
	}
	if response.EmailSent {
		t.Fatal("expected emailSent=false when mailer fails")
	}
	if response.Password == "" {
		t.Fatal("expected password to still be returned when mailer fails")
	}
}

func TestInviteAdmin_MailerNil_StillReturnsPasswordWithEmailSentFalse(t *testing.T) {
	storage := &fakeStorage{}
	// Pass a literal nil, not a *fakeMailer-typed nil (which would produce a
	// non-nil Mailer interface wrapping a nil pointer and panic on use).
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil)

	response, err := svc.InviteAdmin(context.Background(), uuid.New(), "new@example.com", "Имя")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.EmailSent {
		t.Fatal("expected emailSent=false when mailer is not configured")
	}
}
