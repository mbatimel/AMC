package wildberries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"

	customErrors "github.com/mbatimel/AMC/platforms/internal/errors"
)

const cardsUploadPath = "/content/v2/cards/upload"

// Client — клиент Wildberries Content API.
type Client struct {
	baseURL string
	apiKey  string
	timeout time.Duration
	http    *fasthttp.Client
	logger  zerolog.Logger
}

func New(baseURL, apiKey string, timeout time.Duration, logger zerolog.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		timeout: timeout,
		http:    &fasthttp.Client{},
		logger:  logger,
	}
}

func buildVariant(params CreateCardParams) uploadVariant {
	variant := uploadVariant{
		Brand:       params.Brand,
		Title:       params.Title,
		Description: params.Description,
		VendorCode:  params.VendorCode,
		KizMarked:   params.KizMarked,
		Wholesale: &uploadWholesale{
			Enabled: params.WholesaleEnabled,
			Quantum: params.WholesaleQuantum,
		},
		Dimensions: &uploadDimensions{
			Length:       params.Dimensions.LengthCM,
			Width:        params.Dimensions.WidthCM,
			Height:       params.Dimensions.HeightCM,
			WeightBrutto: params.Dimensions.WeightBruttoKG,
		},
		Documents: &uploadDocuments{
			ExcludeDocuments: params.Documents.ExcludeDocuments,
		},
	}
	for _, size := range params.Sizes {
		variant.Sizes = append(variant.Sizes, uploadSize{
			TechSize: size.TechSize,
			WBSize:   size.WBSize,
			Price:    size.Price,
			SKUs:     size.SKUs,
		})
	}
	for _, characteristic := range params.Characteristics {
		variant.Characteristics = append(variant.Characteristics, uploadCharacteristic{
			ID:    characteristic.ID,
			Value: json.RawMessage(characteristic.ValueJSON),
		})
	}
	for _, item := range params.Documents.Items {
		variant.Documents.Items = append(variant.Documents.Items, uploadDocumentItem{
			Type:          item.Type,
			Number:        item.Number,
			ProductNumber: item.ProductNumber,
			TradeName:     item.TradeName,
			Applicant:     item.Applicant,
			StartDate:     item.StartDate,
			EndDate:       item.EndDate,
			IsEndless:     item.IsEndless,
		})
	}
	return variant
}

func mapUpstreamError(statusCode int, body []byte) error {
	var envelope uploadErrorEnvelope
	_ = json.Unmarshal(body, &envelope)
	message := envelope.Detail
	if message == "" {
		message = envelope.Title
	}

	switch statusCode {
	case fasthttp.StatusBadRequest:
		return customErrors.ErrValidation.AddCause("wildberries", message)
	case fasthttp.StatusTooManyRequests:
		return customErrors.ErrRateLimited.AddCause("wildberries", message)
	default:
		return customErrors.ErrUpstreamError.AddCause("wildberries", message, "status", statusCode)
	}
}

func (c *Client) CreateCard(ctx context.Context, params CreateCardParams) (CreateCardResult, error) {
	_ = ctx // fasthttp.Client.DoTimeout не принимает context

	if c.apiKey == "" {
		return CreateCardResult{}, customErrors.ErrInternal.AddCause("field", "wildberriesAPIKey")
	}

	payload := []uploadCardGroup{{
		SubjectID: params.SubjectID,
		Variants:  []uploadVariant{buildVariant(params)},
	}}
	body, err := json.Marshal(payload)
	if err != nil {
		return CreateCardResult{}, customErrors.ErrInternal
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.baseURL + cardsUploadPath)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.Set("Authorization", c.apiKey)
	req.SetBody(body)

	if doErr := c.http.DoTimeout(req, resp, c.timeout); doErr != nil {
		c.logger.Error().Err(doErr).Msg("wildberries cards/upload request failed")
		return CreateCardResult{}, customErrors.ErrUpstreamError.AddCause("wildberries", doErr.Error())
	}

	statusCode := resp.StatusCode()
	respBody := resp.Body()

	if statusCode != fasthttp.StatusOK {
		c.logger.Warn().Int("status", statusCode).Bytes("body", respBody).Msg("wildberries cards/upload unexpected status")
		return CreateCardResult{}, mapUpstreamError(statusCode, respBody)
	}

	var envelope uploadResponse
	if err = json.Unmarshal(respBody, &envelope); err != nil {
		return CreateCardResult{}, fmt.Errorf("wildberries cards/upload: decode response: %w", err)
	}
	if envelope.Error {
		return CreateCardResult{}, customErrors.ErrUpstreamError.AddCause("wildberries", envelope.ErrorText)
	}

	c.logger.Info().Str("vendorCode", params.VendorCode).Int("subjectID", params.SubjectID).Msg("wildberries card upload accepted")
	return CreateCardResult{Accepted: true}, nil
}
