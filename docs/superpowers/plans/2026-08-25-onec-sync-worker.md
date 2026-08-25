# Воркер синхронизации 1С (УТ 10.3) → products — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Новый Go-сервис `back/integrations` — воркер, раз в сутки забирающий из 1С (УТ 10.3, OData) номенклатуру/категории/склады/цены/остатки и upsert'ящий их в БД `products` по колонке `one_c_guid`.

**Architecture:** Отдельный always-on контейнер с тикером внутри процесса (без host-cron). Три слоя: `internal/onec` (чистый HTTP OData-клиент к 1С, ничего не знает про Postgres), `internal/service` (оркестрация — тянет DTO из 1С, маппит, upsert'ит через интерфейс `Storage`), `internal/storage/postgres` (прямая запись в ту же БД, что и `products`-сервис). Наблюдаемость — через уже существующие таблицы `integration_systems`/`sync_jobs`/`sync_logs`.

**Tech Stack:** Go 1.25, `pgx/v4` (та же БД, тот же паттерн, что у `products`), `valyala/fasthttp` для исходящего HTTP-клиента к 1С (как у `back/auth/internal/client/fns`), `gofiber/fiber/v2` для health-эндпоинтов, `rs/zerolog`.

**Spec:** `docs/superpowers/specs/2026-08-25-onec-sync-worker-design.md`

## Global Constraints

- Go `1.25.0` (как во всех остальных `back/*` модулях).
- Никаких новых миграций БД — используем уже существующие колонки `one_c_guid` (`categories`, `warehouses`, `products`) и таблицы `integration_systems`/`sync_jobs`/`sync_logs`. Таблицы `product_prices`/`stock_balances` без уникального составного индекса — upsert через паттерн SELECT-затем-UPDATE/INSERT (как уже делает `products`-сервис в `upsertProductPrice.sql`), новую миграцию под это не заводим.
- Каждый нестабильный внешний вызов (1С) — через `fasthttp.Client`, тестируется через `net/http/httptest.Server` (проверенный паттерн — см. `back/auth/internal/client/fns`).
- Опциональные внешние ключи — `*uuid.UUID` (nil = NULL), а не `uuid.NullUUID` (так делает `products`-сервис в `internal/models/models.go`, `CategoryID *uuid.UUID` и т.д.).
- v1 синкает только: категории (`Catalog_НоменклатурныеГруппы`), склады (`Catalog_Склады`), товары — sku/name/category (`Catalog_Номенклатура`), цены (`InformationRegister_ЦеныНоменклатуры`, `price_group_id=NULL`, `price_type`=raw GUID типа цены из 1С), остатки по складам (`AccumulationRegister_ТоварыНаСкладах`). Brand/GOST/Material/Size/Images/units/price_groups — вне объёма, не трогаем.
- Имена OData entity set'ов/полей в `internal/onec` — лучшее приближение по стандартному именованию 1С (`Ref_Key`/`Parent_Key`/`Description`, суффикс `Balance` у регистров накопления). Сервер 1С ещё не публикует OData (см. `back/integrations/README.md`) — это ожидаемо, свериться после публикации отдельной задачей, вне этого плана.
- Частичный сбой одного шага синка не должен ронять остальные шаги — `sync_jobs.status` = `success`/`partial`/`failed`, ошибки шагов пишутся в `sync_logs`.

---

## Task 1: Модуль, конфиг

**Files:**
- Create: `back/integrations/go.mod` (через `go mod init`, наполняется `go get`/`go mod tidy` по ходу задач)
- Create: `back/integrations/internal/config/config.go`
- Test: `back/integrations/internal/config/config_test.go`

**Interfaces:**
- Produces: `config.Config{PGHost, PGPort, PGDB, PGUser, PGPassword, HealthAddr, OnecBaseURL, OnecUser, OnecPassword string; SyncInterval time.Duration}`, `config.LoadConfig() Config`, `config.LoadEnvFile(path string) error`, `config.GetEnv(key, fallback string) string`.

- [ ] **Step 1: Инициализировать модуль**

```bash
cd back/integrations
go mod init github.com/mbatimel/AMC/integrations
```

- [ ] **Step 2: Написать падающий тест конфига**

Создать `back/integrations/internal/config/config_test.go`:

```go
package config

import (
	"os"
	"testing"
	"time"
)

func setEnv(t *testing.T, key, value string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}

func requiredEnv(t *testing.T) {
	t.Helper()
	setEnv(t, "PG_DB", "amc")
	setEnv(t, "PG_USER", "amc")
	setEnv(t, "PG_PASSWORD", "secret")
	setEnv(t, "ONEC_BASE_URL", "http://pviserver/UT/odata/standard.odata")
	setEnv(t, "ONEC_USER", "site")
	setEnv(t, "ONEC_PASSWORD", "site-pass")
}

func TestLoadConfig_Defaults(t *testing.T) {
	requiredEnv(t)
	os.Unsetenv("SYNC_INTERVAL")
	os.Unsetenv("HEALTH_ADDR")
	os.Unsetenv("PG_HOST")
	os.Unsetenv("PG_PORT")

	cfg := LoadConfig()

	if cfg.PGHost != "localhost" {
		t.Errorf("expected default PGHost=localhost, got %s", cfg.PGHost)
	}
	if cfg.PGPort != "5432" {
		t.Errorf("expected default PGPort=5432, got %s", cfg.PGPort)
	}
	if cfg.HealthAddr != ":9096" {
		t.Errorf("expected default HealthAddr=:9096, got %s", cfg.HealthAddr)
	}
	if cfg.SyncInterval != 24*time.Hour {
		t.Errorf("expected default SyncInterval=24h, got %s", cfg.SyncInterval)
	}
	if cfg.OnecBaseURL != "http://pviserver/UT/odata/standard.odata" {
		t.Errorf("unexpected OnecBaseURL: %s", cfg.OnecBaseURL)
	}
}

func TestLoadConfig_CustomSyncInterval(t *testing.T) {
	requiredEnv(t)
	setEnv(t, "SYNC_INTERVAL", "1h")

	cfg := LoadConfig()

	if cfg.SyncInterval != time.Hour {
		t.Errorf("expected SyncInterval=1h, got %s", cfg.SyncInterval)
	}
}
```

- [ ] **Step 3: Убедиться, что тест падает (пакет ещё не компилируется)**

```bash
go test ./internal/config/... -v
```
Expected: FAIL — `undefined: LoadConfig` (сборка не проходит, `config.go` ещё не создан).

- [ ] **Step 4: Реализовать конфиг**

Создать `back/integrations/internal/config/config.go`:

```go
package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type Config struct {
	PGHost       string
	PGPort       string
	PGDB         string
	PGUser       string
	PGPassword   string
	HealthAddr   string
	OnecBaseURL  string
	OnecUser     string
	OnecPassword string
	SyncInterval time.Duration
}

func LoadConfig() Config {
	cfg := Config{
		PGHost:       GetEnv("PG_HOST", "localhost"),
		PGPort:       GetEnv("PG_PORT", "5432"),
		PGDB:         os.Getenv("PG_DB"),
		PGUser:       os.Getenv("PG_USER"),
		PGPassword:   os.Getenv("PG_PASSWORD"),
		HealthAddr:   GetEnv("HEALTH_ADDR", ":9096"),
		OnecBaseURL:  os.Getenv("ONEC_BASE_URL"),
		OnecUser:     os.Getenv("ONEC_USER"),
		OnecPassword: os.Getenv("ONEC_PASSWORD"),
		SyncInterval: getEnvDuration("SYNC_INTERVAL", 24*time.Hour),
	}
	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	if cfg.OnecBaseURL == "" || cfg.OnecUser == "" || cfg.OnecPassword == "" {
		log.Fatal().Msg("ONEC_BASE_URL, ONEC_USER and ONEC_PASSWORD must be specified")
	}
	return cfg
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		log.Fatal().Err(err).Str("key", key).Msg("invalid duration environment variable")
	}
	return parsed
}

func GetEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		value = os.Expand(value, os.Getenv)
		if err = os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}
```

- [ ] **Step 5: Подтянуть зависимость и запустить тесты**

```bash
go get github.com/rs/zerolog@v1.35.1
go mod tidy
go test ./internal/config/... -v
```
Expected: PASS (`TestLoadConfig_Defaults`, `TestLoadConfig_CustomSyncInterval`).

- [ ] **Step 6: Коммит**

```bash
git add back/integrations/go.mod back/integrations/go.sum back/integrations/internal/config
git commit -m "feat(integrations): add onec-sync module scaffold and config"
```

---

## Task 2: Клиент 1С (OData)

**Files:**
- Create: `back/integrations/internal/onec/models.go`
- Create: `back/integrations/internal/onec/client.go`
- Test: `back/integrations/internal/onec/client_test.go`

**Interfaces:**
- Consumes: ничего из предыдущих задач (пакет самодостаточен, не знает про `config`/`postgres`).
- Produces: `onec.CategoryDTO{RefKey, ParentKey, Description string}`, `onec.WarehouseDTO{RefKey, Description string}`, `onec.ProductDTO{RefKey, CategoryKey, Code, Description string}`, `onec.PriceDTO{ProductKey, PriceTypeKey string; Price float64}`, `onec.StockDTO{ProductKey, WarehouseKey string; Quantity float64}`, `onec.New(baseURL, user, password string, logger zerolog.Logger) *onec.Client`, методы `(*Client) FetchCategories/FetchWarehouses/FetchProducts/FetchPrices/FetchStock(ctx context.Context) ([]DTO, error)`.

- [ ] **Step 1: Написать DTO и константы entity set'ов**

Создать `back/integrations/internal/onec/models.go`:

```go
// Package onec — HTTP OData-клиент к 1С:УТ 10.3.
//
// Имена entity set'ов и полей ниже — лучшее приближение по стандартному
// именованию 1С OData (Ref_Key/Parent_Key/Description для справочников,
// суффикс Balance для регистров накопления). На момент написания сервер
// PVISERVER ещё не публикует OData (см. back/integrations/README.md) —
// сверить с реальными метаданными после публикации.
package onec

type odataEnvelope[T any] struct {
	Value []T `json:"value"`
}

type CategoryDTO struct {
	RefKey      string `json:"Ref_Key"`
	ParentKey   string `json:"Parent_Key"`
	Description string `json:"Description"`
}

type WarehouseDTO struct {
	RefKey      string `json:"Ref_Key"`
	Description string `json:"Description"`
}

type ProductDTO struct {
	RefKey      string `json:"Ref_Key"`
	CategoryKey string `json:"НоменклатурнаяГруппа_Key"`
	Code        string `json:"Code"`
	Description string `json:"Description"`
}

type PriceDTO struct {
	ProductKey   string  `json:"Номенклатура_Key"`
	PriceTypeKey string  `json:"ТипЦен_Key"`
	Price        float64 `json:"Цена"`
}

type StockDTO struct {
	ProductKey   string  `json:"Номенклатура_Key"`
	WarehouseKey string  `json:"Склад_Key"`
	Quantity     float64 `json:"КоличествоBalance"`
}
```

- [ ] **Step 2: Написать падающие тесты клиента**

Создать `back/integrations/internal/onec/client_test.go`:

```go
package onec

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func testLogger() zerolog.Logger { return zerolog.Nop() }

func TestFetchCategories_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Catalog_НоменклатурныеГруппы" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Fatal("expected Authorization header")
		}
		w.Write([]byte(`{"value":[{"Ref_Key":"11111111-1111-1111-1111-111111111111","Parent_Key":"00000000-0000-0000-0000-000000000000","Description":"Инструмент"}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Description != "Инструмент" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchCategories_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	_, err := c.FetchCategories(context.Background())
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}

func TestFetchCategories_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	_, err := c.FetchCategories(context.Background())
	if err == nil {
		t.Fatal("expected error on malformed json")
	}
}

func TestFetchWarehouses_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Ref_Key":"44444444-4444-4444-4444-444444444444","Description":"Склад №1"}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchWarehouses(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Description != "Склад №1" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchProducts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Ref_Key":"22222222-2222-2222-2222-222222222222","НоменклатурнаяГруппа_Key":"11111111-1111-1111-1111-111111111111","Code":"SKU-1","Description":"Дрель"}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchProducts(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Code != "SKU-1" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchPrices_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Номенклатура_Key":"22222222-2222-2222-2222-222222222222","ТипЦен_Key":"33333333-3333-3333-3333-333333333333","Цена":150.5}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchPrices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Price != 150.5 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestFetchStock_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"value":[{"Номенклатура_Key":"22222222-2222-2222-2222-222222222222","Склад_Key":"44444444-4444-4444-4444-444444444444","КоличествоBalance":7}]}`))
	}))
	defer server.Close()

	c := New(server.URL, "user", "pass", testLogger())
	got, err := c.FetchStock(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Quantity != 7 {
		t.Fatalf("unexpected result: %+v", got)
	}
}
```

- [ ] **Step 3: Убедиться, что тесты падают**

```bash
go test ./internal/onec/... -v
```
Expected: FAIL — `undefined: New` (компиляция падает, `client.go` ещё не создан).

- [ ] **Step 4: Реализовать клиент**

Создать `back/integrations/internal/onec/client.go`:

```go
package onec

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const requestTimeout = 30 * time.Second

