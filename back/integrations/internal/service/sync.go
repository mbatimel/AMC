package service

import (
	"context"
	"fmt"
	"sync"

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

// Storage — батч-методы возвращают результаты, выровненные по индексу со
// входным слайсом (errs[i] относится к items[i]), как единственная замена
// per-row вызовов: так шаг синка обрабатывает тысячи строк за один-два
// round-trip'а к Postgres вместо одного round-trip'а на строку.
type Storage interface {
	UpsertIntegrationSystem(ctx context.Context, code, name string) (uuid.UUID, error)
	CreateSyncJob(ctx context.Context, systemID uuid.UUID) (uuid.UUID, error)
	FinishSyncJob(ctx context.Context, jobID uuid.UUID, status string, lastError string) error
	AddSyncLog(ctx context.Context, jobID, systemID uuid.UUID, level models.SyncLogLevel, message string) error

	UpsertCategoriesBatch(ctx context.Context, items []models.CategoryInput) (ids []uuid.UUID, errs []error)
	SetCategoryParent(ctx context.Context, id, parentID uuid.UUID) error
	UpsertWarehousesBatch(ctx context.Context, items []models.WarehouseInput) (ids []uuid.UUID, errs []error)
	UpsertProductsBatch(ctx context.Context, items []models.ProductInput) (ids []uuid.UUID, errs []error)
	UpsertProductPricesBatch(ctx context.Context, items []models.PriceInput) (errs []error)
	UpsertStockBalancesBatch(ctx context.Context, items []models.StockInput) (errs []error)
}

type Service struct {
	onec    OnecClient
	storage Storage
	logger  zerolog.Logger
}

func New(logger zerolog.Logger, onecClient OnecClient, storage Storage) *Service {
	return &Service{onec: onecClient, storage: storage, logger: logger}
}

// fetchResult — сырые ответы 1С по всем пяти entity set'ам одного прогона.
type fetchResult struct {
	categories    []onec.CategoryDTO
	categoriesErr error
	warehouses    []onec.WarehouseDTO
	warehousesErr error
	products      []onec.ProductDTO
	productsErr   error
	prices        []onec.PriceDTO
	pricesErr     error
	stock         []onec.StockDTO
	stockErr      error
}

// fetchAll забирает все пять entity set'ов у 1С параллельно — это
// независимые GET-запросы, последовательное ожидание одного за другим было
// главным источником лишнего времени прогона.
func (s *Service) fetchAll(ctx context.Context) fetchResult {
	var res fetchResult
	var wg sync.WaitGroup
	wg.Add(5)
	go func() { defer wg.Done(); res.categories, res.categoriesErr = s.onec.FetchCategories(ctx) }()
	go func() { defer wg.Done(); res.warehouses, res.warehousesErr = s.onec.FetchWarehouses(ctx) }()
	go func() { defer wg.Done(); res.products, res.productsErr = s.onec.FetchProducts(ctx) }()
	go func() { defer wg.Done(); res.prices, res.pricesErr = s.onec.FetchPrices(ctx) }()
	go func() { defer wg.Done(); res.stock, res.stockErr = s.onec.FetchStock(ctx) }()
	wg.Wait()
	return res
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
	reportSkipped := func(skipped []string) {
		for _, msg := range skipped {
			logStep(models.SyncLogWarn, msg)
			hadErrors = true
		}
	}

	fetched := s.fetchAll(ctx)

	categoryIDs := map[uuid.UUID]uuid.UUID{}
	if fetched.categoriesErr != nil {
		msg := "categories: " + fetched.categoriesErr.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
	} else {
		var skipped []string
		categoryIDs, skipped = s.processCategories(ctx, fetched.categories)
		reportSkipped(skipped)
	}

	warehouseIDs := map[uuid.UUID]uuid.UUID{}
	if fetched.warehousesErr != nil {
		msg := "warehouses: " + fetched.warehousesErr.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
	} else {
		var skipped []string
		warehouseIDs, skipped = s.processWarehouses(ctx, fetched.warehouses)
		reportSkipped(skipped)
	}

	productIDs := map[uuid.UUID]uuid.UUID{}
	if fetched.productsErr != nil {
		msg := "products: " + fetched.productsErr.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
	} else {
		var skipped []string
		productIDs, skipped = s.processProducts(ctx, fetched.products, categoryIDs)
		reportSkipped(skipped)
	}

	if fetched.pricesErr != nil {
		msg := "prices: " + fetched.pricesErr.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
	} else {
		reportSkipped(s.processPrices(ctx, fetched.prices, productIDs))
	}

	if fetched.stockErr != nil {
		msg := "stock: " + fetched.stockErr.Error()
		logStep(models.SyncLogError, msg)
		recordStepFailure(msg)
	} else {
		reportSkipped(s.processStock(ctx, fetched.stock, productIDs, warehouseIDs))
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

func (s *Service) processCategories(ctx context.Context, dtos []onec.CategoryDTO) (map[uuid.UUID]uuid.UUID, []string) {
	items := make([]models.CategoryInput, 0, len(dtos))
	parentByGUID := make(map[uuid.UUID]*uuid.UUID, len(dtos))
	var skipped []string
	for _, dto := range dtos {
		in, parentGUID, mapErr := mapCategory(dto)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("category %s: %s", dto.RefKey, mapErr))
			continue
		}
		items = append(items, in)
		parentByGUID[in.OneCGUID] = parentGUID
	}

	resultIDs, errs := s.storage.UpsertCategoriesBatch(ctx, items)
	ids := make(map[uuid.UUID]uuid.UUID, len(items))
	for i, in := range items {
		if errs[i] != nil {
			skipped = append(skipped, fmt.Sprintf("category %s: %s", in.OneCGUID, errs[i]))
			continue
		}
		ids[in.OneCGUID] = resultIDs[i]
	}

	for oneCGUID, parentGUID := range parentByGUID {
		if parentGUID == nil {
			continue
		}
		id, ok := ids[oneCGUID]
		if !ok {
			continue
		}
		parentID, ok := ids[*parentGUID]
		if !ok {
			continue
		}
		if setErr := s.storage.SetCategoryParent(ctx, id, parentID); setErr != nil {
			skipped = append(skipped, fmt.Sprintf("category %s: set parent: %s", oneCGUID, setErr))
		}
	}
	return ids, capSkipped(skipped)
}

func (s *Service) processWarehouses(ctx context.Context, dtos []onec.WarehouseDTO) (map[uuid.UUID]uuid.UUID, []string) {
	items := make([]models.WarehouseInput, 0, len(dtos))
	var skipped []string
	for _, dto := range dtos {
		in, mapErr := mapWarehouse(dto)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("warehouse %s: %s", dto.RefKey, mapErr))
			continue
		}
		items = append(items, in)
	}

	resultIDs, errs := s.storage.UpsertWarehousesBatch(ctx, items)
	ids := make(map[uuid.UUID]uuid.UUID, len(items))
	for i, in := range items {
		if errs[i] != nil {
			skipped = append(skipped, fmt.Sprintf("warehouse %s: %s", in.OneCGUID, errs[i]))
			continue
		}
		ids[in.OneCGUID] = resultIDs[i]
	}
	return ids, capSkipped(skipped)
}

