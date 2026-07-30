package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/products/pkg/interfaces/externalapi"
	"github.com/mbatimel/AMC/products/pkg/models"
	"github.com/rs/zerolog/log"
)

const ServiceName = "products"

func CreateProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, userID uuid.UUID, sku string, name string, description string, categoryID uuid.UUID, brandID uuid.UUID, gost string, material string, size string, packageQty int, stockQty int, basePrice float64, clientPrice float64, discountPercent float64, images []models.ProductImage, isPublished bool) error {
	return handle(ctx, "post", "/v1/products", "CreateProduct", map[string]interface{}{
		"userID": userID, "sku": sku, "categoryID": categoryID, "brandID": brandID,
	}, func() (interface{}, error) {
		return svc.CreateProduct(ctx.UserContext(), userID, sku, name, description, categoryID, brandID, gost, material, size, packageQty, stockQty, basePrice, clientPrice, discountPercent, images, isPublished)
	})
}

func GetProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, productID uuid.UUID) error {
	return handle(ctx, "get", "/v1/products/{productID}", "GetProduct", map[string]interface{}{
		"productID": productID,
	}, func() (interface{}, error) {
		return svc.GetProduct(ctx.UserContext(), productID)
	})
}

func ListProducts(ctx *fiber.Ctx, svc externalapi.ProductsAPI, q *string, categoryID *string, brandID *string, material *string, size *string, gost *string, inStock *bool, limit *int, offset *int, sort *string) error {
	return handle(ctx, "get", "/v1/products", "ListProducts", map[string]interface{}{
		"categoryID":  categoryID,
		"brandID":     brandID,
		"material":    material,
		"size":        size,
		"gost":        gost,
		"inStock":     inStock,
		"limit":       limit,
		"offset":      offset,
		"sort":        sort,
		"hasQuery":    q != nil,
		"serviceName": ServiceName,
	}, func() (interface{}, error) {
		return svc.ListProducts(ctx.UserContext(), q, categoryID, brandID, material, size, gost, inStock, limit, offset, sort)
	})
}

func UpdateProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, userID uuid.UUID, productID uuid.UUID, sku *string, name *string, description *string, categoryID *uuid.UUID, brandID *uuid.UUID, gost *string, material *string, size *string, packageQty *int, stockQty *int, basePrice *float64, clientPrice *float64, discountPercent *float64, images *[]models.ProductImage, isPublished *bool) error {
	return handle(ctx, "patch", "/v1/products/{productID}", "UpdateProduct", map[string]interface{}{
		"userID": userID, "productID": productID, "categoryID": categoryID, "brandID": brandID,
	}, func() (interface{}, error) {
		return svc.UpdateProduct(ctx.UserContext(), userID, productID, sku, name, description, categoryID, brandID, gost, material, size, packageQty, stockQty, basePrice, clientPrice, discountPercent, images, isPublished)
	})
}

func DeleteProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, userID uuid.UUID, productID uuid.UUID) error {
	return handle(ctx, "delete", "/v1/products/{productID}", "DeleteProduct", map[string]interface{}{
		"userID": userID, "productID": productID,
	}, func() (interface{}, error) {
		return svc.DeleteProduct(ctx.UserContext(), userID, productID)
	})
}

func ListCategories(ctx *fiber.Ctx, svc externalapi.ProductsAPI, limit *int, offset *int) error {
	return handle(ctx, "get", "/v1/categories", "ListCategories", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}, func() (interface{}, error) {
		return svc.ListCategories(ctx.UserContext(), limit, offset)
	})
}

func ListBrands(ctx *fiber.Ctx, svc externalapi.ProductsAPI, limit *int, offset *int) error {
	return handle(ctx, "get", "/v1/brands", "ListBrands", map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}, func() (interface{}, error) {
		return svc.ListBrands(ctx.UserContext(), limit, offset)
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
