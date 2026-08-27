package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/integrations/internal/errors"
	"github.com/mbatimel/AMC/integrations/internal/transport/jsonRPC/internalapi"
)

type fakeWebhookService struct {
	ok  bool
	err error

	gotAPIKey             string
	gotClientOrderID      uuid.UUID
	gotStatus             string
	gotOnecDocumentNumber string
	gotComment            string
	calls                 int
}

func (f *fakeWebhookService) OnecOrderStatusWebhook(_ context.Context, apiKey string, clientOrderID uuid.UUID, status, onecDocumentNumber, comment string) (bool, error) {
	f.calls++
	f.gotAPIKey = apiKey
	f.gotClientOrderID = clientOrderID
	f.gotStatus = status
	f.gotOnecDocumentNumber = onecDocumentNumber
	f.gotComment = comment
	return f.ok, f.err
}

// contractBody is the exact JSON shape the 1С contract documents
// (docs/superpowers/specs/2026-08-26-onec-orders-integration-1c-contract.md).
const contractBody = `{
  "client_order_id": "b7e6c2b0-1234-4d9a-9a0e-8f1a2b3c4d5e",
  "status": "delivered",
  "onec_document_number": "УТ-00456",
  "comment": "Вручено получателю 26.08.2026"
}`

var contractOrderID = uuid.MustParse("b7e6c2b0-1234-4d9a-9a0e-8f1a2b3c4d5e")

func postWebhook(t *testing.T, app *fiber.App, body string, apiKey string) (int, string) {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/v1/onec/orders/status", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set(WebhookAPIKeyHeader, apiKey)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, strings.TrimSpace(string(raw))
}

func TestWebhookRoute_Success_BareOkTrue(t *testing.T) {
	app := fiber.New()
	svc := &fakeWebhookService{ok: true}
	RegisterWebhookRoute(app, svc, zerolog.Nop())

	status, body := postWebhook(t, app, contractBody, "secret-key")

	if status != 200 {
		t.Fatalf("expected 200, got %d (body %s)", status, body)
	}

	// The contract promises a BARE {"ok": true}, not the RestResponse envelope.
	var decoded map[string]any
	if err := json.Unmarshal([]byte(body), &decoded); err != nil {
		t.Fatalf("unmarshal response: %v (body %s)", err, body)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected exactly one field {\"ok\":true}, got %v", decoded)
	}
	if decoded["ok"] != true {
		t.Fatalf("expected ok=true, got %v", decoded)
	}
	for _, envelopeField := range []string{"data", "error", "errorText", "additionalErrors"} {
		if _, present := decoded[envelopeField]; present {
			t.Fatalf("response is wrapped in the RestResponse envelope (field %q present): %s", envelopeField, body)
		}
	}
}

func TestWebhookRoute_ParsesContractJSONBodyFields(t *testing.T) {
	app := fiber.New()
	svc := &fakeWebhookService{ok: true}
	RegisterWebhookRoute(app, svc, zerolog.Nop())

	if status, body := postWebhook(t, app, contractBody, "secret-key"); status != 200 {
		t.Fatalf("expected 200, got %d (%s)", status, body)
	}

	if svc.gotClientOrderID != contractOrderID {
		t.Fatalf("client_order_id not read from JSON body: got %s", svc.gotClientOrderID)
	}
	if svc.gotStatus != "delivered" {
		t.Fatalf("status not read from JSON body: got %q", svc.gotStatus)
	}
	if svc.gotOnecDocumentNumber != "УТ-00456" {
		t.Fatalf("onec_document_number not read from JSON body: got %q", svc.gotOnecDocumentNumber)
	}
	if svc.gotComment != "Вручено получателю 26.08.2026" {
		t.Fatalf("comment not read from JSON body: got %q", svc.gotComment)
	}
	if svc.gotAPIKey != "secret-key" {
		t.Fatalf("X-Onec-Api-Key not read from header: got %q", svc.gotAPIKey)
	}
}

