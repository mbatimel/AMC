package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/products/internal/errors"
	internalModels "github.com/mbatimel/AMC/products/internal/models"
	"github.com/mbatimel/AMC/products/internal/storage/postgres"
	"github.com/mbatimel/AMC/products/pkg/models"
)

type fakeStorage struct {
	createProductFn   func(context.Context, internalModels.CreateProductParams) (internalModels.Product, error)
	getProductByIDFn  func(context.Context, uuid.UUID) (internalModels.Product, error)
	getProductBySKUFn func(context.Context, string) (internalModels.Product, error)
	listProductsFn    func(context.Context, internalModels.ListProductsParams) ([]internalModels.Product, error)
	countProductsFn   func(context.Context, internalModels.ListProductsParams) (int, error)
	updateProductFn   func(context.Context, internalModels.UpdateProductParams) (internalModels.Product, error)
	deleteProductFn   func(context.Context, uuid.UUID) error
	getCategoryByIDFn func(context.Context, uuid.UUID) (internalModels.Category, error)
	listCategoriesFn  func(context.Context, int, int) ([]internalModels.Category, error)
	countCategoriesFn func(context.Context) (int, error)
	getBrandByIDFn    func(context.Context, uuid.UUID) (internalModels.Brand, error)
	listBrandsFn      func(context.Context, int, int) ([]internalModels.Brand, error)
	countBrandsFn     func(context.Context) (int, error)
	lastCreateParams  internalModels.CreateProductParams
	lastListParams    internalModels.ListProductsParams
	lastUpdateParams  internalModels.UpdateProductParams
}

func (f *fakeStorage) CreateProduct(ctx context.Context, params internalModels.CreateProductParams) (internalModels.Product, error) {
	f.lastCreateParams = params
	if f.createProductFn != nil {
		return f.createProductFn(ctx, params)
	}
	return sampleProduct(), nil
}

func (f *fakeStorage) GetProductByID(ctx context.Context, id uuid.UUID) (internalModels.Product, error) {
	if f.getProductByIDFn != nil {
		return f.getProductByIDFn(ctx, id)
	}
	product := sampleProduct()
	product.ID = id
	for index := range product.Images {
		product.Images[index].ProductID = id
	}
	return product, nil
}

func (f *fakeStorage) GetProductBySKU(ctx context.Context, sku string) (internalModels.Product, error) {
	if f.getProductBySKUFn != nil {
		return f.getProductBySKUFn(ctx, sku)
	}
	return internalModels.Product{}, postgres.ErrProductNotFound
}

func (f *fakeStorage) ListProducts(ctx context.Context, params internalModels.ListProductsParams) ([]internalModels.Product, error) {
	f.lastListParams = params
	if f.listProductsFn != nil {
		return f.listProductsFn(ctx, params)
	}
	return []internalModels.Product{sampleProduct()}, nil
}

func (f *fakeStorage) CountProducts(ctx context.Context, params internalModels.ListProductsParams) (int, error) {
	if f.countProductsFn != nil {
		return f.countProductsFn(ctx, params)
	}
	return 1, nil
}

func (f *fakeStorage) UpdateProduct(ctx context.Context, params internalModels.UpdateProductParams) (internalModels.Product, error) {
	f.lastUpdateParams = params
	if f.updateProductFn != nil {
		return f.updateProductFn(ctx, params)
	}
	product := sampleProduct()
	product.ID = params.ProductID
	if params.Name != nil {
		product.Name = *params.Name
	}
	return product, nil
}

func (f *fakeStorage) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if f.deleteProductFn != nil {
		return f.deleteProductFn(ctx, id)
	}
	return nil
}

func (f *fakeStorage) ListProductImages(context.Context, uuid.UUID) ([]internalModels.ProductImage, error) {
	return nil, nil
}

func (f *fakeStorage) AddProductImages(context.Context, uuid.UUID, []internalModels.ProductImage) ([]internalModels.ProductImage, error) {
	return nil, nil
}

func (f *fakeStorage) ReplaceProductImages(context.Context, uuid.UUID, []internalModels.ProductImage) ([]internalModels.ProductImage, error) {
	return nil, nil
}

