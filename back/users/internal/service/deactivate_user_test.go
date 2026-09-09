package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	internalModels "github.com/mbatimel/AMC/users/internal/models"
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

func TestDeactivateUser_StoresReasonAndSendsEmail(t *testing.T) {
	var gotReason, gotContactName, gotContactPhone, gotContactEmail string
	storage := &fakeStorage{deactivateUserFn: func(_ context.Context, _ uuid.UUID, reason, contactName, contactPhone, contactEmail string) (internalModels.User, error) {
		gotReason, gotContactName, gotContactPhone, gotContactEmail = reason, contactName, contactPhone, contactEmail
		return internalModels.User{ID: uuid.New(), Email: "blocked@example.com", Status: "inactive"}, nil
	}}
	mail := &fakeMailer{}
	svc := New(zerolog.Nop(), storage, &fakeAccessClient{}, mail)

	_, err := svc.DeactivateUser(context.Background(), uuid.New(), "Нарушение условий работы", "Иван Иванов", "+79991234567", "manager@company.ru")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotReason != "Нарушение условий работы" || gotContactName != "Иван Иванов" || gotContactPhone != "+79991234567" || gotContactEmail != "manager@company.ru" {
		t.Fatalf("storage did not receive the fields as-is: %q %q %q %q", gotReason, gotContactName, gotContactPhone, gotContactEmail)
	}
	if mail.calls != 1 {
		t.Fatalf("expected mailer.Send called once, got %d", mail.calls)
	}
	if mail.to != "blocked@example.com" {
		t.Fatalf("expected email to blocked@example.com, got %q", mail.to)
	}
	if !strings.Contains(mail.body, "Нарушение условий работы") || !strings.Contains(mail.body, "Иван Иванов") {
		t.Fatalf("expected email body to contain reason and contact name, got: %s", mail.body)
	}
}

func TestDeactivateUser_AllFieldsOptional(t *testing.T) {
	storage := &fakeStorage{deactivateUserFn: func(_ context.Context, _ uuid.UUID, reason, contactName, contactPhone, contactEmail string) (internalModels.User, error) {
		return internalModels.User{ID: uuid.New(), Email: "blocked@example.com", Status: "inactive"}, nil
	}}
	mail := &fakeMailer{}
	svc := New(zerolog.Nop(), storage, &fakeAccessClient{}, mail)

	_, err := svc.DeactivateUser(context.Background(), uuid.New(), "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mail.calls != 1 {
		t.Fatalf("expected mailer.Send still called with no fields set, got %d", mail.calls)
	}
}

func TestDeactivateUser_MailerFails_DoesNotFailRequest(t *testing.T) {
	storage := &fakeStorage{deactivateUserFn: func(context.Context, uuid.UUID, string, string, string, string) (internalModels.User, error) {
		return internalModels.User{ID: uuid.New(), Email: "blocked@example.com"}, nil
	}}
	mail := &fakeMailer{sendErr: errors.New("smtp down")}
	svc := New(zerolog.Nop(), storage, &fakeAccessClient{}, mail)

	_, err := svc.DeactivateUser(context.Background(), uuid.New(), "reason", "", "", "")
	if err != nil {
		t.Fatalf("mailer failure must not fail the request, got: %v", err)
	}
}

func TestDeactivateUser_MailerNil_DoesNotPanic(t *testing.T) {
	storage := &fakeStorage{deactivateUserFn: func(context.Context, uuid.UUID, string, string, string, string) (internalModels.User, error) {
		return internalModels.User{ID: uuid.New(), Email: "blocked@example.com"}, nil
	}}
	svc := New(zerolog.Nop(), storage, &fakeAccessClient{}, nil)

	_, err := svc.DeactivateUser(context.Background(), uuid.New(), "reason", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeactivateUser_RequiresUserID(t *testing.T) {
	svc := New(zerolog.Nop(), &fakeStorage{}, &fakeAccessClient{}, nil)

	_, err := svc.DeactivateUser(context.Background(), uuid.Nil, "reason", "", "", "")
	if err == nil {
		t.Fatal("expected validation error for empty userID")
	}
}

func TestActivateUser_ClearsBlockFields(t *testing.T) {
	var gotActive bool
	storage := &fakeStorage{setUserActiveFn: func(_ context.Context, _ uuid.UUID, active bool) (internalModels.User, error) {
		gotActive = active
		return internalModels.User{ID: uuid.New(), Status: "active"}, nil
	}}
	svc := New(zerolog.Nop(), storage, &fakeAccessClient{}, nil)

	_, err := svc.ActivateUser(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotActive {
		t.Fatal("expected ActivateUser to call SetUserActive(true)")
	}
}
