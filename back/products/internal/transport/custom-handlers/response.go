package custom_handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	customErrors "github.com/mbatimel/AMC/products/internal/errors"
	"github.com/rs/zerolog"
)

const errInternal = "amc.errors.products.internalError"

type RestResponse struct {
	Data             interface{}            `json:"data"`
	Error            bool                   `json:"error"`
	ErrorText        string                 `json:"errorText"`
	AdditionalErrors map[string]interface{} `json:"additionalErrors"`
}

func sendResponse(ctx *fiber.Ctx, log zerolog.Logger, data interface{}, respError error) {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(http.StatusOK)

	response := &RestResponse{
		Data:  data,
		Error: respError != nil,
	}

	if response.Error {
		ctx.Response().SetStatusCode(http.StatusInternalServerError)
		response.ErrorText = errInternal

		if customErr, ok := respError.(*customErrors.Error); ok {
			if customErr.GetStatusCode() != 0 {
				ctx.Response().SetStatusCode(customErr.GetStatusCode())
			}
			response.ErrorText = customErr.ErrorText
			response.AdditionalErrors = customErr.Cause
		}
	}

	respBody, err := json.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal response")
		return
	}

	if _, err = ctx.Write(respBody); err != nil {
		log.Error().Err(err).Msg("failed to send response")
		return
	}
}
