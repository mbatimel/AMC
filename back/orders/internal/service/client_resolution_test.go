package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/orders/internal/errors"
	"github.com/mbatimel/AMC/orders/internal/storage/postgres"
)

type clientResolutionStorage struct {
	Storage
	activeClientID          uuid.UUID
	activeClientErr         error
	counterpartyExists      bool
	counterpartyExistsErr   error
	userHasClient           bool
	userHasClientErr        error
	cartID                  uuid.UUID
	getCartErr              error
	getOrCreateCartErr      error
	listOrdersErr           error
	listOrdersRows          []postgres.OrderRow
	listOrdersTotal         int
	listOrdersParams        postgres.ListOrdersParams
	getCartCalls            int
	getCartItemsCalls       int
	getOrCreateCartCalls    int
	listOrdersCalls         int
	counterpartyExistsCalls int
	userHasClientCalls      int
	getCartCounterpartyID   uuid.NullUUID
	orderStatus             string
	orderStatusErr          error
	productOnecRefs         map[uuid.UUID]postgres.ProductOnecRef
	counterpartyOnecRef     postgres.CounterpartyOnecRef
	cartItems               []postgres.CartItemRow

	insertedAddressID            uuid.UUID
	insertedContactID            uuid.UUID
	deletedAddressID             uuid.UUID
	deletedContactID             uuid.UUID
	deleteAddressAndContactErr   error
	deleteAddressAndContactCalls int
	cleanupCtxErrAtCall          error
	cleanupCtxHadDeadline        bool
	onCreateOrder                func()
}

func (s *clientResolutionStorage) GetProductOnecRefs(context.Context, []uuid.UUID) (map[uuid.UUID]postgres.ProductOnecRef, error) {
	return s.productOnecRefs, nil
}

func (s *clientResolutionStorage) GetCounterpartyOnecRef(context.Context, uuid.UUID) (postgres.CounterpartyOnecRef, error) {
	return s.counterpartyOnecRef, nil
}

func (s *clientResolutionStorage) CreateOrder(
	ctx context.Context,
	params postgres.CreateOrderParams,
	pushToOnec func(context.Context, uuid.UUID, string) (uuid.UUID, string, error),
) (postgres.CreatedOrder, error) {
	orderID := uuid.New()
	if s.onCreateOrder != nil {
		s.onCreateOrder()
	}
	_, _, err := pushToOnec(ctx, orderID, "TEST-0001")
	if err != nil {
		return postgres.CreatedOrder{}, err
	}
	return postgres.CreatedOrder{ID: orderID, Number: "TEST-0001", Status: "processing"}, nil
}

func (s *clientResolutionStorage) GetOrderStatus(context.Context, uuid.UUID) (string, error) {
	if s.orderStatusErr != nil {
		return "", s.orderStatusErr
	}
	return s.orderStatus, nil
}

func (s *clientResolutionStorage) GetActiveClient(context.Context, uuid.UUID) (uuid.UUID, error) {
	return s.activeClientID, s.activeClientErr
}

func (s *clientResolutionStorage) CounterpartyExists(context.Context, uuid.UUID) (bool, error) {
	s.counterpartyExistsCalls++
	return s.counterpartyExists, s.counterpartyExistsErr
}

func (s *clientResolutionStorage) UserHasClient(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	s.userHasClientCalls++
	return s.userHasClient, s.userHasClientErr
}

func (s *clientResolutionStorage) GetCart(_ context.Context, _ uuid.UUID, counterpartyID uuid.NullUUID) (uuid.UUID, error) {
	s.getCartCalls++
	s.getCartCounterpartyID = counterpartyID
	return s.cartID, s.getCartErr
}

func (s *clientResolutionStorage) GetOrCreateCart(context.Context, uuid.UUID, uuid.NullUUID) (uuid.UUID, error) {
	s.getOrCreateCartCalls++
	return uuid.New(), s.getOrCreateCartErr
}