func (f *fakeStorage) DeleteProductImages(context.Context, uuid.UUID) error {
	return nil
}

func (f *fakeStorage) GetCategoryByID(ctx context.Context, id uuid.UUID) (internalModels.Category, error) {
	if f.getCategoryByIDFn != nil {
		return f.getCategoryByIDFn(ctx, id)
	}
	return internalModels.Category{ID: id, Name: "Tools", IsActive: true}, nil
}

func (f *fakeStorage) ListCategories(ctx context.Context, limit int, offset int) ([]internalModels.Category, error) {
	if f.listCategoriesFn != nil {
		return f.listCategoriesFn(ctx, limit, offset)
	}
	return []internalModels.Category{{ID: uuid.New(), Name: "Tools", IsActive: true}}, nil
}

func (f *fakeStorage) CountCategories(ctx context.Context) (int, error) {
	if f.countCategoriesFn != nil {
		return f.countCategoriesFn(ctx)
	}
	return 1, nil
}

func (f *fakeStorage) GetBrandByID(ctx context.Context, id uuid.UUID) (internalModels.Brand, error) {
	if f.getBrandByIDFn != nil {
		return f.getBrandByIDFn(ctx, id)
	}
	return internalModels.Brand{ID: id, Name: "Brand", IsActive: true}, nil
}

func (f *fakeStorage) ListBrands(ctx context.Context, limit int, offset int) ([]internalModels.Brand, error) {
	if f.listBrandsFn != nil {
		return f.listBrandsFn(ctx, limit, offset)
	}
	return []internalModels.Brand{{ID: uuid.New(), Name: "Brand", IsActive: true}}, nil
}

func (f *fakeStorage) CountBrands(ctx context.Context) (int, error) {
	if f.countBrandsFn != nil {
		return f.countBrandsFn(ctx)
	}
	return 1, nil
}

type fakeAccess struct {
	allowed bool
	err     error
	roles   []int
}

func (f *fakeAccess) CheckAccess(_ context.Context, _ uuid.UUID, role int) (bool, error) {
	f.roles = append(f.roles, role)
	return f.allowed, f.err
}

