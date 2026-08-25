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

// totalSyncSteps — количество шагов синхронизации (categories, warehouses,
// products, prices, stock). Если ВСЕ шаги провалились на уровне запроса к
// 1С (Fetch* вернул ошибку), финальный статус — "failed".
const totalSyncSteps = 5

// maxSkippedMessages — максимум отдельных сообщений о пропущенных
// элементах, которые шаг пишет в sync_logs. При превышении лишние
// сообщения схлопываются в одну итоговую запись, чтобы патологический
// прогон не создавал десятки тысяч строк лога.
const maxSkippedMessages = 100

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
	FinishSyncJob(ctx context.Context, jobID uuid.UUID, status string, lastError string) error
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
	stepFailures := 0
	firstStepError := ""
	logStep := func(level models.SyncLogLevel, message string) {
		if logErr := s.storage.AddSyncLog(ctx, jobID, systemID, level, message); logErr != nil {
			s.logger.Error().Err(logErr).Msg("failed to write sync log")
		}
	}
	// recordStepFailure отмечает провал ЦЕЛОГО шага (Fetch* вернул ошибку,
	// т.е. 1С недоступна для этой сущности), в отличие от отдельных
	// item-level пропусков внутри успешно выполненного шага.
	recordStepFailure := func(message string) {
		hadErrors = true
		stepFailures++
		if firstStepError == "" {
			firstStepError = message
		}
	}

	categoryIDs, skipped, err := s.syncCategories(ctx)
	if err != nil {
		msg := "categories: " + err.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
		categoryIDs = map[uuid.UUID]uuid.UUID{}
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	warehouseIDs, skipped, err := s.syncWarehouses(ctx)
	if err != nil {
		msg := "warehouses: " + err.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
		warehouseIDs = map[uuid.UUID]uuid.UUID{}
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	productIDs, skipped, err := s.syncProducts(ctx, categoryIDs)
	if err != nil {
		msg := "products: " + err.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
		productIDs = map[uuid.UUID]uuid.UUID{}
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	if skipped, err = s.syncPrices(ctx, productIDs); err != nil {
		msg := "prices: " + err.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	if skipped, err = s.syncStock(ctx, productIDs, warehouseIDs); err != nil {
		msg := "stock: " + err.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
	}
	for _, msg := range skipped {
		logStep(models.SyncLogWarn, msg)
		hadErrors = true
	}

	status := "success"
	switch {
	case stepFailures == totalSyncSteps:
		status = "failed"
	case hadErrors:
		status = "partial"
	}

	lastError := ""
	if stepFailures > 0 {
		lastError = firstStepError
	}

	return s.storage.FinishSyncJob(ctx, jobID, status, lastError)
}

// capSkipped ограничивает количество отдельных сообщений о пропущенных
// элементах, которые шаг возвращает для записи в sync_logs. Если сообщений
// больше maxSkippedMessages, лишние схлопываются в одну итоговую запись.
func capSkipped(skipped []string) []string {
	if len(skipped) <= maxSkippedMessages {
		return skipped
	}
	capped := make([]string, 0, maxSkippedMessages+1)
	capped = append(capped, skipped[:maxSkippedMessages]...)
	capped = append(capped, fmt.Sprintf("... and %d more skipped (truncated)", len(skipped)-maxSkippedMessages))
	return capped
}

func (s *Service) syncCategories(ctx context.Context) (map[uuid.UUID]uuid.UUID, []string, error) {
	dtos, err := s.onec.FetchCategories(ctx)
	if err != nil {
		return nil, nil, err
	}

	ids := make(map[uuid.UUID]uuid.UUID, len(dtos))
	parents := make(map[uuid.UUID]*uuid.UUID, len(dtos))
	var skipped []string
	for _, dto := range dtos {
		in, parentGUID, mapErr := mapCategory(dto)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("category %s: %s", dto.RefKey, mapErr))
			continue
		}
		id, upsertErr := s.storage.UpsertCategory(ctx, in)
		if upsertErr != nil {
			skipped = append(skipped, fmt.Sprintf("category %s: %s", in.OneCGUID, upsertErr))
			continue
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
		if setErr := s.storage.SetCategoryParent(ctx, ids[oneCGUID], parentID); setErr != nil {
			skipped = append(skipped, fmt.Sprintf("category %s: set parent: %s", oneCGUID, setErr))
			continue
		}
	}
	return ids, capSkipped(skipped), nil
}

func (s *Service) syncWarehouses(ctx context.Context) (map[uuid.UUID]uuid.UUID, []string, error) {
	dtos, err := s.onec.FetchWarehouses(ctx)
	if err != nil {
		return nil, nil, err
	}
	ids := make(map[uuid.UUID]uuid.UUID, len(dtos))
	var skipped []string
	for _, dto := range dtos {
		in, mapErr := mapWarehouse(dto)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("warehouse %s: %s", dto.RefKey, mapErr))
			continue
		}
		id, upsertErr := s.storage.UpsertWarehouse(ctx, in)
		if upsertErr != nil {
			skipped = append(skipped, fmt.Sprintf("warehouse %s: %s", in.OneCGUID, upsertErr))
			continue
		}
		ids[in.OneCGUID] = id
	}
	return ids, capSkipped(skipped), nil
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
	return ids, capSkipped(skipped), nil
}