func (s *clientResolutionStorage) GetCartItems(context.Context, uuid.UUID) ([]postgres.CartItemRow, error) {
	s.getCartItemsCalls++
	return s.cartItems, nil
}

func (s *clientResolutionStorage) InsertDeliveryAddress(context.Context, uuid.NullUUID, string, string) (uuid.UUID, error) {
	id := uuid.New()
	s.insertedAddressID = id
	return id, nil
}

func (s *clientResolutionStorage) InsertContact(context.Context, uuid.NullUUID, string, string, string) (uuid.UUID, error) {
	id := uuid.New()
	s.insertedContactID = id
	return id, nil
}

func (s *clientResolutionStorage) DeleteAddressAndContact(ctx context.Context, addressID uuid.UUID, contactID uuid.UUID) error {
	s.deleteAddressAndContactCalls++
	s.deletedAddressID = addressID
	s.deletedContactID = contactID
	// Captured at call time: the caller cancels the cleanup context via defer as
	// soon as CreateOrder returns, so inspecting it afterwards proves nothing.
	s.cleanupCtxErrAtCall = ctx.Err()
	_, s.cleanupCtxHadDeadline = ctx.Deadline()
	return s.deleteAddressAndContactErr
}

func (s *clientResolutionStorage) GetCounterpartyPriceGroupID(context.Context, uuid.NullUUID) (uuid.NullUUID, error) {
	return uuid.NullUUID{}, nil
}

func (s *clientResolutionStorage) GetVolumeDiscountPercent(context.Context, uuid.NullUUID, uuid.NullUUID, float64) (float64, error) {
	return 0, nil
}

func (s *clientResolutionStorage) ListOrders(_ context.Context, params postgres.ListOrdersParams) ([]postgres.OrderRow, int, error) {
	s.listOrdersCalls++
	s.listOrdersParams = params
	if s.listOrdersRows == nil {
		return []postgres.OrderRow{}, s.listOrdersTotal, s.listOrdersErr
	}
	return s.listOrdersRows, s.listOrdersTotal, s.listOrdersErr
}

func (s *clientResolutionStorage) GetOrderItems(context.Context, uuid.UUID) ([]postgres.OrderItemRow, error) {
	return []postgres.OrderItemRow{}, nil
}

func (s *clientResolutionStorage) GetOrderDocumentsByOrderID(context.Context, uuid.UUID) ([]postgres.OrderDocumentRow, error) {
	return []postgres.OrderDocumentRow{}, nil
}

type allowBuyerAccess struct{}

func (allowBuyerAccess) CheckAccess(context.Context, uuid.UUID, int) (bool, error) {
	return true, nil
}

func newClientResolutionService(storage Storage) *service {
	return NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 0.2, nil).(*service)
}

func requireOrdersError(t *testing.T, err error, status int, trKey string, causeKey, causeValue string) *customErrors.Error {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
	var customErr *customErrors.Error
	if !errors.As(err, &customErr) {
		t.Fatalf("error type = %T, want *errors.Error", err)
	}
	if customErr.GetStatusCode() != status {
		t.Fatalf("status = %d, want %d", customErr.GetStatusCode(), status)
	}
	if customErr.GetTranslationKey() != trKey {
		t.Fatalf("translation key = %q, want %q", customErr.GetTranslationKey(), trKey)
	}
	if causeKey != "" && customErr.Cause[causeKey] != causeValue {
		t.Fatalf("cause[%q] = %#v, want %q", causeKey, customErr.Cause[causeKey], causeValue)
	}
	return customErr
}

