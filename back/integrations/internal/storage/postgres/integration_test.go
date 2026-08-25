package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mbatimel/AMC/integrations/internal/models"
)

// Требует INTEGRATIONS_TEST_DATABASE_URL — Postgres с уже применёнными
// миграциями back/migrations (та же схема, что у products-сервиса).
func TestStorageIntegration(t *testing.T) {
	dsn := os.Getenv("INTEGRATIONS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("INTEGRATIONS_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	storage := New(pool)

	systemID, err := storage.UpsertIntegrationSystem(ctx, "onec_ut_test", "1С:УТ тест")
	if err != nil {
		t.Fatalf("upsert integration system: %v", err)
	}
	systemID2, err := storage.UpsertIntegrationSystem(ctx, "onec_ut_test", "1С:УТ тест обновлено")
	if err != nil {
		t.Fatalf("upsert integration system (idempotent): %v", err)
	}
	if systemID != systemID2 {
		t.Fatalf("expected same system id on repeat upsert, got %s and %s", systemID, systemID2)
	}

	jobID, err := storage.CreateSyncJob(ctx, systemID)
	if err != nil {
		t.Fatalf("create sync job: %v", err)
	}
	if err = storage.AddSyncLog(ctx, jobID, systemID, models.SyncLogInfo, "test log"); err != nil {
		t.Fatalf("add sync log: %v", err)
	}
	if err = storage.FinishSyncJob(ctx, jobID, "success", ""); err != nil {
		t.Fatalf("finish sync job: %v", err)
	}

	parentGUID := uuid.New()
	parentID, err := storage.UpsertCategory(ctx, models.CategoryInput{OneCGUID: parentGUID, Name: "Родитель"})
	if err != nil {
		t.Fatalf("upsert parent category: %v", err)
	}
	childGUID := uuid.New()
	childID, err := storage.UpsertCategory(ctx, models.CategoryInput{OneCGUID: childGUID, Name: "Ребёнок"})
	if err != nil {
		t.Fatalf("upsert child category: %v", err)
	}
	if err = storage.SetCategoryParent(ctx, childID, parentID); err != nil {
		t.Fatalf("set category parent: %v", err)
	}

	childID2, err := storage.UpsertCategory(ctx, models.CategoryInput{OneCGUID: childGUID, Name: "Ребёнок (переименован)"})
	if err != nil {
		t.Fatalf("upsert child category again: %v", err)
	}
	if childID2 != childID {
		t.Fatalf("expected same category id on repeat upsert by one_c_guid, got %s and %s", childID, childID2)
	}

	warehouseGUID := uuid.New()
	warehouseID, err := storage.UpsertWarehouse(ctx, models.WarehouseInput{OneCGUID: warehouseGUID, Name: "Склад тест"})
	if err != nil {
		t.Fatalf("upsert warehouse: %v", err)
	}

	productGUID := uuid.New()
	sku := "TEST-SKU-" + productGUID.String()[:8]
	productID, err := storage.UpsertProduct(ctx, models.ProductInput{
		OneCGUID:   productGUID,
		CategoryID: &childID,
		SKU:        sku,
		Name:       "Тестовый товар",
	})
	if err != nil {
		t.Fatalf("upsert product: %v", err)
	}

	if err = storage.UpsertProductPrice(ctx, models.PriceInput{ProductID: productID, PriceType: "base", Price: 123.45}); err != nil {
		t.Fatalf("upsert product price: %v", err)
	}
	if err = storage.UpsertProductPrice(ctx, models.PriceInput{ProductID: productID, PriceType: "base", Price: 130}); err != nil {
		t.Fatalf("upsert product price (update): %v", err)
	}

	var priceCount int
	if err = pool.QueryRow(ctx,
		`SELECT count(*) FROM product_prices WHERE product_id = $1 AND price_type = $2`,
		productID, "base",
	).Scan(&priceCount); err != nil {
		t.Fatalf("count product prices: %v", err)
	}
	if priceCount != 1 {
		t.Fatalf("expected exactly 1 price row after repeat upsert, got %d", priceCount)
	}

	if err = storage.UpsertStockBalance(ctx, models.StockInput{ProductID: productID, WarehouseID: warehouseID, Quantity: 5}); err != nil {
		t.Fatalf("upsert stock balance: %v", err)
	}
	if err = storage.UpsertStockBalance(ctx, models.StockInput{ProductID: productID, WarehouseID: warehouseID, Quantity: 8}); err != nil {
		t.Fatalf("upsert stock balance (update): %v", err)
	}

	var stockCount int
	var quantity float64
	if err = pool.QueryRow(ctx,
		`SELECT count(*), max(quantity) FROM stock_balances WHERE product_id = $1 AND warehouse_id = $2`,
		productID, warehouseID,
	).Scan(&stockCount, &quantity); err != nil {
		t.Fatalf("count stock balances: %v", err)
	}
	if stockCount != 1 || quantity != 8 {
		t.Fatalf("expected exactly 1 stock row with quantity=8, got count=%d quantity=%v", stockCount, quantity)
	}
}

func TestUpsertProduct_DuplicateSKU_WrapsErrDuplicateSKUWithDetail(t *testing.T) {
	dsn := os.Getenv("INTEGRATIONS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("INTEGRATIONS_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	storage := New(pool)

	sku := "DUP-SKU-" + uuid.New().String()[:8]

	if _, err = storage.UpsertProduct(ctx, models.ProductInput{
		OneCGUID: uuid.New(),
		SKU:      sku,
		Name:     "Товар первый",
	}); err != nil {
		t.Fatalf("upsert first product: %v", err)
	}

	_, err = storage.UpsertProduct(ctx, models.ProductInput{
		OneCGUID: uuid.New(), // другой one_c_guid — конфликт именно по sku
		SKU:      sku,
		Name:     "Товар второй",
	})
	if err == nil {
		t.Fatal("expected error on duplicate sku upsert")
	}
	if !errors.Is(err, ErrDuplicateSKU) {
		t.Fatalf("expected errors.Is(err, ErrDuplicateSKU) to be true, got: %v", err)
	}
	if !strings.Contains(err.Error(), sku) {
		t.Fatalf("expected error message to contain the conflicting sku %q, got: %v", sku, err)
	}
}
