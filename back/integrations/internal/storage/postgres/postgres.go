package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mbatimel/AMC/integrations/internal/models"
)

var ErrDuplicateSKU = errors.New("duplicate sku")

const uniqueViolationCode = "23505"

// skuConstraintName — имя unique-констрейнта на products.sku, автоматически
// сгенерированное Postgres для `sku VARCHAR(255) UNIQUE`
// (см. back/migrations/pkg/migrations/data/20260705171939_catalog.sql).
const skuConstraintName = "products_sku_key"

// defaultBatchSize — размер порции для batch-апсёртов (categories,
// warehouses, products, prices, stock). Каждая порция уходит в Postgres
// одним round-trip'ом вместо одного round-trip'а на строку — на каталоге в
// тысячи SKU это и есть основная экономия времени прогона.
const defaultBatchSize = 500

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
	sqlUpsertIntegrationSystem = query("upsertIntegrationSystem.sql")
	sqlCreateSyncJob           = query("createSyncJob.sql")
	sqlFinishSyncJob           = query("finishSyncJob.sql")
	sqlAddSyncLog              = query("addSyncLog.sql")
	sqlUpsertCategory          = query("upsertCategory.sql")
	sqlSetCategoryParent       = query("setCategoryParent.sql")
	sqlUpsertWarehouse         = query("upsertWarehouse.sql")
	sqlUpsertProduct           = query("upsertProduct.sql")
	sqlUpsertProductPrice      = query("upsertProductPrice.sql")
	sqlUpsertStockBalance      = query("upsertStockBalance.sql")
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) UpsertIntegrationSystem(ctx context.Context, code, name string) (uuid.UUID, error) {
	var id uuid.UUID
	if err := s.pool.QueryRow(ctx, sqlUpsertIntegrationSystem, code, name).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("upsert integration system: %w", err)
	}
	return id, nil
}

func (s *Storage) CreateSyncJob(ctx context.Context, systemID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	if err := s.pool.QueryRow(ctx, sqlCreateSyncJob, systemID).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("create sync job: %w", err)
	}
	return id, nil
}

func (s *Storage) FinishSyncJob(ctx context.Context, jobID uuid.UUID, status string, lastError string) error {
	// Пустая строка означает "ошибок на уровне шага не было" — храним NULL,
	// а не пустую строку, чтобы last_error оставался значимым полем.
	var lastErrorArg interface{}
	if lastError != "" {
		lastErrorArg = lastError
	}
	if _, err := s.pool.Exec(ctx, sqlFinishSyncJob, jobID, status, lastErrorArg); err != nil {
		return fmt.Errorf("finish sync job: %w", err)
	}
	return nil
}

func (s *Storage) AddSyncLog(ctx context.Context, jobID, systemID uuid.UUID, level models.SyncLogLevel, message string) error {
	if _, err := s.pool.Exec(ctx, sqlAddSyncLog, jobID, systemID, string(level), message); err != nil {
		return fmt.Errorf("add sync log: %w", err)
	}
	return nil
}

func (s *Storage) UpsertCategory(ctx context.Context, in models.CategoryInput) (uuid.UUID, error) {
	var id uuid.UUID
	if err := s.pool.QueryRow(ctx, sqlUpsertCategory, in.OneCGUID, in.Name).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("upsert category: %w", err)
	}
	return id, nil
}

func (s *Storage) SetCategoryParent(ctx context.Context, id, parentID uuid.UUID) error {
	if _, err := s.pool.Exec(ctx, sqlSetCategoryParent, id, parentID); err != nil {
		return fmt.Errorf("set category parent: %w", err)
	}
	return nil
}

func (s *Storage) UpsertWarehouse(ctx context.Context, in models.WarehouseInput) (uuid.UUID, error) {
	var id uuid.UUID
	if err := s.pool.QueryRow(ctx, sqlUpsertWarehouse, in.OneCGUID, in.Name).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("upsert warehouse: %w", err)
	}
	return id, nil
}

