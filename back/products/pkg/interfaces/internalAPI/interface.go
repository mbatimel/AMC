// Package internalAPI describes the internal (service-to-service) products
// API contract. Unlike pkg/interfaces/externalapi, this API is not exposed
// through the public gateway — only trusted internal callers (e.g. platforms)
// reach it.
// @tg version=0.0.1
// @tg backend=products
// @tg title=`products-internal`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/internalapi --outSwagger ../../../swaggers/internalapi/swagger.yaml
//go:generate tg client --services . --outPath ../../client/transport -go
package internalAPI

import (
	"context"

	"github.com/google/uuid"
)

// ProductsInternalAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/products/swaggers/internalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/products/swaggers/internalapi/models:Err400
// @tg 500=github.com/mbatimel/AMC/products/swaggers/internalapi/models:Err500
type ProductsInternalAPI interface {
	// SetMarketplaceStatus ...
	// @tg http-method=POST
	// @tg http-path=/v1/products/marketplace-status
	// @tg http-args=productID|productID
	// @tg http-args=marketplace|marketplace
	// @tg http-args=status|status
	// @tg uuidPackage=github.com/google/uuid
	// @tg productID.format=uuid
	// @tg summary=`Установка статуса карточки товара на маркетплейсе`
	// @tg desc=`Отмечает товар как добавленный (или с иным статусом) на указанном маркетплейсе: wildberries, ozon или yandex_market`
	SetMarketplaceStatus(ctx context.Context, productID uuid.UUID, marketplace string, status string) (ok bool, err error)
}