func TestResolveCounterpartyIDValidationAndAccess(t *testing.T) {
	userID := uuid.New()
	clientID := uuid.New()

	tests := []struct {
		name         string
		clientID     string
		storage      *clientResolutionStorage
		wantStatus   int
		wantKey      string
		wantCauseKey string
		wantCause    string
	}{
		{
			name:         "invalid UUID",
			clientID:     "not-a-uuid",
			storage:      &clientResolutionStorage{},
			wantStatus:   http.StatusBadRequest,
			wantKey:      customErrors.ErrBadRequest,
			wantCauseKey: "field",
			wantCause:    "clientID",
		},
		{
			name:         "nil UUID",
			clientID:     uuid.Nil.String(),
			storage:      &clientResolutionStorage{},
			wantStatus:   http.StatusBadRequest,
			wantKey:      customErrors.ErrBadRequest,
			wantCauseKey: "field",
			wantCause:    "clientID",
		},
		{
			name:         "client does not exist",
			clientID:     clientID.String(),
			storage:      &clientResolutionStorage{},
			wantStatus:   http.StatusNotFound,
			wantKey:      customErrors.ErrNotFound,
			wantCauseKey: "field",
			wantCause:    "clientID",
		},
		{
			name:     "client belongs to another user",
			clientID: clientID.String(),
			storage: &clientResolutionStorage{
				counterpartyExists: true,
			},
			wantStatus:   http.StatusForbidden,
			wantKey:      customErrors.ErrForbidden,
			wantCauseKey: "field",
			wantCause:    "clientID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newClientResolutionService(tt.storage).resolveCounterpartyID(
				context.Background(), userID, tt.clientID,
			)
			requireOrdersError(t, err, tt.wantStatus, tt.wantKey, tt.wantCauseKey, tt.wantCause)
		})
	}
}

func TestResolveCounterpartyIDUsesActiveClient(t *testing.T) {
	userID := uuid.New()
	activeClientID := uuid.New()
	storage := &clientResolutionStorage{
		activeClientID:     activeClientID,
		counterpartyExists: true,
		userHasClient:      true,
	}

	got, err := newClientResolutionService(storage).resolveCounterpartyID(context.Background(), userID, "")
	if err != nil {
		t.Fatalf("resolveCounterpartyID() error = %v", err)
	}
	if !got.Valid || got.UUID != activeClientID {
		t.Fatalf("client ID = %v, want active client %s", got, activeClientID)
	}
}

func TestResolveCounterpartyIDReturnsNoClientWhenUserHasNoActiveClient(t *testing.T) {
	userID := uuid.New()
	storage := &clientResolutionStorage{}

	got, err := newClientResolutionService(storage).resolveCounterpartyID(context.Background(), userID, "")
	if err != nil {
		t.Fatalf("resolveCounterpartyID() error = %v", err)
	}
	if got.Valid {
		t.Fatalf("client ID = %v, want no client", got)
	}
}

// Users without an active client (e.g. registered without a company, before the
// personal-client backfill ran) must still be able to use the cart and see their
// orders. resolveCounterpartyID treats a missing active client as "no client
// scoping" instead of failing the request.
func TestCartAndOrdersProceedWithoutActiveClient(t *testing.T) {
	userID := uuid.New()

	t.Run("cart", func(t *testing.T) {
		storage := &clientResolutionStorage{getCartErr: postgres.ErrCartNotFound}
		_, err := newClientResolutionService(storage).GetCart(context.Background(), userID, "")
		if err != nil {
			t.Fatalf("GetCart() error = %v", err)
		}
		if storage.getCartCounterpartyID.Valid {
			t.Fatalf("expected GetCart to be called with no counterparty scoping, got %v", storage.getCartCounterpartyID)
		}
		if storage.counterpartyExistsCalls != 0 || storage.userHasClientCalls != 0 {
			t.Fatalf(
				"expected no counterparty existence/ownership checks, got exists=%d userHasClient=%d",
				storage.counterpartyExistsCalls, storage.userHasClientCalls,
			)
		}
	})

	t.Run("orders", func(t *testing.T) {
		storage := &clientResolutionStorage{}
		_, err := newClientResolutionService(storage).ListOrders(context.Background(), userID, "", "", "", 20, 0, "")
		if err != nil {
			t.Fatalf("ListOrders() error = %v", err)
		}
		if storage.listOrdersParams.CounterpartyID.Valid {
			t.Fatalf("expected ListOrders to be called with no counterparty scoping, got %v", storage.listOrdersParams.CounterpartyID)
		}
	})
}

