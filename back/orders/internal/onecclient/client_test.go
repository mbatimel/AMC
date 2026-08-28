package onecclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/orders/internal/service"
)

func TestPushOrder_Success(t *testing.T) {
	guid := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req wireRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(req.Items) != 1 || req.Items[0].SKU != "SKU-1" {
			t.Fatalf("unexpected items: %+v", req.Items)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wireResponse{OnecDocumentGUID: guid, OnecDocumentNumber: "УТ-1"})
	}))
	defer srv.Close()

	client := New(srv.URL, 5*time.Second)
	onecGUID, onecNumber, err := client.PushOrder(context.Background(), service.OnecPushOrder{
		ClientOrderID: uuid.New(),
		OrderNumber:   "AMC-1",
		Items:         []service.OnecOrderItem{{SKU: "SKU-1", Qty: 1, Price: 100}},
	})
	if err != nil {
		t.Fatalf("PushOrder: %v", err)
	}
	if onecGUID != guid || onecNumber != "УТ-1" {
		t.Fatalf("unexpected result: %s / %s", onecGUID, onecNumber)
	}
}

func TestPushOrder_NonOKStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := New(srv.URL, 5*time.Second)
	_, _, err := client.PushOrder(context.Background(), service.OnecPushOrder{ClientOrderID: uuid.New(), OrderNumber: "AMC-1"})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
