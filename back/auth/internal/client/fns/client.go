package fns

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const requestTimeout = 5 * time.Second

type Client struct {
	addr   string
	key    string
	http   *fasthttp.Client
	logger zerolog.Logger
}

func New(addr, key string, logger zerolog.Logger) *Client {
	return &Client{
		addr:   addr,
		key:    key,
		http:   &fasthttp.Client{},
		logger: logger,
	}
}

type flStatusResponse struct {
	Korrektnost struct {
		KontrSumma     *string `json:"Лицензии"`
	} `json:"Позитив"`
}

// CheckIndividual queries api-fns.ru fl_status and reports whether inn is
// correct according to the Корректность block only. Any transport, HTTP,
// or parsing failure returns valid=false with a non-nil error (fail-closed) —
// the caller must reject registration rather than assume validity.
func (c *Client) CheckIndividual(ctx context.Context, inn string) (valid bool, err error) {
	// если взять подписку или ключ то req заменить на inn
	reqURL := fmt.Sprintf("%s?req=%s&key=%s", c.addr, url.QueryEscape(inn), url.QueryEscape(c.key))

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqURL)
	req.Header.SetMethod(fasthttp.MethodGet)

	doErr := c.http.DoTimeout(req, resp, requestTimeout)

	statusCode := resp.StatusCode()
	body := string(resp.Body())

	logEvent := c.logger.Info().Str("inn", inn).Int("status", statusCode).Str("response", body)
	if doErr != nil {
		logEvent.Err(doErr).Msg("fns fl_status request failed")
		return false, fmt.Errorf("fns fl_status request: %w", doErr)
	}
	logEvent.Msg("fns fl_status response")

	if statusCode != fasthttp.StatusOK {
		return false, fmt.Errorf("fns fl_status: unexpected status %d", statusCode)
	}

	return true, nil
}
