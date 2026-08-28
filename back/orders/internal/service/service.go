package service

import (
	"context"
	"errors"
	"time"

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

	GetActiveClient(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	CounterpartyExists(ctx context.Context, clientID uuid.UUID) (bool, error)
	UserHasClient(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (bool, error)
	GetCounterpartyPriceGroupID(ctx context.Context, counterpartyID uuid.NullUUID) (uuid.NullUUID, error)
	InsertDeliveryAddress(ctx context.Context, counterpartyID uuid.NullUUID, addrType string, address string) (uuid.UUID, error)
	InsertContact(ctx context.Context, counterpartyID uuid.NullUUID, fullName string, phone string, email string) (uuid.UUID, error)
	DeleteAddressAndContact(ctx context.Context, addressID uuid.UUID, contactID uuid.UUID) error

	GetCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.NullUUID) (uuid.UUID, error)
	GetOrCreateCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.NullUUID) (uuid.UUID, error)
	GetCartItems(ctx context.Context, cartID uuid.UUID) ([]postgres.CartItemRow, error)
	ResolveProductPrice(ctx context.Context, productID uuid.UUID, priceGroupID uuid.NullUUID, qty int) (float64, error)
	UpsertCartItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, qty int, price float64) error
	SetCartItemQuantity(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID, qty int) error
	DeleteCartItem(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID) error
	ClearCartItems(ctx context.Context, cartID uuid.UUID) error

	GetVolumeDiscountPercent(ctx context.Context, counterpartyID uuid.NullUUID, priceGroupID uuid.NullUUID, subtotal float64) (float64, error)

	GetProductOnecRefs(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]postgres.ProductOnecRef, error)
	GetCounterpartyOnecRef(ctx context.Context, counterpartyID uuid.UUID) (postgres.CounterpartyOnecRef, error)

	CreateOrder(
		ctx context.Context,
		params postgres.CreateOrderParams,
		pushToOnec func(ctx context.Context, orderID uuid.UUID, orderNumber string) (onecGUID uuid.UUID, onecNumber string, err error),
	) (postgres.CreatedOrder, error)
	ListOrders(ctx context.Context, params postgres.ListOrdersParams) ([]postgres.OrderRow, int, error)
	ListPreviouslyOrderedProducts(ctx context.Context, params postgres.ListPreviouslyOrderedProductsParams) ([]postgres.PreviouslyOrderedProductRow, int, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, counterpartyID uuid.NullUUID) (postgres.OrderDetailRow, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]postgres.OrderItemRow, error)
	GetOrderDocumentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]postgres.OrderDocumentRow, error)
	GetOrderHistory(ctx context.Context, orderID uuid.UUID) ([]postgres.OrderHistoryRow, error)
	CancelOrder(ctx context.Context, orderID uuid.UUID, counterpartyID uuid.NullUUID, changedBy uuid.UUID, comment string) (postgres.OrderDetailRow, error)
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, paymentStatus string, comment string, changedBy uuid.UUID) (postgres.OrderDetailRow, error)
	GetOrderStatus(ctx context.Context, orderID uuid.UUID) (string, error)
}

const (
	defaultOrdersLimit = 20
	maxOrdersLimit     = 100

	// orphanCleanupTimeout bounds the best-effort deletion of the address/contact
	// rows left behind by a failed order creation. Short on purpose: the client is
	// already waiting on an error response.
	orphanCleanupTimeout = 5 * time.Second
)

// AccessClient is implemented by internal/access.Client.
type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error)
}

type service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
	vatRate      float64
	onecPusher   OnecPusher
}

func NewOrdersApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient, vatRate float64, onecPusher OnecPusher) externalapi.OrdersAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
		vatRate:      vatRate,
		onecPusher:   onecPusher,
	}
}

