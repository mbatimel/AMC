package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	customErrors "github.com/mbatimel/AMC/orders/internal/errors"
	"github.com/mbatimel/AMC/orders/internal/storage/postgres"
)

type previouslyOrderedStorage struct {
	*clientResolutionStorage
	rows   []postgres.PreviouslyOrderedProductRow
	total  int
	err    error
	params postgres.ListPreviouslyOrderedProductsParams
	calls  int
}

func (s *previouslyOrderedStorage) ListPreviouslyOrderedProducts(_ context.Context, params postgres.ListPreviouslyOrderedProductsParams) ([]postgres.PreviouslyOrderedProductRow, int, error) {
	s.calls++
	s.params = params
	return s.rows, s.total, s.err
}

func TestListPreviouslyOrderedProductsEmpty(t *testing.T) {
	userID, clientID := uuid.New(), uuid.New()
	storage := &previouslyOrderedStorage{
		clientResolutionStorage: &clientResolutionStorage{counterpartyExists: true, userHasClient: true},
	}

	response, err := newClientResolutionService(storage).ListPreviouslyOrderedProducts(
		context.Background(), userID, clientID.String(), 0, 0,
	)
	if err != nil {
		t.Fatalf("ListPreviouslyOrderedProducts() error = %v", err)
	}
	if response.Items == nil || len(response.Items) != 0 {
		t.Fatalf("items = %#v, want non-nil empty list", response.Items)
	}
	if response.Pagination.Limit != defaultOrdersLimit || response.Pagination.Total != 0 {
		t.Fatalf("pagination = %#v", response.Pagination)
	}
	if storage.params.UserID != userID || !storage.params.CounterpartyID.Valid || storage.params.CounterpartyID.UUID != clientID {
		t.Fatalf("repository scope = %#v", storage.params)
	}
}

func TestListPreviouslyOrderedProductsMapsRepositoryRows(t *testing.T) {
	userID, clientID := uuid.New(), uuid.New()
	firstProductID, secondProductID := uuid.New(), uuid.New()
	firstOrderedAt := time.Now().UTC()
	secondOrderedAt := firstOrderedAt.Add(-time.Hour)
	storage := &previouslyOrderedStorage{
		clientResolutionStorage: &clientResolutionStorage{counterpartyExists: true, userHasClient: true},
		rows: []postgres.PreviouslyOrderedProductRow{
			{ProductID: firstProductID, LastOrderedAt: firstOrderedAt},
			{ProductID: secondProductID, LastOrderedAt: secondOrderedAt},
		},
		total: 7,
	}

	response, err := newClientResolutionService(storage).ListPreviouslyOrderedProducts(
		context.Background(), userID, clientID.String(), 2, 3,
	)
	if err != nil {
		t.Fatalf("ListPreviouslyOrderedProducts() error = %v", err)
	}
	if len(response.Items) != 2 || response.Items[0].ProductID != firstProductID.String() ||
		response.Items[0].LastOrderedAt != firstOrderedAt || response.Items[1].ProductID != secondProductID.String() {
		t.Fatalf("items = %#v", response.Items)
	}
	if response.Pagination.Limit != 2 || response.Pagination.Offset != 3 || response.Pagination.Total != 7 {
		t.Fatalf("pagination = %#v", response.Pagination)
	}
}

func TestListPreviouslyOrderedProductsValidatesPagination(t *testing.T) {
	tests := []struct {
		name       string
		limit      int
		offset     int
		causeField string
	}{
		{name: "negative limit", limit: -1, causeField: "limit"},
		{name: "over max limit", limit: maxOrdersLimit + 1, causeField: "limit"},
		{name: "negative offset", offset: -1, causeField: "offset"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &previouslyOrderedStorage{clientResolutionStorage: &clientResolutionStorage{}}
			_, err := newClientResolutionService(storage).ListPreviouslyOrderedProducts(
				context.Background(), uuid.New(), "", tt.limit, tt.offset,
			)
			requireOrdersError(t, err, http.StatusBadRequest, customErrors.ErrBadRequest, "field", tt.causeField)
			if storage.calls != 0 {
				t.Fatalf("repository calls = %d, want 0", storage.calls)
			}
		})
	}
}

func TestListPreviouslyOrderedProductsRejectsForeignClient(t *testing.T) {
	storage := &previouslyOrderedStorage{
		clientResolutionStorage: &clientResolutionStorage{
			counterpartyExists: true,
			userHasClient:      false,
		},
	}

	_, err := newClientResolutionService(storage).ListPreviouslyOrderedProducts(
		context.Background(), uuid.New(), uuid.NewString(), 10, 0,
	)
	requireOrdersError(t, err, http.StatusForbidden, customErrors.ErrForbidden, "field", "clientID")
	if storage.calls != 0 {
		t.Fatalf("repository calls = %d, want 0", storage.calls)
	}
}

func TestListPreviouslyOrderedProductsRepositoryError(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	storage := &previouslyOrderedStorage{
		clientResolutionStorage: &clientResolutionStorage{
			activeClientID:     uuid.New(),
			counterpartyExists: true,
			userHasClient:      true,
		},
		err: databaseErr,
	}

	_, err := newClientResolutionService(storage).ListPreviouslyOrderedProducts(
		context.Background(), uuid.New(), "", 10, 0,
	)
	customErr := requireOrdersError(t, err, http.StatusInternalServerError, customErrors.ErrInternal, "", "")
	if !errors.Is(customErr, databaseErr) {
		t.Fatalf("error %v does not wrap repository error %v", customErr, databaseErr)
	}
}