func TestCartAndOrdersRejectInvalidClientID(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name string
		call func(*service) error
	}{
		{
			name: "cart",
			call: func(svc *service) error {
				_, err := svc.GetCart(context.Background(), userID, "invalid")
				return err
			},
		},
		{
			name: "orders",
			call: func(svc *service) error {
				_, err := svc.ListOrders(context.Background(), userID, "invalid", "", "", 20, 0, "")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(newClientResolutionService(&clientResolutionStorage{}))
			requireOrdersError(
				t, err, http.StatusBadRequest, customErrors.ErrBadRequest, "field", "clientID",
			)
		})
	}
}

func TestCartAndOrdersReachRepositoryWithValidClient(t *testing.T) {
	userID := uuid.New()
	clientID := uuid.New()

	t.Run("cart", func(t *testing.T) {
		cartID := uuid.New()
		storage := &clientResolutionStorage{
			cartID:             cartID,
			counterpartyExists: true,
			userHasClient:      true,
		}
		response, err := newClientResolutionService(storage).GetCart(context.Background(), userID, clientID.String())
		if err != nil {
			t.Fatalf("GetCart() error = %v", err)
		}
		if response.Cart.ID != cartID.String() || response.Cart.Items == nil || len(response.Cart.Items) != 0 {
			t.Fatalf("GetCart() response = %#v, want existing empty cart", response.Cart)
		}
		if storage.getCartCalls != 1 || storage.getCartItemsCalls != 1 || storage.getOrCreateCartCalls != 0 {
			t.Fatalf(
				"cart repository calls: get=%d items=%d getOrCreate=%d",
				storage.getCartCalls,
				storage.getCartItemsCalls,
				storage.getOrCreateCartCalls,
			)
		}
	})

	t.Run("orders", func(t *testing.T) {
		storage := &clientResolutionStorage{counterpartyExists: true, userHasClient: true}
		response, err := newClientResolutionService(storage).ListOrders(
			context.Background(), userID, clientID.String(), "", "", 0, 0, "",
		)
		if err != nil {
			t.Fatalf("ListOrders() error = %v", err)
		}
		if storage.listOrdersCalls != 1 {
			t.Fatalf("ListOrders calls = %d, want 1", storage.listOrdersCalls)
		}
		if response.Items == nil || len(response.Items) != 0 {
			t.Fatalf("items = %#v, want non-nil empty list", response.Items)
		}
		if response.Pagination.Limit != defaultOrdersLimit || response.Pagination.Offset != 0 || response.Pagination.Total != 0 {
			t.Fatalf("pagination = %#v", response.Pagination)
		}
		if storage.listOrdersParams.Limit != defaultOrdersLimit {
			t.Fatalf("repository limit = %d, want %d", storage.listOrdersParams.Limit, defaultOrdersLimit)
		}
	})
}

func TestListOrdersReturnsExistingOrders(t *testing.T) {
	userID := uuid.New()
	clientID := uuid.New()
	orderID := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	storage := &clientResolutionStorage{
		counterpartyExists: true,
		userHasClient:      true,
		listOrdersRows: []postgres.OrderRow{{
			ID:             orderID,
			Number:         "AMC-1",
			Status:         "processing",
			PaymentStatus:  "not_paid",
			DeliveryMethod: "delivery",
			Subtotal:       100,
			DiscountTotal:  10,
			VATTotal:       19.8,
			Total:          109.8,
			CreatedAt:      createdAt,
		}},
		listOrdersTotal: 1,
	}

	response, err := newClientResolutionService(storage).ListOrders(
		context.Background(), userID, clientID.String(), "", "", 50, 0, "",
	)
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}
	if len(response.Items) != 1 {
		t.Fatalf("items = %#v, want one order", response.Items)
	}
	order := response.Items[0]
	if order.ID != orderID.String() || order.ClientID != clientID.String() || order.Number != "AMC-1" {
		t.Fatalf("order identity = %#v", order)
	}
	if order.Items == nil || order.Documents == nil {
		t.Fatalf("nested collections must be non-nil: %#v", order)
	}
	if response.Pagination.Limit != 50 || response.Pagination.Offset != 0 || response.Pagination.Total != 1 {
		t.Fatalf("pagination = %#v", response.Pagination)
	}
	if storage.listOrdersParams.Limit != 50 || storage.listOrdersParams.Offset != 0 {
		t.Fatalf("repository pagination = %#v", storage.listOrdersParams)
	}
}

