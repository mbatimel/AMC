package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	internalModels "github.com/mbatimel/AMC/products/internal/models"
)

var (
	ErrProductNotFound      = errors.New("product not found")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrBrandNotFound        = errors.New("brand not found")
	ErrSKUTaken             = errors.New("sku already taken")
	ErrInvalidSort          = errors.New("invalid sort")
	ErrProductImageNotFound = errors.New("product image not found")
)

const (
	uniqueViolationCode     = "23505"
	foreignKeyViolationCode = "23503"
)

//go:embed sql/*.sql
var queries embed.FS

func query(name string) string {
	value, err := queries.ReadFile("sql/" + name)
	if err != nil {
		panic(err)
	}
	return string(value)
}

var (
	sqlCreateProduct           = query("createProduct.sql")
	sqlGetProductByID          = query("getProductByID.sql")
	sqlGetProductBySKU         = query("getProductBySKU.sql")
	sqlListProducts            = query("listProducts.sql")
	sqlCountProducts           = query("countProducts.sql")
	sqlUpdateProduct           = query("updateProduct.sql")
	sqlDeleteProduct           = query("deleteProduct.sql")
	sqlListProductImages       = query("listProductImages.sql")
	sqlAddProductImage         = query("addProductImage.sql")
	sqlDeleteProductImages     = query("deleteProductImages.sql")
	sqlAddUploadedProductImage = query("addUploadedProductImage.sql")
	sqlGetProductImage         = query("getProductImage.sql")
	sqlDeleteProductImage      = query("deleteProductImage.sql")
	sqlUpsertProductPrice      = query("upsertProductPrice.sql")
	sqlReplaceProductStock     = query("replaceProductStock.sql")
	sqlGetProductPrice         = query("getProductPrice.sql")
	sqlGetProductStock         = query("getProductStock.sql")
	sqlGetCategoryByID         = query("getCategoryByID.sql")
	sqlListCategories          = query("listCategories.sql")
	sqlCountCategories         = query("countCategories.sql")
	sqlGetBrandByID            = query("getBrandByID.sql")
	sqlListBrands              = query("listBrands.sql")
	sqlCountBrands             = query("countBrands.sql")
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func scanProduct(row rowScanner) (internalModels.Product, error) {
	var product internalModels.Product
	var categoryID uuid.NullUUID
	var brandID uuid.NullUUID
	err := row.Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&categoryID,
		&product.CategoryName,
		&brandID,
		&product.BrandName,
		&product.GOST,
		&product.Material,
		&product.Size,
		&product.PackageQty,
		&product.StockQty,
		&product.BasePrice,
		&product.ClientPrice,
		&product.DiscountPercent,
		&product.IsPublished,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if categoryID.Valid {
		product.CategoryID = categoryID.UUID
	}
	if brandID.Valid {
		product.BrandID = brandID.UUID
	}
	return product, err
}

func getProductByID(ctx context.Context, db queryer, productID uuid.UUID) (internalModels.Product, error) {
	product, err := scanProduct(db.QueryRow(ctx, sqlGetProductByID, productID))
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.Product{}, ErrProductNotFound
	}
	if err != nil {
		return internalModels.Product{}, fmt.Errorf("get product by id: %w", err)
	}
	return product, nil
}

func classifyWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case uniqueViolationCode:
		if strings.Contains(pgErr.ConstraintName, "sku") {
			return ErrSKUTaken
		}
	case foreignKeyViolationCode:
		switch {
		case strings.Contains(pgErr.ConstraintName, "category"):
			return ErrCategoryNotFound
		case strings.Contains(pgErr.ConstraintName, "brand"):
			return ErrBrandNotFound
		}
	}
	return err
}

func addProductImages(
	ctx context.Context,
	tx pgx.Tx,
	productID uuid.UUID,
	images []internalModels.ProductImage,
) ([]internalModels.ProductImage, error) {
	result := make([]internalModels.ProductImage, 0, len(images))
	for _, image := range images {
		image.ID = uuid.Nil
		image.ProductID = productID
		err := tx.QueryRow(
			ctx,
			sqlAddProductImage,
			productID,
			image.URL,
			image.Alt,
			image.SortOrder,
			image.IsPrimary,
		).Scan(&image.ID, &image.CreatedAt, &image.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("add product image: %w", err)
		}
		result = append(result, image)
	}
	return result, nil
}

