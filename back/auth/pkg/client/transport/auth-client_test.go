package transport

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/access/pkg/client/transport/httpclient"
	"github.com/valyala/fasthttp"
)

type roundTripperFunc func(*fasthttp.Request, *fasthttp.Response) error

func (f roundTripperFunc) RoundTrip(
	_ *fasthttp.HostClient,
	request *fasthttp.Request,
	response *fasthttp.Response,
) (bool, error) {
	return false, f(request, response)
}

func TestLoginUser_SendsCredentialsAsJSON(t *testing.T) {
	id := uuid.New()
	transport := roundTripperFunc(func(request *fasthttp.Request, response *fasthttp.Response) error {
		if got := request.URI().QueryString(); len(got) != 0 {
			t.Fatalf("query string = %q, want empty", got)
		}
		if got := string(request.Header.ContentType()); got != "application/json" {
			t.Fatalf("Content-Type = %q", got)
		}

		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(request.Body(), &body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Email != "user@example.com" || body.Password != "password" {
			t.Fatalf("credentials = (%q, %q)", body.Email, body.Password)
		}

		response.SetStatusCode(fasthttp.StatusOK)
		response.SetBodyString(`{"userID":"` + id.String() + `"}`)
		return nil
	})
	client := &fasthttp.Client{Transport: transport}
	authClient := NewClientAuthAPI("http://auth", httpclient.WithClient(client))

	userID, err := authClient.LoginUser(context.Background(), "user@example.com", "password")
	if err != nil {
		t.Fatalf("LoginUser() error = %v", err)
	}
	if userID != id {
		t.Fatalf("userID = %s, want %s", userID, id)
	}
}

func TestLoginUserResponse_UnmarshalsUserID(t *testing.T) {
	id := uuid.New()
	body := []byte(`{"userID":"` + id.String() + `"}`)

	var response loginUserResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}
	if response.UserID != id {
		t.Fatalf("response.UserID = %v, want %v", response.UserID, id)
	}
}

func TestLoginUserResponse_EmptyBodyIsZeroUUID(t *testing.T) {
	var response loginUserResponse
	if err := json.Unmarshal([]byte(`{}`), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}
	if response.UserID != uuid.Nil {
		t.Fatalf("response.UserID = %v, want uuid.Nil", response.UserID)
	}
}

func TestLoginError_CarriesStatusCode(t *testing.T) {
	err := &LoginError{StatusCode: 401}

	if err.StatusCode != 401 {
		t.Fatalf("err.StatusCode = %d, want 401", err.StatusCode)
	}
	want := "auth login failed: HTTP 401"
	if got := err.Error(); got != want {
		t.Fatalf("err.Error() = %q, want %q", got, want)
	}
}

func TestLoginError_ErrorMessageReflectsStatusCode(t *testing.T) {
	err := &LoginError{StatusCode: 500}

	want := "auth login failed: HTTP 500"
	if got := err.Error(); got != want {
		t.Fatalf("err.Error() = %q, want %q", got, want)
	}
}
