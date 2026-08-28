package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/integrations/internal/onecorders"
)

type fakePusher struct {
	result onecorders.PushOrderResult
	err    error
	gotReq onecorders.PushOrderRequest
}

func (f *fakePusher) PushOrder(ctx context.Context, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error) {
	f.gotReq = req
	return f.result, f.err
}

func TestPushOrderRoute_Success(t *testing.T) {
	app := fiber.New()
	pusher := &fakePusher{result: onecorders.PushOrderResult{OnecDocumentGUID: uuid.New(), OnecDocumentNumber: "УТ-1"}}
	RegisterPushOrderRoute(app, pusher, zerolog.Nop())

	body, _ := json.Marshal(onecorders.PushOrderRequest{ClientOrderID: uuid.New(), OrderNumber: "AMC-1", Items: []onecorders.ItemDTO{{SKU: "SKU-1", Qty: 1}}})
	req := httptest.NewRequest("POST", "/api/v1/onec-orders/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if pusher.gotReq.OrderNumber != "AMC-1" || len(pusher.gotReq.Items) != 1 {
		t.Fatalf("unexpected request passed to pusher: %+v", pusher.gotReq)
	}
}

func TestPushOrderRoute_PusherFails_Returns502(t *testing.T) {
	app := fiber.New()
	pusher := &fakePusher{err: errors.New("onec down")}
	RegisterPushOrderRoute(app, pusher, zerolog.Nop())

	body, _ := json.Marshal(onecorders.PushOrderRequest{ClientOrderID: uuid.New(), OrderNumber: "AMC-1"})
	req := httptest.NewRequest("POST", "/api/v1/onec-orders/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 502 {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
}

func TestPushOrderRoute_MalformedBody_Returns400(t *testing.T) {
	app := fiber.New()
	pusher := &fakePusher{}
	RegisterPushOrderRoute(app, pusher, zerolog.Nop())

	req := httptest.NewRequest("POST", "/api/v1/onec-orders/push", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
