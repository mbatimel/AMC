// Package externalapi describes the public platforms API contract.
// @tg version=0.0.1
// @tg backend=platforms
// @tg title=`platforms`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalapi

import (
	"context"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/platforms/pkg/models"
)

// PlatformsAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/platforms/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/platforms/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/platforms/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/platforms/swaggers/externalapi/models:Err403
// @tg 404=github.com/mbatimel/AMC/platforms/swaggers/externalapi/models:Err400
// @tg 409=github.com/mbatimel/AMC/platforms/swaggers/externalapi/models:Err400
// @tg 500=github.com/mbatimel/AMC/platforms/swaggers/externalapi/models:Err500
type PlatformsAPI interface {
	// CreateWildberriesCard ...
	// @tg http-method=POST
	// @tg http-path=/v1/wildberries/cards
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/platforms/internal/transport/custom-handlers:CreateWildberriesCard
	// @tg summary=`Создание карточки товара на Wildberries`
	// @tg desc=`Создание карточки товара в личном кабинете продавца Wildberries (POST /content/v2/cards/upload). Поддерживается создание отдельной (не объединённой) карточки — один вариант на запрос.`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	// @tg productID.format=uuid
	CreateWildberriesCard(ctx context.Context, userID uuid.UUID, productID uuid.UUID, subjectID int, vendorCode string, title string, description string, brand string, kizMarked bool, wholesaleEnabled bool, wholesaleQuantum int, dimensions models.WildberriesDimensions, sizes []models.WildberriesSize, characteristics []models.WildberriesCharacteristic, documents models.WildberriesDocuments) (response models.CreateWildberriesCardResponse, err error)

	// CreateOzonCard ...
	// @tg http-method=POST
	// @tg http-path=/v1/ozon/cards
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/platforms/internal/transport/custom-handlers:CreateOzonCard
	// @tg summary=`Создание/обновление карточки товара на Ozon`
	// @tg desc=`Создание или обновление товара в личном кабинете продавца Ozon (POST /v3/product/import). Обрабатывается асинхронно — ответ содержит номер задания (task_id).`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	// @tg productID.format=uuid
	CreateOzonCard(ctx context.Context, userID uuid.UUID, productID uuid.UUID, offerID string, name string, descriptionCategoryID int64, typeID int64, currencyCode string, price string, oldPrice string, vat string, barcode string, dimensions models.OzonDimensions, images []string, primaryImage string, colorImage string, attributes []models.OzonAttribute) (response models.CreateOzonCardResponse, err error)

	// CreateYandexMarketCard ...
	// @tg http-method=POST
	// @tg http-path=/v1/yandex-market/cards
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/platforms/internal/transport/custom-handlers:CreateYandexMarketCard
	// @tg summary=`Создание/обновление карточки товара на Яндекс Маркете`
	// @tg desc=`Создание или обновление товара в кабинете продавца Яндекс Маркета (POST /v2/businesses/{businessId}/offer-mappings/update). Обрабатывается синхронно — статус и предупреждения приходят сразу в ответе.`
	// @tg uuidPackage=github.com/google/uuid
	// @tg userID.format=uuid
	// @tg productID.format=uuid
	CreateYandexMarketCard(ctx context.Context, userID uuid.UUID, productID uuid.UUID, offerID string, name string, marketCategoryID int64, description string, vendor string, vendorCode string, pictures []string, barcodes []string, dimensions models.YandexMarketDimensions, parameterValues []models.YandexMarketParameterValue) (response models.CreateYandexMarketCardResponse, err error)
}
