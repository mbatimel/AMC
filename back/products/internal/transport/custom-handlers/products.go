package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	externalapi "github.com/mbatimel/AMC/products/pkg/interfaces/externalapi"
	"github.com/mbatimel/AMC/products/pkg/models"
	"github.com/rs/zerolog/log"
)

const ServiceName = "products"

func CreateProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, request models.CreateProductRequest) error {
	return handle(ctx, "post", "/v1/products", "CreateProduct", map[string]interface{}{
		"sku": request.SKU,
	}, func() (interface{}, error) {
		return svc.CreateProduct(ctx.UserContext(), request)
	})
}

func GetProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, request models.GetProductRequest) error {
	return handle(ctx, "get", "/v1/products/{id}", "GetProduct", map[string]interface{}{
		"id": request.ID,
	}, func() (interface{}, error) {
		return svc.GetProduct(ctx.UserContext(), request)
	})
}

func ListProducts(ctx *fiber.Ctx, svc externalapi.ProductsAPI, request models.ListProductsRequest) error {
	return handle(ctx, "get", "/v1/products", "ListProducts", map[string]interface{}{
		"q":           request.Q,
		"categoryID":  request.CategoryID,
		"brandID":     request.BrandID,
		"material":    request.Material,
		"size":        request.Size,
		"gost":        request.GOST,
		"limit":       request.Limit,
		"offset":      request.Offset,
		"sort":        request.Sort,
		"hasInStock":  request.InStock != nil,
		"serviceName": ServiceName,
	}, func() (interface{}, error) {
		return svc.ListProducts(ctx.UserContext(), request)
	})
}

func UpdateProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, request models.UpdateProductRequest) error {
	return handle(ctx, "patch", "/v1/products/{id}", "UpdateProduct", map[string]interface{}{
		"id": request.ID,
	}, func() (interface{}, error) {
		return svc.UpdateProduct(ctx.UserContext(), request)
	})
}

func DeleteProduct(ctx *fiber.Ctx, svc externalapi.ProductsAPI, request models.DeleteProductRequest) error {
	return handle(ctx, "delete", "/v1/products/{id}", "DeleteProduct", map[string]interface{}{
		"id": request.ID,
	}, func() (interface{}, error) {
		return svc.DeleteProduct(ctx.UserContext(), request)
	})
}

func ListCategories(ctx *fiber.Ctx, svc externalapi.ProductsAPI, request models.ListCategoriesRequest) error {
	return handle(ctx, "get", "/v1/categories", "ListCategories", map[string]interface{}{
		"limit":  request.Limit,
		"offset": request.Offset,
	}, func() (interface{}, error) {
		return svc.ListCategories(ctx.UserContext(), request)
	})
}

func ListBrands(ctx *fiber.Ctx, svc externalapi.ProductsAPI, request models.ListBrandsRequest) error {
	return handle(ctx, "get", "/v1/brands", "ListBrands", map[string]interface{}{
		"limit":  request.Limit,
		"offset": request.Offset,
	}, func() (interface{}, error) {
		return svc.ListBrands(ctx.UserContext(), request)
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
