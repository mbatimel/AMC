package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/platforms/internal/errors"
	"github.com/mbatimel/AMC/platforms/internal/ozon"
	"github.com/mbatimel/AMC/platforms/internal/wildberries"
	"github.com/mbatimel/AMC/platforms/internal/yandexmarket"
	"github.com/mbatimel/AMC/platforms/pkg/models"
)

const (
	maxVendorCodeSize  = 72
	maxTitleSize       = 60
	maxDescription     = 5000
	maxSizes           = 30
	maxCharacteristics = 100
	maxDocumentItems   = 30

	maxOfferIDSize    = 250
	maxOzonNameSize   = 500
	maxOzonAttributes = 300
	maxOzonImages     = 50

	maxYandexMarketNameSize    = 256
	maxYandexMarketDescription = 6000
	maxYandexMarketPictures    = 30
	maxYandexMarketBarcodes    = 30
)

var ozonDimensionUnits = map[string]struct{}{"mm": {}, "cm": {}, "in": {}}
var ozonWeightUnits = map[string]struct{}{"g": {}, "kg": {}, "lb": {}}

type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (bool, error)
}

type WildberriesClient interface {
	CreateCard(ctx context.Context, params wildberries.CreateCardParams) (wildberries.CreateCardResult, error)
}

type OzonClient interface {
	CreateCard(ctx context.Context, params ozon.CreateCardParams) (ozon.CreateCardResult, error)
}

type YandexMarketClient interface {
	CreateCard(ctx context.Context, params yandexmarket.CreateCardParams) (yandexmarket.CreateCardResult, error)
}

// ProductsClient — внутренний API products, чтобы отмечать товар каталога
// статусом карточки на маркетплейсе после её создания.
type ProductsClient interface {
	SetMarketplaceStatus(ctx context.Context, productID uuid.UUID, marketplace string, status string) (bool, error)
}

const (
	MarketplaceWildberries = "wildberries"
	MarketplaceOzon        = "ozon"
	MarketplaceYandex      = "yandex_market"

	statusPending = "pending"
)

type Service struct {
	logger             zerolog.Logger
	accessClient       AccessClient
	productsClient     ProductsClient
	wildberriesClient  WildberriesClient
	ozonClient         OzonClient
	yandexMarketClient YandexMarketClient
}

func New(
	logger zerolog.Logger,
	accessClient AccessClient,
	productsClient ProductsClient,
	wildberriesClient WildberriesClient,
	ozonClient OzonClient,
	yandexMarketClient YandexMarketClient,
) *Service {
	return &Service{
		logger:             logger,
		accessClient:       accessClient,
		productsClient:     productsClient,
		wildberriesClient:  wildberriesClient,
		ozonClient:         ozonClient,
		yandexMarketClient: yandexMarketClient,
	}
}

// markProductAdded отмечает товар каталога статусом карточки на
// маркетплейсе. Лучшее усилие: карточка на маркетплейсе уже создана, поэтому
// ошибка простановки флага не должна откатывать успешный результат ручки —
// она только логируется.
func (s *Service) markProductAdded(ctx context.Context, productID uuid.UUID, marketplace string) {
	if productID == uuid.Nil {
		return
	}
	if _, err := s.productsClient.SetMarketplaceStatus(ctx, productID, marketplace, statusPending); err != nil {
		s.logger.Error().Err(err).Str("productID", productID.String()).Str("marketplace", marketplace).
			Msg("failed to mark product as added to marketplace")
	}
}

func validation(field string) error {
	return customErrors.ErrValidation.AddCause("field", field)
}

// upstreamError пробрасывает ошибку клиента маркетплейса как есть, если она
// уже размечена нашим типом (validation/rate-limit/upstream), иначе — ErrInternal.
func upstreamError(err error) error {
	var customErr *customErrors.Error
	if errors.As(err, &customErr) {
		return customErr
	}
	return customErrors.ErrInternal
}

func (s *Service) checkWriteAccess(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return validation("X-User-Id")
	}
	if s.accessClient == nil {
		return customErrors.ErrInternal
	}
	admin, err := s.accessClient.CheckAccess(ctx, userID, roleCodeAdmin)
	if err != nil {
		return customErrors.ErrInternal
	}
	if admin {
		return nil
	}
	supplier, err := s.accessClient.CheckAccess(ctx, userID, roleCodeSupplier)
	if err != nil {
		return customErrors.ErrInternal
	}
	if !supplier {
		return customErrors.ErrForbidden
	}
	return nil
}