func (s *service) checkBuyerAccess(ctx context.Context, userID uuid.UUID) error {
	allowed, err := s.accessClient.CheckAccess(ctx, userID, RoleCodeBuyer)
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}
	if !allowed {
		return customErrors.ForbiddenError()
	}
	return nil
}

func (s *service) checkAdminAccess(ctx context.Context, userID uuid.UUID) error {
	allowed, err := s.accessClient.CheckAccess(ctx, userID, RoleCodeAdmin)
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}
	if !allowed {
		return customErrors.ForbiddenError()
	}
	return nil
}

// resolveCounterpartyID maps a request's clientID (or the user's active client)
// to a counterparty. A user with no active client yet (e.g. registered before the
// personal-client backfill ran) gets a not-Valid NullUUID instead of an error:
// callers proceed unscoped rather than failing the request.
// TODO: once every user is guaranteed to have a client, restore rejecting the
// no-client case here.
func (s *service) resolveCounterpartyID(ctx context.Context, userID uuid.UUID, clientID string) (uuid.NullUUID, error) {
	var id uuid.UUID
	var err error

	if clientID != "" {
		id, err = uuid.Parse(clientID)
		if err != nil || id == uuid.Nil {
			return uuid.NullUUID{}, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "clientID")
		}
	} else {
		id, err = s.storage.GetActiveClient(ctx, userID)
		if err != nil {
			if errors.Is(err, postgres.ErrUserNotFound) {
				return uuid.NullUUID{}, customErrors.NotFoundError().AddCause("entity", "user")
			}
			return uuid.NullUUID{}, customErrors.InternalServerError().SetOuterError(err)
		}
		if id == uuid.Nil {
			return uuid.NullUUID{}, nil
		}
	}

	exists, err := s.storage.CounterpartyExists(ctx, id)
	if err != nil {
		return uuid.NullUUID{}, customErrors.InternalServerError().SetOuterError(err)
	}
	if !exists {
		return uuid.NullUUID{}, customErrors.NotFoundError().AddCause("field", "clientID")
	}

	allowed, err := s.storage.UserHasClient(ctx, userID, id)
	if err != nil {
		return uuid.NullUUID{}, customErrors.InternalServerError().SetOuterError(err)
	}
	if !allowed {
		return uuid.NullUUID{}, customErrors.ForbiddenError().AddCause("field", "clientID")
	}

	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

func clientIDString(counterpartyID uuid.NullUUID) string {
	if !counterpartyID.Valid {
		return ""
	}
	return counterpartyID.UUID.String()
}

func emptyCart(userID uuid.UUID, counterpartyID uuid.NullUUID) models.Cart {
	return models.Cart{
		UserID:   userID.String(),
		ClientID: clientIDString(counterpartyID),
		Items:    make([]models.CartItem, 0),
	}
}

func (s *service) buildCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.NullUUID) (models.Cart, error) {
	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrCounterpartyNotFound) {
			return models.Cart{}, customErrors.NotFoundError().AddCause("field", "clientID")
		}
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}
	return s.buildExistingCart(ctx, userID, counterpartyID, cartID)
}

