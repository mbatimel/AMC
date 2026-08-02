package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/products/internal/errors"
	internalModels "github.com/mbatimel/AMC/products/internal/models"
	"github.com/mbatimel/AMC/products/internal/storage/postgres"
	"github.com/mbatimel/AMC/products/pkg/models"
)

const (
	defaultLimit          = 20
	maxLimit              = 10000
	maxSKUSize            = 255
	maxNameSize           = 255
	maxCharacteristicSize = 255
	maxImageAltSize       = 500
	maxProductDescription = 10000
	maxProductImages      = 50
	maxPromotionName      = 255
)

type Storage interface {
	CreateProduct(ctx context.Context, params internalModels.CreateProductParams) (internalModels.Product, error)
	GetProductByID(ctx context.Context, productID uuid.UUID) (internalModels.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (internalModels.Product, error)
	ListProducts(ctx context.Context, params internalModels.ListProductsParams) ([]internalModels.Product, error)
	CountProducts(ctx context.Context, params internalModels.ListProductsParams) (int, error)
	UpdateProduct(ctx context.Context, params internalModels.UpdateProductParams) (internalModels.Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	ListProductImages(ctx context.Context, productID uuid.UUID) ([]internalModels.ProductImage, error)
	AddProductImages(ctx context.Context, productID uuid.UUID, images []internalModels.ProductImage) ([]internalModels.ProductImage, error)
	ReplaceProductImages(ctx context.Context, productID uuid.UUID, images []internalModels.ProductImage) ([]internalModels.ProductImage, error)
	DeleteProductImages(ctx context.Context, productID uuid.UUID) error
	AddUploadedProductImage(ctx context.Context, productID uuid.UUID, image internalModels.ProductImage) (internalModels.ProductImage, error)
	GetProductImage(ctx context.Context, productID uuid.UUID, imageID uuid.UUID) (internalModels.ProductImage, error)
	DeleteProductImage(ctx context.Context, productID uuid.UUID, imageID uuid.UUID) error
	GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (internalModels.Category, error)
	ListCategories(ctx context.Context, limit int, offset int) ([]internalModels.Category, error)
	CountCategories(ctx context.Context) (int, error)
	GetBrandByID(ctx context.Context, brandID uuid.UUID) (internalModels.Brand, error)
	ListBrands(ctx context.Context, limit int, offset int) ([]internalModels.Brand, error)
	CountBrands(ctx context.Context) (int, error)
	CreatePromotion(ctx context.Context, params internalModels.CreatePromotionParams) (internalModels.Promotion, error)
	GetPromotionByID(ctx context.Context, promotionID uuid.UUID) (internalModels.Promotion, error)
	ListPromotions(ctx context.Context, limit int, offset int) ([]internalModels.Promotion, error)
	CountPromotions(ctx context.Context) (int, error)
	UpdatePromotion(ctx context.Context, params internalModels.UpdatePromotionParams) (internalModels.Promotion, error)
	DeletePromotion(ctx context.Context, promotionID uuid.UUID) error
}

type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (bool, error)
}

type ObjectStorage interface {
	Upload(ctx context.Context, objectKey string, body io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, objectKey string) error
	URL(objectKey string) string
}

type Option func(*Service)

func WithObjectStorage(storage ObjectStorage, maxFileSize int64) Option {
	return func(s *Service) {
		s.objectStorage = storage
		s.maxFileSize = maxFileSize
	}
}

type Service struct {
	logger        zerolog.Logger
	storage       Storage
	accessClient  AccessClient
	objectStorage ObjectStorage
	maxFileSize   int64
}

func New(logger zerolog.Logger, storage Storage, accessClient AccessClient, options ...Option) *Service {
	result := &Service{logger: logger, storage: storage, accessClient: accessClient}
	for _, option := range options {
		option(result)
	}
	return result
}

func validation(field string) error {
	return customErrors.ErrValidation.AddCause("field", field)
}

func mapStorageError(err error) error {
	switch {
	case errors.Is(err, postgres.ErrProductNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "product")
	case errors.Is(err, postgres.ErrCategoryNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "category")
	case errors.Is(err, postgres.ErrBrandNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "brand")
	case errors.Is(err, postgres.ErrSKUTaken):
		return customErrors.ErrConflict.AddCause("field", "sku")
	case errors.Is(err, postgres.ErrInvalidSort):
		return validation("sort")
	case errors.Is(err, postgres.ErrProductImageNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "productImage")
	case errors.Is(err, postgres.ErrPromotionNotFound):
		return customErrors.ErrNotFound.AddCause("entity", "promotion")
	default:
		return customErrors.ErrInternal
	}
}

func (s *Service) checkAdminAccess(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return validation("X-User-Id")
	}
	if s.accessClient == nil {
		return customErrors.ErrInternal
	}
	allowed, err := s.accessClient.CheckAccess(ctx, userID, roleCodeAdmin)
	if err != nil {
		return customErrors.ErrInternal
	}
	if !allowed {
		return customErrors.ErrForbidden
	}
	return nil
}

var imageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func validateImageFile(file models.ImageFile, maxFileSize int64) (string, string, error) {
	name := strings.TrimSpace(file.FileName)
	if name == "" || filepath.Base(name) != name || strings.Contains(name, "..") {
		return "", "", validation("fileName")
	}
	if len(file.Content) == 0 {
		return "", "", validation("file")
	}
	if maxFileSize <= 0 || int64(len(file.Content)) > maxFileSize {
		return "", "", validation("fileSize")
	}
	contentType := http.DetectContentType(file.Content[:min(len(file.Content), 512)])
	extension, ok := imageTypes[contentType]
	if !ok {
		return "", "", validation("contentType")
	}
	suppliedExtension := strings.ToLower(filepath.Ext(name))
	validExtension := suppliedExtension == extension || contentType == "image/jpeg" && suppliedExtension == ".jpeg"
	if !validExtension {
		return "", "", validation("fileName")
	}
	return contentType, extension, nil
}

func (s *Service) uploadProductImage(
	ctx context.Context,
	productID uuid.UUID,
	file models.ImageFile,
) (models.ProductImage, error) {
	if s.objectStorage == nil {
		return models.ProductImage{}, customErrors.ErrInternal
	}
	if productID == uuid.Nil {
		return models.ProductImage{}, validation("productID")
	}
	if _, err := s.storage.GetProductByID(ctx, productID); err != nil {
		return models.ProductImage{}, mapStorageError(err)
	}
	contentType, extension, err := validateImageFile(file, s.maxFileSize)
	if err != nil {
		return models.ProductImage{}, err
	}
	images, err := s.storage.ListProductImages(ctx, productID)
	if err != nil {
		return models.ProductImage{}, mapStorageError(err)
	}
	objectKey := fmt.Sprintf("products/%s/%s%s", productID, uuid.NewString(), extension)
	if err = s.objectStorage.Upload(ctx, objectKey, bytes.NewReader(file.Content), int64(len(file.Content)), contentType); err != nil {
		return models.ProductImage{}, customErrors.ErrInternal
	}
	image, err := s.storage.AddUploadedProductImage(ctx, productID, internalModels.ProductImage{
		ProductID:    productID,
		URL:          s.objectStorage.URL(objectKey),
		ObjectKey:    objectKey,
		OriginalName: file.FileName,
		ContentType:  contentType,
		SizeBytes:    int64(len(file.Content)),
		SortOrder:    len(images),
		IsPrimary:    len(images) == 0,
	})
	if err != nil {
		if cleanupErr := s.objectStorage.Delete(ctx, objectKey); cleanupErr != nil {
			s.logger.Error().Err(cleanupErr).Str("objectKey", objectKey).Msg("failed to compensate product image upload")
		}
		return models.ProductImage{}, mapStorageError(err)
	}
	return modelImage(image), nil
}

func (s *Service) UploadProductImage(ctx context.Context, userID uuid.UUID, productID uuid.UUID, file models.ImageFile) (models.UploadProductImageResponse, error) {
	if err := s.checkAdminAccess(ctx, userID); err != nil {
		return models.UploadProductImageResponse{}, err
	}
	image, err := s.uploadProductImage(ctx, productID, file)
	if err != nil {
		return models.UploadProductImageResponse{}, err
	}
	return models.UploadProductImageResponse{Image: image}, nil
}

func batchErrorText(err error) string {
	var customErr *customErrors.Error
	if errors.As(err, &customErr) {
		return customErr.ErrorText
	}
	return customErrors.ErrInternal.ErrorText
}

func skuFromFileName(fileName string) (string, error) {
	name := strings.TrimSpace(fileName)
	if name == "" || filepath.Base(name) != name || strings.Contains(name, "..") {
		return "", validation("fileName")
	}
	extension := filepath.Ext(name)
	if extension == "" {
		return "", validation("fileName")
	}
	sku := strings.TrimSuffix(name, extension)
	if sku == "" {
		return "", validation("fileName")
	}
	return sku, nil
}

