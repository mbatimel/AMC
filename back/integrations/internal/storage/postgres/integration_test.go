package postgres

import (
	"context"
	"errors"
	"fmt"
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

func TestUpsertCategoriesBatch_IdempotentAndAlignedWithInput(t *testing.T) {
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

	items := []models.CategoryInput{
		{OneCGUID: uuid.New(), Name: "Batch A"},
		{OneCGUID: uuid.New(), Name: "Batch B"},
		{OneCGUID: uuid.New(), Name: "Batch C"},
	}

	ids, errs := storage.UpsertCategoriesBatch(ctx, items)
	for i, e := range errs {
		if e != nil {
			t.Fatalf("item %d: unexpected error: %v", i, e)
		}
		if ids[i] == uuid.Nil {
			t.Fatalf("item %d: expected non-nil id", i)
		}
	}

	// Повторный батч с изменёнными именами по тем же one_c_guid — id должны
	// сохраниться, имена обновиться (тот же контракт, что и у ON CONFLICT
	// DO UPDATE в single-row варианте).
	renamed := make([]models.CategoryInput, len(items))
	for i, in := range items {
		renamed[i] = models.CategoryInput{OneCGUID: in.OneCGUID, Name: in.Name + " renamed"}
	}
	ids2, errs2 := storage.UpsertCategoriesBatch(ctx, renamed)
	for i, e := range errs2 {
		if e != nil {
			t.Fatalf("repeat item %d: unexpected error: %v", i, e)
		}
		if ids2[i] != ids[i] {
			t.Fatalf("repeat item %d: expected same id %s, got %s", i, ids[i], ids2[i])
		}
	}

	var name string
	if err = pool.QueryRow(ctx, `SELECT name FROM categories WHERE id = $1`, ids[1]).Scan(&name); err != nil {
		t.Fatalf("read back category: %v", err)
	}
	if name != "Batch B renamed" {
		t.Fatalf("expected updated name %q, got %q", "Batch B renamed", name)
	}
}

func TestUpsertProductsBatch_DuplicateSKUInChunk_FallsBackAndIsolatesFailure(t *testing.T) {
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

	sku := "DUP-SKU-BATCH-" + uuid.New().String()[:8]
	items := []models.ProductInput{
		{OneCGUID: uuid.New(), SKU: sku, Name: "Первый"},
		{OneCGUID: uuid.New(), SKU: sku, Name: "Второй (дубль sku)"},
		{OneCGUID: uuid.New(), SKU: "OK-SKU-" + uuid.New().String()[:8], Name: "Третий"},
	}

	ids, errs := storage.UpsertProductsBatch(ctx, items)

	// Мульти-строчный INSERT падает целиком (проверено вручную через psql —
	// ON CONFLICT конфликтует по one_c_guid, а sku — отдельный unique-констрейнт,
	// не покрытый conflict target'ом), поэтому чанк обязан откатиться на
	// per-row обработку: ровно один из двух дублей должен успеть встать
	// первым и получить ошибку по второму, третий (без конфликта) — пройти.
	successCount, failCount := 0, 0
	for i, e := range errs {
		if e != nil {
			failCount++
			if !errors.Is(e, ErrDuplicateSKU) {
				t.Fatalf("item %d: expected ErrDuplicateSKU, got: %v", i, e)
			}
			continue
		}
		successCount++
		if ids[i] == uuid.Nil {
			t.Fatalf("item %d: success but nil id", i)
		}
	}
	if successCount != 2 || failCount != 1 {
		t.Fatalf("expected 2 successes and 1 duplicate-sku failure, got successes=%d failures=%d (errs=%v)", successCount, failCount, errs)
	}
}

func TestUpsertProductPricesBatch_And_StockBalancesBatch_Idempotent(t *testing.T) {
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

	warehouseID, err := storage.UpsertWarehouse(ctx, models.WarehouseInput{OneCGUID: uuid.New(), Name: "Склад batch"})
	if err != nil {
		t.Fatalf("upsert warehouse: %v", err)
	}

	var productIDs []uuid.UUID
	for i := 0; i < 3; i++ {
		id, upsertErr := storage.UpsertProduct(ctx, models.ProductInput{
			OneCGUID: uuid.New(),
			SKU:      fmt.Sprintf("BATCH-PRICE-SKU-%s-%d", uuid.New().String()[:8], i),
			Name:     "Товар для batch price/stock",
		})
		if upsertErr != nil {
			t.Fatalf("upsert product %d: %v", i, upsertErr)
		}
		productIDs = append(productIDs, id)
	}

	priceItems := make([]models.PriceInput, len(productIDs))
	stockItems := make([]models.StockInput, len(productIDs))
	for i, pid := range productIDs {
		priceItems[i] = models.PriceInput{ProductID: pid, PriceType: "base", Price: float64(100 + i)}
		stockItems[i] = models.StockInput{ProductID: pid, WarehouseID: warehouseID, Quantity: float64(i)}
	}

	if errs := storage.UpsertProductPricesBatch(ctx, priceItems); anyErr(errs) {
		t.Fatalf("unexpected price batch errors: %v", errs)
	}
	if errs := storage.UpsertStockBalancesBatch(ctx, stockItems); anyErr(errs) {
		t.Fatalf("unexpected stock batch errors: %v", errs)
	}

	// Повтор тем же ключом (product_id+price_type / product_id+warehouse_id)
	// с новыми значениями — должен обновить, а не задублировать строки.
	for i := range priceItems {
		priceItems[i].Price += 1
		stockItems[i].Quantity += 10
	}
	if errs := storage.UpsertProductPricesBatch(ctx, priceItems); anyErr(errs) {
		t.Fatalf("unexpected price batch errors (repeat): %v", errs)
	}
	if errs := storage.UpsertStockBalancesBatch(ctx, stockItems); anyErr(errs) {
		t.Fatalf("unexpected stock batch errors (repeat): %v", errs)
	}

	for i, pid := range productIDs {
		var priceCount int
		var price float64
		if err = pool.QueryRow(ctx,
			`SELECT count(*), max(price) FROM product_prices WHERE product_id = $1 AND price_type = 'base'`,
			pid,
		).Scan(&priceCount, &price); err != nil {
			t.Fatalf("count product prices for item %d: %v", i, err)
		}
		if priceCount != 1 || price != priceItems[i].Price {
			t.Fatalf("item %d: expected exactly 1 price row with price=%v, got count=%d price=%v", i, priceItems[i].Price, priceCount, price)
		}

		var stockCount int
		var qty float64
		if err = pool.QueryRow(ctx,
			`SELECT count(*), max(quantity) FROM stock_balances WHERE product_id = $1 AND warehouse_id = $2`,
			pid, warehouseID,
		).Scan(&stockCount, &qty); err != nil {
			t.Fatalf("count stock balances for item %d: %v", i, err)
		}
		if stockCount != 1 || qty != stockItems[i].Quantity {
			t.Fatalf("item %d: expected exactly 1 stock row with quantity=%v, got count=%d quantity=%v", i, stockItems[i].Quantity, stockCount, qty)
		}
	}
}

func anyErr(errs []error) bool {
	for _, e := range errs {
		if e != nil {
			return true
		}
	}
	return false
}
