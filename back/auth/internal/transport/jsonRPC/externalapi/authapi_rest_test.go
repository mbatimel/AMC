package externalapi

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type recordingAuthAPI struct {
	loginEmail    string
	loginPassword string

	registerIPEmail    string
	registerIPPassword string
	registerIPFullName *string

	registerIndividualFIO             string
	registerIndividualPhone           string
	registerIndividualEmail           string
	registerIndividualDeliveryAddress string
	registerIndividualPassword        string
	registerIndividualCity            string
	registerIndividualINN             *string
}

func (s *recordingAuthAPI) LoginUser(_ context.Context, email string, password string) (uuid.UUID, error) {
	s.loginEmail = email
	s.loginPassword = password
	return uuid.MustParse("72514773-043c-40c0-ae6a-e1bcab0cda0a"), nil
}

func (s *recordingAuthAPI) RegisterIP(
	_ context.Context,
	email string,
	password string,
	fullName, _, _, _, _, _, _, _, _,
	_, _, _, _, _,
	_, _, _, _ *string,
) (uuid.UUID, error) {
	s.registerIPEmail = email
	s.registerIPPassword = password
	s.registerIPFullName = fullName
	return uuid.New(), nil
}

func (s *recordingAuthAPI) RegisterIndividual(
	_ context.Context,
	fio string,
	phone string,
	email string,
	deliveryAddress string,
	password string,
	city string,
	inn *string,
) (uuid.UUID, error) {
	s.registerIndividualFIO = fio
	s.registerIndividualPhone = phone
	s.registerIndividualEmail = email
	s.registerIndividualDeliveryAddress = deliveryAddress
	s.registerIndividualPassword = password
	s.registerIndividualCity = city
	s.registerIndividualINN = inn
	return uuid.New(), nil
}

func (*recordingAuthAPI) LogoutUser(context.Context, uuid.UUID) error {
	return nil
}

func (*recordingAuthAPI) ChangePassword(context.Context, uuid.UUID, string, string) error {
	return nil
}

func (*recordingAuthAPI) VerifyEmailCode(context.Context, uuid.UUID, int64) error {
	return nil
}

func (*recordingAuthAPI) SendEmailVerification(context.Context, uuid.UUID) error {
	return nil
}

func newTestAuthApp(svc *recordingAuthAPI) *fiber.App {
	app := fiber.New()
	NewAuthAPI(svc).SetRoutes(app)
	return app
}

func performJSONRequest(t *testing.T, app *fiber.App, path string, body string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	t.Cleanup(func() {
		_ = response.Body.Close()
	})
	return response
}

func TestLoginUserReadsJSONBody(t *testing.T) {
	svc := &recordingAuthAPI{}
	response := performJSONRequest(
		t,
		newTestAuthApp(svc),
		"/api/v1/auth/login",
		`{"email":"user@example.com","password":"correct-password"}`,
	)

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %q", response.StatusCode, body)
	}
	if svc.loginEmail != "user@example.com" {
		t.Fatalf("email = %q", svc.loginEmail)
	}
	if svc.loginPassword != "correct-password" {
		t.Fatalf("password = %q", svc.loginPassword)
	}
}

func TestLoginUserKeepsQueryCompatibility(t *testing.T) {
	svc := &recordingAuthAPI{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login?email=legacy%40example.com&password=legacy-password",
		nil,
	)
	response, err := newTestAuthApp(svc).Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %q", response.StatusCode, body)
	}
	if svc.loginEmail != "legacy@example.com" || svc.loginPassword != "legacy-password" {
		t.Fatalf("credentials = (%q, %q)", svc.loginEmail, svc.loginPassword)
	}
}

func TestRegisterIPReadsJSONBody(t *testing.T) {
	svc := &recordingAuthAPI{}
	response := performJSONRequest(
		t,
		newTestAuthApp(svc),
		"/api/v1/auth/register/ip",
		`{"email":"ip@example.com","password":"secret","fullName":"ИП Тест"}`,
	)

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %q", response.StatusCode, body)
	}
	if svc.registerIPEmail != "ip@example.com" || svc.registerIPPassword != "secret" {
		t.Fatalf("credentials = (%q, %q)", svc.registerIPEmail, svc.registerIPPassword)
	}
	if svc.registerIPFullName == nil || *svc.registerIPFullName != "ИП Тест" {
		t.Fatalf("fullName = %v", svc.registerIPFullName)
	}
}

func TestRegisterIndividualReadsJSONBody(t *testing.T) {
	svc := &recordingAuthAPI{}
	response := performJSONRequest(
		t,
		newTestAuthApp(svc),
		"/api/v1/auth/register/individual",
		`{"fio":"Иванов Иван","phone":"+79990000000","email":"person@example.com","deliveryAddress":"Москва","password":"secret","city":"Москва","inn":"123456789012"}`,
	)

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %q", response.StatusCode, body)
	}
	if svc.registerIndividualFIO != "Иванов Иван" ||
		svc.registerIndividualPhone != "+79990000000" ||
		svc.registerIndividualEmail != "person@example.com" ||
		svc.registerIndividualDeliveryAddress != "Москва" ||
		svc.registerIndividualPassword != "secret" ||
		svc.registerIndividualCity != "Москва" {
		t.Fatalf("unexpected registration payload: %+v", svc)
	}
	if svc.registerIndividualINN == nil || *svc.registerIndividualINN != "123456789012" {
		t.Fatalf("inn = %v", svc.registerIndividualINN)
	}
}

func TestLoginUserRejectsMalformedJSON(t *testing.T) {
	svc := &recordingAuthAPI{}
	response := performJSONRequest(t, newTestAuthApp(svc), "/api/v1/auth/login", `{"email":`)

	if response.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, body = %q", response.StatusCode, body)
	}
	if svc.loginEmail != "" || svc.loginPassword != "" {
		t.Fatal("service must not be called for malformed JSON")
	}
}