func (s *Service) processProducts(ctx context.Context, dtos []onec.ProductDTO, categoryIDs map[uuid.UUID]uuid.UUID) (map[uuid.UUID]uuid.UUID, []string) {
	items := make([]models.ProductInput, 0, len(dtos))
	var skipped []string
	for _, dto := range dtos {
		in, mapErr := mapProduct(dto, categoryIDs)
		if mapErr != nil {
			skipped = append(skipped, fmt.Sprintf("product %s: %s", dto.RefKey, mapErr))
			continue
		}
		items = append(items, in)
	}

	resultIDs, errs := s.storage.UpsertProductsBatch(ctx, items)
	ids := make(map[uuid.UUID]uuid.UUID, len(items))
	for i, in := range items {
		if errs[i] != nil {
			skipped = append(skipped, fmt.Sprintf("product %s (sku=%s): %s", in.OneCGUID, in.SKU, errs[i]))
			continue
		}
		ids[in.OneCGUID] = resultIDs[i]
	}
	return ids, capSkipped(skipped)
}

func (s *Service) processPrices(ctx context.Context, dtos []onec.PriceDTO, productIDs map[uuid.UUID]uuid.UUID) []string {
	items := make([]models.PriceInput, 0, len(dtos))
	sourceRefs := make([]string, 0, len(dtos))
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
		items = append(items, in)
		sourceRefs = append(sourceRefs, dto.ProductKey)
	}

	errs := s.storage.UpsertProductPricesBatch(ctx, items)
	for i := range items {
		if errs[i] != nil {
			skipped = append(skipped, fmt.Sprintf("price for product %s: %s", sourceRefs[i], errs[i]))
		}
	}
	if droppedUnknownProduct > 0 {
		skipped = append(skipped, fmt.Sprintf("prices: %d rows skipped, referenced product not found in this run", droppedUnknownProduct))
	}
	return capSkipped(skipped)
}

func (s *Service) processStock(ctx context.Context, dtos []onec.StockDTO, productIDs, warehouseIDs map[uuid.UUID]uuid.UUID) []string {
	items := make([]models.StockInput, 0, len(dtos))
	sourceProductRefs := make([]string, 0, len(dtos))
	sourceWarehouseRefs := make([]string, 0, len(dtos))
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
		items = append(items, in)
		sourceProductRefs = append(sourceProductRefs, dto.ProductKey)
		sourceWarehouseRefs = append(sourceWarehouseRefs, dto.WarehouseKey)
	}

	errs := s.storage.UpsertStockBalancesBatch(ctx, items)
	for i := range items {
		if errs[i] != nil {
			skipped = append(skipped, fmt.Sprintf("stock for product %s warehouse %s: %s", sourceProductRefs[i], sourceWarehouseRefs[i], errs[i]))
		}
	}
	if droppedUnknownProduct > 0 {
		skipped = append(skipped, fmt.Sprintf("stock: %d rows skipped, referenced product not found in this run", droppedUnknownProduct))
	}
	if droppedUnknownWarehouse > 0 {
		skipped = append(skipped, fmt.Sprintf("stock: %d rows skipped, referenced warehouse not found in this run", droppedUnknownWarehouse))
	}
	return capSkipped(skipped)
}
