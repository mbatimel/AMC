# Воркер синхронизации 1С (УТ 10.3) → products — дизайн

## Контекст

Нужен воркер, раз в сутки забирающий из 1С (база `UT`, УТ 10.3) номенклатуру, цены
и остатки, и обновляющий ими БД `products`. Разведка канала доступа к 1С уже
проведена и задокументирована в `back/integrations/README.md`: рекомендованный
путь — OData standard interface (`GET /{ib}/odata/standard.odata/...`, Basic Auth),
но на сервере `PVISERVER` он **ещё не опубликован** (нет IIS, галка OData не
включена). Схема БД `products` (миграции `20260705171939_catalog.sql`,
`20260705171943_warehouse.sql`, `20260730120000_products_service.sql`) уже
подготовлена под 1С-синк: `categories`, `products`, `warehouses`, `price_groups`
имеют колонку `one_c_guid UUID UNIQUE` — естественный ключ для upsert.
Также готовы таблицы `integration_systems`/`sync_jobs`/`sync_logs`
(`20260705171945_integration_sync.sql`) — под учёт запусков интеграций, ими же
и воспользуемся.

## Решения (согласованы с пользователем)

- Воркер пишет **напрямую в Postgres** (та же БД, что и `products`-сервис),
  без похода через HTTP API `products`. Upsert по `one_c_guid`.
- Запускается как **отдельный always-on контейнер** с тикером внутри процесса
  (не host-cron, не one-shot + внешний планировщик).
- Протокол к 1С — **OData**, клиент строится под этот протокол уже сейчас
  (URL/креды из конфига), даже если на сервере он пока не включён — это
  согласованный с пользователем следующий шаг после публикации.
- Прямое подключение к Postgres-базе самой 1С — **не делаем** (см. README:
  автоимена таблиц, бинарные UUID, риск для прода).

## Объём (v1)

Синкаем только то, что подтверждено в README:

| Сущность 1С | Таблица `products`-БД | Ключ сопоставления |
|---|---|---|
| `Catalog_НоменклатурныеГруппы` | `categories` | `one_c_guid` |
| `Catalog_Склады` | `warehouses` | `one_c_guid` |
| `Catalog_Номенклатура` | `products` (sku, name, category_id, is_active) | `one_c_guid` |
| `InformationRegister_ЦеныНоменклатуры` | `product_prices` (price_group_id=NULL, price_type=код типа цены) | `product_id + price_type` |
| `AccumulationRegister_ТоварыНаСкладах` | `stock_balances` (per `warehouse_id`) | `product_id + warehouse_id` |

**Вне объёма v1** (не найдено в манифесте 1С, требует уточнения у 1С-разработчика —
см. README): `Brand`, `GOST`, `Material`, `Size`, `Images`, единицы измерения
(`units`), `price_groups`/`volume_discounts` (клиентские группы цен — отдельная
бизнес-логика, не связана с прайс-листом по умолчанию). Эти поля остаются
`NULL`/нетронутыми, помечены `TODO` в маппинге — расширить, когда 1С-разработчик
подтвердит расположение.

## Компоненты

Новый Go-модуль `back/integrations` (по образцу `back/products`):

```
back/integrations/
  go.mod
  cmd/onec-sync/main.go
  internal/config/config.go
  internal/onec/client.go        — ОТДЕЛЬНЫЙ файл: HTTP OData-клиент к 1С
  internal/onec/models.go        — DTO под OData JSON-ответы
  internal/onec/client_test.go
  internal/service/sync.go       — оркестрация синка
  internal/service/mapping.go    — маппинг DTO 1С → внутренние модели
  internal/models/models.go      — внутренние доменные структуры
  internal/storage/postgres/postgres.go
  internal/storage/postgres/sql/*.sql
  internal/storage/postgres/integration_test.go
  README.md                      — обновить статус (было: "не реализован")
```

### `internal/onec/client.go` — клиент 1С

Ничего не знает про Postgres — чистый I/O-слой, мокается `httptest` в тестах.