const (
	entitySetCategories = "Catalog_НоменклатурныеГруппы"
	entitySetWarehouses = "Catalog_Склады"
	entitySetProducts   = "Catalog_Номенклатура"
	entitySetPrices     = "InformationRegister_ЦеныНоменклатуры"
	entitySetStock      = "AccumulationRegister_ТоварыНаСкладахBalance"
)

type Client struct {
	baseURL  string
	user     string
	password string
	http     *fasthttp.Client
	logger   zerolog.Logger
}

func New(baseURL, user, password string, logger zerolog.Logger) *Client {
	return &Client{
		baseURL:  baseURL,
		user:     user,
		password: password,
		http:     &fasthttp.Client{},
		logger:   logger,
	}
}

func fetchEntitySet[T any](ctx context.Context, c *Client, entitySet string) ([]T, error) {
	_ = ctx // резерв на будущее (deadline/cancel), fasthttp.DoTimeout ctx не принимает

	reqURL := fmt.Sprintf("%s/%s?$format=json", c.baseURL, entitySet)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqURL)
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set("Accept", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(c.user + ":" + c.password))
	req.Header.Set("Authorization", "Basic "+auth)

	doErr := c.http.DoTimeout(req, resp, requestTimeout)
	statusCode := resp.StatusCode()
	body := append([]byte(nil), resp.Body()...)

	logEvent := c.logger.Info().Str("entitySet", entitySet).Int("status", statusCode)
	if doErr != nil {
		logEvent.Err(doErr).Msg("onec odata request failed")
		return nil, fmt.Errorf("onec odata request %s: %w", entitySet, doErr)
	}
	logEvent.Str("response", string(body)).Msg("onec odata response")

	if statusCode != fasthttp.StatusOK {
		return nil, fmt.Errorf("onec odata %s: unexpected status %d", entitySet, statusCode)
	}

	var envelope odataEnvelope[T]
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("onec odata %s: decode response: %w", entitySet, err)
	}
	return envelope.Value, nil
}

