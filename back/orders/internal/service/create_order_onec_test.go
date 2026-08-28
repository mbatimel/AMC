package service

import (
	"context"
	"errors"
	"strings"
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

func TestCreateOrder_PushFails_CleansUpOrphanAddressAndContact(t *testing.T) {
	storage := &clientResolutionStorage{
		cartItems: []postgres.CartItemRow{
			{ProductID: uuid.New(), SKU: "SKU-3", ProductName: "Товар 3", Qty: 1, Price: 50},
		},
	}
	pusher := &fakeOnecPusher{err: errors.New("onec down")}
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, pusher)

	if _, err := svc.CreateOrder(context.Background(), uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", ""); err == nil {
		t.Fatal("expected error when onec push fails")
	}

	if storage.deleteAddressAndContactCalls != 1 {
		t.Fatalf("expected exactly one cleanup call, got %d", storage.deleteAddressAndContactCalls)
	}
	if storage.deletedAddressID != storage.insertedAddressID {
		t.Fatalf("cleanup deleted address %s, expected the inserted one %s", storage.deletedAddressID, storage.insertedAddressID)
	}
	if storage.deletedContactID != storage.insertedContactID {
		t.Fatalf("cleanup deleted contact %s, expected the inserted one %s", storage.deletedContactID, storage.insertedContactID)
	}
}

func TestCreateOrder_CleanupFailure_DoesNotMaskPushError(t *testing.T) {
	storage := &clientResolutionStorage{
		cartItems: []postgres.CartItemRow{
			{ProductID: uuid.New(), SKU: "SKU-4", ProductName: "Товар 4", Qty: 1, Price: 50},
		},
		deleteAddressAndContactErr: errors.New("cleanup failed too"),
	}
	pushErr := errors.New("onec down")
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, &fakeOnecPusher{err: pushErr})

	_, err := svc.CreateOrder(context.Background(), uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", "")
	if err == nil {
		t.Fatal("expected error when onec push fails")
	}
	if !errors.Is(err, pushErr) {
		t.Fatalf("cleanup failure masked the original push error: %v", err)
	}
	if strings.Contains(err.Error(), "cleanup failed too") {
		t.Fatalf("cleanup error leaked into the returned error: %v", err)
	}
}

// The 1С push timing out is the most common way to reach the failure path, and
// it leaves the request context already cancelled. The cleanup must still run.
func TestCreateOrder_PushFailsWithCancelledContext_CleanupStillRuns(t *testing.T) {
	storage := &clientResolutionStorage{
		cartItems: []postgres.CartItemRow{
			{ProductID: uuid.New(), SKU: "SKU-6", ProductName: "Товар 6", Qty: 1, Price: 50},
		},
	}
	pusher := &fakeOnecPusher{err: context.DeadlineExceeded}
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, pusher)

	ctx, cancel := context.WithCancel(context.Background())
	storage.onCreateOrder = cancel // cancel the request ctx before the push fails

	if _, err := svc.CreateOrder(ctx, uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", ""); err == nil {
		t.Fatal("expected error when onec push fails")
	}

	if storage.deleteAddressAndContactCalls != 1 {
		t.Fatalf("expected cleanup to run despite the cancelled context, calls=%d", storage.deleteAddressAndContactCalls)
	}
	if storage.cleanupCtxErrAtCall != nil {
		t.Fatalf("cleanup context must not inherit cancellation, got %v", storage.cleanupCtxErrAtCall)
	}
	if !storage.cleanupCtxHadDeadline {
		t.Fatal("cleanup context must carry its own timeout")
	}
}

func TestCreateOrder_PushSucceeds_DoesNotDeleteAddressOrContact(t *testing.T) {
	storage := &clientResolutionStorage{
		cartItems: []postgres.CartItemRow{
			{ProductID: uuid.New(), SKU: "SKU-5", ProductName: "Товар 5", Qty: 1, Price: 50},
		},
	}
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, &fakeOnecPusher{guid: uuid.New(), number: "УТ-1"})

	if _, err := svc.CreateOrder(context.Background(), uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.deleteAddressAndContactCalls != 0 {
		t.Fatalf("cleanup must not run on the success path, calls=%d", storage.deleteAddressAndContactCalls)
	}
}