func (s *service) buildExistingCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.NullUUID, cartID uuid.UUID) (models.Cart, error) {
	rows, err := s.storage.GetCartItems(ctx, cartID)
	if err != nil {
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}

	priceGroupID, err := s.storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}

	calcItems := make([]CartCalcItem, 0, len(rows))
	items := make([]models.CartItem, 0, len(rows))
	for _, row := range rows {
		lineTotal := round2(float64(row.Qty) * row.Price)
		calcItems = append(calcItems, CartCalcItem{Qty: row.Qty, Price: row.Price})
		items = append(items, models.CartItem{
			ID:          row.ID.String(),
			CartID:      cartID.String(),
			ProductID:   row.ProductID.String(),
			SKU:         row.SKU,
			ProductName: row.ProductName,
			Qty:         row.Qty,
			Price:       row.Price,
			Total:       lineTotal,
		})
	}

	subtotal := sumLineTotals(calcItems)

	discountPercent, err := s.storage.GetVolumeDiscountPercent(ctx, counterpartyID, priceGroupID, subtotal)
	if err != nil {
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}

	totals := calcCartTotals(subtotal, discountPercent, s.vatRate)

	return models.Cart{
		ID:            cartID.String(),
		UserID:        userID.String(),
		ClientID:      clientIDString(counterpartyID),
		Items:         items,
		Subtotal:      totals.Subtotal,
		DiscountTotal: totals.DiscountTotal,
		VAT:           totals.VATTotal,
		Total:         totals.Total,
	}, nil
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
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetCart(ctx, userID, counterpartyID)
	if errors.Is(err, postgres.ErrCartNotFound) {
		return models.GetCartResponse{Cart: emptyCart(userID, counterpartyID)}, nil
	}
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildExistingCart(ctx, userID, counterpartyID, cartID)
	if err != nil {
		return response, err
	}

	return models.GetCartResponse{Cart: cart}, nil
}