func (s *Storage) UpsertProduct(ctx context.Context, in models.ProductInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, sqlUpsertProduct, in.OneCGUID, in.CategoryID, in.SKU, in.Name).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode && pgErr.ConstraintName == skuConstraintName {
			return uuid.Nil, fmt.Errorf("%w (constraint=%s): %s", ErrDuplicateSKU, pgErr.ConstraintName, pgErr.Detail)
		}
		return uuid.Nil, fmt.Errorf("upsert product: %w", err)
	}
	return id, nil
}

func (s *Storage) UpsertProductPrice(ctx context.Context, in models.PriceInput) error {
	if _, err := s.pool.Exec(ctx, sqlUpsertProductPrice, in.ProductID, in.PriceType, in.Price); err != nil {
		return fmt.Errorf("upsert product price: %w", err)
	}
	return nil
}

func (s *Storage) UpsertStockBalance(ctx context.Context, in models.StockInput) error {
	if _, err := s.pool.Exec(ctx, sqlUpsertStockBalance, in.ProductID, in.WarehouseID, in.Quantity); err != nil {
		return fmt.Errorf("upsert stock balance: %w", err)
	}
	return nil
}

// chunkRanges режет [0,n) на диапазоны длиной не больше size.
func chunkRanges(n, size int) [][2]int {
	if n == 0 {
		return nil
	}
	ranges := make([][2]int, 0, (n+size-1)/size)
	for start := 0; start < n; start += size {
		end := start + size
		if end > n {
			end = n
		}
		ranges = append(ranges, [2]int{start, end})
	}
	return ranges
}

// UpsertCategoriesBatch — как UpsertCategory, но одним multi-row INSERT на
// порцию до defaultBatchSize строк. Единственный констрейнт на categories —
// UNIQUE(one_c_guid), он же conflict target, поэтому multi-row INSERT
// безопасен и не требует per-row отката в штатном случае.
func (s *Storage) UpsertCategoriesBatch(ctx context.Context, items []models.CategoryInput) ([]uuid.UUID, []error) {
	ids := make([]uuid.UUID, len(items))
	errs := make([]error, len(items))
	for _, r := range chunkRanges(len(items), defaultBatchSize) {
		chunk := items[r[0]:r[1]]
		byGUID, err := s.upsertCategoriesChunk(ctx, chunk)
		if err != nil {
			// Порция целиком не прошла (например сетевой сбой) — откатываемся
			// на per-row обработку той же порции существующим методом, чтобы
			// не терять весь чанк из-за одной проблемной строки.
			for i, in := range chunk {
				id, upsertErr := s.UpsertCategory(ctx, in)
				ids[r[0]+i] = id
				errs[r[0]+i] = upsertErr
			}
			continue
		}
		for i, in := range chunk {
			id, ok := byGUID[in.OneCGUID]
			if !ok {
				errs[r[0]+i] = fmt.Errorf("upsert category: no result returned for %s", in.OneCGUID)
				continue
			}
			ids[r[0]+i] = id
		}
	}
	return ids, errs
}

func (s *Storage) upsertCategoriesChunk(ctx context.Context, chunk []models.CategoryInput) (map[uuid.UUID]uuid.UUID, error) {
	if len(chunk) == 0 {
		return map[uuid.UUID]uuid.UUID{}, nil
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO categories (one_c_guid, name, is_active) VALUES ")
	args := make([]interface{}, 0, len(chunk)*2)
	for i, in := range chunk {
		if i > 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(&sb, "($%d,$%d,TRUE)", i*2+1, i*2+2)
		args = append(args, in.OneCGUID, in.Name)
	}
	sb.WriteString(" ON CONFLICT (one_c_guid) DO UPDATE SET name = EXCLUDED.name RETURNING one_c_guid, id")

	rows, err := s.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("upsert categories batch: %w", err)
	}
	defer rows.Close()
	result := make(map[uuid.UUID]uuid.UUID, len(chunk))
	for rows.Next() {
		var guid, id uuid.UUID
		if err = rows.Scan(&guid, &id); err != nil {
			return nil, fmt.Errorf("upsert categories batch: scan: %w", err)
		}
		result[guid] = id
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("upsert categories batch: %w", err)
	}
	return result, nil
}

