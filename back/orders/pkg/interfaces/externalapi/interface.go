// Package externalapi describes the public cart and orders API contract.
// @tg version=0.0.1
// @tg backend=orders
// @tg title=`orders`
// @tg servers=
package externalapi

import (
	"context"

	"github.com/mbatimel/AMC/orders/pkg/models"
)

// OrdersAPI
// @tg http-server metrics log
// @tg http-prefix=/api
type OrdersAPI interface {
	// GetCart returns the current client cart.
	// @tg http-method=GET
	// @tg http-path=/v1/cart
	// @tg summary=`Получение корзины`
	// @tg desc=`Получение текущей корзины клиента`
	GetCart(ctx context.Context, request models.GetCartRequest) (models.GetCartResponse, error)

	// AddCartItem adds a product to the cart.
	// @tg http-method=POST
	// @tg http-path=/v1/cart/items
	// @tg summary=`Добавление товара в корзину`
	// @tg desc=`Добавление позиции товара в корзину клиента`
	AddCartItem(ctx context.Context, request models.AddCartItemRequest) (models.AddCartItemResponse, error)

	// UpdateCartItem updates cart item quantity.
	// @tg http-method=PATCH
	// @tg http-path=/v1/cart/items/{id}
	// @tg summary=`Обновление позиции корзины`
	// @tg desc=`Изменение количества товара в корзине`
	UpdateCartItem(ctx context.Context, request models.UpdateCartItemRequest) (models.UpdateCartItemResponse, error)

	// DeleteCartItem removes a cart item.
	// @tg http-method=DELETE
	// @tg http-path=/v1/cart/items/{id}
	// @tg summary=`Удаление позиции корзины`
	// @tg desc=`Удаление позиции товара из корзины`
	DeleteCartItem(ctx context.Context, request models.DeleteCartItemRequest) (models.DeleteCartItemResponse, error)

	// ClearCart removes all cart items.
	// @tg http-method=DELETE
	// @tg http-path=/v1/cart
	// @tg summary=`Очистка корзины`
	// @tg desc=`Удаление всех позиций из корзины клиента`
	ClearCart(ctx context.Context, request models.ClearCartRequest) (models.ClearCartResponse, error)

	// CreateOrder creates an order from the cart.
	// @tg http-method=POST
	// @tg http-path=/v1/orders
	// @tg summary=`Создание заказа`
	// @tg desc=`Оформление заказа из корзины клиента`
	CreateOrder(ctx context.Context, request models.CreateOrderRequest) (models.CreateOrderResponse, error)

	// GetOrder returns one order by ID.
	// @tg http-method=GET
	// @tg http-path=/v1/orders/{id}
	// @tg summary=`Получение заказа`
	// @tg desc=`Получение детальной информации по заказу`
	GetOrder(ctx context.Context, request models.GetOrderRequest) (models.GetOrderResponse, error)

	// ListOrders returns client orders.
	// @tg http-method=GET
	// @tg http-path=/v1/orders
	// @tg summary=`Список заказов`
	// @tg desc=`Получение списка заказов клиента`
	ListOrders(ctx context.Context, request models.ListOrdersRequest) (models.ListOrdersResponse, error)

	// CancelOrder cancels an order.
	// @tg http-method=POST
	// @tg http-path=/v1/orders/{id}/cancel
	// @tg summary=`Отмена заказа`
	// @tg desc=`Отмена заказа клиентом или администратором`
	CancelOrder(ctx context.Context, request models.CancelOrderRequest) (models.CancelOrderResponse, error)

	// RepeatOrder repeats an order.
	// @tg http-method=POST
	// @tg http-path=/v1/orders/{id}/repeat
	// @tg summary=`Повтор заказа`
	// @tg desc=`Создание корзины или заказа на основе существующего заказа`
	RepeatOrder(ctx context.Context, request models.RepeatOrderRequest) (models.RepeatOrderResponse, error)

	// GetOrderDocuments returns order documents.
	// @tg http-method=GET
	// @tg http-path=/v1/orders/{id}/documents
	// @tg summary=`Документы заказа`
	// @tg desc=`Получение документов по заказу`
	GetOrderDocuments(ctx context.Context, request models.GetOrderDocumentsRequest) (models.GetOrderDocumentsResponse, error)

	// GetOrderHistory returns order history.
	// @tg http-method=GET
	// @tg http-path=/v1/orders/{id}/history
	// @tg summary=`История заказа`
	// @tg desc=`Получение истории изменения заказа`
	GetOrderHistory(ctx context.Context, request models.GetOrderHistoryRequest) (models.GetOrderHistoryResponse, error)

	// UpdateOrderStatus updates order and payment statuses.
	// @tg http-method=PATCH
	// @tg http-path=/v1/admin/orders/{id}/status
	// @tg summary=`Обновление статуса заказа`
	// @tg desc=`Администраторское изменение статуса заказа и оплаты`
	UpdateOrderStatus(ctx context.Context, request models.UpdateOrderStatusRequest) (models.UpdateOrderStatusResponse, error)
}
