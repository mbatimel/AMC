// back/admin/internal/service/service_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	authTransport "github.com/mbatimel/AMC/auth/pkg/client/transport"
	"github.com/valyala/fasthttp"
)

type fakeStorage struct {
	inserted            []auditCall
	entries             []postgres.AuditLogEntry
	total               int
	listErr             error
	listBannersFn       func(context.Context, bool) ([]postgres.Banner, int, error)
	getBannerFn         func(context.Context, uuid.UUID) (postgres.Banner, error)
	createBannerFn      func(context.Context, postgres.Banner) (postgres.Banner, error)
	updateBannerFn      func(context.Context, postgres.Banner, bool) (postgres.Banner, error)
	deleteBannerFn      func(context.Context, uuid.UUID) error
	updateBannerDelayFn func(context.Context, int) error
	decideSignupFn      func(context.Context, uuid.UUID, string, string) (postgres.SignupRequest, error)
}

type auditCall struct {
	actorUserID uuid.UUID
	actorLabel  string
	action      string
}

func (f *fakeStorage) InsertAuditLogEntry(_ context.Context, actorUserID uuid.UUID, actorLabel string, action string) error {
	f.inserted = append(f.inserted, auditCall{actorUserID, actorLabel, action})
	return nil
}

func (f *fakeStorage) ListAuditLogEntries(_ context.Context, _ int, _ int) ([]postgres.AuditLogEntry, error) {
	return f.entries, f.listErr
}

func (f *fakeStorage) CountAuditLogEntries(_ context.Context) (int, error) {
	return f.total, nil
}

func (f *fakeStorage) ListBanners(ctx context.Context, visibleOnly bool) ([]postgres.Banner, int, error) {
	if f.listBannersFn != nil {
		return f.listBannersFn(ctx, visibleOnly)
	}
	return []postgres.Banner{}, 6, nil
}

func (f *fakeStorage) GetBanner(ctx context.Context, bannerID uuid.UUID) (postgres.Banner, error) {
	if f.getBannerFn != nil {
		return f.getBannerFn(ctx, bannerID)
	}
	return postgres.Banner{ID: bannerID}, nil
}

func (f *fakeStorage) CreateBanner(ctx context.Context, banner postgres.Banner) (postgres.Banner, error) {
	if f.createBannerFn != nil {
		return f.createBannerFn(ctx, banner)
	}
	return banner, nil
}

func (f *fakeStorage) UpdateBanner(ctx context.Context, banner postgres.Banner, replaceImage bool) (postgres.Banner, error) {
	if f.updateBannerFn != nil {
		return f.updateBannerFn(ctx, banner, replaceImage)
	}
	return banner, nil
}

func (f *fakeStorage) DeleteBanner(ctx context.Context, bannerID uuid.UUID) error {
	if f.deleteBannerFn != nil {
		return f.deleteBannerFn(ctx, bannerID)
	}
	return nil
}

func (f *fakeStorage) UpdateBannerDelay(ctx context.Context, delaySec int) error {
	if f.updateBannerDelayFn != nil {
		return f.updateBannerDelayFn(ctx, delaySec)
	}
	return nil
}

func (f *fakeStorage) CreateSignupRequest(_ context.Context, request postgres.SignupRequest) (postgres.SignupRequest, error) {
	return request, nil
}

func (f *fakeStorage) ListSignupRequests(_ context.Context, _ string) ([]postgres.SignupRequest, error) {
	return []postgres.SignupRequest{}, nil
}

func (f *fakeStorage) DecideSignupRequest(ctx context.Context, id uuid.UUID, status string, rejectReason string) (postgres.SignupRequest, error) {
	if f.decideSignupFn != nil {
		return f.decideSignupFn(ctx, id, status, rejectReason)
	}
	return postgres.SignupRequest{ID: id, Email: "applicant@example.com", Status: status, RejectReason: rejectReason}, nil
}

type fakeAuthClient struct {
	userID uuid.UUID
	err    error
}

func (f *fakeAuthClient) LoginUser(_ context.Context, _ string, _ string) (uuid.UUID, error) {
	return f.userID, f.err
}

type fakeAccessClient struct {
	allowed bool
	err     error
}

func (f *fakeAccessClient) CheckAccess(_ context.Context, _ uuid.UUID, _ int) (bool, error) {
	return f.allowed, f.err
}

func TestLogin_Success(t *testing.T) {
	userID := uuid.New()
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{userID: userID}, &fakeAccessClient{allowed: true}, nil, nil)

	resp, err := svc.Login(context.Background(), "admin@volzhsky-instrument.ru", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if resp.UserID != userID {
		t.Fatalf("Login() UserID = %v, want %v", resp.UserID, userID)
	}
	if resp.Role != "admin" {
		t.Fatalf("Login() Role = %q, want %q", resp.Role, "admin")
	}
	if len(storage.inserted) != 1 || storage.inserted[0].action != "Выполнен вход в систему" {
		t.Fatalf("Login() did not write an audit log entry, got %+v", storage.inserted)
	}
}

