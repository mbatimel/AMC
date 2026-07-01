// Package externalapi describes the public products API contract.
// @tg version=0.0.1
// @tg backend=products
// @tg title=`products`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalapi

import (
	"context"

	"github.com/mbatimel/AMC/products/pkg/models"
)

// ProductsAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/products/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/products/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/products/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/products/swaggers/externalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/products/swaggers/externalapi/models:Err500
type ProductsAPI interface {
	// CreateProduct ...
	// @tg http-method=POST
	// @tg http-path=/v1/products
	// @tg http-args=sku|sku
	// @tg http-args=name|name
	// @tg http-args=description|description
	// @tg http-args=categoryID|categoryID
	// @tg http-args=brandID|brandID
	// @tg http-args=gost|gost
	// @tg http-args=material|material
	// @tg http-args=size|size
	// @tg http-args=packageQty|packageQty
	// @tg http-args=stockQty|stockQty
	// @tg http-args=basePrice|basePrice
	// @tg http-args=clientPrice|clientPrice
	// @tg http-args=discountPercent|discountPercent
	// @tg http-args=isPublished|isPublished
	// @tg http-response=github.com/mbatimel/AMC/products/internal/transport/custom-handlers:CreateProduct
	// @tg summary=`Создание товара`
	// @tg desc=`Создание карточки товара каталога`
	CreateProduct(ctx context.Context, sku string, name string, description string, categoryID string, brandID string, gost string, material string, size string, packageQty int, stockQty int, basePrice float64, clientPrice float64, discountPercent float64, isPublished bool) (response models.CreateProductResponse, err error)

	// GetProduct ...
	// @tg http-method=GET
	// @tg http-path=/v1/products/{productID}
	// @tg http-args=productID|productID
	// @tg http-response=github.com/mbatimel/AMC/products/internal/transport/custom-handlers:GetProduct
	// @tg summary=`Получение товара`
	// @tg desc=`Получение карточки товара по идентификатору`
	GetProduct(ctx context.Context, productID string) (response models.GetProductResponse, err error)

	// ListProducts ...
	// @tg http-method=GET
	// @tg http-path=/v1/products
	// @tg http-args=q|q
	// @tg http-args=categoryID|categoryID
	// @tg http-args=brandID|brandID
	// @tg http-args=material|material
	// @tg http-args=size|size
	// @tg http-args=gost|gost
	// @tg http-args=inStock|inStock
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-args=sort|sort
	// @tg http-response=github.com/mbatimel/AMC/products/internal/transport/custom-handlers:ListProducts
	// @tg summary=`Список товаров`
	// @tg desc=`Получение списка товаров с фильтрами, поиском и пагинацией`
	ListProducts(ctx context.Context, q string, categoryID string, brandID string, material string, size string, gost string, inStock bool, limit int, offset int, sort string) (response models.ListProductsResponse, err error)

	// UpdateProduct ...
	// @tg http-method=PATCH
	// @tg http-path=/v1/products/{productID}
	// @tg http-args=productID|productID
	// @tg http-args=sku|sku
	// @tg http-args=name|name
	// @tg http-args=description|description
	// @tg http-args=categoryID|categoryID
	// @tg http-args=brandID|brandID
	// @tg http-args=gost|gost
	// @tg http-args=material|material
	// @tg http-args=size|size
	// @tg http-args=packageQty|packageQty
	// @tg http-args=stockQty|stockQty
	// @tg http-args=basePrice|basePrice
	// @tg http-args=clientPrice|clientPrice
	// @tg http-args=discountPercent|discountPercent
	// @tg http-args=isPublished|isPublished
	// @tg http-response=github.com/mbatimel/AMC/products/internal/transport/custom-handlers:UpdateProduct
	// @tg summary=`Обновление товара`
	// @tg desc=`Обновление карточки товара каталога`
	UpdateProduct(ctx context.Context, productID string, sku string, name string, description string, categoryID string, brandID string, gost string, material string, size string, packageQty int, stockQty int, basePrice float64, clientPrice float64, discountPercent float64, isPublished bool) (response models.UpdateProductResponse, err error)

	// DeleteProduct ...
	// @tg http-method=DELETE
	// @tg http-path=/v1/products/{productID}
	// @tg http-args=productID|productID
	// @tg http-response=github.com/mbatimel/AMC/products/internal/transport/custom-handlers:DeleteProduct
	// @tg summary=`Удаление товара`
	// @tg desc=`Удаление или скрытие карточки товара`
	DeleteProduct(ctx context.Context, productID string) (response models.DeleteProductResponse, err error)

	// ListCategories ...
	// @tg http-method=GET
	// @tg http-path=/v1/categories
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-response=github.com/mbatimel/AMC/products/internal/transport/custom-handlers:ListCategories
	// @tg summary=`Список категорий`
	// @tg desc=`Получение списка категорий каталога`
	ListCategories(ctx context.Context, limit int, offset int) (response models.ListCategoriesResponse, err error)

	// ListBrands ...
	// @tg http-method=GET
	// @tg http-path=/v1/brands
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-response=github.com/mbatimel/AMC/products/internal/transport/custom-handlers:ListBrands
	// @tg summary=`Список брендов`
	// @tg desc=`Получение списка брендов каталога`
	ListBrands(ctx context.Context, limit int, offset int) (response models.ListBrandsResponse, err error)
}