func (s *Service) syncPrices(ctx context.Context, productIDs map[uuid.UUID]uuid.UUID) ([]string, error) {
	dtos, err := s.onec.FetchPrices(ctx)
	if err != nil {
		return nil, err
	}
	var skipped []string
	droppedUnknownProduct := 0
	for _, dto := range dtos {
		in, ok, mapErr := mapPrice(dto, productIDs)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("price for product %s: %s", dto.ProductKey, mapErr))
			continue
		}
		if !ok {
			droppedUnknownProduct++
			continue
		}
		if upsertErr := s.storage.UpsertProductPrice(ctx, in); upsertErr != nil {
			skipped = append(skipped, fmt.Sprintf("price for product %s: %s", dto.ProductKey, upsertErr))
		}
	}
	if droppedUnknownProduct > 0 {
		skipped = append(skipped, fmt.Sprintf("prices: %d rows skipped, referenced product not found in this run", droppedUnknownProduct))
	}
	return capSkipped(skipped), nil
}

func (s *Service) syncStock(ctx context.Context, productIDs, warehouseIDs map[uuid.UUID]uuid.UUID) ([]string, error) {
	dtos, err := s.onec.FetchStock(ctx)
	if err != nil {
		return nil, err
	}
	var skipped []string
	droppedUnknownProduct := 0
	droppedUnknownWarehouse := 0
	for _, dto := range dtos {
		in, ok, mapErr := mapStock(dto, productIDs, warehouseIDs)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("stock for product %s: %s", dto.ProductKey, mapErr))
			continue
		}
		if !ok {
			// mapStock уже успешно распарсил оба ref'а (иначе mapErr != nil
			// выше), поэтому здесь можно безопасно игнорировать ошибку парсинга.
			if productRef, parseErr := uuid.Parse(dto.ProductKey); parseErr == nil {
				if _, known := productIDs[productRef]; !known {
					droppedUnknownProduct++
				}
			}
			if warehouseRef, parseErr := uuid.Parse(dto.WarehouseKey); parseErr == nil {
				if _, known := warehouseIDs[warehouseRef]; !known {
					droppedUnknownWarehouse++
				}
			}
			continue
		}
		if upsertErr := s.storage.UpsertStockBalance(ctx, in); upsertErr != nil {
			skipped = append(skipped, fmt.Sprintf("stock for product %s warehouse %s: %s", dto.ProductKey, dto.WarehouseKey, upsertErr))
		}
	}
	if droppedUnknownProduct > 0 {
		skipped = append(skipped, fmt.Sprintf("stock: %d rows skipped, referenced product not found in this run", droppedUnknownProduct))
	}
	if droppedUnknownWarehouse > 0 {
		skipped = append(skipped, fmt.Sprintf("stock: %d rows skipped, referenced warehouse not found in this run", droppedUnknownWarehouse))
	}
	return capSkipped(skipped), nil
}
