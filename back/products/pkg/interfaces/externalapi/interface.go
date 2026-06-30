// Package externalapi describes the public products API contract.
// @tg version=0.0.1
// @tg backend=products
// @tg title=`products`
// @tg servers=
package externalapi

import (
	"context"

	"github.com/mbatimel/AMC/products/pkg/models"
)

// ProductsAPI
// @tg http-server metrics log
// @tg http-prefix=/api
type ProductsAPI interface {
	// CreateProduct creates a product card for the portal catalog.
	// @tg http-method=POST
	// @tg http-path=/v1/products
	// @tg summary=`Создание товара`
	// @tg desc=`Создание карточки товара каталога`
	CreateProduct(ctx context.Context, request models.CreateProductRequest) (models.CreateProductResponse, error)

	// GetProduct returns one product by ID.
	// @tg http-method=GET
	// @tg http-path=/v1/products/{id}
	// @tg summary=`Получение товара`
	// @tg desc=`Получение карточки товара по идентификатору`
	GetProduct(ctx context.Context, request models.GetProductRequest) (models.GetProductResponse, error)

	// ListProducts returns a filtered product list.
	// @tg http-method=GET
	// @tg http-path=/v1/products
	// @tg summary=`Список товаров`
	// @tg desc=`Получение списка товаров с фильтрами, поиском и пагинацией`
	ListProducts(ctx context.Context, request models.ListProductsRequest) (models.ListProductsResponse, error)

	// UpdateProduct updates portal product content.
	// @tg http-method=PATCH
	// @tg http-path=/v1/products/{id}
	// @tg summary=`Обновление товара`
	// @tg desc=`Обновление карточки товара каталога`
	UpdateProduct(ctx context.Context, request models.UpdateProductRequest) (models.UpdateProductResponse, error)

	// DeleteProduct deletes or hides a product card.
	// @tg http-method=DELETE
	// @tg http-path=/v1/products/{id}
	// @tg summary=`Удаление товара`
	// @tg desc=`Удаление или скрытие карточки товара`
	DeleteProduct(ctx context.Context, request models.DeleteProductRequest) (models.DeleteProductResponse, error)

	// ListCategories returns catalog categories.
	// @tg http-method=GET
	// @tg http-path=/v1/categories
	// @tg summary=`Список категорий`
	// @tg desc=`Получение списка категорий каталога`
	ListCategories(ctx context.Context, request models.ListCategoriesRequest) (models.ListCategoriesResponse, error)

	// ListBrands returns catalog brands.
	// @tg http-method=GET
	// @tg http-path=/v1/brands
	// @tg summary=`Список брендов`
	// @tg desc=`Получение списка брендов каталога`
	ListBrands(ctx context.Context, request models.ListBrandsRequest) (models.ListBrandsResponse, error)
}
