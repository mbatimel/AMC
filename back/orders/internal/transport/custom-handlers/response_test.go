package custom_handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	customErrors "github.com/mbatimel/AMC/orders/internal/errors"
	"github.com/rs/zerolog"
)

func responseForError(t *testing.T, responseErr error) (int, RestResponse) {
	t.Helper()
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		sendResponse(ctx, zerolog.Nop(), nil, responseErr)
		return nil
	})

	request, err := http.NewRequest(http.MethodGet, "http://localhost/", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()

	var body RestResponse
	if err = json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response.StatusCode, body
}

func TestSendResponseMapsValidationError(t *testing.T) {
	parseErr := errors.New("invalid UUID details")
	responseErr := customErrors.BadRequestError().
		SetOuterError(parseErr).
		AddCause("field", "clientID")

	status, body := responseForError(t, fmt.Errorf("handler: %w", responseErr))
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	if body.ErrorText != customErrors.ErrBadRequest {
		t.Fatalf("errorText = %q, want %q", body.ErrorText, customErrors.ErrBadRequest)
	}
	if body.AdditionalErrors["field"] != "clientID" {
		t.Fatalf("additionalErrors = %#v", body.AdditionalErrors)
	}
	if _, leaked := body.AdditionalErrors["reason"]; leaked {
		t.Fatalf("technical reason leaked: %#v", body.AdditionalErrors)
	}
}

func TestSendResponseKeepsInternalKeyForSystemError(t *testing.T) {
	responseErr := customErrors.InternalServerError().SetOuterError(errors.New("database unavailable"))

	status, body := responseForError(t, responseErr)
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if body.ErrorText != errInternal {
		t.Fatalf("errorText = %q, want %q", body.ErrorText, errInternal)
	}
	if body.AdditionalErrors != nil {
		t.Fatalf("additionalErrors = %#v, want nil", body.AdditionalErrors)
	}
}
