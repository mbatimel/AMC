package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type fakeOrdersClient struct {
	status       string
	statusErr    error
	updateErr    error
	updateCalled bool
	lastUserID   uuid.UUID
}

func (f *fakeOrdersClient) GetOrderStatus(ctx context.Context, userID, orderID uuid.UUID) (string, error) {
	return f.status, f.statusErr
}
func (f *fakeOrdersClient) UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status, paymentStatus, comment, changedBy string) error {
	f.updateCalled = true
	f.lastUserID = userID
	return f.updateErr
}

const testSystemUserID = "00000000-0000-0000-0000-0000000a0ec1"
const testAPIKey = "test-api-key"

func TestOnecOrderStatusWebhook_WrongAPIKey_Unauthorized(t *testing.T) {
	client := &fakeOrdersClient{status: "processing"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), "wrong-key", uuid.New(), "delivered", "УТ-1", "")
	if err == nil {
		t.Fatal("expected error for wrong api key")
	}
}

func TestOnecOrderStatusWebhook_InvalidStatus_BadRequest(t *testing.T) {
	client := &fakeOrdersClient{status: "processing"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "cancelled", "УТ-1", "")
	if err == nil {
		t.Fatal("expected error for disallowed status")
	}
	if client.updateCalled {
		t.Fatal("expected UpdateOrderStatus not to be called for invalid status")
	}
}

func TestOnecOrderStatusWebhook_OrderCancelled_Conflict(t *testing.T) {
	client := &fakeOrdersClient{status: "cancelled"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "")
	if err == nil {
		t.Fatal("expected conflict error for cancelled order")
	}
	if client.updateCalled {
		t.Fatal("expected UpdateOrderStatus not to be called for cancelled order")
	}
}

func TestOnecOrderStatusWebhook_AlreadyDelivered_IdempotentNoOp(t *testing.T) {
	client := &fakeOrdersClient{status: "delivered"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	ok, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if client.updateCalled {
		t.Fatal("expected UpdateOrderStatus not to be called when already delivered")
	}
}

func TestOnecOrderStatusWebhook_Processing_AppliesTransition(t *testing.T) {
	client := &fakeOrdersClient{status: "processing"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	ok, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "Вручено")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !client.updateCalled {
		t.Fatal("expected UpdateOrderStatus to be called")
	}
	if client.lastUserID.String() != testSystemUserID {
		t.Fatalf("expected system user id %s, got %s", testSystemUserID, client.lastUserID)
	}
}

func TestOnecOrderStatusWebhook_OrderNotFound(t *testing.T) {
	client := &fakeOrdersClient{statusErr: errors.New("order not found: 404")}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "")
	if err == nil {
		t.Fatal("expected error when order status lookup fails")
	}
}
