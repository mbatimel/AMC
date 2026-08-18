// back/admin/internal/service/signup_requests_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/valyala/fasthttp"
)

type fakeUsersClient struct {
	foundID   uuid.UUID
	found     bool
	findErr   error
	deleteErr error
	deletedID uuid.UUID
	deleted   bool
}

func (f *fakeUsersClient) FindUserIDByEmail(_ context.Context, _ string) (uuid.UUID, bool, error) {
	return f.foundID, f.found, f.findErr
}

func (f *fakeUsersClient) DeleteUser(_ context.Context, userID uuid.UUID) error {
	f.deletedID = userID
	f.deleted = f.deleteErr == nil
	return f.deleteErr
}

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

func TestRejectSignupRequest_DeletesAccountAndSendsEmail(t *testing.T) {
	requestID := uuid.New()
	accountID := uuid.New()
	storage := &fakeStorage{decideSignupFn: func(_ context.Context, id uuid.UUID, status string, reason string) (postgres.SignupRequest, error) {
		if id != requestID || status != signupStatusRejected {
			t.Fatalf("DecideSignupRequest called with id=%v status=%v, want id=%v status=%v", id, status, requestID, signupStatusRejected)
		}
		return postgres.SignupRequest{ID: id, Email: "applicant@example.com", Status: status, RejectReason: reason}, nil
	}}
	users := &fakeUsersClient{foundID: accountID, found: true}
	mail := &fakeMailer{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, users, mail)

	resp, err := svc.RejectSignupRequest(context.Background(), uuid.New(), requestID, "Некорректный ИНН")
	if err != nil {
		t.Fatalf("RejectSignupRequest() error = %v, want nil", err)
	}
	if resp.Status != signupStatusRejected {
		t.Fatalf("RejectSignupRequest() status = %q, want %q", resp.Status, signupStatusRejected)
	}
	if !users.deleted || users.deletedID != accountID {
		t.Fatalf("RejectSignupRequest() did not delete the matched account: deleted=%v id=%v, want true/%v", users.deleted, users.deletedID, accountID)
	}
	if mail.calls != 1 || mail.to != "applicant@example.com" {
		t.Fatalf("RejectSignupRequest() email calls = %d to=%q, want 1 call to applicant@example.com", mail.calls, mail.to)
	}
	if mail.body == "" || !containsReason(mail.body, "Некорректный ИНН") {
		t.Fatalf("RejectSignupRequest() email body = %q, want it to contain the rejection reason", mail.body)
	}
}

func TestRejectSignupRequest_NoMatchingAccount_StillSendsEmail(t *testing.T) {
	requestID := uuid.New()
	storage := &fakeStorage{decideSignupFn: func(_ context.Context, id uuid.UUID, status string, reason string) (postgres.SignupRequest, error) {
		return postgres.SignupRequest{ID: id, Email: "nobody@example.com", Status: status, RejectReason: reason}, nil
	}}
	users := &fakeUsersClient{found: false}
	mail := &fakeMailer{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, users, mail)

	if _, err := svc.RejectSignupRequest(context.Background(), uuid.New(), requestID, ""); err != nil {
		t.Fatalf("RejectSignupRequest() error = %v, want nil", err)
	}
	if users.deleted {
		t.Fatal("RejectSignupRequest() deleted an account that was never found")
	}
	if mail.calls != 1 {
		t.Fatalf("RejectSignupRequest() email calls = %d, want 1", mail.calls)
	}
}

func TestRejectSignupRequest_AlreadyDecided_ReturnsConflict(t *testing.T) {
	storage := &fakeStorage{decideSignupFn: func(context.Context, uuid.UUID, string, string) (postgres.SignupRequest, error) {
		return postgres.SignupRequest{}, postgres.ErrSignupRequestAlreadyDecided
	}}
	users := &fakeUsersClient{}
	mail := &fakeMailer{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, users, mail)

	_, err := svc.RejectSignupRequest(context.Background(), uuid.New(), uuid.New(), "too late")

	var customErr *customErrors.Error
	if !errors.As(err, &customErr) {
		t.Fatalf("RejectSignupRequest() error = %v, want a *customErrors.Error", err)
	}
	if customErr.GetStatusCode() != fasthttp.StatusConflict {
		t.Fatalf("RejectSignupRequest() status code = %d, want %d (guard must block re-deciding a request)", customErr.GetStatusCode(), fasthttp.StatusConflict)
	}
	if mail.calls != 0 || users.deleted {
		t.Fatal("RejectSignupRequest() must not delete accounts or send email when the guard rejects the decision")
	}
}

func TestApproveSignupRequest_AlreadyDecided_ReturnsConflict(t *testing.T) {
	storage := &fakeStorage{decideSignupFn: func(context.Context, uuid.UUID, string, string) (postgres.SignupRequest, error) {
		return postgres.SignupRequest{}, postgres.ErrSignupRequestAlreadyDecided
	}}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, nil)

	_, err := svc.ApproveSignupRequest(context.Background(), uuid.New(), uuid.New())

	var customErr *customErrors.Error
	if !errors.As(err, &customErr) {
		t.Fatalf("ApproveSignupRequest() error = %v, want a *customErrors.Error", err)
	}
	if customErr.GetStatusCode() != fasthttp.StatusConflict {
		t.Fatalf("ApproveSignupRequest() status code = %d, want %d", customErr.GetStatusCode(), fasthttp.StatusConflict)
	}
}

func TestRejectSignupRequest_RequiresAdminRole(t *testing.T) {
	svc := NewAdminApiService(zerolog.Nop(), &fakeStorage{}, &fakeAuthClient{}, &fakeAccessClient{allowed: false}, &fakeUsersClient{}, &fakeMailer{})

	if _, err := svc.RejectSignupRequest(context.Background(), uuid.New(), uuid.New(), "reason"); err == nil {
		t.Fatal("RejectSignupRequest() error = nil, want forbidden error for a non-admin user")
	}
}

func TestCreateSignupRequest_NormalizesAndValidates(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, nil)

	resp, err := svc.CreateSignupRequest(context.Background(), "ООО Ромашка", "", "", "  Applicant@Example.com  ", "", "individual")
	if err != nil {
		t.Fatalf("CreateSignupRequest() error = %v, want nil", err)
	}
	if resp.Email != "applicant@example.com" {
		t.Fatalf("CreateSignupRequest() email = %q, want lowercase-trimmed", resp.Email)
	}
	if resp.Type != signupTypeIndividual {
		t.Fatalf("CreateSignupRequest() type = %q, want %q", resp.Type, signupTypeIndividual)
	}

	if _, err = svc.CreateSignupRequest(context.Background(), "", "", "", "", "", ""); err == nil {
		t.Fatal("CreateSignupRequest() error = nil, want validation error for empty email")
	}
}

func containsReason(body string, reason string) bool {
	return len(body) >= len(reason) && (func() bool {
		for i := 0; i+len(reason) <= len(body); i++ {
			if body[i:i+len(reason)] == reason {
				return true
			}
		}
		return false
	})()
}
