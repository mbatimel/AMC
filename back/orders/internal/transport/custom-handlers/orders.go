package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/orders/pkg/interfaces/externalapi"
	"github.com/rs/zerolog/log"
)

const ServiceName = "orders"

func GetCart(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, clientID string) error {
	return handle(ctx, "get", "/v1/cart", "GetCart", map[string]interface{}{
		"userID":   userID,
		"clientID": clientID,
	}, func() (interface{}, error) {
		return svc.GetCart(ctx.UserContext(), userID, clientID)
	})
}

// func AddCartItem(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.AddCartItemRequest) error {
// 	return handle(ctx, "post", "/v1/cart/items", "AddCartItem", map[string]interface{}{
// 		"userID":    request.UserID,
// 		"clientID":  request.ClientID,
// 		"productID": request.ProductID,
// 		"qty":       request.Qty,
// 	}, func() (interface{}, error) {
// 		return svc.AddCartItem(ctx.UserContext(), request)
// 	})
// }

// func UpdateCartItem(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.UpdateCartItemRequest) error {
// 	return handle(ctx, "patch", "/v1/cart/items/{id}", "UpdateCartItem", map[string]interface{}{
// 		"userID":     request.UserID,
// 		"clientID":   request.ClientID,
// 		"cartItemID": request.CartItemID,
// 		"qty":        request.Qty,
// 	}, func() (interface{}, error) {
// 		return svc.UpdateCartItem(ctx.UserContext(), request)
// 	})
// }

// func DeleteCartItem(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.DeleteCartItemRequest) error {
// 	return handle(ctx, "delete", "/v1/cart/items/{id}", "DeleteCartItem", map[string]interface{}{
// 		"userID":     request.UserID,
// 		"clientID":   request.ClientID,
// 		"cartItemID": request.CartItemID,
// 	}, func() (interface{}, error) {
// 		return svc.DeleteCartItem(ctx.UserContext(), request)
// 	})
// }

// func ClearCart(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.ClearCartRequest) error {
// 	return handle(ctx, "delete", "/v1/cart", "ClearCart", map[string]interface{}{
// 		"userID":   request.UserID,
// 		"clientID": request.ClientID,
// 	}, func() (interface{}, error) {
// 		return svc.ClearCart(ctx.UserContext(), request)
// 	})
// }

// func CreateOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.CreateOrderRequest) error {
// 	return handle(ctx, "post", "/v1/orders", "CreateOrder", map[string]interface{}{
// 		"userID":       request.UserID,
// 		"clientID":     request.ClientID,
// 		"deliveryType": request.DeliveryType,
// 	}, func() (interface{}, error) {
// 		return svc.CreateOrder(ctx.UserContext(), request)
// 	})
// }

// func GetOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.GetOrderRequest) error {
// 	return handle(ctx, "get", "/v1/orders/{id}", "GetOrder", map[string]interface{}{
// 		"id":       request.ID,
// 		"userID":   request.UserID,
// 		"clientID": request.ClientID,
// 	}, func() (interface{}, error) {
// 		return svc.GetOrder(ctx.UserContext(), request)
// 	})
// }

// func ListOrders(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.ListOrdersRequest) error {
// 	return handle(ctx, "get", "/v1/orders", "ListOrders", map[string]interface{}{
// 		"userID":        request.UserID,
// 		"clientID":      request.ClientID,
// 		"status":        request.Status,
// 		"paymentStatus": request.PaymentStatus,
// 		"limit":         request.Limit,
// 		"offset":        request.Offset,
// 		"sort":          request.Sort,
// 	}, func() (interface{}, error) {
// 		return svc.ListOrders(ctx.UserContext(), request)
// 	})
// }

// func CancelOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.CancelOrderRequest) error {
// 	return handle(ctx, "post", "/v1/orders/{id}/cancel", "CancelOrder", map[string]interface{}{
// 		"id":     request.ID,
// 		"userID": request.UserID,
// 	}, func() (interface{}, error) {
// 		return svc.CancelOrder(ctx.UserContext(), request)
// 	})
// }

// func RepeatOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.RepeatOrderRequest) error {
// 	return handle(ctx, "post", "/v1/orders/{id}/repeat", "RepeatOrder", map[string]interface{}{
// 		"id":     request.ID,
// 		"userID": request.UserID,
// 	}, func() (interface{}, error) {
// 		return svc.RepeatOrder(ctx.UserContext(), request)
// 	})
// }

// func GetOrderDocuments(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.GetOrderDocumentsRequest) error {
// 	return handle(ctx, "get", "/v1/orders/{id}/documents", "GetOrderDocuments", map[string]interface{}{
// 		"orderID": request.OrderID,
// 		"userID":  request.UserID,
// 	}, func() (interface{}, error) {
// 		return svc.GetOrderDocuments(ctx.UserContext(), request)
// 	})
// }

// func GetOrderHistory(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.GetOrderHistoryRequest) error {
// 	return handle(ctx, "get", "/v1/orders/{id}/history", "GetOrderHistory", map[string]interface{}{
// 		"orderID": request.OrderID,
// 		"userID":  request.UserID,
// 	}, func() (interface{}, error) {
// 		return svc.GetOrderHistory(ctx.UserContext(), request)
// 	})
// }

// func UpdateOrderStatus(ctx *fiber.Ctx, svc externalapi.OrdersAPI, request models.UpdateOrderStatusRequest) error {
// 	return handle(ctx, "patch", "/v1/admin/orders/{id}/status", "UpdateOrderStatus", map[string]interface{}{
// 		"id":            request.ID,
// 		"status":        request.Status,
// 		"paymentStatus": request.PaymentStatus,
// 		"changedBy":     request.ChangedBy,
// 	}, func() (interface{}, error) {
// 		return svc.UpdateOrderStatus(ctx.UserContext(), request)
// 	})
// }

func handle(
	ctx *fiber.Ctx,
	method string,
	path string,
	methodName string,
	fields map[string]interface{},
	call func() (interface{}, error),
) error {
	var err error

	defer func(begin time.Time) {
		fields["method"] = method
		fields["path"] = path
		fields["methodName"] = methodName
		fields["serviceName"] = ServiceName
		fields["took"] = time.Since(begin).String()

		l := log.Info()
		if err != nil {
			l = log.Error().Err(err)
		}

		l.Fields(fields).Msg("call")
	}(time.Now())

	data, err := call()
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, data, nil)
	return nil
}
