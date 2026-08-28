package onecorders

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const pushOrderPath = "/hs/amc-integration/orders"

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

func (c *Client) PushOrder(ctx context.Context, in PushOrderRequest) (PushOrderResult, error) {
	_ = ctx

	body, err := json.Marshal(in)
	if err != nil {
		return PushOrderResult{}, fmt.Errorf("push order: marshal request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.baseURL + pushOrderPath)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(c.user + ":" + c.password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.SetBody(body)

	if err = c.http.DoTimeout(req, resp, c.timeout); err != nil {
		c.logger.Error().Str("clientOrderID", in.ClientOrderID.String()).Err(err).Msg("onec push order request failed")
		return PushOrderResult{}, fmt.Errorf("push order request: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		c.logger.Error().Str("clientOrderID", in.ClientOrderID.String()).Int("status", resp.StatusCode()).Str("response", string(resp.Body())).Msg("onec push order unexpected status")
		return PushOrderResult{}, fmt.Errorf("push order: unexpected status %d: %s", resp.StatusCode(), resp.Body())
	}

	var out pushOrderSuccessResponse
	if err = json.Unmarshal(resp.Body(), &out); err != nil {
		return PushOrderResult{}, fmt.Errorf("push order: decode response: %w", err)
	}
	return PushOrderResult{OnecDocumentGUID: out.OnecDocumentGUID, OnecDocumentNumber: out.OnecDocumentNumber}, nil
}
