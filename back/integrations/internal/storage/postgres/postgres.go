package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mbatimel/AMC/integrations/internal/models"
)

var ErrDuplicateSKU = errors.New("duplicate sku")

const uniqueViolationCode = "23505"

// skuConstraintName — имя unique-констрейнта на products.sku, автоматически
// сгенерированное Postgres для `sku VARCHAR(255) UNIQUE`
// (см. back/migrations/pkg/migrations/data/20260705171939_catalog.sql).
const skuConstraintName = "products_sku_key"

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
