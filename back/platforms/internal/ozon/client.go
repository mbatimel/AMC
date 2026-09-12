package ozon

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	customErrors "github.com/mbatimel/AMC/platforms/internal/errors"
)

const productImportPath = "/v3/product/import"

// Client — клиент Ozon Seller API.
type Client struct {
	baseURL  string
	clientID string
	apiKey   string
	timeout  time.Duration
	http     *fasthttp.Client
	logger   zerolog.Logger
}

func New(baseURL, clientID, apiKey string, timeout time.Duration, logger zerolog.Logger) *Client {
	return &Client{
		baseURL:  baseURL,
		clientID: clientID,
		apiKey:   apiKey,
		timeout:  timeout,
		http:     &fasthttp.Client{},
		logger:   logger,
	}
}

func buildItem(params CreateCardParams) importItem {
	item := importItem{
		Barcode:               params.Barcode,
		DescriptionCategoryID: params.DescriptionCategoryID,
		ColorImage:            params.ColorImage,
		CurrencyCode:          params.CurrencyCode,
		Depth:                 params.Dimensions.DepthMM,
		DimensionUnit:         params.Dimensions.DimensionUnit,
		Height:                params.Dimensions.HeightMM,
		Images:                params.Images,
		Name:                  params.Name,
		OfferID:               params.OfferID,
		OldPrice:              params.OldPrice,
		Price:                 params.Price,
		PrimaryImage:          params.PrimaryImage,
		TypeID:                params.TypeID,
		VAT:                   params.VAT,
		Weight:                params.Dimensions.Weight,
		WeightUnit:            params.Dimensions.WeightUnit,
		Width:                 params.Dimensions.WidthMM,
	}
	for _, attribute := range params.Attributes {
		values := make([]importAttributeValue, 0, len(attribute.Values))
		for _, value := range attribute.Values {
			values = append(values, importAttributeValue{
				DictionaryValueID: value.DictionaryValueID,
				Value:             value.Value,
			})
		}
		item.Attributes = append(item.Attributes, importAttribute{
			ComplexID: attribute.ComplexID,
			ID:        attribute.ID,
			Values:    values,
		})
	}
	return item
}

func mapUpstreamError(statusCode int, body []byte) error {
	var envelope errorEnvelope
	_ = json.Unmarshal(body, &envelope)

	switch statusCode {
	case fasthttp.StatusBadRequest:
		return customErrors.ErrValidation.AddCause("ozon", envelope.Message)
	case fasthttp.StatusTooManyRequests:
		return customErrors.ErrRateLimited.AddCause("ozon", envelope.Message)
	default:
		return customErrors.ErrUpstreamError.AddCause("ozon", envelope.Message, "status", statusCode)
	}
}

func (c *Client) CreateCard(ctx context.Context, params CreateCardParams) (CreateCardResult, error) {
	_ = ctx // fasthttp.Client.DoTimeout не принимает context

	if c.clientID == "" || c.apiKey == "" {
		return CreateCardResult{}, customErrors.ErrInternal.AddCause("field", "ozonCredentials")
	}

	body, err := json.Marshal(importRequest{Items: []importItem{buildItem(params)}})
	if err != nil {
		return CreateCardResult{}, customErrors.ErrInternal
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.baseURL + productImportPath)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Client-Id", c.clientID)
	req.Header.Set("Api-Key", c.apiKey)
	req.SetBody(body)

	if doErr := c.http.DoTimeout(req, resp, c.timeout); doErr != nil {
		c.logger.Error().Err(doErr).Msg("ozon product/import request failed")
		return CreateCardResult{}, customErrors.ErrUpstreamError.AddCause("ozon", doErr.Error())
	}

	statusCode := resp.StatusCode()
	respBody := resp.Body()

	if statusCode != fasthttp.StatusOK {
		c.logger.Warn().Int("status", statusCode).Bytes("body", respBody).Msg("ozon product/import unexpected status")
		return CreateCardResult{}, mapUpstreamError(statusCode, respBody)
	}

	var envelope importResponse
	if err = json.Unmarshal(respBody, &envelope); err != nil {
		return CreateCardResult{}, fmt.Errorf("ozon product/import: decode response: %w", err)
	}

	c.logger.Info().Str("offerID", params.OfferID).Int64("taskID", envelope.Result.TaskID).Msg("ozon product/import accepted")
	return CreateCardResult{Accepted: true, TaskID: envelope.Result.TaskID}, nil
}
