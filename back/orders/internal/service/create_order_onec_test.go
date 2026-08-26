package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/orders/internal/storage/postgres"
)

type fakeOnecPusher struct {
	called    bool
	lastOrder OnecPushOrder
	guid      uuid.UUID
	number    string
	err       error
}

func (f *fakeOnecPusher) PushOrder(ctx context.Context, order OnecPushOrder) (uuid.UUID, string, error) {
	f.called = true
	f.lastOrder = order
	return f.guid, f.number, f.err
}

func TestCreateOrder_CallsOnecPusherWithItemsAndReturnsProcessingStatus(t *testing.T) {
	productID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	storage := &clientResolutionStorage{
		cartItems: []postgres.CartItemRow{
			{ProductID: productID, SKU: "SKU-1", ProductName: "Товар 1", Qty: 2, Price: 100},
		},
		productOnecRefs: map[uuid.UUID]postgres.ProductOnecRef{
			productID: {SKU: "SKU-1"},
		},
	}
	pusher := &fakeOnecPusher{guid: uuid.New(), number: "УТ-00099"}
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, pusher)

	resp, err := svc.CreateOrder(context.Background(), uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", "коммент")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pusher.called {
		t.Fatal("expected OnecPusher.PushOrder to be called")
	}
	if len(pusher.lastOrder.Items) != 1 || pusher.lastOrder.Items[0].SKU != "SKU-1" || pusher.lastOrder.Items[0].Qty != 2 {
		t.Fatalf("unexpected items passed to pusher: %+v", pusher.lastOrder.Items)
	}
	if resp.Order.Status != "processing" {
		t.Fatalf("expected order status processing, got %q", resp.Order.Status)
	}
}

func TestCreateOrder_OnecPushFails_ReturnsError(t *testing.T) {
	storage := &clientResolutionStorage{
		cartItems: []postgres.CartItemRow{
			{ProductID: uuid.New(), SKU: "SKU-2", ProductName: "Товар 2", Qty: 1, Price: 50},
		},
	}
	pusher := &fakeOnecPusher{err: errors.New("onec down")}
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, pusher)

	_, err := svc.CreateOrder(context.Background(), uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", "")
	if err == nil {
		t.Fatal("expected error when onec push fails")
	}
}