// UpsertWarehousesBatch — тот же паттерн, что UpsertCategoriesBatch:
// warehouses тоже имеет единственный констрейнт UNIQUE(one_c_guid).
func (s *Storage) UpsertWarehousesBatch(ctx context.Context, items []models.WarehouseInput) ([]uuid.UUID, []error) {
	ids := make([]uuid.UUID, len(items))
	errs := make([]error, len(items))
	for _, r := range chunkRanges(len(items), defaultBatchSize) {
		chunk := items[r[0]:r[1]]
		byGUID, err := s.upsertWarehousesChunk(ctx, chunk)
		if err != nil {
			for i, in := range chunk {
				id, upsertErr := s.UpsertWarehouse(ctx, in)
				ids[r[0]+i] = id
				errs[r[0]+i] = upsertErr
			}
			continue
		}
		for i, in := range chunk {
			id, ok := byGUID[in.OneCGUID]
			if !ok {
				errs[r[0]+i] = fmt.Errorf("upsert warehouse: no result returned for %s", in.OneCGUID)
				continue
			}
			ids[r[0]+i] = id
		}
	}
	return ids, errs
}

func (s *Storage) upsertWarehousesChunk(ctx context.Context, chunk []models.WarehouseInput) (map[uuid.UUID]uuid.UUID, error) {
	if len(chunk) == 0 {
		return map[uuid.UUID]uuid.UUID{}, nil
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO warehouses (one_c_guid, name, is_active) VALUES ")
	args := make([]interface{}, 0, len(chunk)*2)
	for i, in := range chunk {
		if i > 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(&sb, "($%d,$%d,TRUE)", i*2+1, i*2+2)
		args = append(args, in.OneCGUID, in.Name)
	}
	sb.WriteString(" ON CONFLICT (one_c_guid) DO UPDATE SET name = EXCLUDED.name RETURNING one_c_guid, id")

	rows, err := s.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("upsert warehouses batch: %w", err)
	}
	defer rows.Close()
	result := make(map[uuid.UUID]uuid.UUID, len(chunk))
	for rows.Next() {
		var guid, id uuid.UUID
		if err = rows.Scan(&guid, &id); err != nil {
			return nil, fmt.Errorf("upsert warehouses batch: scan: %w", err)
		}
		result[guid] = id
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("upsert warehouses batch: %w", err)
	}
	return result, nil
}

// UpsertProductsBatch — как UpsertCategoriesBatch, но products вдобавок
// несёт UNIQUE(sku), который НЕ является conflict target'ом ON CONFLICT
// (тот — one_c_guid). Дубль/пустой sku внутри порции валит весь multi-row
// INSERT целиком (атомарность statement'а) — в этом случае вся порция
// откатывается на per-row обработку через UpsertProduct, которая уже умеет
// пропускать именно проблемную строку и оборачивать её в ErrDuplicateSKU
// (см. TestUpsertProduct_DuplicateSKU_WrapsErrDuplicateSKUWithDetail).
func (s *Storage) UpsertProductsBatch(ctx context.Context, items []models.ProductInput) ([]uuid.UUID, []error) {
	ids := make([]uuid.UUID, len(items))
	errs := make([]error, len(items))
	for _, r := range chunkRanges(len(items), defaultBatchSize) {
		chunk := items[r[0]:r[1]]
		byGUID, err := s.upsertProductsChunk(ctx, chunk)
		if err != nil {
			for i, in := range chunk {
				id, upsertErr := s.UpsertProduct(ctx, in)
				ids[r[0]+i] = id
				errs[r[0]+i] = upsertErr
			}
			continue
		}
		for i, in := range chunk {
			id, ok := byGUID[in.OneCGUID]
			if !ok {
				errs[r[0]+i] = fmt.Errorf("upsert product: no result returned for %s", in.OneCGUID)
				continue
			}
			ids[r[0]+i] = id
		}
	}
	return ids, errs
}

