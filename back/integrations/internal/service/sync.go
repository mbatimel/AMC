package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/integrations/internal/models"
	"github.com/mbatimel/AMC/integrations/internal/onec"
)

const (
	systemCode = "onec_ut"
	systemName = "1С:Управление торговлей 10.3 (UT)"
)

type OnecClient interface {
	FetchCategories(ctx context.Context) ([]onec.CategoryDTO, error)
	FetchWarehouses(ctx context.Context) ([]onec.WarehouseDTO, error)
	FetchProducts(ctx context.Context) ([]onec.ProductDTO, error)
	FetchPrices(ctx context.Context) ([]onec.PriceDTO, error)
	FetchStock(ctx context.Context) ([]onec.StockDTO, error)
}

type Storage interface {
	UpsertIntegrationSystem(ctx context.Context, code, name string) (uuid.UUID, error)
	CreateSyncJob(ctx context.Context, systemID uuid.UUID) (uuid.UUID, error)
	FinishSyncJob(ctx context.Context, jobID uuid.UUID, status string) error
	AddSyncLog(ctx context.Context, jobID, systemID uuid.UUID, level models.SyncLogLevel, message string) error

	UpsertCategory(ctx context.Context, in models.CategoryInput) (uuid.UUID, error)
	SetCategoryParent(ctx context.Context, id, parentID uuid.UUID) error
	UpsertWarehouse(ctx context.Context, in models.WarehouseInput) (uuid.UUID, error)
	UpsertProduct(ctx context.Context, in models.ProductInput) (uuid.UUID, error)
	UpsertProductPrice(ctx context.Context, in models.PriceInput) error
	UpsertStockBalance(ctx context.Context, in models.StockInput) error
}

type Service struct {
	onec    OnecClient
	storage Storage
	logger  zerolog.Logger
}

func New(logger zerolog.Logger, onecClient OnecClient, storage Storage) *Service {
	return &Service{onec: onecClient, storage: storage, logger: logger}
}