func validateWildberriesSizes(sizes []models.WildberriesSize) ([]wildberries.Size, error) {
	if len(sizes) > maxSizes {
		return nil, validation("sizes")
	}
	result := make([]wildberries.Size, 0, len(sizes))
	for _, size := range sizes {
		if size.Price < 0 {
			return nil, validation("sizes.price")
		}
		result = append(result, wildberries.Size{
			TechSize: strings.TrimSpace(size.TechSize),
			WBSize:   strings.TrimSpace(size.WBSize),
			Price:    size.Price,
			SKUs:     size.SKUs,
		})
	}
	return result, nil
}

func validateWildberriesCharacteristics(characteristics []models.WildberriesCharacteristic) ([]wildberries.Characteristic, error) {
	if len(characteristics) > maxCharacteristics {
		return nil, validation("characteristics")
	}
	result := make([]wildberries.Characteristic, 0, len(characteristics))
	for _, characteristic := range characteristics {
		if characteristic.ID <= 0 {
			return nil, validation("characteristics.id")
		}
		valueJSON := strings.TrimSpace(characteristic.ValueJSON)
		if valueJSON == "" || !json.Valid([]byte(valueJSON)) {
			return nil, validation("characteristics.valueJSON")
		}
		result = append(result, wildberries.Characteristic{ID: characteristic.ID, ValueJSON: valueJSON})
	}
	return result, nil
}

func validateWildberriesDocuments(documents models.WildberriesDocuments) (wildberries.Documents, error) {
	if len(documents.Items) > maxDocumentItems {
		return wildberries.Documents{}, validation("documents.items")
	}
	items := make([]wildberries.DocumentItem, 0, len(documents.Items))
	for _, item := range documents.Items {
		if item.Type < 0 {
			return wildberries.Documents{}, validation("documents.items.type")
		}
		items = append(items, wildberries.DocumentItem{
			Type:          item.Type,
			Number:        strings.TrimSpace(item.Number),
			ProductNumber: strings.TrimSpace(item.ProductNumber),
			TradeName:     strings.TrimSpace(item.TradeName),
			Applicant:     strings.TrimSpace(item.Applicant),
			StartDate:     strings.TrimSpace(item.StartDate),
			EndDate:       strings.TrimSpace(item.EndDate),
			IsEndless:     item.IsEndless,
		})
	}
	return wildberries.Documents{Items: items, ExcludeDocuments: documents.ExcludeDocuments}, nil
}

func (s *Service) CreateWildberriesCard(
	ctx context.Context,
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
) (response models.CreateWildberriesCardResponse, err error) {
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if productID == uuid.Nil {
		return response, validation("productID")
	}
	if subjectID <= 0 {
		return response, validation("subjectID")
	}
	vendorCode = strings.TrimSpace(vendorCode)
	if vendorCode == "" || len(vendorCode) > maxVendorCodeSize {
		return response, validation("vendorCode")
	}
	title = strings.TrimSpace(title)
	if len(title) > maxTitleSize {
		return response, validation("title")
	}
	description = strings.TrimSpace(description)
	if len(description) > maxDescription {
		return response, validation("description")
	}
	if dimensions.LengthCM < 0 || dimensions.WidthCM < 0 || dimensions.HeightCM < 0 || dimensions.WeightBruttoKG < 0 {
		return response, validation("dimensions")
	}
	if wholesaleQuantum < 0 {
		return response, validation("wholesaleQuantum")
	}
	wbSizes, err := validateWildberriesSizes(sizes)
	if err != nil {
		return response, err
	}
	wbCharacteristics, err := validateWildberriesCharacteristics(characteristics)
	if err != nil {
		return response, err
	}
	wbDocuments, err := validateWildberriesDocuments(documents)
	if err != nil {
		return response, err
	}

	_, err = s.wildberriesClient.CreateCard(ctx, wildberries.CreateCardParams{
		SubjectID:        subjectID,
		VendorCode:       vendorCode,
		Title:            title,
		Description:      description,
		Brand:            strings.TrimSpace(brand),
		KizMarked:        kizMarked,
		WholesaleEnabled: wholesaleEnabled,
		WholesaleQuantum: wholesaleQuantum,
		Dimensions: wildberries.Dimensions{
			LengthCM:       dimensions.LengthCM,
			WidthCM:        dimensions.WidthCM,
			HeightCM:       dimensions.HeightCM,
			WeightBruttoKG: dimensions.WeightBruttoKG,
		},
		Sizes:           wbSizes,
		Characteristics: wbCharacteristics,
		Documents:       wbDocuments,
	})
	if err != nil {
		return response, upstreamError(err)
	}
	s.markProductAdded(ctx, productID, MarketplaceWildberries)

	response.Accepted = true
	response.VendorCode = vendorCode
	response.SubjectID = subjectID
	return response, nil
}

