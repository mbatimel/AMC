package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/integrations/internal/errors"
)

// OrdersClient describes the subset of the orders service that the 1С
// webhook needs. It is a local interface — no concrete implementation is
// wired in yet; that happens once the generated orders client exists
// (see Task 14).
type OrdersClient interface {
	GetOrderStatus(ctx context.Context, userID, orderID uuid.UUID) (status string, err error)
	UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status, paymentStatus, comment, changedBy string) error
}

// allowedWebhookStatuses — единственный допустимый статус от 1С в v1.
// Расширять только через этот список + docs/superpowers/specs/2026-08-26-onec-orders-integration-1c-contract.md.
var allowedWebhookStatuses = map[string]bool{
	"delivered": true,
}

type OnecOrdersService struct {
	logger       zerolog.Logger
	ordersClient OrdersClient
	systemUserID uuid.UUID
	apiKey       string
}

func NewOnecOrdersService(logger zerolog.Logger, ordersClient OrdersClient, systemUserID uuid.UUID, apiKey string) *OnecOrdersService {
	return &OnecOrdersService{logger: logger, ordersClient: ordersClient, systemUserID: systemUserID, apiKey: apiKey}
}

func (s *OnecOrdersService) OnecOrderStatusWebhook(ctx context.Context, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) (ok bool, err error) {
	if apiKey != s.apiKey {
		return false, customErrors.UnauthorizedError()
	}
	if !allowedWebhookStatuses[status] {
		return false, customErrors.BadRequestError().AddCause("field", "status")
	}

	currentStatus, err := s.ordersClient.GetOrderStatus(ctx, s.systemUserID, clientOrderID)
	if err != nil {
		return false, err
	}

	switch currentStatus {
	case status:
		return true, nil
	case "cancelled":
		return false, customErrors.ConflictError().AddCause("field", "status")
	}

	fullComment := comment
	if onecDocumentNumber != "" {
		fullComment = "1С документ " + onecDocumentNumber + ": " + comment
	}
	if err = s.ordersClient.UpdateOrderStatus(ctx, s.systemUserID, clientOrderID, status, "", fullComment, ""); err != nil {
		return false, err
	}
	return true, nil
}
