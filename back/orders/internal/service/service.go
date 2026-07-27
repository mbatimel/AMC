package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/orders/internal/errors"
	"github.com/mbatimel/AMC/orders/internal/storage/postgres"
	externalapi "github.com/mbatimel/AMC/orders/pkg/interfaces/externalapi"
	"github.com/mbatimel/AMC/orders/pkg/models"
)

// Storage is implemented by internal/storage/postgres.Storage.
type Storage interface {
	GetCities(ctx context.Context) ([]postgres.City, error)
}

// AccessClient is implemented by internal/access.Client.
type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error)
}

type service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
}

func NewOrdersApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient) externalapi.OrdersAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
	}
}

// GetCities returns all cities from the cities table. Only buyers may call it.
func (s *service) GetCities(ctx context.Context) (response []models.GetCities, err error) {
	
	cities, err := s.storage.GetCities(ctx)
	if err != nil {
		return nil, customErrors.InternalServerError().SetOuterError(err)
	}

	response = make([]models.GetCities, 0, len(cities))
	for _, city := range cities {
		response = append(response, models.GetCities{
			ID:   city.ID,
			City: city.Name,
		})
	}

	return response, nil
}

func (s *service) GetCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.GetCartResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) AddCartItem(ctx context.Context, userID uuid.UUID, clientID string, productID string, qty int) (response models.AddCartItemResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) UpdateCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string, qty int) (response models.UpdateCartItemResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) DeleteCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string) (response models.DeleteCartItemResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) ClearCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.ClearCartResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) CreateOrder(ctx context.Context, userID uuid.UUID, clientID string, deliveryType string, deliveryAddress string, contactName string, phone string, email string, comment string) (response models.CreateOrderResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) GetOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, clientID string) (response models.GetOrderResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) ListOrders(ctx context.Context, userID uuid.UUID, clientID string, status string, paymentStatus string, limit int, offset int, sort string) (response models.ListOrdersResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) CancelOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, comment string) (response models.CancelOrderResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) RepeatOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, clientID string) (response models.RepeatOrderResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) GetOrderDocuments(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (response models.GetOrderDocumentsResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) GetOrderHistory(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (response models.GetOrderHistoryResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, paymentStatus string, comment string, changedBy string) (response models.UpdateOrderStatusResponse, err error) {
	return response, customErrors.NotImplementedError()
}
