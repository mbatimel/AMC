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
	// @tg http-response=github.com/mbatimel/AMC/orders/orders/internal/transport/custom-handlers:GetCart
	// @tg summary=`Получение корзины`
	// @tg desc=`Получение текущей корзины клиента`
	GetCart(ctx context.Context, userID uuid.UUID, clientID string) (models.GetCartResponse, error)

	// 	// AddCartItem adds a product to the cart.
	// 	// @tg http-method=POST
	// 	// @tg http-path=/v1/cart/items
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Добавление товара в корзину`
	// 	// @tg desc=`Добавление позиции товара в корзину клиента`
	// 	AddCartItem(ctx context.Context,  userID uuid.UUID, requestmodels.AddCartItemRequest) (models.AddCartItemResponse, error)

	// 	// UpdateCartItem updates cart item quantity.
	// 	// @tg http-method=PATCH
	// 	// @tg http-path=/v1/cart/items/{id}
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Обновление позиции корзины`
	// 	// @tg desc=`Изменение количества товара в корзине`
	// 	UpdateCartItem(ctx context.Context,   userID uuid.UUID, requestmodels.UpdateCartItemRequest) (models.UpdateCartItemResponse, error)

	// 	// DeleteCartItem removes a cart item.
	// 	// @tg http-method=DELETE
	// 	// @tg http-path=/v1/cart/items/{id}
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Удаление позиции корзины`
	// 	// @tg desc=`Удаление позиции товара из корзины`
	// 	DeleteCartItem(ctx context.Context,   userID uuid.UUID, requestmodels.DeleteCartItemRequest) (models.DeleteCartItemResponse, error)

	// 	// ClearCart removes all cart items.
	// 	// @tg http-method=DELETE
	// 	// @tg http-path=/v1/cart
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Очистка корзины`
	// 	// @tg desc=`Удаление всех позиций из корзины клиента`
	// 	ClearCart(ctx context.Context, userID uuid.UUID, requestmodels.ClearCartRequest) (models.ClearCartResponse, error)

	// 	// CreateOrder creates an order from the cart.
	// 	// @tg http-method=POST
	// 	// @tg http-path=/v1/orders
	// 	// @tg summary=`Создание заказа`
	// 	// @tg desc=`Оформление заказа из корзины клиента`
	// 	CreateOrder(ctx context.Context,   userID uuid.UUID, requestmodels.CreateOrderRequest) (models.CreateOrderResponse, error)

	// 	// GetOrder returns one order by ID.
	// 	// @tg http-method=GET
	// 	// @tg http-path=/v1/orders/{id}
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Получение заказа`
	// 	// @tg desc=`Получение детальной информации по заказу`
	// 	GetOrder(ctx context.Context,  userID uuid.UUID, requestmodels.GetOrderRequest) (models.GetOrderResponse, error)

	// 	// ListOrders returns client orders.
	// 	// @tg http-method=GET
	// 	// @tg http-path=/v1/orders
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Список заказов`
	// 	// @tg desc=`Получение списка заказов клиента`
	// 	ListOrders(ctx context.Context,  userID uuid.UUID, requestmodels.ListOrdersRequest) (models.ListOrdersResponse, error)

	// 	// CancelOrder cancels an order.
	// 	// @tg http-method=POST
	// 	// @tg http-path=/v1/orders/{id}/cancel
	// 	// @tg summary=`Отмена заказа`
	// 	// @tg desc=`Отмена заказа клиентом или администратором`
	// 	CancelOrder(ctx context.Context,  userID uuid.UUID, requestmodels.CancelOrderRequest) (models.CancelOrderResponse, error)

	// 	// RepeatOrder repeats an order.
	// 	// @tg http-method=POST
	// 	// @tg http-path=/v1/orders/{id}/repeat
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Повтор заказа`
	// 	// @tg desc=`Создание корзины или заказа на основе существующего заказа`
	// 	RepeatOrder(ctx context.Context,  userID uuid.UUID, requestmodels.RepeatOrderRequest) (models.RepeatOrderResponse, error)

	// 	// GetOrderDocuments returns order documents.
	// 	// @tg http-method=GET
	// 	// @tg http-path=/v1/orders/{id}/documents
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`Документы заказа`
	// 	// @tg desc=`Получение документов по заказу`
	// 	GetOrderDocuments(ctx context.Context,  userID uuid.UUID, requestmodels.GetOrderDocumentsRequest) (models.GetOrderDocumentsResponse, error)

	// 	// GetOrderHistory returns order history.
	// 	// @tg http-method=GET
	// 	// @tg http-path=/v1/orders/{id}/history
	// 	// @tg http-headers=userID|X-User-Id
	// 	// @tg summary=`История заказа`
	// 	// @tg desc=`Получение истории изменения заказа`
	// 	GetOrderHistory(ctx context.Context,  userID uuid.UUID, requestmodels.GetOrderHistoryRequest) (models.GetOrderHistoryResponse, error)

	//		// UpdateOrderStatus updates order and payment statuses.
	//		// @tg http-method=PATCH
	//		// @tg http-path=/v1/admin/orders/{id}/status
	//		// @tg http-headers=userID|X-User-Id
	//		// @tg summary=`Обновление статуса заказа`
	//		// @tg desc=`Администраторское изменение статуса заказа и оплаты`
	//		UpdateOrderStatus(ctx context.Context,  userID uuid.UUID, requestmodels.UpdateOrderStatusRequest) (models.UpdateOrderStatusResponse, error)
	//	}
}