func (s *Service) UploadProductImagesBatch(ctx context.Context, userID uuid.UUID, files []models.ImageFile) (models.UploadProductImagesBatchResponse, error) {
	if err := s.checkAdminAccess(ctx, userID); err != nil {
		return models.UploadProductImagesBatchResponse{}, err
	}
	response := models.UploadProductImagesBatchResponse{Items: make([]models.ProductImageBatchResult, 0, len(files))}
	for _, file := range files {
		item := models.ProductImageBatchResult{FileName: file.FileName}
		sku, err := skuFromFileName(file.FileName)
		item.SKU = sku
		if err == nil {
			var product internalModels.Product
			product, err = s.storage.GetProductBySKU(ctx, sku)
			if err == nil {
				var image models.ProductImage
				image, err = s.uploadProductImage(ctx, product.ID, file)
				if err == nil {
					item.Success = true
					item.Image = &image
				}
			} else {
				err = mapStorageError(err)
			}
		}
		if err != nil {
			item.ErrorText = batchErrorText(err)
		}
		response.Items = append(response.Items, item)
	}
	return response, nil
}

func (s *Service) DeleteProductImage(ctx context.Context, userID uuid.UUID, productID uuid.UUID, imageID uuid.UUID) (models.DeleteProductImageResponse, error) {
	if err := s.checkAdminAccess(ctx, userID); err != nil {
		return models.DeleteProductImageResponse{}, err
	}
	if productID == uuid.Nil || imageID == uuid.Nil {
		return models.DeleteProductImageResponse{}, validation("productImageID")
	}
	image, err := s.storage.GetProductImage(ctx, productID, imageID)
	if err != nil {
		return models.DeleteProductImageResponse{}, mapStorageError(err)
	}
	if image.ObjectKey != "" {
		if err = s.objectStorage.Delete(ctx, image.ObjectKey); err != nil {
			return models.DeleteProductImageResponse{}, customErrors.ErrInternal
		}
	}
	if err = s.storage.DeleteProductImage(ctx, productID, imageID); err != nil {
		return models.DeleteProductImageResponse{}, mapStorageError(err)
	}
	return models.DeleteProductImageResponse{Deleted: true}, nil
}

func validateLength(field, value string, maxLength int) error {
	if utf8.RuneCountInString(value) > maxLength {
		return validation(field)
	}
	return nil
}

func validateRequired(field, value string, maxLength int) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", validation(field)
	}
	if err := validateLength(field, value, maxLength); err != nil {
		return "", err
	}
	return value, nil
}

func validateMoney(field string, value float64) error {
	if value < 0 {
		return validation(field)
	}
	return nil
}

func validateDiscount(value float64) error {
	if value < 0 || value > 100 {
		return validation("discountPercent")
	}
	return nil
}

