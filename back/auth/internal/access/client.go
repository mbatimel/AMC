package access

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type Client struct {
	baseURL string
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

type checkAccessResponse struct {
	Allowed bool `json:"allowed"`
}

// CheckAccess calls access service over HTTP to verify that userID's role is
// at least as privileged as role.
func (c *Client) CheckAccess(ctx context.Context, userID uuid.UUID, role int) (bool, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(fmt.Sprintf("%s/api/v1/access/check", c.baseURL))
	req.Header.SetMethod(fasthttp.MethodPost)
	req.URI().QueryArgs().Set("userID", userID.String())
	req.URI().QueryArgs().Set("role", fmt.Sprint(role))

	timeout := 5 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	if err := fasthttp.DoTimeout(req, resp, timeout); err != nil {
		return false, fmt.Errorf("call access CheckAccess: %w", err)
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return false, fmt.Errorf("access CheckAccess returned status %d: %s", resp.StatusCode(), resp.Body())
	}

	var body checkAccessResponse
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return false, fmt.Errorf("decode access CheckAccess response: %w", err)
	}
	return body.Allowed, nil
}
