package fns

import (
	"context"
	"encoding/json"
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
		KontrSumma     *bool `json:"КонтрСумма"`
		Nedeystvitelny *bool `json:"Недействительный"`
	} `json:"Корректность"`
}

// CheckIndividual queries api-fns.ru fl_status and reports whether inn is
// correct according to the Корректность block only. Any transport, HTTP,
// or parsing failure returns valid=false with a non-nil error (fail-closed) —
// the caller must reject registration rather than assume validity.
func (c *Client) CheckIndividual(ctx context.Context, inn string) (valid bool, err error) {
	reqURL := fmt.Sprintf("%s?inn=%s&key=%s", c.addr, url.QueryEscape(inn), url.QueryEscape(c.key))

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

	var parsed flStatusResponse
	if err = json.Unmarshal(resp.Body(), &parsed); err != nil {
		return false, fmt.Errorf("fns fl_status: parse response: %w", err)
	}

	if parsed.Korrektnost.KontrSumma == nil || parsed.Korrektnost.Nedeystvitelny == nil {
		return false, fmt.Errorf("fns fl_status: Корректность fields are null, retry later")
	}

	return *parsed.Korrektnost.KontrSumma && !*parsed.Korrektnost.Nedeystvitelny, nil
}
