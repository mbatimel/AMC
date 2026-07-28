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
func GetCities(ctx *fiber.Ctx, svc externalapi.OrdersAPI) error {
	return handle(ctx, "get", "/v1/orders/cities", "GetCities", map[string]interface{}{
	}, func() (interface{}, error) {
		return svc.GetCities(ctx.UserContext())
	})
}

func AddCartItem(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, clientID string, productID string, qty int) error {
	return handle(ctx, "post", "/v1/cart/items", "AddCartItem", map[string]interface{}{
		"userID":    userID,
		"clientID":  clientID,
		"productID": productID,
		"qty":       qty,
	}, func() (interface{}, error) {
		return svc.AddCartItem(ctx.UserContext(), userID, clientID, productID, qty)
	})
}

func UpdateCartItem(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, clientID string, cartItemID string, qty int) error {
	return handle(ctx, "patch", "/v1/cart/items/{cartItemID}", "UpdateCartItem", map[string]interface{}{
		"userID":     userID,
		"clientID":   clientID,
		"cartItemID": cartItemID,
		"qty":        qty,
	}, func() (interface{}, error) {
		return svc.UpdateCartItem(ctx.UserContext(), userID, clientID, cartItemID, qty)
	})
}

func DeleteCartItem(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, clientID string, cartItemID string) error {
	return handle(ctx, "delete", "/v1/cart/items/{cartItemID}", "DeleteCartItem", map[string]interface{}{
		"userID":     userID,
		"clientID":   clientID,
		"cartItemID": cartItemID,
	}, func() (interface{}, error) {
		return svc.DeleteCartItem(ctx.UserContext(), userID, clientID, cartItemID)
	})
}

func ClearCart(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, clientID string) error {
	return handle(ctx, "delete", "/v1/cart", "ClearCart", map[string]interface{}{
		"userID":   userID,
		"clientID": clientID,
	}, func() (interface{}, error) {
		return svc.ClearCart(ctx.UserContext(), userID, clientID)
	})
}

func CreateOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, clientID string, deliveryType string, deliveryAddress string, contactName string, phone string, email string, comment string) error {
	return handle(ctx, "post", "/v1/orders", "CreateOrder", map[string]interface{}{
		"userID":          userID,
		"clientID":        clientID,
		"deliveryType":    deliveryType,
		"deliveryAddress": deliveryAddress,
		"contactName":     contactName,
		"phone":           phone,
		"email":           email,
		"comment":         comment,
	}, func() (interface{}, error) {
		return svc.CreateOrder(ctx.UserContext(), userID, clientID, deliveryType, deliveryAddress, contactName, phone, email, comment)
	})
}

func GetOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, orderID uuid.UUID, userID uuid.UUID, clientID string) error {
	return handle(ctx, "get", "/v1/orders/{orderID}", "GetOrder", map[string]interface{}{
		"orderID":  orderID,
		"userID":   userID,
		"clientID": clientID,
	}, func() (interface{}, error) {
		return svc.GetOrder(ctx.UserContext(), orderID, userID, clientID)
	})
}

func ListOrders(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, clientID string, status string, paymentStatus string, limit int, offset int, sort string) error {
	return handle(ctx, "get", "/v1/orders", "ListOrders", map[string]interface{}{
		"userID":        userID,
		"clientID":      clientID,
		"status":        status,
		"paymentStatus": paymentStatus,
		"limit":         limit,
		"offset":        offset,
		"sort":          sort,
	}, func() (interface{}, error) {
		return svc.ListOrders(ctx.UserContext(), userID, clientID, status, paymentStatus, limit, offset, sort)
	})
}

func CancelOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, orderID uuid.UUID, userID uuid.UUID, comment string) error {
	return handle(ctx, "post", "/v1/orders/{orderID}/cancel", "CancelOrder", map[string]interface{}{
		"orderID": orderID,
		"userID":  userID,
		"comment": comment,
	}, func() (interface{}, error) {
		return svc.CancelOrder(ctx.UserContext(), orderID, userID, comment)
	})
}

func RepeatOrder(ctx *fiber.Ctx, svc externalapi.OrdersAPI, orderID uuid.UUID, userID uuid.UUID, clientID string) error {
	return handle(ctx, "post", "/v1/orders/{orderID}/repeat", "RepeatOrder", map[string]interface{}{
		"orderID":  orderID,
		"userID":   userID,
		"clientID": clientID,
	}, func() (interface{}, error) {
		return svc.RepeatOrder(ctx.UserContext(), orderID, userID, clientID)
	})
}

func GetOrderDocuments(ctx *fiber.Ctx, svc externalapi.OrdersAPI, orderID uuid.UUID, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/orders/{orderID}/documents", "GetOrderDocuments", map[string]interface{}{
		"orderID": orderID,
		"userID":  userID,
	}, func() (interface{}, error) {
		return svc.GetOrderDocuments(ctx.UserContext(), orderID, userID)
	})
}

func GetOrderHistory(ctx *fiber.Ctx, svc externalapi.OrdersAPI, orderID uuid.UUID, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/orders/{orderID}/history", "GetOrderHistory", map[string]interface{}{
		"orderID": orderID,
		"userID":  userID,
	}, func() (interface{}, error) {
		return svc.GetOrderHistory(ctx.UserContext(), orderID, userID)
	})
}

func UpdateOrderStatus(ctx *fiber.Ctx, svc externalapi.OrdersAPI,userID uuid.UUID, orderID uuid.UUID, status string, paymentStatus string, comment string, changedBy string) error {
	return handle(ctx, "patch", "/v1/admin/orders/{orderID}/status", "UpdateOrderStatus", map[string]interface{}{
		"orderID":       orderID,
		"status":        status,
		"paymentStatus": paymentStatus,
		"comment":       comment,
		"changedBy":     changedBy,
	}, func() (interface{}, error) {
		return svc.UpdateOrderStatus(ctx.UserContext(),userID, orderID, status, paymentStatus, comment, changedBy)
	})
}

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