func TestListOrdersRejectsInvalidPagination(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		offset     int
		causeField string
	}{
		{name: "negative limit", limit: -1, causeField: "limit"},
		{name: "limit above maximum", limit: maxOrdersLimit + 1, causeField: "limit"},
		{name: "negative offset", limit: 50, offset: -1, causeField: "offset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &clientResolutionStorage{}
			_, err := newClientResolutionService(storage).ListOrders(
				context.Background(), uuid.New(), "", "", "", tt.limit, tt.offset, "",
			)
			requireOrdersError(
				t, err, http.StatusBadRequest, customErrors.ErrBadRequest, "field", tt.causeField,
			)
			if storage.listOrdersCalls != 0 {
				t.Fatalf("repository called %d times for invalid pagination", storage.listOrdersCalls)
			}
		})
	}
}

func TestGetCartReturnsEmptyModelWhenCartDoesNotExist(t *testing.T) {
	userID := uuid.New()
	clientID := uuid.New()
	storage := &clientResolutionStorage{
		counterpartyExists: true,
		userHasClient:      true,
		getCartErr:         postgres.ErrCartNotFound,
	}

	response, err := newClientResolutionService(storage).GetCart(
		context.Background(), userID, clientID.String(),
	)
	if err != nil {
		t.Fatalf("GetCart() error = %v", err)
	}
	if response.Cart.ID != "" || response.Cart.UserID != userID.String() || response.Cart.ClientID != clientID.String() {
		t.Fatalf("cart identity = %#v", response.Cart)
	}
	if response.Cart.Items == nil || len(response.Cart.Items) != 0 {
		t.Fatalf("items = %#v, want non-nil empty list", response.Cart.Items)
	}
	if response.Cart.Subtotal != 0 || response.Cart.DiscountTotal != 0 || response.Cart.VAT != 0 || response.Cart.Total != 0 {
		t.Fatalf("empty cart totals = %#v", response.Cart)
	}
	if storage.getCartItemsCalls != 0 || storage.getOrCreateCartCalls != 0 {
		t.Fatalf(
			"missing cart triggered extra repository calls: items=%d getOrCreate=%d",
			storage.getCartItemsCalls,
			storage.getOrCreateCartCalls,
		)
	}
}

func TestCartAndOrdersMapRepositoryFailureToInternalError(t *testing.T) {
	userID := uuid.New()
	clientID := uuid.New().String()
	databaseErr := errors.New("database unavailable")

	tests := []struct {
		name    string
		storage *clientResolutionStorage
		call    func(*service) error
	}{
		{
			name: "cart",
			storage: &clientResolutionStorage{
				counterpartyExists: true,
				userHasClient:      true,
				getCartErr:         databaseErr,
			},
			call: func(svc *service) error {
				_, err := svc.GetCart(context.Background(), userID, clientID)
				return err
			},
		},
		{
			name: "orders",
			storage: &clientResolutionStorage{
				counterpartyExists: true,
				userHasClient:      true,
				listOrdersErr:      databaseErr,
			},
			call: func(svc *service) error {
				_, err := svc.ListOrders(context.Background(), userID, clientID, "", "", 20, 0, "")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customErr := requireOrdersError(
				t,
				tt.call(newClientResolutionService(tt.storage)),
				http.StatusInternalServerError,
				customErrors.ErrInternal,
				"",
				"",
			)
			if !errors.Is(customErr, databaseErr) {
				t.Fatalf("internal error does not wrap repository error: %v", customErr)
			}
		})
	}
}