// computeClientPrice derives the final (post-promo) price from the base
// price and discount percentage, so callers never need to keep the two in
// sync by hand.
func computeClientPrice(basePrice float64, discountPercent float64) float64 {
	return math.Round(basePrice*(1-discountPercent/100)*100) / 100
}

func upsertProductPrices(
	ctx context.Context,
	tx pgx.Tx,
	productID uuid.UUID,
	basePrice float64,
	discountPercent float64,
) error {
	if _, err := tx.Exec(ctx, sqlUpsertProductPrice, productID, "base", basePrice, 0); err != nil {
		return fmt.Errorf("upsert base price: %w", err)
	}
	clientPrice := computeClientPrice(basePrice, discountPercent)
	if _, err := tx.Exec(ctx, sqlUpsertProductPrice, productID, "client", clientPrice, discountPercent); err != nil {
		return fmt.Errorf("upsert client price: %w", err)
	}
	return nil
}

func (s *Storage) CreateProduct(
	ctx context.Context,
	params internalModels.CreateProductParams,
) (internalModels.Product, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return internalModels.Product{}, fmt.Errorf("begin create product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var productID uuid.UUID
	err = tx.QueryRow(
		ctx,
		sqlCreateProduct,
		params.SKU,
		params.Name,
		params.Description,
		params.CategoryID,
		params.BrandID,
		params.GOST,
		params.Material,
		params.Size,
		params.PackageQty,
		params.IsPublished,
	).Scan(&productID)
	if err != nil {
		return internalModels.Product{}, classifyWriteError(fmt.Errorf("insert product: %w", err))
	}
	if err = upsertProductPrices(
		ctx,
		tx,
		productID,
		params.BasePrice,
		params.DiscountPercent,
	); err != nil {
		return internalModels.Product{}, classifyWriteError(err)
	}
	if _, err = tx.Exec(ctx, sqlReplaceProductStock, productID, params.StockQty); err != nil {
		return internalModels.Product{}, fmt.Errorf("replace product stock: %w", err)
	}
	if _, err = addProductImages(ctx, tx, productID, params.Images); err != nil {
		return internalModels.Product{}, classifyWriteError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return internalModels.Product{}, fmt.Errorf("commit create product: %w", err)
	}
	return s.GetProductByID(ctx, productID)
}

func (s *Storage) GetProductByID(ctx context.Context, productID uuid.UUID) (internalModels.Product, error) {
	product, err := getProductByID(ctx, s.pool, productID)
	if err != nil {
		return internalModels.Product{}, err
	}
	product.Images, err = s.ListProductImages(ctx, productID)
	if err != nil {
		return internalModels.Product{}, err
	}
	return product, nil
}

func (s *Storage) GetProductBySKU(ctx context.Context, sku string) (internalModels.Product, error) {
	product, err := scanProduct(s.pool.QueryRow(ctx, sqlGetProductBySKU, sku))
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.Product{}, ErrProductNotFound
	}
	if err != nil {
		return internalModels.Product{}, fmt.Errorf("get product by sku: %w", err)
	}
	product.Images, err = s.ListProductImages(ctx, product.ID)
	if err != nil {
		return internalModels.Product{}, err
	}
	return product, nil
}

