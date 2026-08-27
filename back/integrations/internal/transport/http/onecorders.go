package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/integrations/internal/errors"
	"github.com/mbatimel/AMC/integrations/internal/onecorders"
)

// OnecPusher pushes an order to 1С and returns the resulting document
// identifiers. It is implemented by onecorders.Client.
type OnecPusher interface {
	PushOrder(ctx context.Context, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error)
}

// RegisterPushOrderRoute registers a hand-written POST /api/v1/onec-orders/push
// route on app. This route is registered directly on *fiber.App rather than
// through the tg code generator because the request body carries an array of
// order line items — a shape the generator can't express.
func RegisterPushOrderRoute(app *fiber.App, pusher OnecPusher, logger zerolog.Logger) {
	app.Post("/api/v1/onec-orders/push", func(ctx *fiber.Ctx) error {
		var req onecorders.PushOrderRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
		}

		result, err := pusher.PushOrder(ctx.UserContext(), req)
		if err != nil {
			logger.Error().Str("clientOrderID", req.ClientOrderID.String()).Err(err).Msg("push order to onec failed")
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to push order to 1С"})
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"onec_document_guid":   result.OnecDocumentGUID,
			"onec_document_number": result.OnecDocumentNumber,
		})
	})
}

// WebhookAPIKeyHeader is the header 1С sends the static webhook key in.
const WebhookAPIKeyHeader = "X-Onec-Api-Key"

// OnecStatusWebhook applies an order-status change reported by 1С. It is
// implemented by service.OnecOrdersService.
type OnecStatusWebhook interface {
	OnecOrderStatusWebhook(ctx context.Context, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) (ok bool, err error)
}

// onecStatusWebhookRequest mirrors the JSON body documented in
// docs/superpowers/specs/2026-08-26-onec-orders-integration-1c-contract.md
// (раздел "Ручка №2"). Field names are snake_case because that is what the
// contract promises 1С — do not rename without updating that document.
type onecStatusWebhookRequest struct {
	ClientOrderID      uuid.UUID `json:"client_order_id"`
	Status             string    `json:"status"`
	OnecDocumentNumber string    `json:"onec_document_number"`
	Comment            string    `json:"comment"`
}

// RegisterWebhookRoute registers a hand-written POST /api/v1/onec/orders/status
// route on app.
//
// Like RegisterPushOrderRoute this route bypasses the tg code generator, for two
// reasons the generator cannot work around:
//
//   - tg's `http-args` annotations map to QUERY STRING parameters, so the
//     generated route (internal/transport/jsonRPC/internalapi/onecordersapi-rest.go)
//     reads ctx.Query("clientOrderID") and never sees the JSON body the contract
//     tells 1С to send;
//   - tg responses go through custom-handlers.sendResponse, which wraps every
//     payload in the RestResponse envelope, while the contract promises a bare
//     {"ok": true}.
//
// The @tg annotations on pkg/interfaces/internalAPI are kept as documentation of
// intent (and for potential `tg client` consumers), but the generated HTTP route
// for this method is deliberately NOT wired into the server.
func RegisterWebhookRoute(app *fiber.App, svc OnecStatusWebhook, logger zerolog.Logger) {
	app.Post("/api/v1/onec/orders/status", func(ctx *fiber.Ctx) error {
		var req onecStatusWebhookRequest
		// json.Unmarshal rather than ctx.BodyParser: the body shape is fixed by
		// the contract and this does not depend on 1С setting Content-Type.
		if err := json.Unmarshal(ctx.Body(), &req); err != nil {
			logger.Error().Err(err).Msg("onec status webhook: malformed request body")
			return writeWebhookError(ctx, http.StatusBadRequest, customErrors.ErrBadRequest)
		}
		if req.ClientOrderID == uuid.Nil {
			logger.Error().Str("status", req.Status).Msg("onec status webhook: missing client_order_id")
			return writeWebhookError(ctx, http.StatusBadRequest, customErrors.ErrBadRequest)
		}

		apiKey := ctx.Get(WebhookAPIKeyHeader)

		ok, err := svc.OnecOrderStatusWebhook(ctx.UserContext(), apiKey, req.ClientOrderID, req.Status, req.OnecDocumentNumber, req.Comment)
		if err != nil {
			// Log unconditionally: the tg dispatcher only logged errors that
			// carried a wrapped outer error, so plain failures left no trace.
			statusCode, trKey := webhookErrorStatus(err)
			logger.Error().
				Err(err).
				Str("clientOrderID", req.ClientOrderID.String()).
				Str("status", req.Status).
				Int("responseStatus", statusCode).
				Msg("onec status webhook failed")
			return writeWebhookError(ctx, statusCode, trKey)
		}
		if !ok {
			logger.Error().Str("clientOrderID", req.ClientOrderID.String()).Msg("onec status webhook returned not-ok without an error")
			return writeWebhookError(ctx, http.StatusInternalServerError, customErrors.ErrInternal)
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"ok": true})
	})
}

// webhookErrorStatus maps a service error to the HTTP status code the 1С
// contract promises. Typed *customErrors.Error values carry their own status
// (400/401/404/409/500); anything else is an unexpected failure — 500, which is
// also the only class the contract tells 1С to retry.
func webhookErrorStatus(err error) (int, string) {
	var customErr *customErrors.Error
	if errors.As(err, &customErr) {
		if code := customErr.GetStatusCode(); code != 0 {
			return code, customErr.GetTranslationKey()
		}
	}
	return http.StatusInternalServerError, customErrors.ErrInternal
}

func writeWebhookError(ctx *fiber.Ctx, statusCode int, trKey string) error {
	return ctx.Status(statusCode).JSON(fiber.Map{"ok": false, "error": trKey})
}