func (s *service) AddCartItem(ctx context.Context, userID uuid.UUID, clientID string, productID string, qty int) (response models.AddCartItemResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if qty <= 0 {
		return response, customErrors.BadRequestError().AddCause("field", "qty")
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return response, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "productID")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrCounterpartyNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "clientID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	priceGroupID, err := s.storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	price, err := s.storage.ResolveProductPrice(ctx, productUUID, priceGroupID, qty)
	if err != nil {
		if errors.Is(err, postgres.ErrProductPriceNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "productID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.UpsertCartItem(ctx, cartID, productUUID, qty, price); err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.AddCartItemResponse{Cart: cart}, nil
}

func (s *service) UpdateCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string, qty int) (response models.UpdateCartItemResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if qty <= 0 {
		return response, customErrors.BadRequestError().AddCause("field", "qty")
	}

	itemUUID, err := uuid.Parse(cartItemID)
	if err != nil {
		return response, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "cartItemID")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrCounterpartyNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "clientID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.SetCartItemQuantity(ctx, itemUUID, cartID, qty); err != nil {
		if errors.Is(err, postgres.ErrCartItemNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "cartItemID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.UpdateCartItemResponse{Cart: cart}, nil
}

func (s *service) DeleteCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string) (response models.DeleteCartItemResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	itemUUID, err := uuid.Parse(cartItemID)
	if err != nil {
		return response, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "cartItemID")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrCounterpartyNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "clientID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.DeleteCartItem(ctx, itemUUID, cartID); err != nil {
		if errors.Is(err, postgres.ErrCartItemNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "cartItemID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.DeleteCartItemResponse{Deleted: true, Cart: cart}, nil
}

func (s *service) ClearCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.ClearCartResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrCounterpartyNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "clientID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.ClearCartItems(ctx, cartID); err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.ClearCartResponse{Cleared: true, Cart: cart}, nil
}

func (s *service) CreateOrder(ctx context.Context, userID uuid.UUID, clientID string, deliveryType string, deliveryAddress string, contactName string, phone string, email string, comment string) (response models.CreateOrderResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if deliveryType == "" {
		return response, customErrors.BadRequestError().AddCause("field", "deliveryType")
	}
	if deliveryAddress == "" {
		return response, customErrors.BadRequestError().AddCause("field", "deliveryAddress")
	}
	if contactName == "" {
		return response, customErrors.BadRequestError().AddCause("field", "contactName")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrCounterpartyNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "clientID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cartRows, err := s.storage.GetCartItems(ctx, cartID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}
	if len(cartRows) == 0 {
		return response, customErrors.BadRequestError().AddCause("field", "cart")
	}

	priceGroupID, err := s.storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	calcItems := make([]CartCalcItem, 0, len(cartRows))
	for _, row := range cartRows {
		calcItems = append(calcItems, CartCalcItem{Qty: row.Qty, Price: row.Price})
	}
	subtotal := sumLineTotals(calcItems)

	discountPercent, err := s.storage.GetVolumeDiscountPercent(ctx, counterpartyID, priceGroupID, subtotal)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	totals := calcCartTotals(subtotal, discountPercent, s.vatRate)

	addressID, err := s.storage.InsertDeliveryAddress(ctx, counterpartyID, deliveryType, deliveryAddress)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	contactID, err := s.storage.InsertContact(ctx, counterpartyID, contactName, phone, email)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	orderItems := make([]postgres.OrderItemInput, 0, len(cartRows))
	responseItems := make([]models.OrderItem, 0, len(cartRows))
	for _, row := range cartRows {
		lineTotal := round2(float64(row.Qty) * row.Price)
		orderItems = append(orderItems, postgres.OrderItemInput{
			ProductID:       row.ProductID,
			SKU:             row.SKU,
			Name:            row.ProductName,
			Quantity:        row.Qty,
			UnitPrice:       row.Price,
			DiscountPercent: discountPercent,
			VATRate:         s.vatRate,
			LineTotal:       lineTotal,
		})
		responseItems = append(responseItems, models.OrderItem{
			ProductID:   row.ProductID.String(),
			SKU:         row.SKU,
			ProductName: row.ProductName,
			Qty:         row.Qty,
			Price:       row.Price,
			Total:       lineTotal,
		})
	}

	productIDs := make([]uuid.UUID, 0, len(cartRows))
	for _, row := range cartRows {
		productIDs = append(productIDs, row.ProductID)
	}
	productRefs, err := s.storage.GetProductOnecRefs(ctx, productIDs)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}
	var counterpartyRef postgres.CounterpartyOnecRef
	if counterpartyID.Valid {
		counterpartyRef, err = s.storage.GetCounterpartyOnecRef(ctx, counterpartyID.UUID)
		if err != nil {
			return response, customErrors.InternalServerError().SetOuterError(err)
		}
	}

	pushItems := make([]OnecOrderItem, 0, len(cartRows))
	for _, row := range cartRows {
		ref := productRefs[row.ProductID]
		pushItems = append(pushItems, OnecOrderItem{
			OneCGUID: ref.OneCGUID,
			SKU:      row.SKU,
			Name:     row.ProductName,
			Qty:      row.Qty,
			Price:    row.Price,
			VATRate:  s.vatRate,
		})
	}

	created, err := s.storage.CreateOrder(ctx, postgres.CreateOrderParams{
		UserID:            userID,
		CounterpartyID:    counterpartyID,
		CartID:            cartID,
		DeliveryMethod:    deliveryType,
		DeliveryAddressID: addressID,
		ContactID:         contactID,
		Comment:           comment,
		Subtotal:          totals.Subtotal,
		DiscountTotal:     totals.DiscountTotal,
		VATTotal:          totals.VATTotal,
		Total:             totals.Total,
		Items:             orderItems,
	}, func(ctx context.Context, orderID uuid.UUID, orderNumber string) (uuid.UUID, string, error) {
		return s.onecPusher.PushOrder(ctx, OnecPushOrder{
			ClientOrderID:    orderID,
			OrderNumber:      orderNumber,
			CounterpartyGUID: counterpartyRef.OneCGUID,
			CounterpartyINN:  counterpartyRef.INN,
			CounterpartyName: counterpartyRef.Name,
			DeliveryType:     deliveryType,
			DeliveryAddress:  deliveryAddress,
			ContactName:      contactName,
			Phone:            phone,
			Email:            email,
			Comment:          comment,
			Items:            pushItems,
		})
	})
	if err != nil {
		// InsertDeliveryAddress/InsertContact above committed on their own,
		// outside storage.CreateOrder's transaction, so its rollback leaves them
		// behind as orphans attached to the counterparty with no order referencing
		// them. Clean them up best-effort; a cleanup failure is logged, never
		// returned — the caller needs the original (push) error.
		//
		// The cleanup runs on a context detached from cancellation: the most
		// common way to get here is the 1С push timing out, which leaves ctx
		// already expired — reusing it would make the cleanup fail exactly when
		// orphans are being produced fastest.
		cleanupCtx, cancelCleanup := context.WithTimeout(context.WithoutCancel(ctx), orphanCleanupTimeout)
		defer cancelCleanup()

		if cleanupErr := s.storage.DeleteAddressAndContact(cleanupCtx, addressID, contactID); cleanupErr != nil {
			s.logger.Error().
				Err(cleanupErr).
				Str("addressID", addressID.String()).
				Str("contactID", contactID.String()).
				Msg("failed to clean up orphan address/contact after failed order creation")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	order := models.Order{
		ID:              created.ID.String(),
		Number:          created.Number,
		UserID:          userID.String(),
		ClientID:        clientIDString(counterpartyID),
		Items:           responseItems,
		Subtotal:        totals.Subtotal,
		DiscountTotal:   totals.DiscountTotal,
		Total:           totals.Total,
		VAT:             totals.VATTotal,
		DeliveryType:    deliveryType,
		DeliveryAddress: deliveryAddress,
		ContactName:     contactName,
		Phone:           phone,
		Email:           email,
		Comment:         comment,
		Status:          created.Status,
		PaymentStatus:   "not_paid",
		CreatedAt:       created.CreatedAt,
	}

	return models.CreateOrderResponse{Order: order}, nil
}

func (s *service) buildOrderModel(ctx context.Context, row postgres.OrderDetailRow) (models.Order, error) {
	itemRows, err := s.storage.GetOrderItems(ctx, row.ID)
	if err != nil {
		return models.Order{}, customErrors.InternalServerError().SetOuterError(err)
	}

	docRows, err := s.storage.GetOrderDocumentsByOrderID(ctx, row.ID)
	if err != nil {
		return models.Order{}, customErrors.InternalServerError().SetOuterError(err)
	}

	items := make([]models.OrderItem, 0, len(itemRows))
	for _, item := range itemRows {
		items = append(items, models.OrderItem{
			ID:          item.ID.String(),
			OrderID:     row.ID.String(),
			ProductID:   item.ProductID.String(),
			SKU:         item.SKU,
			ProductName: item.Name,
			Qty:         item.Quantity,
			Price:       item.UnitPrice,
			Total:       item.LineTotal,
		})
	}

	docs := make([]models.OrderDocument, 0, len(docRows))
	for _, doc := range docRows {
		docs = append(docs, models.OrderDocument{
			ID:        doc.ID.String(),
			OrderID:   row.ID.String(),
			Type:      doc.Type,
			Name:      doc.Number,
			URL:       doc.URL,
			CreatedAt: doc.CreatedAt,
		})
	}

	return models.Order{
		ID:              row.ID.String(),
		Number:          row.Number,
		UserID:          row.UserID.String(),
		ClientID:        clientIDString(row.CounterpartyID),
		Items:           items,
		Subtotal:        row.Subtotal,
		DiscountTotal:   row.DiscountTotal,
		Total:           row.Total,
		VAT:             row.VATTotal,
		DeliveryType:    row.DeliveryMethod,
		DeliveryAddress: row.DeliveryAddress,
		ContactName:     row.ContactName,
		Phone:           row.Phone,
		Email:           row.Email,
		Comment:         row.Comment,
		Status:          row.Status,
		PaymentStatus:   row.PaymentStatus,
		Documents:       docs,
		CreatedAt:       row.CreatedAt,
	}, nil
}

func (s *service) GetOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, clientID string) (response models.GetOrderResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	row, err := s.storage.GetOrderByID(ctx, orderID, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	order, err := s.buildOrderModel(ctx, row)
	if err != nil {
		return response, err
	}

	return models.GetOrderResponse{Order: order}, nil
}

func (s *service) ListOrders(ctx context.Context, userID uuid.UUID, clientID string, status string, paymentStatus string, limit int, offset int, sort string) (response models.ListOrdersResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}
	if limit <= 0 {
		limit = defaultOrdersLimit
	}

	rows, total, err := s.storage.ListOrders(ctx, postgres.ListOrdersParams{
		UserID:         userID,
		CounterpartyID: counterpartyID,
		Status:         status,
		PaymentStatus:  paymentStatus,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
	})
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	orders := make([]models.Order, 0, len(rows))
	for _, row := range rows {
		itemRows, itemsErr := s.storage.GetOrderItems(ctx, row.ID)
		if itemsErr != nil {
			return response, customErrors.InternalServerError().SetOuterError(itemsErr)
		}

		docRows, docsErr := s.storage.GetOrderDocumentsByOrderID(ctx, row.ID)
		if docsErr != nil {
			return response, customErrors.InternalServerError().SetOuterError(docsErr)
		}

		items := make([]models.OrderItem, 0, len(itemRows))
		for _, item := range itemRows {
			items = append(items, models.OrderItem{
				ID:          item.ID.String(),
				OrderID:     row.ID.String(),
				ProductID:   item.ProductID.String(),
				SKU:         item.SKU,
				ProductName: item.Name,
				Qty:         item.Quantity,
				Price:       item.UnitPrice,
				Total:       item.LineTotal,
			})
		}

		docs := make([]models.OrderDocument, 0, len(docRows))
		for _, doc := range docRows {
			docs = append(docs, models.OrderDocument{
				ID:        doc.ID.String(),
				OrderID:   row.ID.String(),
				Type:      doc.Type,
				Name:      doc.Number,
				URL:       doc.URL,
				CreatedAt: doc.CreatedAt,
			})
		}

		orders = append(orders, models.Order{
			ID:            row.ID.String(),
			Number:        row.Number,
			ClientID:      clientIDString(counterpartyID),
			Items:         items,
			Subtotal:      row.Subtotal,
			DiscountTotal: row.DiscountTotal,
			Total:         row.Total,
			VAT:           row.VATTotal,
			DeliveryType:  row.DeliveryMethod,
			Status:        row.Status,
			PaymentStatus: row.PaymentStatus,
			Documents:     docs,
			CreatedAt:     row.CreatedAt,
		})
	}

	return models.ListOrdersResponse{
		Items: orders,
		Pagination: models.Pagination{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}, nil
}

func (s *service) ListPreviouslyOrderedProducts(ctx context.Context, userID uuid.UUID, clientID string, limit int, offset int) (response models.ListPreviouslyOrderedProductsResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if limit == 0 {
		limit = defaultOrdersLimit
	}
	if limit < 0 || limit > maxOrdersLimit {
		return response, customErrors.BadRequestError().AddCause("field", "limit")
	}
	if offset < 0 {
		return response, customErrors.BadRequestError().AddCause("field", "offset")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}
	rows, total, err := s.storage.ListPreviouslyOrderedProducts(ctx, postgres.ListPreviouslyOrderedProductsParams{
		UserID:         userID,
		CounterpartyID: counterpartyID,
		Limit:          limit,
		Offset:         offset,
	})
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	items := make([]models.PreviouslyOrderedProduct, 0, len(rows))
	for _, row := range rows {
		items = append(items, models.PreviouslyOrderedProduct{
			ProductID:     row.ProductID.String(),
			LastOrderedAt: row.LastOrderedAt,
		})
	}
	return models.ListPreviouslyOrderedProductsResponse{
		Items: items,
		Pagination: models.Pagination{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}, nil
}

func (s *service) CancelOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, comment string) (response models.CancelOrderResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, "")
	if err != nil {
		return response, err
	}

	row, err := s.storage.CancelOrder(ctx, orderID, counterpartyID, userID, comment)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	order, err := s.buildOrderModel(ctx, row)
	if err != nil {
		return response, err
	}

	return models.CancelOrderResponse{Order: order}, nil
}

func (s *service) RepeatOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, clientID string) (response models.RepeatOrderResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	row, err := s.storage.GetOrderByID(ctx, orderID, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	itemRows, err := s.storage.GetOrderItems(ctx, row.ID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrCounterpartyNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "clientID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	priceGroupID, err := s.storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	for _, item := range itemRows {
		price, priceErr := s.storage.ResolveProductPrice(ctx, item.ProductID, priceGroupID, item.Quantity)
		if priceErr != nil {
			if errors.Is(priceErr, postgres.ErrProductPriceNotFound) {
				continue
			}
			return response, customErrors.InternalServerError().SetOuterError(priceErr)
		}
		if err = s.storage.UpsertCartItem(ctx, cartID, item.ProductID, item.Quantity, price); err != nil {
			return response, customErrors.InternalServerError().SetOuterError(err)
		}
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	order, err := s.buildOrderModel(ctx, row)
	if err != nil {
		return response, err
	}

	return models.RepeatOrderResponse{Order: order, Cart: cart}, nil
}

func (s *service) GetOrderDocuments(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (response models.GetOrderDocumentsResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, "")
	if err != nil {
		return response, err
	}

	row, err := s.storage.GetOrderByID(ctx, orderID, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	docRows, err := s.storage.GetOrderDocumentsByOrderID(ctx, row.ID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	docs := make([]models.OrderDocument, 0, len(docRows))
	for _, doc := range docRows {
		docs = append(docs, models.OrderDocument{
			ID:        doc.ID.String(),
			OrderID:   row.ID.String(),
			Type:      doc.Type,
			Name:      doc.Number,
			URL:       doc.URL,
			CreatedAt: doc.CreatedAt,
		})
	}

	return models.GetOrderDocumentsResponse{Items: docs}, nil
}

func (s *service) GetOrderHistory(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (response models.GetOrderHistoryResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, "")
	if err != nil {
		return response, err
	}

	row, err := s.storage.GetOrderByID(ctx, orderID, userID, counterpartyID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	historyRows, err := s.storage.GetOrderHistory(ctx, row.ID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	items := make([]models.OrderHistoryItem, 0, len(historyRows))
	for _, h := range historyRows {
		changedBy := ""
		if h.ChangedBy.Valid {
			changedBy = h.ChangedBy.UUID.String()
		}
		items = append(items, models.OrderHistoryItem{
			ID:            h.ID.String(),
			OrderID:       h.OrderID.String(),
			Status:        h.Status,
			PaymentStatus: h.PaymentStatus,
			Comment:       h.Comment,
			ChangedBy:     changedBy,
			CreatedAt:     h.CreatedAt,
		})
	}

	return models.GetOrderHistoryResponse{Items: items}, nil
}

func (s *service) UpdateOrderStatus(ctx context.Context, userID uuid.UUID, orderID uuid.UUID, status string, paymentStatus string, comment string, changedBy string) (response models.UpdateOrderStatusResponse, err error) {
	if err = s.checkAdminAccess(ctx, userID); err != nil {
		return response, err
	}

	// changedBy is intentionally ignored: the actor recorded in history must be
	// the authenticated caller, not a client-supplied value (avoids attribution spoofing).
	row, err := s.storage.UpdateOrderStatus(ctx, orderID, status, paymentStatus, comment, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	order, err := s.buildOrderModel(ctx, row)
	if err != nil {
		return response, err
	}

	return models.UpdateOrderStatusResponse{Order: order}, nil
}

func (s *service) GetOrderStatus(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (response models.GetOrderStatusResponse, err error) {
	if err = s.checkAdminAccess(ctx, userID); err != nil {
		return response, err
	}
	status, err := s.storage.GetOrderStatus(ctx, orderID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}
	return models.GetOrderStatusResponse{Status: status}, nil
}
