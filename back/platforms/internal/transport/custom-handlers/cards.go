package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/platforms/pkg/interfaces/externalapi"
	"github.com/mbatimel/AMC/platforms/pkg/models"
	"github.com/rs/zerolog/log"
)

const ServiceName = "platforms"

func CreateWildberriesCard(
	ctx *fiber.Ctx,
	svc externalapi.PlatformsAPI,
	userID uuid.UUID,
	productID uuid.UUID,
	subjectID int,
	vendorCode string,
	title string,
	description string,
	brand string,
	kizMarked bool,
	wholesaleEnabled bool,
	wholesaleQuantum int,
	dimensions models.WildberriesDimensions,
	sizes []models.WildberriesSize,
	characteristics []models.WildberriesCharacteristic,
	documents models.WildberriesDocuments,
) error {
	return handle(ctx, "post", "/v1/wildberries/cards", "CreateWildberriesCard", map[string]interface{}{
		"userID": userID, "productID": productID, "subjectID": subjectID, "vendorCode": vendorCode,
	}, func() (interface{}, error) {
		return svc.CreateWildberriesCard(ctx.UserContext(), userID, productID, subjectID, vendorCode, title, description, brand, kizMarked, wholesaleEnabled, wholesaleQuantum, dimensions, sizes, characteristics, documents)
	})
}

func CreateOzonCard(
	ctx *fiber.Ctx,
	svc externalapi.PlatformsAPI,
	userID uuid.UUID,
	productID uuid.UUID,
	offerID string,
	name string,
	descriptionCategoryID int64,
	typeID int64,
	currencyCode string,
	price string,
	oldPrice string,
	vat string,
	barcode string,
	dimensions models.OzonDimensions,
	images []string,
	primaryImage string,
	colorImage string,
	attributes []models.OzonAttribute,
) error {
	return handle(ctx, "post", "/v1/ozon/cards", "CreateOzonCard", map[string]interface{}{
		"userID": userID, "productID": productID, "offerID": offerID, "descriptionCategoryID": descriptionCategoryID,
	}, func() (interface{}, error) {
		return svc.CreateOzonCard(ctx.UserContext(), userID, productID, offerID, name, descriptionCategoryID, typeID, currencyCode, price, oldPrice, vat, barcode, dimensions, images, primaryImage, colorImage, attributes)
	})
}

func CreateYandexMarketCard(
	ctx *fiber.Ctx,
	svc externalapi.PlatformsAPI,
	userID uuid.UUID,
	productID uuid.UUID,
	offerID string,
	name string,
	marketCategoryID int64,
	description string,
	vendor string,
	vendorCode string,
	pictures []string,
	barcodes []string,
	dimensions models.YandexMarketDimensions,
	parameterValues []models.YandexMarketParameterValue,
) error {
	return handle(ctx, "post", "/v1/yandex-market/cards", "CreateYandexMarketCard", map[string]interface{}{
		"userID": userID, "productID": productID, "offerID": offerID, "marketCategoryID": marketCategoryID,
	}, func() (interface{}, error) {
		return svc.CreateYandexMarketCard(ctx.UserContext(), userID, productID, offerID, name, marketCategoryID, description, vendor, vendorCode, pictures, barcodes, dimensions, parameterValues)
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