func sampleProduct() internalModels.Product {
	productID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	now := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	return internalModels.Product{
		ID:              productID,
		SKU:             "SKU-1",
		Name:            "Product",
		CategoryID:      uuid.MustParse("20000000-0000-0000-0000-000000000001"),
		CategoryName:    "Tools",
		BrandID:         uuid.MustParse("30000000-0000-0000-0000-000000000001"),
		BrandName:       "Brand",
		GOST:            "GOST",
		Material:        "Steel",
		Size:            "M6",
		PackageQty:      10,
		StockQty:        5,
		BasePrice:       100,
		ClientPrice:     90,
		DiscountPercent: 10,
		Images: []internalModels.ProductImage{{
			ID:        uuid.MustParse("40000000-0000-0000-0000-000000000001"),
			ProductID: productID,
			URL:       "https://s3.example/products/main.jpg",
			IsPrimary: true,
			CreatedAt: now,
			UpdatedAt: now,
		}},
		IsPublished: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func newTestService(storage *fakeStorage) *Service {
	return New(zerolog.Nop(), storage, &fakeAccess{allowed: true})
}

func callCreate(ctx context.Context, svc *Service) (models.CreateProductResponse, error) {
	product := sampleProduct()
	return svc.CreateProduct(
		ctx,
		uuid.New(),
		product.SKU,
		product.Name,
		"Description",
		product.CategoryID,
		product.BrandID,
		product.GOST,
		product.Material,
		product.Size,
		product.PackageQty,
		product.StockQty,
		product.BasePrice,
		product.ClientPrice,
		product.DiscountPercent,
		[]models.ProductImage{
			{URL: "https://s3.example/products/one.jpg", SortOrder: 1},
			{URL: "https://s3.example/products/two.jpg", SortOrder: 2},
		},
		true,
	)
}

func TestCreateProduct(t *testing.T) {
	storage := &fakeStorage{}
	svc := newTestService(storage)

	response, err := callCreate(context.Background(), svc)
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	if response.Product.ID == "" || len(response.Product.Images) != 1 {
		t.Fatalf("CreateProduct() response = %+v", response)
	}
	if len(storage.lastCreateParams.Images) != 2 {
		t.Fatalf("saved images = %d, want 2", len(storage.lastCreateParams.Images))
	}
	if !storage.lastCreateParams.Images[0].IsPrimary {
		t.Fatal("first image must become primary when no primary image is supplied")
	}
}

func TestCreateProductSKUConflict(t *testing.T) {
	storage := &fakeStorage{
		getProductBySKUFn: func(context.Context, string) (internalModels.Product, error) {
			return sampleProduct(), nil
		},
	}
	_, err := callCreate(context.Background(), newTestService(storage))
	if !errors.Is(err, customErrors.ErrConflict) {
		t.Fatalf("CreateProduct() error = %v, want conflict", err)
	}
}

func TestCreateProductImageFailure(t *testing.T) {
	imageErr := errors.New("image insert failed")
	storage := &fakeStorage{
		createProductFn: func(context.Context, internalModels.CreateProductParams) (internalModels.Product, error) {
			return internalModels.Product{}, imageErr
		},
	}
	response, err := callCreate(context.Background(), newTestService(storage))
	if !errors.Is(err, customErrors.ErrInternal) {
		t.Fatalf("CreateProduct() error = %v, want internal", err)
	}
	if response.Product.ID != "" {
		t.Fatalf("CreateProduct() returned a partially created product: %+v", response)
	}
}

func TestGetProduct(t *testing.T) {
	productID := uuid.New()
	response, err := newTestService(&fakeStorage{}).GetProduct(context.Background(), productID)
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}
	if response.Product.ID != productID.String() {
		t.Fatalf("GetProduct() id = %q, want %q", response.Product.ID, productID)
	}
	if len(response.Product.Images) != 1 || response.Product.Images[0].ProductID != productID.String() {
		t.Fatalf("GetProduct() images = %+v", response.Product.Images)
	}
}

func TestGetProductNotFound(t *testing.T) {
	storage := &fakeStorage{
		getProductByIDFn: func(context.Context, uuid.UUID) (internalModels.Product, error) {
			return internalModels.Product{}, postgres.ErrProductNotFound
		},
	}
	_, err := newTestService(storage).GetProduct(context.Background(), uuid.New())
	if !errors.Is(err, customErrors.ErrNotFound) {
		t.Fatalf("GetProduct() error = %v, want not found", err)
	}
}

func TestListProductsDefaultsAndNilStock(t *testing.T) {
	storage := &fakeStorage{}
	response, err := newTestService(storage).ListProducts(
		context.Background(), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	if storage.lastListParams.Limit != defaultLimit || storage.lastListParams.Offset != 0 {
		t.Fatalf("pagination = %d/%d", storage.lastListParams.Limit, storage.lastListParams.Offset)
	}
	if storage.lastListParams.InStock != nil {
		t.Fatalf("inStock = %v, want nil", storage.lastListParams.InStock)
	}
	if response.Pagination.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("ListProducts() response = %+v", response)
	}
}

func TestListProductsOptionalFilters(t *testing.T) {
	categoryID := uuid.NewString()
	brandID := uuid.NewString()
	q, material, size, gost := " bolt ", " steel ", " M6 ", " GOST "
	inStock := true
	limit, offset := 15, 4
	sort := "price_asc"
	storage := &fakeStorage{}

	_, err := newTestService(storage).ListProducts(
		context.Background(),
		&q,
		&categoryID,
		&brandID,
		&material,
		&size,
		&gost,
		&inStock,
		&limit,
		&offset,
		&sort,
	)
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	params := storage.lastListParams
	if params.Q == nil || *params.Q != "bolt" ||
		params.CategoryID == nil || params.CategoryID.String() != categoryID ||
		params.BrandID == nil || params.BrandID.String() != brandID ||
		params.Material == nil || *params.Material != "steel" ||
		params.Size == nil || *params.Size != "M6" ||
		params.GOST == nil || *params.GOST != "GOST" ||
		params.InStock == nil || !*params.InStock ||
		params.Limit != limit || params.Offset != offset || params.Sort != sort {
		t.Fatalf("ListProducts() params = %+v", params)
	}
}

func TestListProductsEachOptionalFilter(t *testing.T) {
	stringValue := "value"
	categoryID := uuid.NewString()
	brandID := uuid.NewString()
	trueValue := true
	limitValue := 5
	offsetValue := 2
	sortValue := "name_asc"

	tests := []struct {
		name   string
		call   func(*Service) error
		assert func(t *testing.T, params internalModels.ListProductsParams)
	}{
		{
			name: "q",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), &stringValue, nil, nil, nil, nil, nil, nil, nil, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.Q == nil {
					t.Fatal("q pointer was lost")
				}
			},
		},
		{
			name: "categoryID",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, &categoryID, nil, nil, nil, nil, nil, nil, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.CategoryID == nil {
					t.Fatal("categoryID pointer was lost")
				}
			},
		},
		{
			name: "brandID",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, &brandID, nil, nil, nil, nil, nil, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.BrandID == nil {
					t.Fatal("brandID pointer was lost")
				}
			},
		},
		{
			name: "material",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, nil, &stringValue, nil, nil, nil, nil, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.Material == nil {
					t.Fatal("material pointer was lost")
				}
			},
		},
		{
			name: "size",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, nil, nil, &stringValue, nil, nil, nil, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.Size == nil {
					t.Fatal("size pointer was lost")
				}
			},
		},
		{
			name: "gost",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, nil, nil, nil, &stringValue, nil, nil, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.GOST == nil {
					t.Fatal("gost pointer was lost")
				}
			},
		},
		{
			name: "inStock",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, nil, nil, nil, nil, &trueValue, nil, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.InStock == nil {
					t.Fatal("inStock pointer was lost")
				}
			},
		},
		{
			name: "limit",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, nil, nil, nil, nil, nil, &limitValue, nil, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.Limit != limitValue {
					t.Fatalf("limit = %d", params.Limit)
				}
			},
		},
		{
			name: "offset",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, nil, nil, nil, nil, nil, nil, &offsetValue, nil)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.Offset != offsetValue {
					t.Fatalf("offset = %d", params.Offset)
				}
			},
		},
		{
			name: "sort",
			call: func(svc *Service) error {
				_, err := svc.ListProducts(context.Background(), nil, nil, nil, nil, nil, nil, nil, nil, nil, &sortValue)
				return err
			},
			assert: func(t *testing.T, params internalModels.ListProductsParams) {
				if params.Sort != sortValue {
					t.Fatalf("sort = %q", params.Sort)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			storage := &fakeStorage{}
			if err := test.call(newTestService(storage)); err != nil {
				t.Fatalf("ListProducts() error = %v", err)
			}
			test.assert(t, storage.lastListParams)
		})
	}
}

