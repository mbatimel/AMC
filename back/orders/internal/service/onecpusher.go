package service

import (
	"context"

	"github.com/google/uuid"
)

type OnecOrderItem struct {
	OneCGUID uuid.NullUUID
	SKU      string
	Name     string
	Qty      int
	Price    float64
	VATRate  float64
}

type OnecPushOrder struct {
	ClientOrderID    uuid.UUID
	OrderNumber      string
	CounterpartyGUID uuid.NullUUID
	CounterpartyINN  string
	CounterpartyName string
	DeliveryType     string
	DeliveryAddress  string
	ContactName      string
	Phone            string
	Email            string
	Comment          string
	Items            []OnecOrderItem
}

// OnecPusher is implemented by internal/onecclient.Client.
type OnecPusher interface {
	PushOrder(ctx context.Context, order OnecPushOrder) (onecGUID uuid.UUID, onecNumber string, err error)
}