func productFilter(params internalModels.ListProductsParams) (string, []interface{}) {
	clauses := []string{"p.deleted_at IS NULL"}
	args := make([]interface{}, 0, 7)
	add := func(clause string, value interface{}) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if params.Q != nil {
		add(`(
			COALESCE(p.sku, '') ILIKE $%[1]d OR
			COALESCE(p.name, '') ILIKE $%[1]d OR
			COALESCE(p.gost_tu, '') ILIKE $%[1]d OR
			COALESCE(p.material, '') ILIKE $%[1]d OR
			COALESCE(p.size, '') ILIKE $%[1]d
		)`, "%"+*params.Q+"%")
	}
	if params.CategoryID != nil {
		add("p.category_id = $%d", *params.CategoryID)
	}
	if params.BrandID != nil {
		add("p.brand_id = $%d", *params.BrandID)
	}
	if params.Material != nil {
		add("p.material = $%d", *params.Material)
	}
	if params.Size != nil {
		add("p.size = $%d", *params.Size)
	}
	if params.GOST != nil {
		add("p.gost_tu = $%d", *params.GOST)
	}
	if params.InStock != nil {
		if *params.InStock {
			clauses = append(clauses, "COALESCE(stock.stock_qty, 0) > 0")
		} else {
			clauses = append(clauses, "COALESCE(stock.stock_qty, 0) <= 0")
		}
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func safeSort(sort string) (string, error) {
	whitelist := map[string]string{
		"":                "p.created_at DESC, p.id",
		"created_at_desc": "p.created_at DESC, p.id",
		"price_asc":       "COALESCE(prices.client_price, prices.base_price, 0) ASC, p.id",
		"price_desc":      "COALESCE(prices.client_price, prices.base_price, 0) DESC, p.id",
		"name_asc":        "p.name ASC NULLS LAST, p.id",
		"name_desc":       "p.name DESC NULLS LAST, p.id",
		"stock_desc":      "COALESCE(stock.stock_qty, 0) DESC, p.id",
	}
	order, ok := whitelist[sort]
	if !ok {
		return "", ErrInvalidSort
	}
	return order, nil
}

func (s *Storage) ListProducts(
	ctx context.Context,
	params internalModels.ListProductsParams,
) ([]internalModels.Product, error) {
	order, err := safeSort(params.Sort)
	if err != nil {
		return nil, err
	}
	where, args := productFilter(params)
	args = append(args, params.Limit, params.Offset)
	statement := sqlListProducts + where +
		fmt.Sprintf(" ORDER BY %s LIMIT $%d OFFSET $%d", order, len(args)-1, len(args))
	rows, err := s.pool.Query(ctx, statement, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	products := make([]internalModels.Product, 0)
	for rows.Next() {
		product, scanErr := scanProductWithMainImage(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan product: %w", scanErr)
		}
		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}
	return products, nil
}

func scanProductWithMainImage(row rowScanner) (internalModels.Product, error) {
	var product internalModels.Product
	var categoryID uuid.NullUUID
	var brandID uuid.NullUUID
	var image internalModels.ProductImage
	err := row.Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&categoryID,
		&product.CategoryName,
		&brandID,
		&product.BrandName,
		&product.GOST,
		&product.Material,
		&product.Size,
		&product.PackageQty,
		&product.StockQty,
		&product.BasePrice,
		&product.ClientPrice,
		&product.DiscountPercent,
		&product.IsPublished,
		&product.CreatedAt,
		&product.UpdatedAt,
		&image.ID,
		&image.URL,
		&image.Alt,
		&image.SortOrder,
		&image.IsPrimary,
		&image.CreatedAt,
		&image.UpdatedAt,
	)
	if categoryID.Valid {
		product.CategoryID = categoryID.UUID
	}
	if brandID.Valid {
		product.BrandID = brandID.UUID
	}
	if image.ID != uuid.Nil {
		image.ProductID = product.ID
		product.Images = []internalModels.ProductImage{image}
	} else {
		product.Images = make([]internalModels.ProductImage, 0)
	}
	return product, err
}

func (s *Storage) CountProducts(ctx context.Context, params internalModels.ListProductsParams) (int, error) {
	where, args := productFilter(params)
	statement := sqlCountProducts +
		` LEFT JOIN LATERAL (
			SELECT SUM(COALESCE(sb.quantity, 0) - COALESCE(sb.reserved_quantity, 0)) AS stock_qty
			FROM stock_balances sb
			WHERE sb.product_id = p.id
		) stock ON TRUE` + where
	var total int
	if err := s.pool.QueryRow(ctx, statement, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}
	return total, nil
}

func pointerValue[T any](value *T) T {
	var zero T
	if value == nil {
		return zero
	}
	return *value
}

func (s *Storage) UpdateProduct(
	ctx context.Context,
	params internalModels.UpdateProductParams,
) (internalModels.Product, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return internalModels.Product{}, fmt.Errorf("begin update product: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var updatedID uuid.UUID
	err = tx.QueryRow(
		ctx,
		sqlUpdateProduct,
		params.ProductID,
		params.SKU != nil,
		pointerValue(params.SKU),
		params.Name != nil,
		pointerValue(params.Name),
		params.Description != nil,
		pointerValue(params.Description),
		params.CategoryID != nil,
		pointerValue(params.CategoryID),
		params.BrandID != nil,
		pointerValue(params.BrandID),
		params.GOST != nil,
		pointerValue(params.GOST),
		params.Material != nil,
		pointerValue(params.Material),
		params.Size != nil,
		pointerValue(params.Size),
		params.PackageQty != nil,
		pointerValue(params.PackageQty),
		params.IsPublished != nil,
		pointerValue(params.IsPublished),
	).Scan(&updatedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.Product{}, ErrProductNotFound
	}
	if err != nil {
		return internalModels.Product{}, classifyWriteError(fmt.Errorf("update product: %w", err))
	}

	if params.BasePrice != nil || params.DiscountPercent != nil {
		current, getErr := getProductByID(ctx, tx, params.ProductID)
		if getErr != nil {
			return internalModels.Product{}, getErr
		}
		basePrice := current.BasePrice
		discountPercent := current.DiscountPercent
		if params.BasePrice != nil {
			basePrice = *params.BasePrice
		}
		if params.DiscountPercent != nil {
			discountPercent = *params.DiscountPercent
		}
		if err = upsertProductPrices(ctx, tx, params.ProductID, basePrice, discountPercent); err != nil {
			return internalModels.Product{}, classifyWriteError(err)
		}
	}
	if params.StockQty != nil {
		if _, err = tx.Exec(ctx, sqlReplaceProductStock, params.ProductID, *params.StockQty); err != nil {
			return internalModels.Product{}, fmt.Errorf("replace product stock: %w", err)
		}
	}
	if params.Images != nil {
		if _, err = tx.Exec(ctx, sqlDeleteProductImages, params.ProductID); err != nil {
			return internalModels.Product{}, fmt.Errorf("delete product images: %w", err)
		}
		if _, err = addProductImages(ctx, tx, params.ProductID, *params.Images); err != nil {
			return internalModels.Product{}, classifyWriteError(err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return internalModels.Product{}, fmt.Errorf("commit update product: %w", err)
	}
	return s.GetProductByID(ctx, updatedID)
}

func (s *Storage) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, sqlDeleteProduct, productID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (s *Storage) ListProductImages(
	ctx context.Context,
	productID uuid.UUID,
) ([]internalModels.ProductImage, error) {
	rows, err := s.pool.Query(ctx, sqlListProductImages, productID)
	if err != nil {
		return nil, fmt.Errorf("list product images: %w", err)
	}
	defer rows.Close()
	images := make([]internalModels.ProductImage, 0)
	for rows.Next() {
		var image internalModels.ProductImage
		if err = rows.Scan(
			&image.ID,
			&image.ProductID,
			&image.FileID,
			&image.URL,
			&image.ObjectKey,
			&image.OriginalName,
			&image.ContentType,
			&image.SizeBytes,
			&image.Alt,
			&image.SortOrder,
			&image.IsPrimary,
			&image.CreatedAt,
			&image.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan product image: %w", err)
		}
		images = append(images, image)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product images: %w", err)
	}
	return images, nil
}

func (s *Storage) AddUploadedProductImage(
	ctx context.Context,
	productID uuid.UUID,
	image internalModels.ProductImage,
) (internalModels.ProductImage, error) {
	image.ID = uuid.Nil
	image.ProductID = productID
	err := s.pool.QueryRow(
		ctx,
		sqlAddUploadedProductImage,
		productID,
		image.ObjectKey,
		image.OriginalName,
		image.ContentType,
		image.SizeBytes,
		image.URL,
		image.Alt,
		image.SortOrder,
		image.IsPrimary,
	).Scan(&image.ID, &image.FileID, &image.CreatedAt, &image.UpdatedAt)
	if err != nil {
		return internalModels.ProductImage{}, classifyWriteError(fmt.Errorf("add uploaded product image: %w", err))
	}
	return image, nil
}

func (s *Storage) GetProductImage(
	ctx context.Context,
	productID uuid.UUID,
	imageID uuid.UUID,
) (internalModels.ProductImage, error) {
	var image internalModels.ProductImage
	err := s.pool.QueryRow(ctx, sqlGetProductImage, productID, imageID).Scan(
		&image.ID,
		&image.ProductID,
		&image.FileID,
		&image.URL,
		&image.ObjectKey,
		&image.OriginalName,
		&image.ContentType,
		&image.SizeBytes,
		&image.Alt,
		&image.SortOrder,
		&image.IsPrimary,
		&image.CreatedAt,
		&image.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.ProductImage{}, ErrProductImageNotFound
	}
	if err != nil {
		return internalModels.ProductImage{}, fmt.Errorf("get product image: %w", err)
	}
	return image, nil
}

func (s *Storage) DeleteProductImage(ctx context.Context, productID uuid.UUID, imageID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete product image: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var fileID uuid.NullUUID
	if err = tx.QueryRow(ctx, `SELECT file_id FROM product_images WHERE product_id = $1 AND id = $2`, productID, imageID).Scan(&fileID); errors.Is(err, pgx.ErrNoRows) {
		return ErrProductImageNotFound
	} else if err != nil {
		return fmt.Errorf("select product image file: %w", err)
	}
	tag, err := tx.Exec(ctx, sqlDeleteProductImage, productID, imageID)
	if err != nil {
		return fmt.Errorf("delete product image: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrProductImageNotFound
	}
	if fileID.Valid {
		if _, err = tx.Exec(ctx, `DELETE FROM files WHERE id = $1`, fileID.UUID); err != nil {
			return fmt.Errorf("delete product image file metadata: %w", err)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit delete product image: %w", err)
	}
	return nil
}

func (s *Storage) AddProductImages(
	ctx context.Context,
	productID uuid.UUID,
	images []internalModels.ProductImage,
) ([]internalModels.ProductImage, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin add product images: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	result, err := addProductImages(ctx, tx, productID, images)
	if err != nil {
		return nil, classifyWriteError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit add product images: %w", err)
	}
	return result, nil
}

func (s *Storage) ReplaceProductImages(
	ctx context.Context,
	productID uuid.UUID,
	images []internalModels.ProductImage,
) ([]internalModels.ProductImage, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin replace product images: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = tx.Exec(ctx, sqlDeleteProductImages, productID); err != nil {
		return nil, fmt.Errorf("delete product images: %w", err)
	}
	result, err := addProductImages(ctx, tx, productID, images)
	if err != nil {
		return nil, classifyWriteError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit replace product images: %w", err)
	}
	return result, nil
}

func (s *Storage) DeleteProductImages(ctx context.Context, productID uuid.UUID) error {
	if _, err := s.pool.Exec(ctx, sqlDeleteProductImages, productID); err != nil {
		return fmt.Errorf("delete product images: %w", err)
	}
	return nil
}

func (s *Storage) GetProductPrice(
	ctx context.Context,
	productID uuid.UUID,
) (internalModels.ProductPrice, error) {
	var price internalModels.ProductPrice
	price.ProductID = productID
	err := s.pool.QueryRow(ctx, sqlGetProductPrice, productID).Scan(
		&price.ProductID,
		&price.BasePrice,
		&price.ClientPrice,
		&price.DiscountPercent,
		&price.Currency,
		&price.CreatedAt,
		&price.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.ProductPrice{}, ErrProductNotFound
	}
	if err != nil {
		return internalModels.ProductPrice{}, fmt.Errorf("get product price: %w", err)
	}
	return price, nil
}

func (s *Storage) UpdateProductPrice(
	ctx context.Context,
	productID uuid.UUID,
	basePrice float64,
	discountPercent float64,
) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin update product price: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err = getProductByID(ctx, tx, productID); err != nil {
		return err
	}
	if err = upsertProductPrices(ctx, tx, productID, basePrice, discountPercent); err != nil {
		return classifyWriteError(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit update product price: %w", err)
	}
	return nil
}

func (s *Storage) GetProductStock(
	ctx context.Context,
	productID uuid.UUID,
) (internalModels.ProductStock, error) {
	var stock internalModels.ProductStock
	stock.ProductID = productID
	err := s.pool.QueryRow(ctx, sqlGetProductStock, productID).Scan(
		&stock.ProductID,
		&stock.StockQty,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.ProductStock{}, ErrProductNotFound
	}
	if err != nil {
		return internalModels.ProductStock{}, fmt.Errorf("get product stock: %w", err)
	}
	return stock, nil
}

func (s *Storage) UpdateProductStock(ctx context.Context, productID uuid.UUID, stockQty int) error {
	if _, err := getProductByID(ctx, s.pool, productID); err != nil {
		return err
	}
	if _, err := s.pool.Exec(ctx, sqlReplaceProductStock, productID, stockQty); err != nil {
		return fmt.Errorf("update product stock: %w", err)
	}
	return nil
}

func scanCategory(row rowScanner) (internalModels.Category, error) {
	var category internalModels.Category
	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.SortOrder,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	return category, err
}

func (s *Storage) GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (internalModels.Category, error) {
	category, err := scanCategory(s.pool.QueryRow(ctx, sqlGetCategoryByID, categoryID))
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.Category{}, ErrCategoryNotFound
	}
	if err != nil {
		return internalModels.Category{}, fmt.Errorf("get category by id: %w", err)
	}
	return category, nil
}

func (s *Storage) ListCategories(ctx context.Context, limit int, offset int) ([]internalModels.Category, error) {
	rows, err := s.pool.Query(ctx, sqlListCategories, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()
	categories := make([]internalModels.Category, 0)
	for rows.Next() {
		category, scanErr := scanCategory(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan category: %w", scanErr)
		}
		categories = append(categories, category)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}

func (s *Storage) CountCategories(ctx context.Context) (int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, sqlCountCategories).Scan(&total); err != nil {
		return 0, fmt.Errorf("count categories: %w", err)
	}
	return total, nil
}

func scanBrand(row rowScanner) (internalModels.Brand, error) {
	var brand internalModels.Brand
	err := row.Scan(
		&brand.ID,
		&brand.Name,
		&brand.Slug,
		&brand.IsActive,
		&brand.CreatedAt,
		&brand.UpdatedAt,
	)
	return brand, err
}

func (s *Storage) GetBrandByID(ctx context.Context, brandID uuid.UUID) (internalModels.Brand, error) {
	brand, err := scanBrand(s.pool.QueryRow(ctx, sqlGetBrandByID, brandID))
	if errors.Is(err, pgx.ErrNoRows) {
		return internalModels.Brand{}, ErrBrandNotFound
	}
	if err != nil {
		return internalModels.Brand{}, fmt.Errorf("get brand by id: %w", err)
	}
	return brand, nil
}

func (s *Storage) ListBrands(ctx context.Context, limit int, offset int) ([]internalModels.Brand, error) {
	rows, err := s.pool.Query(ctx, sqlListBrands, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list brands: %w", err)
	}
	defer rows.Close()
	brands := make([]internalModels.Brand, 0)
	for rows.Next() {
		brand, scanErr := scanBrand(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan brand: %w", scanErr)
		}
		brands = append(brands, brand)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate brands: %w", err)
	}
	return brands, nil
}

func (s *Storage) CountBrands(ctx context.Context) (int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, sqlCountBrands).Scan(&total); err != nil {
		return 0, fmt.Errorf("count brands: %w", err)
	}
	return total, nil
}