func TestListProductsInStockFalse(t *testing.T) {
	inStock := false
	storage := &fakeStorage{}
	_, err := newTestService(storage).ListProducts(
		context.Background(), nil, nil, nil, nil, nil, nil, &inStock, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	if storage.lastListParams.InStock == nil || *storage.lastListParams.InStock {
		t.Fatalf("inStock = %v, want false pointer", storage.lastListParams.InStock)
	}
}

func TestListProductsRejectsInvalidPagination(t *testing.T) {
	tests := []struct {
		name   string
		limit  *int
		offset *int
	}{
		{name: "negative limit", limit: intPointer(-1)},
		{name: "over max limit", limit: intPointer(maxLimit + 1)},
		{name: "negative offset", offset: intPointer(-1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := newTestService(&fakeStorage{}).ListProducts(
				context.Background(), nil, nil, nil, nil, nil, nil, nil,
				test.limit, test.offset, nil,
			)
			if !errors.Is(err, customErrors.ErrValidation) {
				t.Fatalf("ListProducts() error = %v, want validation", err)
			}
		})
	}
}

func TestListProductsSort(t *testing.T) {
	for _, sort := range []string{
		"price_asc", "price_desc", "name_asc", "name_desc", "stock_desc", "created_at_desc",
	} {
		t.Run(sort, func(t *testing.T) {
			_, err := newTestService(&fakeStorage{}).ListProducts(
				context.Background(), nil, nil, nil, nil, nil, nil, nil, nil, nil, &sort,
			)
			if err != nil {
				t.Fatalf("ListProducts(sort=%q) error = %v", sort, err)
			}
		})
	}
	invalid := "name; DROP TABLE products"
	_, err := newTestService(&fakeStorage{}).ListProducts(
		context.Background(), nil, nil, nil, nil, nil, nil, nil, nil, nil, &invalid,
	)
	if !errors.Is(err, customErrors.ErrValidation) {
		t.Fatalf("ListProducts(invalid sort) error = %v, want validation", err)
	}
}

func TestUpdateProduct(t *testing.T) {
	storage := &fakeStorage{}
	svc := newTestService(storage)
	name := " Updated "
	stock := 0
	published := false
	images := []models.ProductImage{
		{URL: "s3://catalog/product/new.jpg", SortOrder: 0, IsPrimary: true},
	}
	productID := uuid.New()

	response, err := svc.UpdateProduct(
		context.Background(), uuid.New(), productID,
		nil, &name, nil, nil, nil, nil, nil, nil, nil, &stock,
		nil, nil, nil, &images, &published,
	)
	if err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}
	if storage.lastUpdateParams.Name == nil || *storage.lastUpdateParams.Name != "Updated" {
		t.Fatalf("name = %v", storage.lastUpdateParams.Name)
	}
	if storage.lastUpdateParams.StockQty == nil || *storage.lastUpdateParams.StockQty != 0 {
		t.Fatalf("stockQty = %v, zero value was lost", storage.lastUpdateParams.StockQty)
	}
	if storage.lastUpdateParams.IsPublished == nil || *storage.lastUpdateParams.IsPublished {
		t.Fatalf("isPublished = %v, false value was lost", storage.lastUpdateParams.IsPublished)
	}
	if storage.lastUpdateParams.Images == nil || len(*storage.lastUpdateParams.Images) != 1 {
		t.Fatalf("images = %v", storage.lastUpdateParams.Images)
	}
	if response.Product.Name != "Updated" {
		t.Fatalf("UpdateProduct() response = %+v", response)
	}
}