func TestWebhookRoute_QueryParamsAreIgnored(t *testing.T) {
	// Guards against silently falling back to the tg-generated query-string
	// route: the body must win, and query params must not be read at all.
	app := fiber.New()
	svc := &fakeWebhookService{ok: true}
	RegisterWebhookRoute(app, svc, zerolog.Nop())

	req := httptest.NewRequest("POST",
		"/api/v1/onec/orders/status?clientOrderID="+uuid.New().String()+"&status=cancelled",
		bytes.NewReader([]byte(contractBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(WebhookAPIKeyHeader, "secret-key")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if svc.gotClientOrderID != contractOrderID || svc.gotStatus != "delivered" {
		t.Fatalf("query params leaked into the request: id=%s status=%q", svc.gotClientOrderID, svc.gotStatus)
	}
}

func TestWebhookRoute_ErrorStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		svcErr     error
		wantStatus int
	}{
		{"bad request", customErrors.BadRequestError().AddCause("field", "status"), 400},
		{"unauthorized", customErrors.UnauthorizedError(), 401},
		{"order not found", customErrors.NotFoundError(), 404},
		{"already cancelled", customErrors.ConflictError().AddCause("field", "status"), 409},
		{"internal", customErrors.InternalServerError().SetOuterError(errors.New("orders down")), 500},
		{"untyped error", errors.New("boom"), 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			RegisterWebhookRoute(app, &fakeWebhookService{err: tt.svcErr}, zerolog.Nop())

			status, body := postWebhook(t, app, contractBody, "secret-key")
			if status != tt.wantStatus {
				t.Fatalf("expected %d, got %d (body %s)", tt.wantStatus, status, body)
			}

			var decoded map[string]any
			if err := json.Unmarshal([]byte(body), &decoded); err != nil {
				t.Fatalf("unmarshal error response: %v (body %s)", err, body)
			}
			if _, present := decoded["additionalErrors"]; present {
				t.Fatalf("error response is wrapped in the RestResponse envelope: %s", body)
			}
			if decoded["ok"] != false {
				t.Fatalf("expected ok=false in error response, got %s", body)
			}
		})
	}
}

func TestWebhookRoute_MalformedBody_Returns400(t *testing.T) {
	app := fiber.New()
	svc := &fakeWebhookService{ok: true}
	RegisterWebhookRoute(app, svc, zerolog.Nop())

	if status, _ := postWebhook(t, app, "not json", "secret-key"); status != 400 {
		t.Fatalf("expected 400 for malformed body, got %d", status)
	}
	if svc.calls != 0 {
		t.Fatalf("service must not be called for a malformed body, calls=%d", svc.calls)
	}
}

func TestWebhookRoute_MissingClientOrderID_Returns400(t *testing.T) {
	app := fiber.New()
	svc := &fakeWebhookService{ok: true}
	RegisterWebhookRoute(app, svc, zerolog.Nop())

	if status, _ := postWebhook(t, app, `{"status":"delivered"}`, "secret-key"); status != 400 {
		t.Fatalf("expected 400 for missing client_order_id, got %d", status)
	}
	if svc.calls != 0 {
		t.Fatalf("service must not be called without client_order_id, calls=%d", svc.calls)
	}
}

func TestWebhookRoute_MissingAPIKeyHeader_ReachesServiceWithEmptyKey(t *testing.T) {
	// Auth is the service's decision (it compares against the configured key),
	// so the route must forward an absent header as an empty string rather than
	// short-circuiting with its own 401.
	app := fiber.New()
	svc := &fakeWebhookService{err: customErrors.UnauthorizedError()}
	RegisterWebhookRoute(app, svc, zerolog.Nop())

	status, _ := postWebhook(t, app, contractBody, "")
	if status != 401 {
		t.Fatalf("expected 401, got %d", status)
	}
	if svc.gotAPIKey != "" {
		t.Fatalf("expected empty api key, got %q", svc.gotAPIKey)
	}
}

// TestWebhookRoute_OnInternalAPIServer replicates cmd/onec-orders-api's wiring:
// the fiber app comes from internalapi.New (middleware only, no typed service
// registered) and the webhook route is hand-registered on it. It proves a JSON
// POST at the real path reaches the hand-written handler, not the tg-generated
// query-string one.
func TestWebhookRoute_OnInternalAPIServer(t *testing.T) {
	srv := internalapi.New(zerolog.Nop()).WithLog().WithMetrics()
	svc := &fakeWebhookService{ok: true}
	RegisterWebhookRoute(srv.Fiber(), svc, zerolog.Nop())

	status, body := postWebhook(t, srv.Fiber(), contractBody, "secret-key")
	if status != 200 {
		t.Fatalf("expected 200, got %d (body %s)", status, body)
	}
	if body != `{"ok":true}` {
		t.Fatalf("expected bare {\"ok\":true}, got %s", body)
	}
	if svc.calls != 1 || svc.gotClientOrderID != contractOrderID {
		t.Fatalf("request did not reach the hand-written handler: calls=%d id=%s", svc.calls, svc.gotClientOrderID)
	}
}