func validateQuantity(field string, value int) error {
	if value < 0 {
		return validation(field)
	}
	return nil
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

func normalizeImages(images []models.ProductImage) ([]internalModels.ProductImage, error) {
	if len(images) > maxProductImages {
		return nil, validation("images")
	}
	result := make([]internalModels.ProductImage, 0, len(images))
	primaryCount := 0
	for _, image := range images {
		rawURL := strings.TrimSpace(image.URL)
		if rawURL == "" {
			return nil, validation("images.url")
		}
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Scheme == "" {
			return nil, validation("images.url")
		}
		switch parsed.Scheme {
		case "http", "https":
			if parsed.Host == "" {
				return nil, validation("images.url")
			}
		case "s3":
			if parsed.Host == "" && strings.Trim(parsed.Path, "/") == "" {
				return nil, validation("images.url")
			}
		default:
			return nil, validation("images.url")
		}
		alt := strings.TrimSpace(image.Alt)
		if err = validateLength("images.alt", alt, maxImageAltSize); err != nil {
			return nil, err
		}
		if image.SortOrder < 0 {
			return nil, validation("images.sortOrder")
		}
		if image.IsPrimary {
			primaryCount++
		}
		result = append(result, internalModels.ProductImage{
			URL:       rawURL,
			Alt:       alt,
			SortOrder: image.SortOrder,
			IsPrimary: image.IsPrimary,
		})
	}
	if primaryCount > 1 {
		return nil, validation("images.isPrimary")
	}
	if len(result) > 0 && primaryCount == 0 {
		result[0].IsPrimary = true
	}
	return result, nil
}

func modelImage(image internalModels.ProductImage) models.ProductImage {
	return models.ProductImage{
		ID:        image.ID.String(),
		ProductID: image.ProductID.String(),
		URL:       image.URL,
		Alt:       image.Alt,
		SortOrder: image.SortOrder,
		IsPrimary: image.IsPrimary,
		CreatedAt: image.CreatedAt,
		UpdatedAt: image.UpdatedAt,
	}
}

func modelImages(images []internalModels.ProductImage) []models.ProductImage {
	result := make([]models.ProductImage, 0, len(images))
	for _, image := range images {
		result = append(result, modelImage(image))
	}
	return result
}

func uuidOrEmpty(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

func modelProduct(product internalModels.Product) models.Product {
	return models.Product{
		ID:              product.ID.String(),
		SKU:             product.SKU,
		Name:            product.Name,
		Description:     product.Description,
		CategoryID:      product.CategoryID.String(),
		CategoryName:    product.CategoryName,
		BrandID:         uuidOrEmpty(product.BrandID),
		BrandName:       product.BrandName,
		GOST:            product.GOST,
		Material:        product.Material,
		Size:            product.Size,
		PackageQty:      product.PackageQty,
		StockQty:        product.StockQty,
		BasePrice:       product.BasePrice,
		ClientPrice:     product.ClientPrice,
		DiscountPercent: product.DiscountPercent,
		Images:          modelImages(product.Images),
		IsPublished:     product.IsPublished,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}
}

func modelProductListItem(product internalModels.Product) models.ProductListItem {
	return models.ProductListItem{
		ID:              product.ID.String(),
		SKU:             product.SKU,
		Name:            product.Name,
		CategoryID:      product.CategoryID.String(),
		CategoryName:    product.CategoryName,
		BrandID:         uuidOrEmpty(product.BrandID),
		BrandName:       product.BrandName,
		GOST:            product.GOST,
		Material:        product.Material,
		Size:            product.Size,
		PackageQty:      product.PackageQty,
		StockQty:        product.StockQty,
		BasePrice:       product.BasePrice,
		ClientPrice:     product.ClientPrice,
		DiscountPercent: product.DiscountPercent,
		Images:          modelImages(product.Images),
		IsPublished:     product.IsPublished,
		CreatedAt:       product.CreatedAt,
		UpdatedAt:       product.UpdatedAt,
	}
}

func modelCategory(category internalModels.Category) models.Category {
	result := models.Category{
		ID:        category.ID.String(),
		Name:      category.Name,
		Slug:      category.Slug,
		SortOrder: category.SortOrder,
		IsActive:  category.IsActive,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
	if category.ParentID.Valid {
		result.ParentID = category.ParentID.UUID.String()
	}
	return result
}

func modelBrand(brand internalModels.Brand) models.Brand {
	return models.Brand{
		ID:        brand.ID.String(),
		Name:      brand.Name,
		Slug:      brand.Slug,
		IsActive:  brand.IsActive,
		CreatedAt: brand.CreatedAt,
		UpdatedAt: brand.UpdatedAt,
	}
}

func (s *Service) CreateProduct(
	ctx context.Context,
	userID uuid.UUID,
	sku string,
	name string,
	description string,
	categoryID uuid.UUID,
	brandID uuid.UUID,
	gost string,
	material string,
	size string,
	packageQty int,
	stockQty int,
	basePrice float64,
	_ float64, // clientPrice: derived from basePrice/discountPercent, see computeClientPrice
	discountPercent float64,
	images []models.ProductImage,
	isPublished bool,
) (response models.CreateProductResponse, err error) {
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if sku, err = validateRequired("sku", sku, maxSKUSize); err != nil {
		return response, err
	}
	if name, err = validateRequired("name", name, maxNameSize); err != nil {
		return response, err
	}
	description = strings.TrimSpace(description)
	if err = validateLength("description", description, maxProductDescription); err != nil {
		return response, err
	}
	if categoryID == uuid.Nil {
		return response, validation("categoryID")
	}
	if brandID == uuid.Nil {
		return response, validation("brandID")
	}
	for field, value := range map[string]string{
		"gost": gost, "material": material, "size": size,
	} {
		if err = validateLength(field, strings.TrimSpace(value), maxCharacteristicSize); err != nil {
			return response, err
		}
	}
	if err = validateQuantity("packageQty", packageQty); err != nil {
		return response, err
	}
	if err = validateQuantity("stockQty", stockQty); err != nil {
		return response, err
	}
	if err = validateMoney("basePrice", basePrice); err != nil {
		return response, err
	}
	if err = validateDiscount(discountPercent); err != nil {
		return response, err
	}
	normalizedImages, err := normalizeImages(images)
	if err != nil {
		return response, err
	}
	if _, err = s.storage.GetProductBySKU(ctx, sku); err == nil {
		return response, customErrors.ErrConflict.AddCause("field", "sku")
	} else if !errors.Is(err, postgres.ErrProductNotFound) {
		return response, mapStorageError(err)
	}
	if _, err = s.storage.GetCategoryByID(ctx, categoryID); err != nil {
		return response, mapStorageError(err)
	}
	if _, err = s.storage.GetBrandByID(ctx, brandID); err != nil {
		return response, mapStorageError(err)
	}
	product, err := s.storage.CreateProduct(ctx, internalModels.CreateProductParams{
		SKU:             sku,
		Name:            name,
		Description:     description,
		CategoryID:      categoryID,
		BrandID:         brandID,
		GOST:            strings.TrimSpace(gost),
		Material:        strings.TrimSpace(material),
		Size:            strings.TrimSpace(size),
		PackageQty:      packageQty,
		StockQty:        stockQty,
		BasePrice:       basePrice,
		DiscountPercent: discountPercent,
		Images:          normalizedImages,
		IsPublished:     isPublished,
	})
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Product = modelProduct(product)
	return response, nil
}

func (s *Service) GetProduct(
	ctx context.Context,
	productID uuid.UUID,
) (response models.GetProductResponse, err error) {
	if productID == uuid.Nil {
		return response, validation("productID")
	}
	product, err := s.storage.GetProductByID(ctx, productID)
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Product = modelProduct(product)
	return response, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	return &normalized
}

func parseOptionalUUID(value *string, field string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	parsed, err := uuid.Parse(strings.TrimSpace(*value))
	if err != nil || parsed == uuid.Nil {
		return nil, validation(field)
	}
	return &parsed, nil
}

func pagination(limit *int, offset *int) (int, int, error) {
	resultLimit := defaultLimit
	resultOffset := 0
	if limit != nil {
		if *limit < 0 || *limit > maxLimit {
			return 0, 0, validation("limit")
		}
		resultLimit = *limit
	}
	if offset != nil {
		if *offset < 0 {
			return 0, 0, validation("offset")
		}
		resultOffset = *offset
	}
	return resultLimit, resultOffset, nil
}

func normalizeSort(sort *string) (string, error) {
	if sort == nil {
		return "", nil
	}
	value := strings.TrimSpace(*sort)
	allowed := map[string]struct{}{
		"": {}, "created_at_desc": {}, "price_asc": {}, "price_desc": {},
		"name_asc": {}, "name_desc": {}, "stock_desc": {},
	}
	if _, ok := allowed[value]; !ok {
		return "", validation("sort")
	}
	return value, nil
}

func (s *Service) ListProducts(
	ctx context.Context,
	q *string,
	categoryID *string,
	brandID *string,
	material *string,
	size *string,
	gost *string,
	inStock *bool,
	limit *int,
	offset *int,
	sort *string,
) (response models.ListProductsResponse, err error) {
	resultLimit, resultOffset, err := pagination(limit, offset)
	if err != nil {
		return response, err
	}
	parsedCategoryID, err := parseOptionalUUID(categoryID, "categoryID")
	if err != nil {
		return response, err
	}
	parsedBrandID, err := parseOptionalUUID(brandID, "brandID")
	if err != nil {
		return response, err
	}
	sortValue, err := normalizeSort(sort)
	if err != nil {
		return response, err
	}
	params := internalModels.ListProductsParams{
		Q:          normalizeOptionalString(q),
		CategoryID: parsedCategoryID,
		BrandID:    parsedBrandID,
		Material:   normalizeOptionalString(material),
		Size:       normalizeOptionalString(size),
		GOST:       normalizeOptionalString(gost),
		InStock:    inStock,
		Limit:      resultLimit,
		Offset:     resultOffset,
		Sort:       sortValue,
	}
	items, err := s.storage.ListProducts(ctx, params)
	if err != nil {
		return response, mapStorageError(err)
	}
	total, err := s.storage.CountProducts(ctx, params)
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Items = make([]models.ProductListItem, 0, len(items))
	for _, item := range items {
		response.Items = append(response.Items, modelProductListItem(item))
	}
	response.Pagination = models.Pagination{
		Limit: resultLimit, Offset: resultOffset, Total: total,
	}
	return response, nil
}

func hasUpdate(params internalModels.UpdateProductParams) bool {
	return params.SKU != nil ||
		params.Name != nil ||
		params.Description != nil ||
		params.CategoryID != nil ||
		params.BrandID != nil ||
		params.GOST != nil ||
		params.Material != nil ||
		params.Size != nil ||
		params.PackageQty != nil ||
		params.StockQty != nil ||
		params.BasePrice != nil ||
		params.DiscountPercent != nil ||
		params.Images != nil ||
		params.IsPublished != nil
}

func normalizeOptionalField(value *string, field string, maxLength int, required bool) (*string, error) {
	if value == nil {
		return nil, nil
	}
	normalized := strings.TrimSpace(*value)
	if required && normalized == "" {
		return nil, validation(field)
	}
	if err := validateLength(field, normalized, maxLength); err != nil {
		return nil, err
	}
	return &normalized, nil
}

func (s *Service) UpdateProduct(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
	sku *string,
	name *string,
	description *string,
	categoryID *uuid.UUID,
	brandID *uuid.UUID,
	gost *string,
	material *string,
	size *string,
	packageQty *int,
	stockQty *int,
	basePrice *float64,
	_ *float64, // clientPrice: derived from basePrice/discountPercent, see computeClientPrice
	discountPercent *float64,
	images *[]models.ProductImage,
	isPublished *bool,
) (response models.UpdateProductResponse, err error) {
	if productID == uuid.Nil {
		return response, validation("productID")
	}
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	params := internalModels.UpdateProductParams{
		ProductID:       productID,
		CategoryID:      categoryID,
		BrandID:         brandID,
		PackageQty:      packageQty,
		StockQty:        stockQty,
		BasePrice:       basePrice,
		DiscountPercent: discountPercent,
		IsPublished:     isPublished,
	}
	if params.SKU, err = normalizeOptionalField(sku, "sku", maxSKUSize, true); err != nil {
		return response, err
	}
	if params.Name, err = normalizeOptionalField(name, "name", maxNameSize, true); err != nil {
		return response, err
	}
	if params.Description, err = normalizeOptionalField(description, "description", maxProductDescription, false); err != nil {
		return response, err
	}
	if params.GOST, err = normalizeOptionalField(gost, "gost", maxCharacteristicSize, false); err != nil {
		return response, err
	}
	if params.Material, err = normalizeOptionalField(material, "material", maxCharacteristicSize, false); err != nil {
		return response, err
	}
	if params.Size, err = normalizeOptionalField(size, "size", maxCharacteristicSize, false); err != nil {
		return response, err
	}
	if categoryID != nil && *categoryID == uuid.Nil {
		return response, validation("categoryID")
	}
	if brandID != nil && *brandID == uuid.Nil {
		return response, validation("brandID")
	}
	if packageQty != nil {
		if err = validateQuantity("packageQty", *packageQty); err != nil {
			return response, err
		}
	}
	if stockQty != nil {
		if err = validateQuantity("stockQty", *stockQty); err != nil {
			return response, err
		}
	}
	if basePrice != nil {
		if err = validateMoney("basePrice", *basePrice); err != nil {
			return response, err
		}
	}
	if discountPercent != nil {
		if err = validateDiscount(*discountPercent); err != nil {
			return response, err
		}
	}
	if images != nil {
		var normalized []internalModels.ProductImage
		normalized, err = normalizeImages(*images)
		if err != nil {
			return response, err
		}
		params.Images = &normalized
	}
	if !hasUpdate(params) {
		return response, validation("request")
	}
	if categoryID != nil {
		if _, err = s.storage.GetCategoryByID(ctx, *categoryID); err != nil {
			return response, mapStorageError(err)
		}
	}
	if brandID != nil {
		if _, err = s.storage.GetBrandByID(ctx, *brandID); err != nil {
			return response, mapStorageError(err)
		}
	}
	product, err := s.storage.UpdateProduct(ctx, params)
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Product = modelProduct(product)
	return response, nil
}

func (s *Service) DeleteProduct(
	ctx context.Context,
	userID uuid.UUID,
	productID uuid.UUID,
) (response models.DeleteProductResponse, err error) {
	if productID == uuid.Nil {
		return response, validation("productID")
	}
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if err = s.storage.DeleteProduct(ctx, productID); err != nil {
		return response, mapStorageError(err)
	}
	response.Deleted = true
	return response, nil
}

func (s *Service) ListCategories(
	ctx context.Context,
	limit *int,
	offset *int,
) (response models.ListCategoriesResponse, err error) {
	resultLimit, resultOffset, err := pagination(limit, offset)
	if err != nil {
		return response, err
	}
	items, err := s.storage.ListCategories(ctx, resultLimit, resultOffset)
	if err != nil {
		return response, mapStorageError(err)
	}
	total, err := s.storage.CountCategories(ctx)
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Items = make([]models.Category, 0, len(items))
	for _, item := range items {
		response.Items = append(response.Items, modelCategory(item))
	}
	response.Pagination = models.Pagination{
		Limit: resultLimit, Offset: resultOffset, Total: total,
	}
	return response, nil
}

func promotionStatus(startsAt, endsAt time.Time) string {
	now := time.Now().UTC()
	switch {
	case now.Before(startsAt):
		return "scheduled"
	case now.After(endsAt):
		return "ended"
	default:
		return "active"
	}
}

func modelPromotion(promotion internalModels.Promotion) models.Promotion {
	products := make([]models.PromotionProduct, 0, len(promotion.Products))
	for _, product := range promotion.Products {
		products = append(products, models.PromotionProduct{
			ProductID: product.ProductID.String(),
			MinQty:    product.MinQty,
		})
	}
	return models.Promotion{
		ID:              promotion.ID.String(),
		Name:            promotion.Name,
		DiscountPercent: promotion.DiscountPercent,
		StartsAt:        promotion.StartsAt,
		EndsAt:          promotion.EndsAt,
		Status:          promotionStatus(promotion.StartsAt, promotion.EndsAt),
		Products:        products,
		CreatedAt:       promotion.CreatedAt,
		UpdatedAt:       promotion.UpdatedAt,
	}
}

func parsePromotionPeriod(startsAt, endsAt string) (time.Time, time.Time, error) {
	startsAtTime, err := time.Parse(time.RFC3339, strings.TrimSpace(startsAt))
	if err != nil {
		return time.Time{}, time.Time{}, validation("startsAt")
	}
	endsAtTime, err := time.Parse(time.RFC3339, strings.TrimSpace(endsAt))
	if err != nil {
		return time.Time{}, time.Time{}, validation("endsAt")
	}
	if !endsAtTime.After(startsAtTime) {
		return time.Time{}, time.Time{}, validation("endsAt")
	}
	return startsAtTime, endsAtTime, nil
}

func parsePromotionProducts(products []models.PromotionProduct) ([]internalModels.PromotionProduct, error) {
	if len(products) == 0 {
		return nil, validation("products")
	}
	result := make([]internalModels.PromotionProduct, 0, len(products))
	seen := make(map[uuid.UUID]struct{}, len(products))
	for _, product := range products {
		productID, err := uuid.Parse(strings.TrimSpace(product.ProductID))
		if err != nil || productID == uuid.Nil {
			return nil, validation("products.productID")
		}
		if _, ok := seen[productID]; ok {
			return nil, validation("products.productID")
		}
		seen[productID] = struct{}{}
		if product.MinQty < 1 {
			return nil, validation("products.minQty")
		}
		result = append(result, internalModels.PromotionProduct{ProductID: productID, MinQty: product.MinQty})
	}
	return result, nil
}

func (s *Service) checkPromotionProductsExist(ctx context.Context, products []internalModels.PromotionProduct) error {
	for _, product := range products {
		if _, err := s.storage.GetProductByID(ctx, product.ProductID); err != nil {
			return mapStorageError(err)
		}
	}
	return nil
}

func (s *Service) CreatePromotion(
	ctx context.Context,
	userID uuid.UUID,
	name string,
	discountPercent float64,
	startsAt string,
	endsAt string,
	products []models.PromotionProduct,
) (response models.CreatePromotionResponse, err error) {
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if name, err = validateRequired("name", name, maxPromotionName); err != nil {
		return response, err
	}
	if err = validateDiscount(discountPercent); err != nil {
		return response, err
	}
	startsAtTime, endsAtTime, err := parsePromotionPeriod(startsAt, endsAt)
	if err != nil {
		return response, err
	}
	parsedProducts, err := parsePromotionProducts(products)
	if err != nil {
		return response, err
	}
	if err = s.checkPromotionProductsExist(ctx, parsedProducts); err != nil {
		return response, err
	}
	promotion, err := s.storage.CreatePromotion(ctx, internalModels.CreatePromotionParams{
		Name:            name,
		DiscountPercent: discountPercent,
		StartsAt:        startsAtTime,
		EndsAt:          endsAtTime,
		Products:        parsedProducts,
	})
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Promotion = modelPromotion(promotion)
	return response, nil
}

func (s *Service) GetPromotion(
	ctx context.Context,
	promotionID uuid.UUID,
) (response models.GetPromotionResponse, err error) {
	if promotionID == uuid.Nil {
		return response, validation("promotionID")
	}
	promotion, err := s.storage.GetPromotionByID(ctx, promotionID)
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Promotion = modelPromotion(promotion)
	return response, nil
}

func (s *Service) ListPromotions(
	ctx context.Context,
	limit *int,
	offset *int,
) (response models.ListPromotionsResponse, err error) {
	resultLimit, resultOffset, err := pagination(limit, offset)
	if err != nil {
		return response, err
	}
	items, err := s.storage.ListPromotions(ctx, resultLimit, resultOffset)
	if err != nil {
		return response, mapStorageError(err)
	}
	total, err := s.storage.CountPromotions(ctx)
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Items = make([]models.Promotion, 0, len(items))
	for _, item := range items {
		response.Items = append(response.Items, modelPromotion(item))
	}
	response.Pagination = models.Pagination{
		Limit: resultLimit, Offset: resultOffset, Total: total,
	}
	return response, nil
}

func (s *Service) UpdatePromotion(
	ctx context.Context,
	userID uuid.UUID,
	promotionID uuid.UUID,
	name string,
	discountPercent float64,
	startsAt string,
	endsAt string,
	products []models.PromotionProduct,
) (response models.UpdatePromotionResponse, err error) {
	if promotionID == uuid.Nil {
		return response, validation("promotionID")
	}
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if name, err = validateRequired("name", name, maxPromotionName); err != nil {
		return response, err
	}
	if err = validateDiscount(discountPercent); err != nil {
		return response, err
	}
	startsAtTime, endsAtTime, err := parsePromotionPeriod(startsAt, endsAt)
	if err != nil {
		return response, err
	}
	parsedProducts, err := parsePromotionProducts(products)
	if err != nil {
		return response, err
	}
	if err = s.checkPromotionProductsExist(ctx, parsedProducts); err != nil {
		return response, err
	}
	promotion, err := s.storage.UpdatePromotion(ctx, internalModels.UpdatePromotionParams{
		PromotionID:     promotionID,
		Name:            name,
		DiscountPercent: discountPercent,
		StartsAt:        startsAtTime,
		EndsAt:          endsAtTime,
		Products:        parsedProducts,
	})
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Promotion = modelPromotion(promotion)
	return response, nil
}

func (s *Service) DeletePromotion(
	ctx context.Context,
	userID uuid.UUID,
	promotionID uuid.UUID,
) (response models.DeletePromotionResponse, err error) {
	if promotionID == uuid.Nil {
		return response, validation("promotionID")
	}
	if err = s.checkWriteAccess(ctx, userID); err != nil {
		return response, err
	}
	if err = s.storage.DeletePromotion(ctx, promotionID); err != nil {
		return response, mapStorageError(err)
	}
	response.Deleted = true
	return response, nil
}

func (s *Service) ListBrands(
	ctx context.Context,
	limit *int,
	offset *int,
) (response models.ListBrandsResponse, err error) {
	resultLimit, resultOffset, err := pagination(limit, offset)
	if err != nil {
		return response, err
	}
	items, err := s.storage.ListBrands(ctx, resultLimit, resultOffset)
	if err != nil {
		return response, mapStorageError(err)
	}
	total, err := s.storage.CountBrands(ctx)
	if err != nil {
		return response, mapStorageError(err)
	}
	response.Items = make([]models.Brand, 0, len(items))
	for _, item := range items {
		response.Items = append(response.Items, modelBrand(item))
	}
	response.Pagination = models.Pagination{
		Limit: resultLimit, Offset: resultOffset, Total: total,
	}
	return response, nil
}
