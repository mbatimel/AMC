package yandexmarket

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	customErrors "github.com/mbatimel/AMC/platforms/internal/errors"
)

// statusMethodFailure — нестандартный код Яндекс Маркета: превышен лимит запросов.
const statusMethodFailure = 420

// Client — клиент Yandex Market Partner API.
type Client struct {
	baseURL    string
	businessID int64
	apiKey     string
	timeout    time.Duration
	http       *fasthttp.Client
	logger     zerolog.Logger
}

func New(baseURL string, businessID int64, apiKey string, timeout time.Duration, logger zerolog.Logger) *Client {
	return &Client{
		baseURL:    baseURL,
		businessID: businessID,
		apiKey:     apiKey,
		timeout:    timeout,
		http:       &fasthttp.Client{},
		logger:     logger,
	}
}

func buildOffer(params CreateCardParams) offerDTO {
	offer := offerDTO{
		OfferID:          params.OfferID,
		Name:             params.Name,
		MarketCategoryID: params.MarketCategoryID,
		Description:      params.Description,
		Vendor:           params.Vendor,
		VendorCode:       params.VendorCode,
		Pictures:         params.Pictures,
		Barcodes:         params.Barcodes,
	}
	if d := params.Dimensions; d.LengthCM > 0 || d.WidthCM > 0 || d.HeightCM > 0 || d.WeightKG > 0 {
		offer.WeightDimensions = &weightDimensions{
			Length: d.LengthCM,
			Width:  d.WidthCM,
			Height: d.HeightCM,
			Weight: d.WeightKG,
		}
	}
	for _, value := range params.ParameterValues {
		offer.ParameterValues = append(offer.ParameterValues, parameterValueDTO{
			ParameterID: value.ParameterID,
			ValueID:     value.ValueID,
			UnitID:      value.UnitID,
			Value:       value.Value,
		})
	}
	return offer
}

func formatMappingErrors(errs []offerMappingError) string {
	messages := make([]string, 0, len(errs))
	for _, e := range errs {
		messages = append(messages, e.Type+": "+e.Message)
	}
	return strings.Join(messages, "; ")
}

func mapUpstreamError(statusCode int, body []byte) error {
	var envelope errorEnvelope
	_ = json.Unmarshal(body, &envelope)
	message := ""
	if len(envelope.Errors) > 0 {
		message = envelope.Errors[0].Code + ": " + envelope.Errors[0].Message
	}

	switch statusCode {
	case fasthttp.StatusBadRequest:
		return customErrors.ErrValidation.AddCause("yandexmarket", message)
	case statusMethodFailure:
		return customErrors.ErrRateLimited.AddCause("yandexmarket", message)
	default:
		return customErrors.ErrUpstreamError.AddCause("yandexmarket", message, "status", statusCode)
	}
}

func (c *Client) CreateCard(ctx context.Context, params CreateCardParams) (CreateCardResult, error) {
	_ = ctx // fasthttp.Client.DoTimeout не принимает context

	if c.businessID <= 0 || c.apiKey == "" {
		return CreateCardResult{}, customErrors.ErrInternal.AddCause("field", "yandexMarketCredentials")
	}

	body, err := json.Marshal(updateRequest{
		OfferMappings: []offerMapping{{Offer: buildOffer(params)}},
	})
	if err != nil {
		return CreateCardResult{}, customErrors.ErrInternal
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	url := fmt.Sprintf("%s/v2/businesses/%s/offer-mappings/update", c.baseURL, strconv.FormatInt(c.businessID, 10))
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Api-Key", c.apiKey)
	req.SetBody(body)

	if doErr := c.http.DoTimeout(req, resp, c.timeout); doErr != nil {
		c.logger.Error().Err(doErr).Msg("yandex market offer-mappings/update request failed")
		return CreateCardResult{}, customErrors.ErrUpstreamError.AddCause("yandexmarket", doErr.Error())
	}

	statusCode := resp.StatusCode()
	respBody := resp.Body()

	if statusCode != fasthttp.StatusOK {
		c.logger.Warn().Int("status", statusCode).Bytes("body", respBody).Msg("yandex market offer-mappings/update unexpected status")
		return CreateCardResult{}, mapUpstreamError(statusCode, respBody)
	}

	var envelope updateResponse
	if err = json.Unmarshal(respBody, &envelope); err != nil {
		return CreateCardResult{}, fmt.Errorf("yandex market offer-mappings/update: decode response: %w", err)
	}

	var warnings []string
	for _, result := range envelope.Results {
		if len(result.Errors) > 0 {
			return CreateCardResult{}, customErrors.ErrValidation.AddCause("yandexmarket", formatMappingErrors(result.Errors))
		}
		if len(result.Warnings) > 0 {
			warnings = append(warnings, formatMappingErrors(result.Warnings))
		}
	}
	if envelope.Status != "OK" {
		return CreateCardResult{}, customErrors.ErrUpstreamError.AddCause("yandexmarket", envelope.Status)
	}

	c.logger.Info().Str("offerID", params.OfferID).Msg("yandex market offer accepted")
	return CreateCardResult{Accepted: true, Warnings: warnings}, nil
}
