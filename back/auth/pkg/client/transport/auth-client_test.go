package transport

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

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
