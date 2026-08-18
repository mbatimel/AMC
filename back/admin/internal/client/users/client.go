// back/admin/internal/client/users/client.go
package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// Client is a minimal hand-written HTTP client to the `users` service:
// admin only needs to look up an account by email and delete it (used when
// a signup request is rejected), so there is no need for the generated
// tg-transport client used by auth/access.
type Client struct {
	httpClient *fasthttp.Client
	baseURL    string
}

func New(baseURL string) *Client {
	return &Client{
		httpClient: &fasthttp.Client{ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

type listUsersItem struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type listUsersEnvelope struct {
	Error     bool   `json:"error"`
	ErrorText string `json:"errorText"`
	Data      struct {
		Items []listUsersItem `json:"items"`
	} `json:"data"`
}

// FindUserIDByEmail looks up an account by exact email (case-insensitive).
// found is false both when the account does not exist and when the lookup
// itself could not be performed reliably enough to say otherwise.
func (c *Client) FindUserIDByEmail(ctx context.Context, email string) (userID uuid.UUID, found bool, err error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return uuid.Nil, false, nil
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(fmt.Sprintf("%s/api/v1/users?q=%s&limit=50", c.baseURL, url.QueryEscape(normalized)))
	req.Header.SetMethod(fasthttp.MethodGet)

	if deadline, ok := ctx.Deadline(); ok {
		c.httpClient.ReadTimeout = time.Until(deadline)
	}
	if err = c.httpClient.Do(req, resp); err != nil {
		return uuid.Nil, false, fmt.Errorf("list users: %w", err)
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return uuid.Nil, false, fmt.Errorf("users service returned status %d", resp.StatusCode())
	}

	var parsed listUsersEnvelope
	if err = json.Unmarshal(resp.Body(), &parsed); err != nil {
		return uuid.Nil, false, fmt.Errorf("decode users response: %w", err)
	}
	if parsed.Error {
		return uuid.Nil, false, fmt.Errorf("users service error: %s", parsed.ErrorText)
	}

	for _, item := range parsed.Data.Items {
		if strings.ToLower(item.Email) == normalized {
			id, parseErr := uuid.Parse(item.ID)
			if parseErr != nil {
				continue
			}
			return id, true, nil
		}
	}
	return uuid.Nil, false, nil
}

func (c *Client) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID))
	req.Header.SetMethod(fasthttp.MethodDelete)

	if deadline, ok := ctx.Deadline(); ok {
		c.httpClient.ReadTimeout = time.Until(deadline)
	}
	if err := c.httpClient.Do(req, resp); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if resp.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("users service returned status %d on delete", resp.StatusCode())
	}
	return nil
}
