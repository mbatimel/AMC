// Package externalapi describes the public cart and orders API contract.
// @tg version=0.0.1
// @tg backend=orders
// @tg title=`orders`
// @tg servers=
package externalapi

//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
import (
	"context"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/orders/pkg/models"
)

// OrdersAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/orders/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/orders/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/orders/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/orders/swaggers/externalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/orders/swaggers/externalapi/models:Err500
type OrdersAPI interface {
	// GetCart ...
	// @tg http-method=GET
	// @tg http-path=/v1/cart
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:GetCart
	// @tg summary=`Получение корзины`
	// @tg desc=`Получение текущей корзины клиента`
	// @tg uuidPackage=github.com/google/uuid
	GetCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.GetCartResponse, err error)

	// AddCartItem adds a product to the cart.
	// @tg http-method=POST
	// @tg http-path=/v1/cart/items
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-args=productID|productID
	// @tg http-args=qty|qty
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:AddCartItem
	// @tg summary=`Добавление товара в корзину`
	// @tg desc=`Добавление позиции товара в корзину клиента`
	// @tg uuidPackage=github.com/google/uuid
	AddCartItem(ctx context.Context, userID uuid.UUID, clientID string, productID string, qty int) (response models.AddCartItemResponse, err error)

	// UpdateCartItem updates cart item quantity.
	// @tg http-method=PATCH
	// @tg http-path=/v1/cart/items/{cartItemID}
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-args=cartItemID|cartItemID
	// @tg http-args=qty|qty
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:UpdateCartItem
	// @tg summary=`Обновление позиции корзины`
	// @tg desc=`Изменение количества товара в корзине`
	// @tg uuidPackage=github.com/google/uuid
	UpdateCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string, qty int) (response models.UpdateCartItemResponse, err error)

	// DeleteCartItem removes a cart item.
	// @tg http-method=DELETE
	// @tg http-path=/v1/cart/items/{cartItemID}
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-args=cartItemID|cartItemID
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:DeleteCartItem
	// @tg summary=`Удаление позиции корзины`
	// @tg desc=`Удаление позиции товара из корзины`
	// @tg uuidPackage=github.com/google/uuid
	DeleteCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string) (response models.DeleteCartItemResponse, err error)

	// ClearCart removes all cart items.
	// @tg http-method=DELETE
	// @tg http-path=/v1/cart
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:ClearCart
	// @tg summary=`Очистка корзины`
	// @tg desc=`Удаление всех позиций из корзины клиента`
	// @tg uuidPackage=github.com/google/uuid
	ClearCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.ClearCartResponse, err error)

	// CreateOrder creates an order from the cart.
	// @tg http-method=POST
	// @tg http-path=/v1/orders
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-args=deliveryType|deliveryType
	// @tg http-args=deliveryAddress|deliveryAddress
	// @tg http-args=contactName|contactName
	// @tg http-args=phone|phone
	// @tg http-args=email|email
	// @tg http-args=comment|comment
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:CreateOrder
	// @tg summary=`Создание заказа`
	// @tg desc=`Оформление заказа из корзины клиента`
	// @tg uuidPackage=github.com/google/uuid
	CreateOrder(ctx context.Context, userID uuid.UUID, clientID string, deliveryType string, deliveryAddress string, contactName string, phone string, email string, comment string) (response models.CreateOrderResponse, err error)

	// GetOrder returns one order by ID.
	// @tg http-method=GET
	// @tg http-path=/v1/orders/{orderID}
	// @tg http-args=orderID|orderID
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:GetOrder
	// @tg summary=`Получение заказа`
	// @tg desc=`Получение детальной информации по заказу`
	// @tg uuidPackage=github.com/google/uuid
	GetOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, clientID string) (response models.GetOrderResponse, err error)

	// ListOrders returns client orders.
	// @tg http-method=GET
	// @tg http-path=/v1/orders
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-args=status|status
	// @tg http-args=paymentStatus|paymentStatus
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-args=sort|sort
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:ListOrders
	// @tg summary=`Список заказов`
	// @tg desc=`Получение списка заказов клиента`
	// @tg uuidPackage=github.com/google/uuid
	ListOrders(ctx context.Context, userID uuid.UUID, clientID string, status string, paymentStatus string, limit int, offset int, sort string) (response models.ListOrdersResponse, err error)

	// CancelOrder cancels an order.
	// @tg http-method=POST
	// @tg http-path=/v1/orders/{orderID}/cancel
	// @tg http-args=orderID|orderID
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=comment|comment
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:CancelOrder
	// @tg summary=`Отмена заказа`
	// @tg desc=`Отмена заказа клиентом или администратором`
	// @tg uuidPackage=github.com/google/uuid
	CancelOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, comment string) (response models.CancelOrderResponse, err error)

	// RepeatOrder repeats an order.
	// @tg http-method=POST
	// @tg http-path=/v1/orders/{orderID}/repeat
	// @tg http-args=orderID|orderID
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=clientID|clientID
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:RepeatOrder
	// @tg summary=`Повтор заказа`
	// @tg desc=`Создание корзины или заказа на основе существующего заказа`
	// @tg uuidPackage=github.com/google/uuid
	RepeatOrder(ctx context.Context, orderID uuid.UUID, userID uuid.UUID, clientID string) (response models.RepeatOrderResponse, err error)

	// GetOrderDocuments returns order documents.
	// @tg http-method=GET
	// @tg http-path=/v1/orders/{orderID}/documents
	// @tg http-args=orderID|orderID
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:GetOrderDocuments
	// @tg summary=`Документы заказа`
	// @tg desc=`Получение документов по заказу`
	// @tg uuidPackage=github.com/google/uuid
	GetOrderDocuments(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (response models.GetOrderDocumentsResponse, err error)

	// GetOrderHistory returns order history.
	// @tg http-method=GET
	// @tg http-path=/v1/orders/{orderID}/history
	// @tg http-args=orderID|orderID
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:GetOrderHistory
	// @tg summary=`История заказа`
	// @tg desc=`Получение истории изменения заказа`
	// @tg uuidPackage=github.com/google/uuid
	GetOrderHistory(ctx context.Context, orderID uuid.UUID, userID uuid.UUID) (response models.GetOrderHistoryResponse, err error)

	// UpdateOrderStatus updates order and payment statuses.
	// @tg http-method=PATCH
	// @tg http-path=/v1/admin/orders/{orderID}/status
	// @tg http-args=orderID|orderID
	// @tg http-args=status|status
	// @tg http-args=paymentStatus|paymentStatus
	// @tg http-args=comment|comment
	// @tg http-args=changedBy|changedBy
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:UpdateOrderStatus
	// @tg summary=`Обновление статуса заказа`
	// @tg desc=`Администраторское изменение статуса заказа и оплаты`
	// @tg uuidPackage=github.com/google/uuid
	UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string, paymentStatus string, comment string, changedBy string) (response models.UpdateOrderStatusResponse, err error)
}
