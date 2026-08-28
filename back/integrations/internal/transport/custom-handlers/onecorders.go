package custom_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	internalapi "github.com/mbatimel/AMC/integrations/pkg/interfaces/internalAPI"
)

func OnecOrderStatusWebhook(ctx *fiber.Ctx, svc internalapi.OnecOrdersAPI, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) error {
	ok, err := svc.OnecOrderStatusWebhook(ctx.UserContext(), apiKey, clientOrderID, status, onecDocumentNumber, comment)
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, fiber.Map{"ok": ok}, nil)
	return nil
}