func validateOzonImages(images []string, primaryImage string) ([]string, error) {
	limit := maxOzonImages
	if primaryImage != "" {
		limit = maxOzonImages - 1
	}
	if len(images) > limit {
		return nil, validation("images")
	}
	for _, image := range images {
		if strings.TrimSpace(image) == "" {
			return nil, validation("images")
		}
	}
	return images, nil
}

func validateOzonAttributes(attributes []models.OzonAttribute) ([]ozon.Attribute, error) {
	if len(attributes) > maxOzonAttributes {
		return nil, validation("attributes")
	}
	result := make([]ozon.Attribute, 0, len(attributes))
	for _, attribute := range attributes {
		if attribute.ID <= 0 {
			return nil, validation("attributes.id")
		}
		if len(attribute.Values) == 0 {
			return nil, validation("attributes.values")
		}
		values := make([]ozon.AttributeValue, 0, len(attribute.Values))
		for _, value := range attribute.Values {
			if value.DictionaryValueID <= 0 && strings.TrimSpace(value.Value) == "" {
				return nil, validation("attributes.values.value")
			}
			values = append(values, ozon.AttributeValue{
				DictionaryValueID: value.DictionaryValueID,
				Value:             strings.TrimSpace(value.Value),
			})
		}
		result = append(result, ozon.Attribute{
			ComplexID: attribute.ComplexID,
			ID:        attribute.ID,
			Values:    values,
		})
	}
	return result, nil
}

