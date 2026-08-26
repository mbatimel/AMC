package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/orders/internal/errors"
	"github.com/mbatimel/AMC/orders/internal/storage/postgres"
)

// denyAccess is a local AccessClient fake that always denies, used to exercise
// the forbidden-case branch of checkAdminAccess.
type denyAccess struct{}

func (denyAccess) CheckAccess(context.Context, uuid.UUID, int) (bool, error) {
	return false, nil
}

func TestGetOrderStatus_AdminAllowed(t *testing.T) {
	storage := &clientResolutionStorage{orderStatus: "processing"}
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, nil)

	resp, err := svc.GetOrderStatus(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "processing" {
		t.Fatalf("expected status processing, got %q", resp.Status)
	}
}

func TestGetOrderStatus_NonAdminForbidden(t *testing.T) {
	storage := &clientResolutionStorage{orderStatus: "processing"}
	svc := NewOrdersApiService(zerolog.Nop(), storage, denyAccess{}, 20, nil)

	_, err := svc.GetOrderStatus(context.Background(), uuid.New(), uuid.New())
	var custErr *customErrors.Error
	if !errors.As(err, &custErr) || custErr.GetStatusCode() != 403 {
		t.Fatalf("expected 403 forbidden, got %v", err)
	}
}

func TestGetOrderStatus_NotFound(t *testing.T) {
	storage := &clientResolutionStorage{orderStatusErr: postgres.ErrOrderNotFound}
	svc := NewOrdersApiService(zerolog.Nop(), storage, allowBuyerAccess{}, 20, nil)

	_, err := svc.GetOrderStatus(context.Background(), uuid.New(), uuid.New())
	var custErr *customErrors.Error
	if !errors.As(err, &custErr) || custErr.GetStatusCode() != 404 {
		t.Fatalf("expected 404 not found, got %v", err)
	}
}