func TestLogin_NotAdmin_Forbidden(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{userID: uuid.New()}, &fakeAccessClient{allowed: false}, nil, nil)

	_, err := svc.Login(context.Background(), "buyer@example.com", "secret")
	if err == nil {
		t.Fatal("Login() error = nil, want forbidden error")
	}
	if len(storage.inserted) != 0 {
		t.Fatalf("Login() wrote an audit log entry for a rejected login: %+v", storage.inserted)
	}
}

func TestLogin_AuthClientError(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{err: errors.New("boom")}, &fakeAccessClient{allowed: true}, nil, nil)

	_, err := svc.Login(context.Background(), "admin@volzhsky-instrument.ru", "wrong")
	if err == nil {
		t.Fatal("Login() error = nil, want an error when the auth client fails")
	}

	var customErr *customErrors.Error
	if !errors.As(err, &customErr) {
		t.Fatalf("Login() error = %v, want a *customErrors.Error", err)
	}
	if customErr.GetStatusCode() != fasthttp.StatusInternalServerError {
		t.Fatalf("Login() status code = %d, want %d (network/other errors must still map to 500)", customErr.GetStatusCode(), fasthttp.StatusInternalServerError)
	}
}

func TestLogin_AuthClientInvalidCredentials_Returns401(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{err: &authTransport.LoginError{StatusCode: fasthttp.StatusUnauthorized}}, &fakeAccessClient{allowed: true}, nil, nil)

	_, err := svc.Login(context.Background(), "admin@volzhsky-instrument.ru", "wrong")
	if err == nil {
		t.Fatal("Login() error = nil, want an invalid-credentials error")
	}

	var customErr *customErrors.Error
	if !errors.As(err, &customErr) {
		t.Fatalf("Login() error = %v, want a *customErrors.Error", err)
	}
	if customErr.GetStatusCode() != fasthttp.StatusUnauthorized {
		t.Fatalf("Login() status code = %d, want %d", customErr.GetStatusCode(), fasthttp.StatusUnauthorized)
	}
	if len(storage.inserted) != 0 {
		t.Fatalf("Login() wrote an audit log entry for a rejected login: %+v", storage.inserted)
	}
}

func TestLogin_AuthClientOtherLoginError_Returns500(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{err: &authTransport.LoginError{StatusCode: fasthttp.StatusInternalServerError}}, &fakeAccessClient{allowed: true}, nil, nil)

	_, err := svc.Login(context.Background(), "admin@volzhsky-instrument.ru", "wrong")
	if err == nil {
		t.Fatal("Login() error = nil, want an error when auth returns a 500")
	}

	var customErr *customErrors.Error
	if !errors.As(err, &customErr) {
		t.Fatalf("Login() error = %v, want a *customErrors.Error", err)
	}
	if customErr.GetStatusCode() != fasthttp.StatusInternalServerError {
		t.Fatalf("Login() status code = %d, want %d (non-401 LoginError must still map to 500)", customErr.GetStatusCode(), fasthttp.StatusInternalServerError)
	}
}

func TestGetSession_RequiresAdminRole(t *testing.T) {
	svc := NewAdminApiService(zerolog.Nop(), &fakeStorage{}, &fakeAuthClient{}, &fakeAccessClient{allowed: false}, nil, nil)

	_, err := svc.GetSession(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetSession() error = nil, want forbidden error for a non-admin user")
	}
}

func TestListAuditLog_ReturnsItemsAndTotal(t *testing.T) {
	entry := postgres.AuditLogEntry{ID: uuid.New(), ActorUserID: uuid.New(), ActorLabel: "Админ портала", Action: "Портал запущен"}
	storage := &fakeStorage{entries: []postgres.AuditLogEntry{entry}, total: 1}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, nil)

	resp, err := svc.ListAuditLog(context.Background(), uuid.New(), 10, 0)
	if err != nil {
		t.Fatalf("ListAuditLog() error = %v, want nil", err)
	}
	if resp.Total != 1 {
		t.Fatalf("ListAuditLog() Total = %d, want 1", resp.Total)
	}
	if len(resp.Items) != 1 || resp.Items[0].Action != "Портал запущен" {
		t.Fatalf("ListAuditLog() Items = %+v, want the entry from storage", resp.Items)
	}
}

func TestListAuditLog_ClampsLimit(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, nil, nil)

	if _, err := svc.ListAuditLog(context.Background(), uuid.New(), 0, -5); err != nil {
		t.Fatalf("ListAuditLog() error = %v, want nil", err)
	}
}