```go
type Client struct {
    baseURL  string // например http://PVISERVER/UT/odata/standard.odata
    user     string
    password string
    http     *http.Client // timeout 30s
}

func New(baseURL, user, password string) *Client

func (c *Client) FetchCategories(ctx context.Context) ([]CategoryDTO, error)
func (c *Client) FetchWarehouses(ctx context.Context) ([]WarehouseDTO, error)
func (c *Client) FetchProducts(ctx context.Context) ([]ProductDTO, error)
func (c *Client) FetchPrices(ctx context.Context) ([]PriceDTO, error)
func (c *Client) FetchStock(ctx context.Context) ([]StockDTO, error)
```

Каждый метод — `GET {baseURL}/{EntitySet}?$format=json`, Basic Auth, декод JSON
(`{"value": [...]}`  — стандартная обёртка OData), возврат слайса DTO. Имена
entity set'ов и полей (`Catalog_НоменклатурныеГруппы`, `Артикул`, `Ссылка` и т.д.)
— именованные константы вверху файла с комментарием, что подлежат сверке с
реальным сервером после публикации OData (сейчас взяты по стандартному
именованию 1С: `Catalog_<Имя>`, `InformationRegister_<Имя>`,
`AccumulationRegister_<Имя>`).

Ошибка сети/не-200/битый JSON — оборачивается и возвращается как есть, ретраев
внутри клиента нет (это уровень `service`).

### `internal/service/sync.go` — оркестрация

```go
type Service struct {
    onec    OnecClient // интерфейс из internal/onec, для мока в тестах
    storage Storage
    logger  zerolog.Logger
}

func (s *Service) RunSync(ctx context.Context) error
```

Порядок шагов одного рана (каждый — отдельная запись в `sync_logs`, ошибка шага
не прерывает весь ран — логируется, шаг пропускается, ран в конце помечается
`partial`/`failed` по наличию ошибок):

1. `integration_systems` upsert по `code='onec_ut'` (создаётся при первом запуске,
   если ещё нет) → получить `system_id`.
2. Создать `sync_jobs` (`system_id`, `direction='inbound'`,
   `entity_type='onec_full_sync'`, `status='running'`).
3. Категории: `onec.FetchCategories` → upsert без `parent_id` (по `one_c_guid`) →
   второй проход: проставить `parent_id` по карте `one_c_guid → id` (двухпроходно,
   т.к. родитель может идти в списке позже потомка).
4. Склады: `onec.FetchWarehouses` → upsert по `one_c_guid`.
5. Товары: `onec.FetchProducts` → upsert по `one_c_guid` (`sku`, `name`,
   `category_id` через карту категорий, `is_active`).
6. Цены: `onec.FetchPrices` → per-товар upsert в `product_prices`
   (`price_group_id IS NULL`, ключ `product_id + price_type`, тот же паттерн
   SELECT-then-UPDATE/INSERT, что уже использует `products`-сервис в
   `upsertProductPrice.sql`).
7. Остатки: `onec.FetchStock` → per-(товар, склад) upsert в `stock_balances`
   (ключ `product_id + warehouse_id`).
8. `sync_jobs.status = 'success'` (или `'failed'`, если хоть один шаг упал
   полностью — например сам 1С недоступен на шаге 3) / `processed_at = now()`.

Товар/цена/остаток, для которых `one_c_guid`-родитель не нашёлся (например
`product_id` для строки цены, которой нет в `products`) — пропускается с записью
в `sync_logs` (`level='warn'`), не валит весь ран.

### `internal/storage/postgres/postgres.go`

Тот же паттерн, что в `products`: `embed.FS` для `sql/*.sql`, `pgxpool.Pool`.
Методы: `UpsertCategory`, `SetCategoryParent`, `UpsertWarehouse`, `UpsertProduct`,
`UpsertProductPrice`, `UpsertStockBalance`, плюс `UpsertIntegrationSystem`,
`CreateSyncJob`, `FinishSyncJob`, `AddSyncLog`. Категории/склады/товары —
`INSERT ... ON CONFLICT (one_c_guid) DO UPDATE` (колонка `UNIQUE`, конфликт
работает напрямую). Цены/остатки — SELECT существующей строки +
UPDATE-или-INSERT (как `upsertProductPrice.sql` в products), т.к. на этих
таблицах нет unique-индекса под нужный составной ключ — новую миграцию под это
не заводим, переиспользуем уже работающий в проекте паттерн.

### `cmd/onec-sync/main.go`

