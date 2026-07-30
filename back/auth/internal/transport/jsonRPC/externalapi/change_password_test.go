package externalapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type changePasswordRecorder struct {
	userID      uuid.UUID
	oldPassword string
	newPassword string
	called      bool
}

func (*changePasswordRecorder) LoginUser(context.Context, string, string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (*changePasswordRecorder) RegisterIP(
	context.Context,
	string, string,
	*string, *string, *string, *string, *string, *string, *string, *string, *string,
	*string, *string, *string, *string, *string,
	*string, *string, *string, *string,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (*changePasswordRecorder) RegisterIndividual(
	context.Context,
	string, string, string, string, string, string,
	*string,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (*changePasswordRecorder) LogoutUser(context.Context, uuid.UUID) error {
	return nil
}

func (s *changePasswordRecorder) ChangePassword(
	_ context.Context,
	userID uuid.UUID,
	oldPassword string,
	newPassword string,
) error {
	s.userID = userID
	s.oldPassword = oldPassword
	s.newPassword = newPassword
	s.called = true
	return nil
}

func (*changePasswordRecorder) VerifyEmailCode(context.Context, uuid.UUID, int64) error {
	return nil
}

func (*changePasswordRecorder) SendEmailVerification(context.Context, uuid.UUID) error {
	return nil
}

func changePasswordTestApp(svc *changePasswordRecorder) *fiber.App {
	app := fiber.New()
	NewAuthAPI(svc).SetRoutes(app)
	return app
}

func TestChangePasswordAcceptsFrontendQueryContract(t *testing.T) {
	userID := uuid.New()
	svc := &changePasswordRecorder{}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/change-password?oldPassword=current-password&newPassword=new-password",
		nil,
	)
	request.Header.Set("X-User-Id", userID.String())

	response, err := changePasswordTestApp(svc).Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if !svc.called ||
		svc.userID != userID ||
		svc.oldPassword != "current-password" ||
		svc.newPassword != "new-password" {
		t.Fatalf("unexpected service call: %+v", svc)
	}
}

func TestChangePasswordAlsoAcceptsJSONBody(t *testing.T) {
	userID := uuid.New()
	svc := &changePasswordRecorder{}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/change-password",
		strings.NewReader(`{"oldPassword":"current-password","newPassword":"new-password"}`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	request.Header.Set("X-User-Id", userID.String())

	response, err := changePasswordTestApp(svc).Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if !svc.called ||
		svc.userID != userID ||
		svc.oldPassword != "current-password" ||
		svc.newPassword != "new-password" {
		t.Fatalf("unexpected service call: %+v", svc)
	}
}

func TestChangePasswordRejectsMalformedJSON(t *testing.T) {
	svc := &changePasswordRecorder{}
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/change-password",
		strings.NewReader(`{"oldPassword":`),
	)
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	response, err := changePasswordTestApp(svc).Test(request)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if svc.called {
		t.Fatal("service must not be called for malformed JSON")
	}
}
