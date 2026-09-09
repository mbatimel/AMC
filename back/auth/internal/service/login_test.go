package service

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"
	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return string(hash)
}

func TestLoginUser_BlockedUser_RejectsBeforePasswordCheck(t *testing.T) {
	userID := uuid.New()
	storage := &stubStorage{getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
		return postgres.User{
			ID:                  userID,
			Email:               "blocked@example.com",
			Password:            hashPassword(t, "correct-password"),
			IsActive:            false,
			BlockedReason:       sql.NullString{String: "Нарушение условий работы", Valid: true},
			BlockedContactName:  sql.NullString{String: "Иван Иванов", Valid: true},
			BlockedContactPhone: sql.NullString{String: "+79991234567", Valid: true},
			BlockedContactEmail: sql.NullString{String: "manager@company.ru", Valid: true},
		}, nil
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	// Even a WRONG password must still surface "blocked", not "invalid credentials" —
	// blocked status is reported unconditionally (see docs/superpowers/specs).
	_, err := svc.LoginUser(context.Background(), "blocked@example.com", "wrong-password")
	if err == nil {
		t.Fatal("expected an error for a blocked user")
	}
	var custErr *customErrors.Error
	if !errAs(err, &custErr) {
		t.Fatalf("expected *customErrors.Error, got %T: %v", err, err)
	}
	if custErr.Cause["reason"] != "Нарушение условий работы" {
		t.Fatalf("expected reason in cause, got %+v", custErr.Cause)
	}
	if custErr.Cause["contactName"] != "Иван Иванов" ||
		custErr.Cause["contactPhone"] != "+79991234567" ||
		custErr.Cause["contactEmail"] != "manager@company.ru" {
		t.Fatalf("expected contact fields in cause, got %+v", custErr.Cause)
	}
}

func TestLoginUser_BlockedUser_NoStoredReason_OmitsCause(t *testing.T) {
	storage := &stubStorage{getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
		return postgres.User{ID: uuid.New(), Email: "blocked@example.com", IsActive: false}, nil
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	_, err := svc.LoginUser(context.Background(), "blocked@example.com", "any-password")
	var custErr *customErrors.Error
	if !errAs(err, &custErr) {
		t.Fatalf("expected *customErrors.Error, got %T: %v", err, err)
	}
	if len(custErr.Cause) != 0 {
		t.Fatalf("expected no cause fields when nothing was stored, got %+v", custErr.Cause)
	}
}

func TestLoginUser_ActiveUser_CorrectPassword_Succeeds(t *testing.T) {
	userID := uuid.New()
	storage := &stubStorage{getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
		return postgres.User{ID: userID, Email: "user@example.com", Password: hashPassword(t, "correct-password"), IsActive: true}, nil
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	gotID, err := svc.LoginUser(context.Background(), "user@example.com", "correct-password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != userID {
		t.Fatalf("expected userID %v, got %v", userID, gotID)
	}
}

func TestLoginUser_ActiveUser_WrongPassword_InvalidCredentials(t *testing.T) {
	storage := &stubStorage{getUserByEmailFn: func(context.Context, string) (postgres.User, error) {
		return postgres.User{ID: uuid.New(), Email: "user@example.com", Password: hashPassword(t, "correct-password"), IsActive: true}, nil
	}}
	svc := &service{logger: zerolog.Nop(), storage: storage}

	_, err := svc.LoginUser(context.Background(), "user@example.com", "wrong-password")
	if err == nil {
		t.Fatal("expected invalid credentials error")
	}
}

func errAs(err error, target **customErrors.Error) bool {
	custErr, ok := err.(*customErrors.Error)
	if !ok {
		return false
	}
	*target = custErr
	return true
}
