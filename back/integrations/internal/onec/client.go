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

// maxRedirects ограничивает число переходов по Location, чтобы не зациклиться
// на редиректе на редирект (например, http -> https -> http на кривом прокси).
const maxRedirects = 5

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

	origURI := fasthttp.AcquireURI()
	origURI.Update(reqURL)
	origHost := string(origURI.Host())
	origScheme := string(origURI.Scheme())
	fasthttp.ReleaseURI(origURI)

	var statusCode int
	var body []byte

	// fasthttp.Client.DoTimeout (в отличие от Get/DoRedirects) редиректы не
	// следует — 1C/IIS на 301/302 отдаёт пустое тело, и без ручного перехода
	// по Location запрос всегда падал бы с "unexpected status 301".
	for redirect := 0; ; redirect++ {
		doErr := c.http.DoTimeout(req, resp, c.timeout)
		if doErr != nil {
			c.logger.Error().Str("entitySet", entitySet).Err(doErr).Msg("onec odata request failed")
			return nil, fmt.Errorf("onec odata request %s: %w", entitySet, doErr)
		}

		statusCode = resp.StatusCode()
		if !fasthttp.StatusCodeIsRedirect(statusCode) {
			body = append([]byte(nil), resp.Body()...)
			break
		}

		location := resp.Header.Peek("Location")
		if redirect >= maxRedirects || len(location) == 0 {
			c.logger.Error().Str("entitySet", entitySet).Int("status", statusCode).Bytes("location", location).Msg("onec odata redirect loop or missing location")
			return nil, fmt.Errorf("onec odata %s: unexpected status %d (redirect not followed)", entitySet, statusCode)
		}

		// Location может быть относительным — резолвим его так же, как это
		// делает сам fasthttp в DoRedirects.
		redirectURI := fasthttp.AcquireURI()
		redirectURI.Update(reqURL)
		redirectURI.UpdateBytes(location)
		redirectURL := redirectURI.String()
		redirectHost := string(redirectURI.Host())
		redirectScheme := string(redirectURI.Scheme())
		fasthttp.ReleaseURI(redirectURI)

		c.logger.Warn().Str("entitySet", entitySet).Int("status", statusCode).Str("location", redirectURL).Msg("onec odata redirected")

		// Basic-auth уходит только на исходный host и не понижается до http:
		// редирект на другой хост или https->http не должен палить учётку.
		if redirectHost != origHost || (origScheme == "https" && redirectScheme == "http") {
			c.logger.Warn().Str("entitySet", entitySet).Str("origHost", origHost).Str("redirectHost", redirectHost).Msg("onec odata cross-host/downgrade redirect, stripping Authorization")
			req.Header.Del("Authorization")
		}

		resp.Reset()
		req.SetRequestURI(redirectURL)
		req.Header.SetMethod(fasthttp.MethodGet)
		reqURL = redirectURL
	}

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