func (s *Storage) upsertProductsChunk(ctx context.Context, chunk []models.ProductInput) (map[uuid.UUID]uuid.UUID, error) {
	if len(chunk) == 0 {
		return map[uuid.UUID]uuid.UUID{}, nil
	}
	var sb strings.Builder
	sb.WriteString("INSERT INTO products (one_c_guid, category_id, sku, name, is_active) VALUES ")
	args := make([]interface{}, 0, len(chunk)*4)
	for i, in := range chunk {
		if i > 0 {
			sb.WriteString(",")
		}
		fmt.Fprintf(&sb, "($%d,$%d,$%d,$%d,TRUE)", i*4+1, i*4+2, i*4+3, i*4+4)
		args = append(args, in.OneCGUID, in.CategoryID, in.SKU, in.Name)
	}
	sb.WriteString(` ON CONFLICT (one_c_guid) DO UPDATE
		SET category_id = EXCLUDED.category_id,
		    sku = EXCLUDED.sku,
		    name = EXCLUDED.name,
		    updated_at = now()
		RETURNING one_c_guid, id`)

	rows, err := s.pool.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("upsert products batch: %w", err)
	}
	defer rows.Close()
	result := make(map[uuid.UUID]uuid.UUID, len(chunk))
	for rows.Next() {
		var guid, id uuid.UUID
		if err = rows.Scan(&guid, &id); err != nil {
			return nil, fmt.Errorf("upsert products batch: scan: %w", err)
		}
		result[guid] = id
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("upsert products batch: %w", err)
	}
	return result, nil
}

// UpsertProductPricesBatch — пайплайнит порцию до defaultBatchSize строк
// через pgx.Batch (SendBatch): N SQL-команд уходят в Postgres одним
// round-trip'ом. product_prices не несёт уникальных констрейнтов (см.
// upsertProductPrice.sql — SELECT-then-UPDATE-or-INSERT), а productID в
// items уже гарантированно существует (mapPrice отсеивает неизвестные
// продукты до вызова этого метода) — поэтому в штатной работе батч не
// падает. Если всё же падает (сеть/БД) — вся порция откатывается на
// per-row UpsertProductPrice, как и для остальных batch-методов.
func (s *Storage) UpsertProductPricesBatch(ctx context.Context, items []models.PriceInput) []error {
	errs := make([]error, len(items))
	for _, r := range chunkRanges(len(items), defaultBatchSize) {
		chunk := items[r[0]:r[1]]
		if chunkErr := s.upsertPricesChunk(ctx, chunk); chunkErr != nil {
			for i, in := range chunk {
				errs[r[0]+i] = s.UpsertProductPrice(ctx, in)
			}
		}
	}
	return errs
}

func (s *Storage) upsertPricesChunk(ctx context.Context, chunk []models.PriceInput) error {
	if len(chunk) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, in := range chunk {
		batch.Queue(sqlUpsertProductPrice, in.ProductID, in.PriceType, in.Price)
	}
	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range chunk {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert prices batch: %w", err)
		}
	}
	return nil
}

// UpsertStockBalancesBatch — тот же паттерн, что UpsertProductPricesBatch.
func (s *Storage) UpsertStockBalancesBatch(ctx context.Context, items []models.StockInput) []error {
	errs := make([]error, len(items))
	for _, r := range chunkRanges(len(items), defaultBatchSize) {
		chunk := items[r[0]:r[1]]
		if chunkErr := s.upsertStockChunk(ctx, chunk); chunkErr != nil {
			for i, in := range chunk {
				errs[r[0]+i] = s.UpsertStockBalance(ctx, in)
			}
		}
	}
	return errs
}

func (s *Storage) upsertStockChunk(ctx context.Context, chunk []models.StockInput) error {
	if len(chunk) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, in := range chunk {
		batch.Queue(sqlUpsertStockBalance, in.ProductID, in.WarehouseID, in.Quantity)
	}
	br := s.pool.SendBatch(ctx, batch)
	defer br.Close()
	for range chunk {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("upsert stock batch: %w", err)
		}
	}
	return nil
}
