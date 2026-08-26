package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

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
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"onec_document_guid":   result.OnecDocumentGUID,
			"onec_document_number": result.OnecDocumentNumber,
		})
	})
}