```go
func main() {
    cfg := config.LoadConfig()
    pool, _ := postgres.NewPool(cfg)
    storage := postgres.New(pool)
    onecClient := onec.New(cfg.OnecBaseURL, cfg.OnecUser, cfg.OnecPassword)
    svc := service.New(log.Logger, onecClient, storage)

    ticker := time.NewTicker(cfg.SyncInterval) // по умолчанию 24h
    // + первый запуск сразу при старте контейнера
    for {
        if err := svc.RunSync(ctx); err != nil {
            log.Error().Err(err).Msg("onec sync run failed")
        }
        select {
        case <-ticker.C:
        case <-shutdown:
            return
        }
    }
}
```

Health-сервер — минимальный (`/healthz` живости процесса), без бизнес-эндпоинтов,
как у остальных сервисов (`transport/http/health.go`).

## Конфиг / деплой

`internal/config/config.go` — новые обязательные переменные (fatal, если пустые,
как у остальных сервисов):

| Переменная | Назначение |
|---|---|
| `PG_HOST`/`PG_PORT`/`PG_DB`/`PG_USER`/`PG_PASSWORD` | та же БД, что у products |
| `ONEC_BASE_URL` | напр. `http://PVISERVER/UT/odata/standard.odata` |
| `ONEC_USER` / `ONEC_PASSWORD` | технический пользователь 1С (read-only) |
| `SYNC_INTERVAL` | по умолчанию `24h` |
| `HEALTH_ADDR` | по умолчанию `:9096` |

`deploy/docker-compose.yml` — новый сервис `onec-sync`, без публикуемого порта,
`depends_on: migrations (service_completed_successfully)`, сеть `amc_net`.

`deploy/.env.example` — добавить `ONEC_BASE_URL`, `ONEC_USER`, `ONEC_PASSWORD`.

`back/integrations/README.md` — обновить раздел "Статус": воркер реализован,
ждёт публикации OData на стороне 1С для первого реального прогона.

## Ошибки / edge cases

- 1С недоступна целиком (сеть/таймаут на первом же fetch) — ран помечается
  `failed`, БД не трогается вообще на этом шаге, следующая попытка — через
  `SYNC_INTERVAL`. Ретраев внутри одного рана нет (YAGNI — раз в сутки нет
  смысла городить backoff).
- Частичная ошибка (например `FetchPrices` упал, остальное успешно) — шаги до
  и после не блокируются, `sync_jobs.status` в конце = `partial`.
- Дубликат/пустой `sku` у товара из 1С (упадёт `UNIQUE` на `products.sku`) —
  строка пропускается, `sync_logs` с `level='error'` и телом ошибки, ран
  продолжается для остальных товаров.
- `category_id`/`warehouse_id`, отсутствующий в уже засинканной карте — то же:
  skip + warn-лог, не падение всего рана.

## Тестирование

- `internal/onec/client_test.go` — `httptest.Server`, мок OData JSON-ответов
  (`{"value": [...]}`), проверка парсинга DTO, обработки не-200 и битого JSON.
- `internal/service/sync_test.go` — мок `OnecClient` (интерфейс) + мок `Storage`
  (интерфейс), проверка порядка шагов, skip-логики при отсутствующих ссылках,
  `partial`/`failed`/`success` статусов.
- `internal/storage/postgres/integration_test.go` — по паттерну `products`:
  реальный Postgres через `INTEGRATIONS_TEST_DATABASE_URL`, `t.Skip`, если не
  задан. Проверка upsert-идемпотentности (повторный прогон с тем же
  `one_c_guid` не плодит дубликаты, корректно обновляет поля).

## Что не входит в этот спек

- Реальная сверка имён OData entity set'ов/полей с сервером 1С — невозможна,
  пока там не включена публикация (см. README, черновик запроса 1С-разработчику
  уже подготовлен). Константы в `internal/onec/client.go` — лучшее известное
  приближение по стандартному именованию 1С, потребуют правки после первого
  реального похода.
- Brand/GOST/Material/Size/Images/units/price_groups — см. "Объём (v1)" выше.
- Инкрементальная синхронизация (тянуть только изменённое) — v1 всегда полный
  проход по всем сущностям; README отмечает это как вопрос к 1С-разработчику
  (есть ли поле даты изменения) — можно добавить отдельным спеком позже.