func (c *Client) FetchCategories(ctx context.Context) ([]CategoryDTO, error) {
	return fetchEntitySet[CategoryDTO](ctx, c, entitySetCategories)
}

func (c *Client) FetchWarehouses(ctx context.Context) ([]WarehouseDTO, error) {
	return fetchEntitySet[WarehouseDTO](ctx, c, entitySetWarehouses)
}

func (c *Client) FetchProducts(ctx context.Context) ([]ProductDTO, error) {
	return fetchEntitySet[ProductDTO](ctx, c, entitySetProducts)
}

func (c *Client) FetchPrices(ctx context.Context) ([]PriceDTO, error) {
	return fetchEntitySet[PriceDTO](ctx, c, entitySetPrices)
}

func (c *Client) FetchStock(ctx context.Context) ([]StockDTO, error) {
	return fetchEntitySet[StockDTO](ctx, c, entitySetStock)
}
```

- [ ] **Step 5: Подтянуть зависимости и прогнать тесты**

```bash
go get github.com/valyala/fasthttp@v1.72.0
go mod tidy
go test ./internal/onec/... -v
```
Expected: PASS — все 7 тестов.

- [ ] **Step 6: Коммит**

```bash
git add back/integrations/go.mod back/integrations/go.sum back/integrations/internal/onec
git commit -m "feat(integrations): add 1C OData client"
```

---

## Task 3: Доменные модели и маппинг DTO → модели

**Files:**
- Create: `back/integrations/internal/models/models.go`
- Create: `back/integrations/internal/service/mapping.go`
- Test: `back/integrations/internal/service/mapping_test.go`

**Interfaces:**
- Consumes: `onec.CategoryDTO/WarehouseDTO/ProductDTO/PriceDTO/StockDTO` (Task 2).
- Produces: `models.SyncLogLevel` (+ `SyncLogInfo/SyncLogWarn/SyncLogError`), `models.CategoryInput{OneCGUID uuid.UUID; Name string}`, `models.WarehouseInput{OneCGUID uuid.UUID; Name string}`, `models.ProductInput{OneCGUID uuid.UUID; CategoryID *uuid.UUID; SKU, Name string}`, `models.PriceInput{ProductID uuid.UUID; PriceType string; Price float64}`, `models.StockInput{ProductID, WarehouseID uuid.UUID; Quantity float64}`. В пакете `service`: `zeroGUID` (константа), `mapCategory(dto onec.CategoryDTO) (models.CategoryInput, *uuid.UUID, error)`, `mapWarehouse(dto onec.WarehouseDTO) (models.WarehouseInput, error)`, `mapProduct(dto onec.ProductDTO, categoryIDs map[uuid.UUID]uuid.UUID) (models.ProductInput, error)`, `mapPrice(dto onec.PriceDTO, productIDs map[uuid.UUID]uuid.UUID) (models.PriceInput, bool, error)`, `mapStock(dto onec.StockDTO, productIDs, warehouseIDs map[uuid.UUID]uuid.UUID) (models.StockInput, bool, error)`.

- [ ] **Step 1: Создать пакет моделей**

Создать `back/integrations/internal/models/models.go`:

```go
package models

import "github.com/google/uuid"

type SyncLogLevel string

const (
	SyncLogInfo  SyncLogLevel = "info"
	SyncLogWarn  SyncLogLevel = "warn"
	SyncLogError SyncLogLevel = "error"
)

type CategoryInput struct {
	OneCGUID uuid.UUID
	Name     string
}

type WarehouseInput struct {
	OneCGUID uuid.UUID
	Name     string
}

type ProductInput struct {
	OneCGUID   uuid.UUID
	CategoryID *uuid.UUID
	SKU        string
	Name       string
}

type PriceInput struct {
	ProductID uuid.UUID
	PriceType string
	Price     float64
}

type StockInput struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Quantity    float64
}
```

- [ ] **Step 2: Написать падающие тесты маппинга**

Создать `back/integrations/internal/service/mapping_test.go`:

```go
package service

import (
	"testing"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/integrations/internal/onec"
)

func TestMapCategory_WithParent(t *testing.T) {
	dto := onec.CategoryDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		ParentKey:   "22222222-2222-2222-2222-222222222222",
		Description: "Дрели",
	}
	in, parent, err := mapCategory(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.OneCGUID.String() != dto.RefKey || in.Name != "Дрели" {
		t.Fatalf("unexpected input: %+v", in)
	}
	if parent == nil || parent.String() != dto.ParentKey {
		t.Fatalf("expected parent %s, got %v", dto.ParentKey, parent)
	}
}

func TestMapCategory_ZeroParentIsRoot(t *testing.T) {
	dto := onec.CategoryDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		ParentKey:   zeroGUID,
		Description: "Инструмент",
	}
	_, parent, err := mapCategory(dto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parent != nil {
		t.Fatalf("expected nil parent for zero guid, got %v", parent)
	}
}

