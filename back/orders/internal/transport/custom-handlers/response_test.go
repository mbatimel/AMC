package custom_handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	customErrors "github.com/mbatimel/AMC/orders/internal/errors"
	"github.com/mbatimel/AMC/orders/pkg/models"
	"github.com/rs/zerolog"
)

func testResponse(t *testing.T, data interface{}, responseErr error) (int, RestResponse) {
	t.Helper()
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		sendResponse(ctx, zerolog.Nop(), data, responseErr)
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

	status, body := testResponse(t, nil, fmt.Errorf("handler: %w", responseErr))
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

	status, body := testResponse(t, nil, responseErr)
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

func TestSendResponseSerializesEmptyCollections(t *testing.T) {
	tests := []struct {
		name       string
		data       interface{}
		collection func(map[string]interface{}) interface{}
	}{
		{
			name: "cart items",
			data: models.GetCartResponse{Cart: models.Cart{
				Items: make([]models.CartItem, 0),
			}},
			collection: func(data map[string]interface{}) interface{} {
				return data["cart"].(map[string]interface{})["items"]
			},
		},
		{
			name: "orders",
			data: models.ListOrdersResponse{
				Items: make([]models.Order, 0),
				Pagination: models.Pagination{
					Limit: 20,
				},
			},
			collection: func(data map[string]interface{}) interface{} {
				return data["items"]
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := testResponse(t, tt.data, nil)
			if status != http.StatusOK || body.Error {
				t.Fatalf("status = %d, error = %v", status, body.Error)
			}
			if body.AdditionalErrors == nil || len(body.AdditionalErrors) != 0 {
				t.Fatalf("additionalErrors = %#v, want {}", body.AdditionalErrors)
			}
			data, ok := body.Data.(map[string]interface{})
			if !ok {
				t.Fatalf("data type = %T", body.Data)
			}
			items, ok := tt.collection(data).([]interface{})
			if !ok || len(items) != 0 {
				t.Fatalf("collection = %#v, want []", tt.collection(data))
			}
		})
	}
}
