package onecorders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestPushOrder_Success(t *testing.T) {
	guid := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hs/amc-integration/orders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatal("expected Authorization header")
		}
		var req PushOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(req.Items) != 1 || req.Items[0].SKU != "SKU-1" {
			t.Fatalf("unexpected items: %+v", req.Items)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pushOrderSuccessResponse{OnecDocumentGUID: guid, OnecDocumentNumber: "УТ-00042"})
	}))
	defer srv.Close()

	client := New(srv.URL, "user", "pass", 5*time.Second, zerolog.Nop())
	result, err := client.PushOrder(context.Background(), PushOrderRequest{
		ClientOrderID: uuid.New(),
		OrderNumber:   "AMC-1",
		Items:         []ItemDTO{{SKU: "SKU-1", Qty: 1, Price: 100}},
	})
	if err != nil {
		t.Fatalf("PushOrder: %v", err)
	}
	if result.OnecDocumentGUID != guid || result.OnecDocumentNumber != "УТ-00042" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestPushOrder_NonOKStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad counterparty"}`))
	}))
	defer srv.Close()

	client := New(srv.URL, "user", "pass", 5*time.Second, zerolog.Nop())
	_, err := client.PushOrder(context.Background(), PushOrderRequest{ClientOrderID: uuid.New(), OrderNumber: "AMC-2"})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
