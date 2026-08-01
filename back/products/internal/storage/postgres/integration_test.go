package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	internalModels "github.com/mbatimel/AMC/products/internal/models"
)

func TestStorageProductsIntegration(t *testing.T) {
	dsn := os.Getenv("PRODUCTS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PRODUCTS_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	var categoryID uuid.UUID
	if err = pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug, sort_order)
		VALUES ('Integration category', $1, 10)
		RETURNING id
	`, uuid.NewString()).Scan(&categoryID); err != nil {
		t.Fatalf("insert category: %v", err)
	}
	var brandID uuid.UUID
	if err = pool.QueryRow(ctx, `
		INSERT INTO brands (name, slug)
		VALUES ('Integration brand', $1)
		RETURNING id
	`, uuid.NewString()).Scan(&brandID); err != nil {
		t.Fatalf("insert brand: %v", err)
	}

	productIDs := make([]uuid.UUID, 0, 2)
	t.Cleanup(func() {
		if len(productIDs) > 0 {
			_, _ = pool.Exec(ctx, `DELETE FROM product_images WHERE product_id = ANY($1::uuid[])`, productIDs)
			_, _ = pool.Exec(ctx, `DELETE FROM product_prices WHERE product_id = ANY($1::uuid[])`, productIDs)
			_, _ = pool.Exec(ctx, `DELETE FROM stock_balances WHERE product_id = ANY($1::uuid[])`, productIDs)
			_, _ = pool.Exec(ctx, `DELETE FROM products WHERE id = ANY($1::uuid[])`, productIDs)
		}
		_, _ = pool.Exec(ctx, `DELETE FROM brands WHERE id = $1`, brandID)
		_, _ = pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
	})

	storage := New(pool)
	sku := "integration-" + uuid.NewString()
	created, err := storage.CreateProduct(ctx, internalModels.CreateProductParams{
		SKU:             sku,
		Name:            "Integration product",
		Description:     "Description",
		CategoryID:      categoryID,
		BrandID:         brandID,
		GOST:            "GOST-1",
		Material:        "Steel",
		Size:            "M6",
		PackageQty:      10,
		StockQty:        7,
		BasePrice:       150,
		DiscountPercent: 20,
		Images: []internalModels.ProductImage{
			{URL: "https://s3.example/products/main.jpg", SortOrder: 1, IsPrimary: true},
			{URL: "https://s3.example/products/second.jpg", SortOrder: 2},
		},
		IsPublished: true,
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	productIDs = append(productIDs, created.ID)
	if created.SKU != sku || created.StockQty != 7 || len(created.Images) != 2 {
		t.Fatalf("CreateProduct() = %+v", created)
	}
	for _, image := range created.Images {
		if image.ProductID != created.ID {
			t.Fatalf("image productID = %s, want %s", image.ProductID, created.ID)
		}
	}

	got, err := storage.GetProductByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetProductByID() error = %v", err)
	}
	if got.CategoryID != categoryID || got.BrandID != brandID ||
		got.BasePrice != 150 || got.ClientPrice != 120 || got.DiscountPercent != 20 {
		t.Fatalf("GetProductByID() = %+v", got)
	}
	price, err := storage.GetProductPrice(ctx, created.ID)
	if err != nil || price.BasePrice != 150 || price.ClientPrice != 120 || price.Currency != "RUB" {
		t.Fatalf("GetProductPrice() price=%+v error=%v", price, err)
	}
	stock, err := storage.GetProductStock(ctx, created.ID)
	if err != nil || stock.StockQty != 7 {
		t.Fatalf("GetProductStock() stock=%+v error=%v", stock, err)
	}

	query := "Integration"
	material := "Steel"
	inStock := true
	items, err := storage.ListProducts(ctx, internalModels.ListProductsParams{
		Q:          &query,
		CategoryID: &categoryID,
		BrandID:    &brandID,
		Material:   &material,
		InStock:    &inStock,
		Limit:      10,
		Sort:       "price_asc",
	})
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != created.ID || len(items[0].Images) != 1 {
		t.Fatalf("ListProducts() = %+v; product rows must not be duplicated by images", items)
	}
	total, err := storage.CountProducts(ctx, internalModels.ListProductsParams{
		Q: &query, CategoryID: &categoryID, BrandID: &brandID, Material: &material, InStock: &inStock,
	})
	if err != nil || total != 1 {
		t.Fatalf("CountProducts() total=%d error=%v", total, err)
	}

	updatedName := "Updated integration product"
	zeroStock := 0
	replacementImages := []internalModels.ProductImage{
		{URL: "s3://catalog/updated.jpg", IsPrimary: true},
	}
	updated, err := storage.UpdateProduct(ctx, internalModels.UpdateProductParams{
		ProductID: created.ID,
		Name:      &updatedName,
		StockQty:  &zeroStock,
		Images:    &replacementImages,
	})
	if err != nil {
		t.Fatalf("UpdateProduct() error = %v", err)
	}
	if updated.Name != updatedName || updated.StockQty != 0 || len(updated.Images) != 1 {
		t.Fatalf("UpdateProduct() = %+v", updated)
	}
	inStock = false
	items, err = storage.ListProducts(ctx, internalModels.ListProductsParams{
		InStock: &inStock, Limit: 10, Sort: "stock_desc",
	})
	if err != nil {
		t.Fatalf("ListProducts(inStock=false) error = %v", err)
	}
	found := false
	for _, item := range items {
		if item.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("ListProducts(inStock=false) did not return the zero-stock product")
	}

	if err = storage.DeleteProduct(ctx, created.ID); err != nil {
		t.Fatalf("DeleteProduct() error = %v", err)
	}
	if _, err = storage.GetProductByID(ctx, created.ID); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("GetProductByID(deleted) error = %v, want not found", err)
	}
}

func TestCreateProductRollsBackWhenImageInsertFails(t *testing.T) {
	dsn := os.Getenv("PRODUCTS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PRODUCTS_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	var categoryID, brandID uuid.UUID
	if err = pool.QueryRow(ctx,
		`INSERT INTO categories (name, slug) VALUES ('Rollback category', $1) RETURNING id`,
		uuid.NewString(),
	).Scan(&categoryID); err != nil {
		t.Fatalf("insert category: %v", err)
	}
	if err = pool.QueryRow(ctx,
		`INSERT INTO brands (name, slug) VALUES ('Rollback brand', $1) RETURNING id`,
		uuid.NewString(),
	).Scan(&brandID); err != nil {
		t.Fatalf("insert brand: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM brands WHERE id = $1`, brandID)
		_, _ = pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
	})

	sku := "rollback-" + uuid.NewString()
	_, err = New(pool).CreateProduct(ctx, internalModels.CreateProductParams{
		SKU:        sku,
		Name:       "Rollback product",
		CategoryID: categoryID,
		BrandID:    brandID,
		Images: []internalModels.ProductImage{
			{URL: "https://s3.example/products/one.jpg", IsPrimary: true},
			{URL: "https://s3.example/products/two.jpg", IsPrimary: true},
		},
	})
	if err == nil {
		t.Fatal("CreateProduct() error = nil, want image uniqueness error")
	}

	var count int
	if scanErr := pool.QueryRow(ctx, `SELECT COUNT(*) FROM products WHERE sku = $1`, sku).Scan(&count); scanErr != nil {
		t.Fatalf("count rolled-back product: %v", scanErr)
	}
	if count != 0 {
		t.Fatalf("partially created products = %d, want 0", count)
	}
}