func TestMapCategory_InvalidRef(t *testing.T) {
	dto := onec.CategoryDTO{RefKey: "not-a-guid", Description: "x"}
	_, _, err := mapCategory(dto)
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestMapProduct_ResolvesCategory(t *testing.T) {
	categoryRef := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	categoryID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	categoryIDs := map[uuid.UUID]uuid.UUID{categoryRef: categoryID}

	dto := onec.ProductDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		CategoryKey: categoryRef.String(),
		Code:        "SKU-1",
		Description: "Дрель",
	}
	in, err := mapProduct(dto, categoryIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.CategoryID == nil || *in.CategoryID != categoryID {
		t.Fatalf("expected category %s, got %v", categoryID, in.CategoryID)
	}
	if in.SKU != "SKU-1" || in.Name != "Дрель" {
		t.Fatalf("unexpected input: %+v", in)
	}
}

func TestMapProduct_UnknownCategory_LeavesNil(t *testing.T) {
	dto := onec.ProductDTO{
		RefKey:      "11111111-1111-1111-1111-111111111111",
		CategoryKey: "99999999-9999-9999-9999-999999999999",
		Code:        "SKU-1",
		Description: "Дрель",
	}
	in, err := mapProduct(dto, map[uuid.UUID]uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if in.CategoryID != nil {
		t.Fatalf("expected nil category, got %v", in.CategoryID)
	}
}

func TestMapProduct_InvalidRef(t *testing.T) {
	dto := onec.ProductDTO{RefKey: "not-a-guid"}
	_, err := mapProduct(dto, map[uuid.UUID]uuid.UUID{})
	if err == nil {
		t.Fatal("expected error for invalid ref")
	}
}

func TestMapPrice_SkipsUnknownProduct(t *testing.T) {
	dto := onec.PriceDTO{ProductKey: "11111111-1111-1111-1111-111111111111", PriceTypeKey: "x", Price: 10}
	_, ok, err := mapPrice(dto, map[uuid.UUID]uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for unknown product")
	}
}

func TestMapPrice_ResolvesProduct(t *testing.T) {
	productRef := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	productID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	dto := onec.PriceDTO{ProductKey: productRef.String(), PriceTypeKey: "type-a", Price: 99.5}
	in, ok, err := mapPrice(dto, map[uuid.UUID]uuid.UUID{productRef: productID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || in.ProductID != productID || in.PriceType != "type-a" || in.Price != 99.5 {
		t.Fatalf("unexpected result: %+v", in)
	}
}

func TestMapStock_ResolvesProductAndWarehouse(t *testing.T) {
	productRef := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	productID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	warehouseRef := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	warehouseID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

	dto := onec.StockDTO{ProductKey: productRef.String(), WarehouseKey: warehouseRef.String(), Quantity: 7}
	in, ok, err := mapStock(dto,
		map[uuid.UUID]uuid.UUID{productRef: productID},
		map[uuid.UUID]uuid.UUID{warehouseRef: warehouseID},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || in.ProductID != productID || in.WarehouseID != warehouseID || in.Quantity != 7 {
		t.Fatalf("unexpected result: %+v", in)
	}
}

func TestMapStock_SkipsUnknownWarehouse(t *testing.T) {
	productRef := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	productID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	dto := onec.StockDTO{ProductKey: productRef.String(), WarehouseKey: "77777777-7777-7777-7777-777777777777", Quantity: 3}
	_, ok, err := mapStock(dto, map[uuid.UUID]uuid.UUID{productRef: productID}, map[uuid.UUID]uuid.UUID{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for unknown warehouse")
	}
}
```

- [ ] **Step 3: Убедиться, что тесты падают**

```bash
go test ./internal/service/... -v
```
Expected: FAIL — `undefined: mapCategory` (пакет `service` ещё не содержит `mapping.go`).

- [ ] **Step 4: Реализовать маппинг**

Создать `back/integrations/internal/service/mapping.go`:

```go
package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/integrations/internal/models"
	"github.com/mbatimel/AMC/integrations/internal/onec"
)

const zeroGUID = "00000000-0000-0000-0000-000000000000"

func mapCategory(dto onec.CategoryDTO) (models.CategoryInput, *uuid.UUID, error) {
	ref, err := uuid.Parse(dto.RefKey)
	if err != nil {
		return models.CategoryInput{}, nil, fmt.Errorf("parse category ref %q: %w", dto.RefKey, err)
	}
	var parent *uuid.UUID
	if dto.ParentKey != "" && dto.ParentKey != zeroGUID {
		parentRef, parseErr := uuid.Parse(dto.ParentKey)
		if parseErr != nil {
			return models.CategoryInput{}, nil, fmt.Errorf("parse category parent %q: %w", dto.ParentKey, parseErr)
		}
		parent = &parentRef
	}
	return models.CategoryInput{OneCGUID: ref, Name: dto.Description}, parent, nil
}

func mapWarehouse(dto onec.WarehouseDTO) (models.WarehouseInput, error) {
	ref, err := uuid.Parse(dto.RefKey)
	if err != nil {
		return models.WarehouseInput{}, fmt.Errorf("parse warehouse ref %q: %w", dto.RefKey, err)
	}
	return models.WarehouseInput{OneCGUID: ref, Name: dto.Description}, nil
}

func mapProduct(dto onec.ProductDTO, categoryIDs map[uuid.UUID]uuid.UUID) (models.ProductInput, error) {
	ref, err := uuid.Parse(dto.RefKey)
	if err != nil {
		return models.ProductInput{}, fmt.Errorf("parse product ref %q: %w", dto.RefKey, err)
	}
	var categoryID *uuid.UUID
	if dto.CategoryKey != "" && dto.CategoryKey != zeroGUID {
		categoryRef, parseErr := uuid.Parse(dto.CategoryKey)
		if parseErr != nil {
			return models.ProductInput{}, fmt.Errorf("parse product category %q: %w", dto.CategoryKey, parseErr)
		}
		if id, ok := categoryIDs[categoryRef]; ok {
			categoryID = &id
		}
	}
	return models.ProductInput{OneCGUID: ref, CategoryID: categoryID, SKU: dto.Code, Name: dto.Description}, nil
}

func mapPrice(dto onec.PriceDTO, productIDs map[uuid.UUID]uuid.UUID) (models.PriceInput, bool, error) {
	productRef, err := uuid.Parse(dto.ProductKey)
	if err != nil {
		return models.PriceInput{}, false, fmt.Errorf("parse price product ref %q: %w", dto.ProductKey, err)
	}
	productID, ok := productIDs[productRef]
	if !ok {
		return models.PriceInput{}, false, nil
	}
	return models.PriceInput{ProductID: productID, PriceType: dto.PriceTypeKey, Price: dto.Price}, true, nil
}

func mapStock(dto onec.StockDTO, productIDs, warehouseIDs map[uuid.UUID]uuid.UUID) (models.StockInput, bool, error) {
	productRef, err := uuid.Parse(dto.ProductKey)
	if err != nil {
		return models.StockInput{}, false, fmt.Errorf("parse stock product ref %q: %w", dto.ProductKey, err)
	}
	warehouseRef, err := uuid.Parse(dto.WarehouseKey)
	if err != nil {
		return models.StockInput{}, false, fmt.Errorf("parse stock warehouse ref %q: %w", dto.WarehouseKey, err)
	}
	productID, okP := productIDs[productRef]
	warehouseID, okW := warehouseIDs[warehouseRef]
	if !okP || !okW {
		return models.StockInput{}, false, nil
	}
	return models.StockInput{ProductID: productID, WarehouseID: warehouseID, Quantity: dto.Quantity}, true, nil
}
```

- [ ] **Step 5: Подтянуть зависимости и прогнать тесты**

```bash
go get github.com/google/uuid@v1.6.0
go mod tidy
go test ./internal/models/... ./internal/service/... -v
```
Expected: PASS — все тесты маппинга.

- [ ] **Step 6: Коммит**

```bash
git add back/integrations/go.mod back/integrations/go.sum back/integrations/internal/models back/integrations/internal/service
git commit -m "feat(integrations): add domain models and onec DTO mapping"
```

---

## Task 4: Postgres storage (upsert по one_c_guid)

**Files:**
- Create: `back/integrations/internal/storage/postgres/connectManager.go`
- Create: `back/integrations/internal/storage/postgres/postgres.go`
- Create: `back/integrations/internal/storage/postgres/sql/upsertIntegrationSystem.sql`
- Create: `back/integrations/internal/storage/postgres/sql/createSyncJob.sql`
- Create: `back/integrations/internal/storage/postgres/sql/finishSyncJob.sql`
- Create: `back/integrations/internal/storage/postgres/sql/addSyncLog.sql`
- Create: `back/integrations/internal/storage/postgres/sql/upsertCategory.sql`
- Create: `back/integrations/internal/storage/postgres/sql/setCategoryParent.sql`
- Create: `back/integrations/internal/storage/postgres/sql/upsertWarehouse.sql`
- Create: `back/integrations/internal/storage/postgres/sql/upsertProduct.sql`
- Create: `back/integrations/internal/storage/postgres/sql/upsertProductPrice.sql`
- Create: `back/integrations/internal/storage/postgres/sql/upsertStockBalance.sql`
- Test: `back/integrations/internal/storage/postgres/integration_test.go`

**Interfaces:**
- Consumes: `config.Config` (Task 1), `models.*Input` типы (Task 3).
- Produces: `postgres.NewPool(cfg config.Config) (*pgxpool.Pool, error)`, `postgres.New(pool *pgxpool.Pool) *Storage`, `postgres.ErrDuplicateSKU error`, методы `(*Storage) UpsertIntegrationSystem(ctx, code, name string) (uuid.UUID, error)`, `CreateSyncJob(ctx, systemID uuid.UUID) (uuid.UUID, error)`, `FinishSyncJob(ctx, jobID uuid.UUID, status string) error`, `AddSyncLog(ctx, jobID, systemID uuid.UUID, level models.SyncLogLevel, message string) error`, `UpsertCategory(ctx, in models.CategoryInput) (uuid.UUID, error)`, `SetCategoryParent(ctx, id, parentID uuid.UUID) error`, `UpsertWarehouse(ctx, in models.WarehouseInput) (uuid.UUID, error)`, `UpsertProduct(ctx, in models.ProductInput) (uuid.UUID, error)`, `UpsertProductPrice(ctx, in models.PriceInput) error`, `UpsertStockBalance(ctx, in models.StockInput) error`.

- [ ] **Step 1: Написать SQL-файлы**

`back/integrations/internal/storage/postgres/sql/upsertIntegrationSystem.sql`:
```sql
INSERT INTO integration_systems (code, name, is_active)
VALUES ($1, $2, TRUE)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
RETURNING id;
```

`back/integrations/internal/storage/postgres/sql/createSyncJob.sql`:
```sql
INSERT INTO sync_jobs (system_id, direction, entity_type, status, attempts)
VALUES ($1, 'inbound', 'onec_full_sync', 'running', 1)
RETURNING id;
```

`back/integrations/internal/storage/postgres/sql/finishSyncJob.sql`:
```sql
UPDATE sync_jobs
SET status = $2,
    processed_at = now()
WHERE id = $1;
```

`back/integrations/internal/storage/postgres/sql/addSyncLog.sql`:
```sql
INSERT INTO sync_logs (job_id, system_id, level, message)
VALUES ($1, $2, $3, $4);
```

`back/integrations/internal/storage/postgres/sql/upsertCategory.sql`:
```sql
INSERT INTO categories (one_c_guid, name, is_active)
VALUES ($1, $2, TRUE)
ON CONFLICT (one_c_guid) DO UPDATE SET name = EXCLUDED.name
RETURNING id;
```

`back/integrations/internal/storage/postgres/sql/setCategoryParent.sql`:
```sql
UPDATE categories
SET parent_id = $2
WHERE id = $1;
```

`back/integrations/internal/storage/postgres/sql/upsertWarehouse.sql`:
```sql
INSERT INTO warehouses (one_c_guid, name, is_active)
VALUES ($1, $2, TRUE)
ON CONFLICT (one_c_guid) DO UPDATE SET name = EXCLUDED.name
RETURNING id;
```

`back/integrations/internal/storage/postgres/sql/upsertProduct.sql`:
```sql
INSERT INTO products (one_c_guid, category_id, sku, name, is_active)
VALUES ($1, $2, $3, $4, TRUE)
ON CONFLICT (one_c_guid) DO UPDATE
SET category_id = EXCLUDED.category_id,
    sku = EXCLUDED.sku,
    name = EXCLUDED.name,
    updated_at = now()
RETURNING id;
```

`back/integrations/internal/storage/postgres/sql/upsertProductPrice.sql`:
```sql
WITH updated AS (
    UPDATE product_prices
    SET price = $3,
        currency = 'RUB',
        valid_from = now(),
        synced_at = now(),
        updated_at = now()
    WHERE id = (
        SELECT id
        FROM product_prices
        WHERE product_id = $1
          AND price_group_id IS NULL
          AND price_type = $2
        ORDER BY valid_from DESC NULLS LAST, id
        LIMIT 1
    )
    RETURNING id
)
INSERT INTO product_prices (
    product_id, price_group_id, price_type, price, discount_percent, currency, valid_from, synced_at
)
SELECT $1, NULL, $2, $3, 0, 'RUB', now(), now()
WHERE NOT EXISTS (SELECT 1 FROM updated);
```

`back/integrations/internal/storage/postgres/sql/upsertStockBalance.sql`:
```sql
WITH updated AS (
    UPDATE stock_balances
    SET quantity = $3,
        reserved_quantity = 0,
        synced_at = now()
    WHERE id = (
        SELECT id
        FROM stock_balances
        WHERE product_id = $1
          AND warehouse_id = $2
        LIMIT 1
    )
    RETURNING id
)
INSERT INTO stock_balances (product_id, warehouse_id, quantity, reserved_quantity, synced_at)
SELECT $1, $2, $3, 0, now()
WHERE NOT EXISTS (SELECT 1 FROM updated);
```

- [ ] **Step 2: Написать падающий интеграционный тест**

Создать `back/integrations/internal/storage/postgres/integration_test.go`:

```go
package postgres

import (
	"context"
	"os"
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
	if err = storage.FinishSyncJob(ctx, jobID, "success"); err != nil {
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
```

- [ ] **Step 3: Убедиться, что пакет не компилируется**

```bash
go build ./internal/storage/...
```
Expected: FAIL — `undefined: New` (`postgres.go` ещё не создан).

- [ ] **Step 4: Реализовать `connectManager.go`**

Создать `back/integrations/internal/storage/postgres/connectManager.go`:

```go
package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mbatimel/AMC/integrations/internal/config"
)

func NewPool(cfg config.Config) (*pgxpool.Pool, error) {
	port, err := strconv.Atoi(cfg.PGPort)
	if err != nil {
		return nil, fmt.Errorf("parse PG_PORT: %w", err)
	}
	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s sslmode=disable user=%s password=%s",
		cfg.PGHost, port, cfg.PGDB, cfg.PGUser, cfg.PGPassword,
	)
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
```

- [ ] **Step 5: Реализовать `postgres.go`**

Создать `back/integrations/internal/storage/postgres/postgres.go`:

```go
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

func (s *Storage) FinishSyncJob(ctx context.Context, jobID uuid.UUID, status string) error {
	if _, err := s.pool.Exec(ctx, sqlFinishSyncJob, jobID, status); err != nil {
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
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return uuid.Nil, ErrDuplicateSKU
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
```

- [ ] **Step 6: Собрать пакет, подтянуть зависимости**

```bash
go get github.com/jackc/pgx/v4@v4.18.3
go get github.com/jackc/pgconn@v1.14.3
go mod tidy
go build ./...
```
Expected: сборка проходит без ошибок.

- [ ] **Step 7: Прогнать интеграционный тест (если есть тестовая БД)**

```bash
INTEGRATIONS_TEST_DATABASE_URL="postgres://amc:change-me@localhost:5432/amc?sslmode=disable" \
  go test ./internal/storage/postgres/... -run TestStorageIntegration -v
```
Expected: PASS, если переменная задана и БД с миграциями `back/migrations` доступна; иначе `SKIP` — это ожидаемо, не блокирует задачу (`go build ./...` из Step 6 — обязательное условие).

- [ ] **Step 8: Коммит**

```bash
git add back/integrations/go.mod back/integrations/go.sum back/integrations/internal/storage
git commit -m "feat(integrations): add postgres storage with one_c_guid upserts"
```

---

## Task 5: Оркестрация синка (`service.RunSync`)

**Files:**
- Create: `back/integrations/internal/service/sync.go`
- Test: `back/integrations/internal/service/sync_test.go`

**Interfaces:**
- Consumes: `onec.CategoryDTO/.../StockDTO` (Task 2), `models.*` (Task 3), `mapCategory/mapWarehouse/mapProduct/mapPrice/mapStock` (Task 3, тот же пакет).
- Produces: `service.OnecClient` interface, `service.Storage` interface, `service.New(logger zerolog.Logger, onecClient OnecClient, storage Storage) *Service`, `(*Service) RunSync(ctx context.Context) error`. `postgres.Storage` (Task 4) должен структурно удовлетворять `service.Storage` — сигнатуры уже совпадают.

- [ ] **Step 1: Написать падающие тесты оркестрации (с фейками)**

Создать `back/integrations/internal/service/sync_test.go`:

```go
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/integrations/internal/models"
	"github.com/mbatimel/AMC/integrations/internal/onec"
)

type fakeOnecClient struct {
	categories []onec.CategoryDTO
	warehouses []onec.WarehouseDTO
	products   []onec.ProductDTO
	prices     []onec.PriceDTO
	stock      []onec.StockDTO

	categoriesErr error
}

func (f *fakeOnecClient) FetchCategories(context.Context) ([]onec.CategoryDTO, error) {
	return f.categories, f.categoriesErr
}
func (f *fakeOnecClient) FetchWarehouses(context.Context) ([]onec.WarehouseDTO, error) {
	return f.warehouses, nil
}
func (f *fakeOnecClient) FetchProducts(context.Context) ([]onec.ProductDTO, error) {
	return f.products, nil
}
func (f *fakeOnecClient) FetchPrices(context.Context) ([]onec.PriceDTO, error) {
	return f.prices, nil
}
func (f *fakeOnecClient) FetchStock(context.Context) ([]onec.StockDTO, error) {
	return f.stock, nil
}

type fakeStorage struct {
	nextID       int
	categoryIDs  map[uuid.UUID]uuid.UUID
	warehouseIDs map[uuid.UUID]uuid.UUID
	productIDs   map[uuid.UUID]uuid.UUID
	parents      map[uuid.UUID]uuid.UUID
	prices       []models.PriceInput
	stocks       []models.StockInput
	logs         []string
	finalStatus  string
	failSKU      string
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{
		categoryIDs:  map[uuid.UUID]uuid.UUID{},
		warehouseIDs: map[uuid.UUID]uuid.UUID{},
		productIDs:   map[uuid.UUID]uuid.UUID{},
		parents:      map[uuid.UUID]uuid.UUID{},
	}
}

func (f *fakeStorage) newID() uuid.UUID {
	f.nextID++
	return uuid.MustParse(fmt.Sprintf("99999999-9999-9999-9999-%012d", f.nextID))
}

func (f *fakeStorage) UpsertIntegrationSystem(context.Context, string, string) (uuid.UUID, error) {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111"), nil
}
func (f *fakeStorage) CreateSyncJob(context.Context, uuid.UUID) (uuid.UUID, error) {
	return uuid.MustParse("22222222-2222-2222-2222-222222222222"), nil
}
func (f *fakeStorage) FinishSyncJob(_ context.Context, _ uuid.UUID, status string) error {
	f.finalStatus = status
	return nil
}
func (f *fakeStorage) AddSyncLog(_ context.Context, _, _ uuid.UUID, level models.SyncLogLevel, message string) error {
	f.logs = append(f.logs, string(level)+": "+message)
	return nil
}
func (f *fakeStorage) UpsertCategory(_ context.Context, in models.CategoryInput) (uuid.UUID, error) {
	id := f.newID()
	f.categoryIDs[in.OneCGUID] = id
	return id, nil
}
func (f *fakeStorage) SetCategoryParent(_ context.Context, id, parentID uuid.UUID) error {
	f.parents[id] = parentID
	return nil
}
func (f *fakeStorage) UpsertWarehouse(_ context.Context, in models.WarehouseInput) (uuid.UUID, error) {
	id := f.newID()
	f.warehouseIDs[in.OneCGUID] = id
	return id, nil
}
func (f *fakeStorage) UpsertProduct(_ context.Context, in models.ProductInput) (uuid.UUID, error) {
	if in.SKU == f.failSKU {
		return uuid.Nil, errors.New("duplicate sku")
	}
	id := f.newID()
	f.productIDs[in.OneCGUID] = id
	return id, nil
}
func (f *fakeStorage) UpsertProductPrice(_ context.Context, in models.PriceInput) error {
	f.prices = append(f.prices, in)
	return nil
}
func (f *fakeStorage) UpsertStockBalance(_ context.Context, in models.StockInput) error {
	f.stocks = append(f.stocks, in)
	return nil
}

const (
	parentGUID    = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	childGUID     = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	productGUID   = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	warehouseGUID = "dddddddd-dddd-dddd-dddd-dddddddddddd"
)

func TestRunSync_HappyPath(t *testing.T) {
	onecClient := &fakeOnecClient{
		categories: []onec.CategoryDTO{
			{RefKey: parentGUID, ParentKey: zeroGUID, Description: "Инструмент"},
			{RefKey: childGUID, ParentKey: parentGUID, Description: "Дрели"},
		},
		warehouses: []onec.WarehouseDTO{{RefKey: warehouseGUID, Description: "Склад №1"}},
		products: []onec.ProductDTO{
			{RefKey: productGUID, CategoryKey: childGUID, Code: "SKU-1", Description: "Дрель"},
		},
		prices: []onec.PriceDTO{{ProductKey: productGUID, PriceTypeKey: "type-a", Price: 100}},
		stock:  []onec.StockDTO{{ProductKey: productGUID, WarehouseKey: warehouseGUID, Quantity: 5}},
	}
	storage := newFakeStorage()
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "success" {
		t.Fatalf("expected success, got %s", storage.finalStatus)
	}
	if len(storage.prices) != 1 || storage.prices[0].Price != 100 {
		t.Fatalf("unexpected prices: %+v", storage.prices)
	}
	if len(storage.stocks) != 1 || storage.stocks[0].Quantity != 5 {
		t.Fatalf("unexpected stocks: %+v", storage.stocks)
	}
	childID := storage.categoryIDs[uuid.MustParse(childGUID)]
	parentID := storage.categoryIDs[uuid.MustParse(parentGUID)]
	if storage.parents[childID] != parentID {
		t.Fatalf("expected child category parent to be set to parent id")
	}
}

func TestRunSync_DuplicateSKU_SkippedButOthersContinue(t *testing.T) {
	onecClient := &fakeOnecClient{
		products: []onec.ProductDTO{
			{RefKey: "cccccccc-cccc-cccc-cccc-cccccccccccc", Code: "SKU-DUP", Description: "A"},
			{RefKey: "ffffffff-ffff-ffff-ffff-ffffffffffff", Code: "SKU-OK", Description: "B"},
		},
	}
	storage := newFakeStorage()
	storage.failSKU = "SKU-DUP"
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "partial" {
		t.Fatalf("expected partial, got %s", storage.finalStatus)
	}
	if len(storage.productIDs) != 1 {
		t.Fatalf("expected 1 product synced, got %d", len(storage.productIDs))
	}
	found := false
	for _, l := range storage.logs {
		if strings.Contains(l, "SKU-DUP") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sync log about duplicate sku, got %v", storage.logs)
	}
}

func TestRunSync_CategoriesFetchError_ContinuesOtherSteps(t *testing.T) {
	onecClient := &fakeOnecClient{
		categoriesErr: errors.New("network down"),
		warehouses:    []onec.WarehouseDTO{{RefKey: warehouseGUID, Description: "Склад №1"}},
	}
	storage := newFakeStorage()
	svc := New(zerolog.Nop(), onecClient, storage)

	if err := svc.RunSync(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if storage.finalStatus != "partial" {
		t.Fatalf("expected partial, got %s", storage.finalStatus)
	}
	if len(storage.warehouseIDs) != 1 {
		t.Fatalf("expected warehouses step to still run, got %d", len(storage.warehouseIDs))
	}
	found := false
	for _, l := range storage.logs {
		if strings.Contains(l, "network down") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sync log about categories error, got %v", storage.logs)
	}
}
```

- [ ] **Step 2: Убедиться, что тесты падают**

```bash
go test ./internal/service/... -run TestRunSync -v
```
Expected: FAIL — `undefined: New` (в пакете `service` ещё нет `sync.go`).

- [ ] **Step 3: Реализовать оркестрацию**

Создать `back/integrations/internal/service/sync.go`:

```go
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
```

- [ ] **Step 4: Прогнать тесты**

```bash
go test ./internal/service/... -v
```
Expected: PASS — все тесты `TestMap*` (Task 3) и `TestRunSync_*` (этот таск).

- [ ] **Step 5: Проверить, что `*postgres.Storage` реально удовлетворяет `service.Storage`**

Добавить временную строку в конец `back/integrations/internal/storage/postgres/postgres.go` (после `New`) для компайл-тайм проверки — **нет**, вместо временного кода просто собрать весь модуль целиком, раз `cmd` (Task 6) ещё не существует. На этом шаге — точечная проверка через `go vet`:

```bash
go build ./...
go vet ./...
```
Expected: успешно (пока `cmd/main.go` не создан, реальной проверки присваивания `*postgres.Storage` в `service.Storage` не будет — она произойдёт в Task 6 при компиляции `main.go`, где выполняется `service.New(logger, onecClient, storage)`).

- [ ] **Step 6: Коммит**

```bash
git add back/integrations/internal/service
git commit -m "feat(integrations): add onec sync orchestration"
```

---

## Task 6: `cmd/main.go`, health-сервер, цикл по тикеру

**Files:**
- Create: `back/integrations/internal/transport/http/health.go`
- Create: `back/integrations/cmd/loop.go`
- Create: `back/integrations/cmd/main.go`
- Test: `back/integrations/cmd/loop_test.go`

**Interfaces:**
- Consumes: `config.LoadConfig/LoadEnvFile` (Task 1), `onec.New` (Task 2), `service.New`, `(*Service).RunSync` (Task 5), `postgres.NewPool`, `postgres.New` (Task 4).
- Produces: `transportHTTP.NewHealthServer() *HealthServer` с `Start(addr string) error`/`Stop() error` (копия паттерна `products`), package-`main` функция `runLoop(ctx context.Context, interval time.Duration, run func(context.Context) error, onError func(error))`.

- [ ] **Step 1: Health-сервер (копия паттерна `products`)**

Создать `back/integrations/internal/transport/http/health.go`:

```go
package http

import "github.com/gofiber/fiber/v2"

type HealthServer struct {
	server *fiber.App
}

func NewHealthServer() *HealthServer {
	health := &HealthServer{server: fiber.New(fiber.Config{DisableStartupMessage: true})}
	health.server.Get("/liveness", probesHandler)
	health.server.Get("/readiness", probesHandler)
	return health
}

func probesHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

func (h *HealthServer) Start(bindURL string) error {
	return h.server.Listen(bindURL)
}

func (h *HealthServer) Stop() error {
	return h.server.Shutdown()
}
```

- [ ] **Step 2: Написать падающие тесты цикла-тикера**

Создать `back/integrations/cmd/loop_test.go`:

```go
package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRunLoop_RunsImmediatelyAndOnInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var mu sync.Mutex
	calls := 0
	done := make(chan struct{})

	run := func(context.Context) error {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 2 {
			close(done)
		}
		return nil
	}

	go runLoop(ctx, 10*time.Millisecond, run, func(error) {})

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for second run")
	}

	mu.Lock()
	defer mu.Unlock()
	if calls < 2 {
		t.Fatalf("expected at least 2 calls, got %d", calls)
	}
}

func TestRunLoop_ReportsErrorFromRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	run := func(context.Context) error { return errors.New("boom") }
	onError := func(err error) { errCh <- err }

	go runLoop(ctx, time.Hour, run, onError)

	select {
	case err := <-errCh:
		if err.Error() != "boom" {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for error callback")
	}
}

func TestRunLoop_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var mu sync.Mutex
	calls := 0
	run := func(context.Context) error {
		mu.Lock()
		calls++
		mu.Unlock()
		return nil
	}

	loopDone := make(chan struct{})
	go func() {
		runLoop(ctx, time.Hour, run, func(error) {})
		close(loopDone)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-loopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for runLoop to return after cancel")
	}

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected exactly 1 immediate call before cancel, got %d", calls)
	}
}
```

- [ ] **Step 3: Убедиться, что тесты падают**

```bash
go test ./cmd/... -v
```
Expected: FAIL — `undefined: runLoop` (`loop.go` ещё не создан).

- [ ] **Step 4: Реализовать цикл**

Создать `back/integrations/cmd/loop.go`:

```go
package main

import (
	"context"
	"time"
)

// runLoop выполняет run немедленно, затем повторяет по interval,
// пока ctx не отменят. Каждая ошибка run (включая первый немедленный
// вызов) передаётся в onError; сам цикл при этом не останавливается.
func runLoop(ctx context.Context, interval time.Duration, run func(context.Context) error, onError func(error)) {
	if err := run(ctx); err != nil {
		onError(err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := run(ctx); err != nil {
				onError(err)
			}
		}
	}
}
```

- [ ] **Step 5: Прогнать тесты**

```bash
go test ./cmd/... -v
```
Expected: PASS — все 3 теста.

- [ ] **Step 6: Написать `main.go`**

Создать `back/integrations/cmd/main.go`:

```go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/mbatimel/AMC/integrations/internal/config"
	"github.com/mbatimel/AMC/integrations/internal/onec"
	"github.com/mbatimel/AMC/integrations/internal/service"
	"github.com/mbatimel/AMC/integrations/internal/storage/postgres"
	transportHTTP "github.com/mbatimel/AMC/integrations/internal/transport/http"
)

const serviceName = "onec-sync"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Str("serviceName", serviceName).Logger()

	if err := config.LoadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}
	cfg := config.LoadConfig()

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()

	storageImpl := postgres.New(pool)
	onecClient := onec.New(cfg.OnecBaseURL, cfg.OnecUser, cfg.OnecPassword, log.Logger)
	svc := service.New(log.Logger, onecClient, storageImpl)

	healthServer := transportHTTP.NewHealthServer()
	go func() {
		if healthErr := healthServer.Start(cfg.HealthAddr); healthErr != nil {
			log.Error().Err(healthErr).Msg("onec-sync health server stopped")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-shutdown
		cancel()
	}()

	log.Info().Dur("interval", cfg.SyncInterval).Msg("onec sync worker started")
	runLoop(ctx, cfg.SyncInterval, svc.RunSync, func(err error) {
		log.Error().Err(err).Msg("onec sync run failed")
	})

	if err = healthServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop health server")
	}
}
```

- [ ] **Step 7: Собрать весь модуль (проверка, что `*postgres.Storage` удовлетворяет `service.Storage`)**

```bash
go get github.com/gofiber/fiber/v2@v2.52.13
go mod tidy
go build ./...
go vet ./...
go test ./...
```
Expected: сборка и все тесты модуля проходят. Если `service.New(log.Logger, onecClient, storageImpl)` не компилируется — сверить сигнатуры методов `*postgres.Storage` (Task 4) с интерфейсом `service.Storage` (Task 5) один в один.

- [ ] **Step 8: Коммит**

```bash
git add back/integrations/go.mod back/integrations/go.sum back/integrations/cmd back/integrations/internal/transport
git commit -m "feat(integrations): add main entrypoint with daily sync loop"
```

---

## Task 7: Деплой — Dockerfile, docker-compose, .env, CI

**Files:**
- Create: `back/integrations/Dockerfile`
- Modify: `deploy/docker-compose.yml`
- Modify: `deploy/.env.example`
- Modify: `.github/workflows/go.yml`
- Modify: `.github/workflows/build.yml`

**Interfaces:**
- Consumes: бинарь `/out/onec-sync` из `back/integrations/cmd` (Task 6), переменные окружения из `config.Config` (Task 1).
- Produces: образ `ghcr.io/mbatimel/amc-onec-sync`, сервис `onec-sync` в compose-стеке.

- [ ] **Step 1: Dockerfile**

Создать `back/integrations/Dockerfile`:

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/integrations/go.mod back/integrations/go.sum ./integrations/
WORKDIR /src/integrations
RUN go mod download

COPY back/integrations/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/onec-sync ./cmd

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/onec-sync /usr/local/bin/onec-sync

USER app
EXPOSE 9096

ENTRYPOINT ["/usr/local/bin/onec-sync"]
```

- [ ] **Step 2: `docker-compose.yml` — новый сервис**

В `deploy/docker-compose.yml`, сразу после блока `products:` (перед `front:`), добавить:

```yaml
  onec-sync:
    image: ghcr.io/mbatimel/amc-onec-sync:${IMAGE_TAG}
    restart: unless-stopped
    environment:
      PG_HOST: postgres
      PG_PORT: "5432"
      PG_DB: ${PG_DB}
      PG_USER: ${PG_USER}
      PG_PASSWORD: ${PG_PASSWORD}
      HEALTH_ADDR: ":9096"
      ONEC_BASE_URL: ${ONEC_BASE_URL}
      ONEC_USER: ${ONEC_USER}
      ONEC_PASSWORD: ${ONEC_PASSWORD}
      SYNC_INTERVAL: "24h"
    depends_on:
      migrations:
        condition: service_completed_successfully
    networks: [amc_net]
```

- [ ] **Step 3: `.env.example` — новые переменные**

В конец `deploy/.env.example` добавить:

```
# 1С:УТ 10.3 (сервер PVISERVER) — OData standard interface.
# Пока не опубликован на стороне 1С (см. back/integrations/README.md) —
# воркер будет падать на каждом ране до публикации, это ожидаемо.
ONEC_BASE_URL=http://PVISERVER/UT/odata/standard.odata
ONEC_USER=
ONEC_PASSWORD=
```

- [ ] **Step 4: Проверить, что compose-файл валиден**

```bash
cd deploy
docker compose config --quiet
```
Expected: без ошибок (может ругаться на пустые `ONEC_USER`/`ONEC_PASSWORD` только предупреждением, не ошибкой — это ожидаемо для локального `.env`, если он не заполнен; при необходимости временно скопировать `.env.example` в `.env` для проверки).

- [ ] **Step 5: CI — добавить модуль в матрицу тестов**

В `.github/workflows/go.yml`, в список `matrix.module`, добавить строку (по алфавиту, после `back/auth` и перед `back/migrations` — сохраняя текущий порядок файла):

```yaml
        module:
          - back/access
          - back/admin
          - back/auth
          - back/integrations
          - back/migrations
          - back/orders
          - back/products
          - back/users
```

- [ ] **Step 6: CI — добавить сборку образа**

В `.github/workflows/build.yml`, в `matrix.service`, добавить (после `admin`, перед `migrations` — по алфавиту):

```yaml
          - name: integrations
            dockerfile: ./back/integrations/Dockerfile
            image: ghcr.io/mbatimel/amc-onec-sync
```

- [ ] **Step 7: Коммит**

```bash
git add back/integrations/Dockerfile deploy/docker-compose.yml deploy/.env.example .github/workflows/go.yml .github/workflows/build.yml
git commit -m "feat(integrations): wire onec-sync into deploy and CI"
```

---

## Task 8: Обновить `back/integrations/README.md`

**Files:**
- Modify: `back/integrations/README.md`

**Interfaces:**
- Consumes: ничего (документация).
- Produces: актуальный статус для читателей README.

- [ ] **Step 1: Заменить раздел "## Статус"**

Найти в `back/integrations/README.md` блок:

```
## Статус

Исследование доступа проведено, воркер ещё не реализован. Следующий шаг — согласовать с
1С-разработчиком доступ (см. черновик выше), затем спроектировать воркер под конкретный
протокол (OData/HTTP-сервис).
```

Заменить на:

```
## Статус

Воркер реализован (`back/integrations`, сервис `onec-sync` в `deploy/docker-compose.yml`).
Дизайн и объём v1 — `docs/superpowers/specs/2026-08-25-onec-sync-worker-design.md`.

Синкает раз в сутки (или после рестарта контейнера сразу): категории
(`Catalog.НоменклатурныеГруппы`), склады (`Catalog.Склады`), товары — артикул/наименование/
категория (`Catalog.Номенклатура`), цены (`InformationRegister.ЦеныНоменклатуры`), остатки по
складам (`AccumulationRegister.ТоварыНаСкладах`). Brand/GOST/Material/Size/Images/единицы
измерения/группы цен — вне объёма v1, см. спек.

Заблокировано на **публикации OData на сервере `PVISERVER`** (см. раздел "Рекомендуемый путь"
выше — IIS + галка "Стандартный интерфейс OData" в Конфигураторе, плюс технический
пользователь на чтение). Имена OData entity set'ов/полей в `internal/onec/models.go` —
best-effort по стандартному именованию 1С, не сверены с реальным сервером. После публикации:
1. Заполнить `ONEC_BASE_URL`/`ONEC_USER`/`ONEC_PASSWORD` в `deploy/.env`.
2. Прогнать воркер разово (`docker compose up onec-sync` или локально `go run ./cmd`
   из `back/integrations`) и свериться с логами (`sync_logs` в БД, либо просто вывод —
   воркер логирует каждый OData-ответ целиком).
3. Поправить константы `entitySet*` и JSON-теги DTO в `internal/onec/models.go`, если реальные
   имена отличаются от предположенных.
```

- [ ] **Step 2: Коммит**

```bash
git add back/integrations/README.md
git commit -m "docs(integrations): mark onec-sync worker as implemented"
```

---

## Self-Review (проведён)

- **Покрытие спека:** объём v1 (категории/склады/товары/цены/остатки) — Tasks 2–5; наблюдаемость через `integration_systems`/`sync_jobs`/`sync_logs` — Task 4/5; запуск раз в сутки отдельным контейнером с тикером — Task 6/7; отсутствие новых миграций — соблюдено везде (только существующие колонки/таблицы); частичный сбой не валит весь ран — Task 5 (`hadErrors`/`status`), покрыто тестами.
- **Плейсхолдеры:** не найдено — весь код и SQL приведены полностью, включая тестовые фикстуры.
- **Согласованность типов:** `models.ProductInput.CategoryID *uuid.UUID` — одинаково в Task 3 (mapping), Task 4 (storage SQL-параметр), Task 5 (тесты); `service.Storage`/`service.OnecClient` интерфейсы (Task 5) сигнатурно совпадают с `*postgres.Storage` (Task 4) и `*onec.Client` (Task 2) — проверяется компиляцией `cmd/main.go` в Task 6, Step 7.

---

**Plan complete and saved to `docs/superpowers/plans/2026-08-25-onec-sync-worker.md`. Two execution options:**

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