func (s *Service) CreateOzonCard(
	ctx context.Context,
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
) (response models.CreateOzonCardResponse, err error) {
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if productID == uuid.Nil {
		return response, validation("productID")
	}
	offerID = strings.TrimSpace(offerID)
	if offerID == "" || len(offerID) > maxOfferIDSize {
		return response, validation("offerID")
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxOzonNameSize {
		return response, validation("name")
	}
	if descriptionCategoryID <= 0 {
		return response, validation("descriptionCategoryID")
	}
	if typeID <= 0 {
		return response, validation("typeID")
	}
	price = strings.TrimSpace(price)
	if price == "" {
		return response, validation("price")
	}
	vat = strings.TrimSpace(vat)
	if vat == "" {
		return response, validation("vat")
	}
	if currencyCode = strings.TrimSpace(currencyCode); currencyCode == "" {
		currencyCode = "RUB"
	}
	// Ozon требует реальные объёмно-весовые характеристики: нулевые значения
	// приводят к отклонению товара при модерации.
	if dimensions.DepthMM <= 0 || dimensions.WidthMM <= 0 || dimensions.HeightMM <= 0 || dimensions.Weight <= 0 {
		return response, validation("dimensions")
	}
	if _, ok := ozonDimensionUnits[dimensions.DimensionUnit]; !ok {
		return response, validation("dimensions.dimensionUnit")
	}
	if _, ok := ozonWeightUnits[dimensions.WeightUnit]; !ok {
		return response, validation("dimensions.weightUnit")
	}
	validImages, err := validateOzonImages(images, strings.TrimSpace(primaryImage))
	if err != nil {
		return response, err
	}
	ozonAttributes, err := validateOzonAttributes(attributes)
	if err != nil {
		return response, err
	}

	result, err := s.ozonClient.CreateCard(ctx, ozon.CreateCardParams{
		OfferID:               offerID,
		Name:                  name,
		DescriptionCategoryID: descriptionCategoryID,
		TypeID:                typeID,
		CurrencyCode:          currencyCode,
		Price:                 price,
		OldPrice:              strings.TrimSpace(oldPrice),
		VAT:                   vat,
		Barcode:               strings.TrimSpace(barcode),
		Dimensions: ozon.Dimensions{
			DepthMM:       dimensions.DepthMM,
			WidthMM:       dimensions.WidthMM,
			HeightMM:      dimensions.HeightMM,
			DimensionUnit: dimensions.DimensionUnit,
			Weight:        dimensions.Weight,
			WeightUnit:    dimensions.WeightUnit,
		},
		Images:       validImages,
		PrimaryImage: strings.TrimSpace(primaryImage),
		ColorImage:   strings.TrimSpace(colorImage),
		Attributes:   ozonAttributes,
	})
	if err != nil {
		return response, upstreamError(err)
	}
	s.markProductAdded(ctx, productID, MarketplaceOzon)

	response.Accepted = true
	response.OfferID = offerID
	response.TaskID = result.TaskID
	return response, nil
}

func validateYandexMarketPictures(pictures []string) ([]string, error) {
	if len(pictures) == 0 || len(pictures) > maxYandexMarketPictures {
		return nil, validation("pictures")
	}
	for _, picture := range pictures {
		if strings.TrimSpace(picture) == "" {
			return nil, validation("pictures")
		}
	}
	return pictures, nil
}

func validateYandexMarketBarcodes(barcodes []string) ([]string, error) {
	if len(barcodes) > maxYandexMarketBarcodes {
		return nil, validation("barcodes")
	}
	for _, barcode := range barcodes {
		if strings.TrimSpace(barcode) == "" {
			return nil, validation("barcodes")
		}
	}
	return barcodes, nil
}

func validateYandexMarketParameterValues(values []models.YandexMarketParameterValue) ([]yandexmarket.ParameterValue, error) {
	result := make([]yandexmarket.ParameterValue, 0, len(values))
	for _, value := range values {
		if value.ParameterID <= 0 {
			return nil, validation("parameterValues.parameterId")
		}
		result = append(result, yandexmarket.ParameterValue{
			ParameterID: value.ParameterID,
			ValueID:     value.ValueID,
			UnitID:      value.UnitID,
			Value:       strings.TrimSpace(value.Value),
		})
	}
	return result, nil
}

func (s *Service) CreateYandexMarketCard(
	ctx context.Context,
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
) (response models.CreateYandexMarketCardResponse, err error) {
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if productID == uuid.Nil {
		return response, validation("productID")
	}
	offerID = strings.TrimSpace(offerID)
	if offerID == "" || len(offerID) > maxOfferIDSize {
		return response, validation("offerID")
	}
	name = strings.TrimSpace(name)
	if name == "" || len(name) > maxYandexMarketNameSize {
		return response, validation("name")
	}
	if marketCategoryID <= 0 {
		return response, validation("marketCategoryID")
	}
	description = strings.TrimSpace(description)
	if description == "" || len(description) > maxYandexMarketDescription {
		return response, validation("description")
	}
	vendor = strings.TrimSpace(vendor)
	if vendor == "" {
		return response, validation("vendor")
	}
	if dimensions.LengthCM < 0 || dimensions.WidthCM < 0 || dimensions.HeightCM < 0 || dimensions.WeightKG < 0 {
		return response, validation("dimensions")
	}
	validPictures, err := validateYandexMarketPictures(pictures)
	if err != nil {
		return response, err
	}
	validBarcodes, err := validateYandexMarketBarcodes(barcodes)
	if err != nil {
		return response, err
	}
	values, err := validateYandexMarketParameterValues(parameterValues)
	if err != nil {
		return response, err
	}

	result, err := s.yandexMarketClient.CreateCard(ctx, yandexmarket.CreateCardParams{
		OfferID:          offerID,
		Name:             name,
		MarketCategoryID: marketCategoryID,
		Description:      description,
		Vendor:           vendor,
		VendorCode:       strings.TrimSpace(vendorCode),
		Pictures:         validPictures,
		Barcodes:         validBarcodes,
		Dimensions: yandexmarket.Dimensions{
			LengthCM: dimensions.LengthCM,
			WidthCM:  dimensions.WidthCM,
			HeightCM: dimensions.HeightCM,
			WeightKG: dimensions.WeightKG,
		},
		ParameterValues: values,
	})
	if err != nil {
		return response, upstreamError(err)
	}
	s.markProductAdded(ctx, productID, MarketplaceYandex)

	response.Accepted = true
	response.OfferID = offerID
	response.Warnings = result.Warnings
	return response, nil
}