func (s *Service) RunSync(ctx context.Context) error {
	systemID, err := s.storage.UpsertIntegrationSystem(ctx, systemCode, systemName)
	if err != nil {
		return fmt.Errorf("upsert integration system: %w", err)
	}
	jobID, err := s.storage.CreateSyncJob(ctx, systemID)
	if err != nil {
		return fmt.Errorf("create sync job: %w", err)
	}

	hadErrors := false
	logStep := func(level models.SyncLogLevel, message string) {
		if logErr := s.storage.AddSyncLog(ctx, jobID, systemID, level, message); logErr != nil {
			s.logger.Error().Err(logErr).Msg("failed to write sync log")
		}
	}

	categoryIDs, err := s.syncCategories(ctx)
	if err != nil {
		logStep(models.SyncLogError, "categories: "+err.Error())
		hadErrors = true
		categoryIDs = map[uuid.UUID]uuid.UUID{}
	}

	warehouseIDs, err := s.syncWarehouses(ctx)
	if err != nil {
		logStep(models.SyncLogError, "warehouses: "+err.Error())
		hadErrors = true
		warehouseIDs = map[uuid.UUID]uuid.UUID{}
	}

	productIDs, skipped, err := s.syncProducts(ctx, categoryIDs)
	if err != nil {
		logStep(models.SyncLogError, "products: "+err.Error())
		hadErrors = true
		productIDs = map[uuid.UUID]uuid.UUID{}
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	if skipped, err = s.syncPrices(ctx, productIDs); err != nil {
		logStep(models.SyncLogError, "prices: "+err.Error())
		hadErrors = true
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	if skipped, err = s.syncStock(ctx, productIDs, warehouseIDs); err != nil {
		logStep(models.SyncLogError, "stock: "+err.Error())
		hadErrors = true
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	status := "success"
	if hadErrors {
		status = "partial"
	}
	return s.storage.FinishSyncJob(ctx, jobID, status)
}

func (s *Service) syncCategories(ctx context.Context) (map[uuid.UUID]uuid.UUID, error) {
	dtos, err := s.onec.FetchCategories(ctx)
	if err != nil {
		return nil, err
	}

	ids := make(map[uuid.UUID]uuid.UUID, len(dtos))
	parents := make(map[uuid.UUID]*uuid.UUID, len(dtos))
	for _, dto := range dtos {
		in, parentGUID, mapErr := mapCategory(dto)
		if mapErr != nil {
			return nil, mapErr
		}
		id, upsertErr := s.storage.UpsertCategory(ctx, in)
		if upsertErr != nil {
			return nil, fmt.Errorf("upsert category %s: %w", in.OneCGUID, upsertErr)
		}
		ids[in.OneCGUID] = id
		parents[in.OneCGUID] = parentGUID
	}
	for oneCGUID, parentGUID := range parents {
		if parentGUID == nil {
			continue
		}
		parentID, ok := ids[*parentGUID]
		if !ok {
			continue
		}
		if err = s.storage.SetCategoryParent(ctx, ids[oneCGUID], parentID); err != nil {
			return nil, fmt.Errorf("set category parent %s: %w", oneCGUID, err)
		}
	}
	return ids, nil
}

func (s *Service) syncWarehouses(ctx context.Context) (map[uuid.UUID]uuid.UUID, error) {
	dtos, err := s.onec.FetchWarehouses(ctx)
	if err != nil {
		return nil, err
	}
	ids := make(map[uuid.UUID]uuid.UUID, len(dtos))
	for _, dto := range dtos {
		in, mapErr := mapWarehouse(dto)
		if mapErr != nil {
			return nil, mapErr
		}
		id, upsertErr := s.storage.UpsertWarehouse(ctx, in)
		if upsertErr != nil {
			return nil, fmt.Errorf("upsert warehouse %s: %w", in.OneCGUID, upsertErr)
		}
		ids[in.OneCGUID] = id
	}
	return ids, nil
}

func (s *Service) syncProducts(ctx context.Context, categoryIDs map[uuid.UUID]uuid.UUID) (map[uuid.UUID]uuid.UUID, []string, error) {
	dtos, err := s.onec.FetchProducts(ctx)
	if err != nil {
		return nil, nil, err
	}
	ids := make(map[uuid.UUID]uuid.UUID, len(dtos))
	var skipped []string
	for _, dto := range dtos {
		in, mapErr := mapProduct(dto, categoryIDs)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("product %s: %s", dto.RefKey, mapErr))
			continue
		}
		id, upsertErr := s.storage.UpsertProduct(ctx, in)
		if upsertErr != nil {
			skipped = append(skipped, fmt.Sprintf("product %s (sku=%s): %s", in.OneCGUID, in.SKU, upsertErr))
			continue
		}
		ids[in.OneCGUID] = id
	}
	return ids, skipped, nil
}

func (s *Service) syncPrices(ctx context.Context, productIDs map[uuid.UUID]uuid.UUID) ([]string, error) {
	dtos, err := s.onec.FetchPrices(ctx)
	if err != nil {
		return nil, err
	}
	var skipped []string
	for _, dto := range dtos {
		in, ok, mapErr := mapPrice(dto, productIDs)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("price for product %s: %s", dto.ProductKey, mapErr))
			continue
		}
		if !ok {
			continue
		}
		if upsertErr := s.storage.UpsertProductPrice(ctx, in); upsertErr != nil {
			skipped = append(skipped, fmt.Sprintf("price for product %s: %s", dto.ProductKey, upsertErr))
		}
	}
	return skipped, nil
}

func (s *Service) syncStock(ctx context.Context, productIDs, warehouseIDs map[uuid.UUID]uuid.UUID) ([]string, error) {
	dtos, err := s.onec.FetchStock(ctx)
	if err != nil {
		return nil, err
	}
	var skipped []string
	for _, dto := range dtos {
		in, ok, mapErr := mapStock(dto, productIDs, warehouseIDs)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("stock for product %s: %s", dto.ProductKey, mapErr))
			continue
		}
		if !ok {
			continue
		}
		if upsertErr := s.storage.UpsertStockBalance(ctx, in); upsertErr != nil {
			skipped = append(skipped, fmt.Sprintf("stock for product %s warehouse %s: %s", dto.ProductKey, dto.WarehouseKey, upsertErr))
		}
	}
	return skipped, nil
}
