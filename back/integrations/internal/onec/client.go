package onec

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const (
	entitySetCategories = "Catalog_НоменклатурныеГруппы"
	entitySetWarehouses = "Catalog_Склады"
	entitySetProducts   = "Catalog_Номенклатура"
	entitySetPrices     = "InformationRegister_ЦеныНоменклатуры"
	entitySetStock      = "AccumulationRegister_ТоварыНаСкладахBalance"
)

type Client struct {
	baseURL  string
	user     string
	password string
	timeout  time.Duration
	http     *fasthttp.Client
	logger   zerolog.Logger
}

func New(baseURL, user, password string, timeout time.Duration, logger zerolog.Logger) *Client {
	return &Client{
		baseURL:  baseURL,
		user:     user,
		password: password,
		timeout:  timeout,
		http:     &fasthttp.Client{},
		logger:   logger,
	}
}

func fetchEntitySet[T any](ctx context.Context, c *Client, entitySet string) ([]T, error) {
	_ = ctx // резерв на будущее (deadline/cancel), fasthttp.DoTimeout ctx не принимает

	reqURL := fmt.Sprintf("%s/%s?$format=json", c.baseURL, entitySet)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqURL)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("Accept", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(c.user + ":" + c.password))
	req.Header.Set("Authorization", "Basic "+auth)

	doErr := c.http.DoTimeout(req, resp, c.timeout)
	if doErr != nil {
		c.logger.Error().Str("entitySet", entitySet).Err(doErr).Msg("onec odata request failed")
		return nil, fmt.Errorf("onec odata request %s: %w", entitySet, doErr)
	}

	statusCode := resp.StatusCode()
	// resp.Body() остаётся валидным до fasthttp.ReleaseResponse(resp) (см.
	// defer выше), а весь код ниже, использующий body, выполняется до
	// возврата из функции — отдельная копия буфера не нужна.
	body := resp.Body()

	// Полное тело ответа логируем только на Debug: для крупного каталога
	// это может быть один огромный лог на каждый прогон (5 раз в день).
	c.logger.Debug().Str("entitySet", entitySet).Int("status", statusCode).Str("response", string(body)).Msg("onec odata response body")

	if statusCode != fasthttp.StatusOK {
		c.logger.Error().Str("entitySet", entitySet).Int("status", statusCode).Msg("onec odata unexpected status")
		return nil, fmt.Errorf("onec odata %s: unexpected status %d", entitySet, statusCode)
	}
	c.logger.Info().Str("entitySet", entitySet).Int("status", statusCode).Msg("onec odata response")

	var envelope odataEnvelope[T]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("onec odata %s: decode response: %w", entitySet, err)
	}
	return envelope.Value, nil
}

func (c *Client) FetchCategories(ctx context.Context) ([]CategoryDTO, error) {
	return fetchEntitySet[CategoryDTO](ctx, c, entitySetCategories)
}

func (c *Client) FetchWarehouses(ctx context.Context) ([]WarehouseDTO, error) {
	return fetchEntitySet[WarehouseDTO](ctx, c, entitySetWarehouses)
}

func (c *Client) FetchProducts(ctx context.Context) ([]ProductDTO, error) {
	return fetchEntitySet[ProductDTO](ctx, c, entitySetProducts)
}

func (c *Client) FetchPrices(ctx context.Context) ([]PriceDTO, error) {
	return fetchEntitySet[PriceDTO](ctx, c, entitySetPrices)
}

func (c *Client) FetchStock(ctx context.Context) ([]StockDTO, error) {
	return fetchEntitySet[StockDTO](ctx, c, entitySetStock)
}