func TestDeleteProduct(t *testing.T) {
	response, err := newTestService(&fakeStorage{}).DeleteProduct(
		context.Background(), uuid.New(), uuid.New(),
	)
	if err != nil {
		t.Fatalf("DeleteProduct() error = %v", err)
	}
	if !response.Deleted {
		t.Fatal("DeleteProduct() deleted = false")
	}
}

func TestListCategories(t *testing.T) {
	response, err := newTestService(&fakeStorage{}).ListCategories(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}
	if len(response.Items) != 1 || response.Pagination.Limit != defaultLimit {
		t.Fatalf("ListCategories() response = %+v", response)
	}
}

func TestListBrands(t *testing.T) {
	response, err := newTestService(&fakeStorage{}).ListBrands(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("ListBrands() error = %v", err)
	}
	if len(response.Items) != 1 || response.Pagination.Limit != defaultLimit {
		t.Fatalf("ListBrands() response = %+v", response)
	}
}

func TestWriteAccessAllowsSupplierAndRejectsBuyer(t *testing.T) {
	storage := &fakeStorage{}
	supplierAccess := &roleAccess{allowedRole: roleCodeSupplier}
	svc := New(zerolog.Nop(), storage, supplierAccess)
	if _, err := callCreate(context.Background(), svc); err != nil {
		t.Fatalf("supplier CreateProduct() error = %v", err)
	}
	if len(supplierAccess.roles) != 2 {
		t.Fatalf("checked roles = %v, want admin and supplier", supplierAccess.roles)
	}

	buyerSvc := New(zerolog.Nop(), storage, &fakeAccess{allowed: false})
	_, err := callCreate(context.Background(), buyerSvc)
	if !errors.Is(err, customErrors.ErrForbidden) {
		t.Fatalf("buyer CreateProduct() error = %v, want forbidden", err)
	}
}

type roleAccess struct {
	allowedRole int
	roles       []int
}

func (r *roleAccess) CheckAccess(_ context.Context, _ uuid.UUID, role int) (bool, error) {
	r.roles = append(r.roles, role)
	return role == r.allowedRole, nil
}

func intPointer(value int) *int {
	return &value
}
