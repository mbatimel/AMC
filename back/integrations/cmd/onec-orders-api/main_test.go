package main

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	customErrors "github.com/mbatimel/AMC/integrations/internal/errors"
)

// generatedClientError reproduces, verbatim, the error the generated orders
// client emits for a non-200 response (orders/pkg/client/transport/
// ordersapi-http-client.go — GENERATED, DO NOT EDIT).
func generatedClientError(statusCode int) error {
	return fmt.Errorf("HTTP error: %d. URL: %s, Method: %s, Body: %s",
		statusCode, "http://orders:8082/api/v1/orders/status", "POST", `{"error":true}`)
}

func TestMapOrdersClientError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"404 becomes NotFound", generatedClientError(http.StatusNotFound), http.StatusNotFound},
		{"401 becomes Unauthorized", generatedClientError(http.StatusUnauthorized), http.StatusUnauthorized},
		{"403 becomes Unauthorized", generatedClientError(http.StatusForbidden), http.StatusUnauthorized},
		{"400 becomes InternalServerError", generatedClientError(http.StatusBadRequest), http.StatusInternalServerError},
		{"500 becomes InternalServerError", generatedClientError(http.StatusInternalServerError), http.StatusInternalServerError},
		{"unparseable becomes InternalServerError", errors.New("dial tcp: connection refused"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := mapOrdersClientError(tt.err)

			var customErr *customErrors.Error
			if !errors.As(mapped, &customErr) {
				t.Fatalf("expected *customErrors.Error, got %T (%v)", mapped, mapped)
			}
			if got := customErr.GetStatusCode(); got != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, got)
			}
			// The original error must survive for server-side logging (I3).
			if customErr.GetOuterError() == nil {
				t.Fatal("expected the original error to be wrapped as the outer error")
			}
			if !errors.Is(mapped, tt.err) {
				t.Fatalf("expected mapped error to unwrap to the original, got %v", mapped)
			}
		})
	}
}

func TestMapOrdersClientError_NilStaysNil(t *testing.T) {
	if err := mapOrdersClientError(nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}
