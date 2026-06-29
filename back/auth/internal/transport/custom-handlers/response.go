package custom_handlers

import (
	"encoding/json"
	"net/http"

	customErrors "github.com/mbatimel/AMC/auth/internal/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

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
		ctx.Response().SetStatusCode(fasthttp.StatusInternalServerError)
		response.ErrorText = customErrors.ErrInternal

		customErr, ok := respError.(*customErrors.Error)
		if ok {
			if customErr.GetOuterError() != nil {
				response.ErrorText = customErr.GetOuterError().Error()
			}

			if customErr.GetStatusCode() != 0 {
				ctx.Response().SetStatusCode(customErr.GetStatusCode())
			}
		}

		if customErr != nil {
			response.AdditionalErrors = make(map[string]interface{})
			response.AdditionalErrors["reason"] = customErr.ErrorText
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
