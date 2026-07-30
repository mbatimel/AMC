package service

import (
	"context"
	"errors"
	"net/url"
	"strings"
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
	maxLimit              = 100
	maxSKUSize            = 255
	maxNameSize           = 255
	maxCharacteristicSize = 255
	maxImageAltSize       = 500
	maxProductDescription = 10000
	maxProductImages      = 50
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
	GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (internalModels.Category, error)
	ListCategories(ctx context.Context, limit int, offset int) ([]internalModels.Category, error)
	CountCategories(ctx context.Context) (int, error)
	GetBrandByID(ctx context.Context, brandID uuid.UUID) (internalModels.Brand, error)
	ListBrands(ctx context.Context, limit int, offset int) ([]internalModels.Brand, error)
	CountBrands(ctx context.Context) (int, error)
}

type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (bool, error)
}

type Service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
}

func New(logger zerolog.Logger, storage Storage, accessClient AccessClient) *Service {
	return &Service{logger: logger, storage: storage, accessClient: accessClient}
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
	default:
		return customErrors.ErrInternal
	}
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

func modelProduct(product internalModels.Product) models.Product {
	return models.Product{
		ID:              product.ID.String(),
		SKU:             product.SKU,
		Name:            product.Name,
		Description:     product.Description,
		CategoryID:      product.CategoryID.String(),
		CategoryName:    product.CategoryName,
		BrandID:         product.BrandID.String(),
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
		BrandID:         product.BrandID.String(),
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
	clientPrice float64,
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
	if err = validateMoney("clientPrice", clientPrice); err != nil {
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
		ClientPrice:     clientPrice,
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
		params.ClientPrice != nil ||
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
	clientPrice *float64,
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
		ClientPrice:     clientPrice,
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
	if clientPrice != nil {
		if err = validateMoney("clientPrice", *clientPrice); err != nil {
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
