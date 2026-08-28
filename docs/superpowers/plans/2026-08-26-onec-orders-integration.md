# Заказ ↔ 1С Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Заказ при оформлении синхронно уходит в 1С (статус `new`→`processing`), 1С по вебхуку переводит заказ в `delivered`.

**Architecture:** `orders.CreateOrder` держит одну DB-транзакцию открытой на весь вызов, включая синхронный HTTP-поход в новый процесс `onec-orders-api` (модуль `back/integrations`), который либо проксирует данные в кастомный HTTP-service 1С (ручка "создание заказа" — исходящий вызов AMC→1С), либо принимает вебхук от 1С (входящий, публичный, статус `delivered`) и дёргает уже существующий `orders.UpdateOrderStatus` от имени системного admin-пользователя.

**Tech Stack:** Go 1.25, Postgres (pgx/v4), fiber (через `tg`-сгенерированный транспорт для CRUD-ручек orders и вебхука; **вручную** — fiber-роут для `PushOrder`, т.к. `tg` не умеет параметры-структуры/слайсы структур, см. ниже), fasthttp (клиенты).

**Spec:**
- `docs/superpowers/specs/2026-08-26-onec-orders-integration-design.md`
- `docs/superpowers/specs/2026-08-26-onec-orders-integration-1c-contract.md`

## Global Constraints

- Стейт-машина заказа: `new` (только внутри незакоммиченной транзакции) → `processing` (после успешного `PushOrder`) → `delivered` (по вебхуку). `cancelled` — только `CancelOrder`, 1С не может его выставить и не может перевести уже `cancelled`-заказ в `delivered` (409).
- 1С недоступна/ответила ошибкой при создании заказа → **вся транзакция создания заказа откатывается** (не только сам инсерт заказа — вставка `order_items`, `order_status_history`, `DELETE FROM cart_items` тоже откатываются).
- **`tg`-кодогенерацию (`tg transport`, `tg client`) в этом плане никто не запускает и generated-файлы не пишет.** Каждая задача, где нужен сгенерированный транспорт/клиент, готовит `@tg`-аннотированный интерфейс и custom-handler (обычный Go-файл, не генерируется), и явно помечена «⚠ требует `tg`-кодогенерации перед сборкой этого пакета» — пользователь запускает генерацию сам между задачами.
- `PushOrder` (AMC→1С, несёт массив позиций заказа) и клиент к нему на стороне `orders` — **не через `tg`**, обычный `encoding/json` HTTP поверх `*fiber.App`/fasthttp. Экспериментально проверено: `tg transport` ломается на параметре-структуре и на `[]struct` (тип теряется, кодогенерация падает с ошибкой синтаксиса).
- Сервис-к-сервису внутри `amc_net` — без доп. auth поверх сетевой изоляции (как `orders→access` сегодня). Публичный вебхук от 1С — `X-Onec-Api-Key`, статический ключ.

---

### Task 1: Сид-миграция системного пользователя "1С"

**Files:**
- Create: `back/migrations/pkg/migrations/data/20260826150000_onec_system_user.sql`

**Interfaces:**
- Produces: фиксированный UUID `00000000-0000-0000-0000-0000000a0ec1` — системный пользователь с ролью admin (code=0). Используется как `ORDERS_SYSTEM_USER_ID` в конфиге `onec-orders-api` (Task 13) для вызовов `orders.UpdateOrderStatus`/`GetOrderStatus`.

- [ ] **Step 1: Написать миграцию**

```sql
-- +goose Up
-- +goose StatementBegin
INSERT INTO users (id, email, name, status)
VALUES ('00000000-0000-0000-0000-0000000a0ec1', 'onec-integration@system.local', '1С интеграция', 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT '00000000-0000-0000-0000-0000000a0ec1', id FROM roles WHERE code = 0
ON CONFLICT (user_id, role_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM user_roles WHERE user_id = '00000000-0000-0000-0000-0000000a0ec1';
DELETE FROM users WHERE id = '00000000-0000-0000-0000-0000000a0ec1';
-- +goose StatementEnd
```

- [ ] **Step 2: Применить локально и проверить**

Требует запущенный Postgres с уже применёнными предыдущими миграциями. Из `back/migrations`:

```bash
PG_ADDRESS=localhost:5432 PG_DB=<your_test_db> PG_USER=<user> PG_PASSWORD=<pass> go run ./cmd
```

Ожидаемо: миграция применяется без ошибок (проверить логи — новая версия в списке применённых). Затем:

```sql
SELECT u.email, r.code FROM users u JOIN user_roles ur ON ur.user_id = u.id JOIN roles r ON r.id = ur.role_id WHERE u.id = '00000000-0000-0000-0000-0000000a0ec1';
```

Ожидаемо: одна строка, `code = 0`.

- [ ] **Step 3: Commit**

```bash
git add back/migrations/pkg/migrations/data/20260826150000_onec_system_user.sql
git commit -m "feat(migrations): seed system user for 1С order-status webhook"
```

---

### Task 2: orders — чтение one_c_guid для товаров и контрагента

**Files:**
- Modify: `back/orders/internal/storage/postgres/orders.go`
- Test: `back/orders/internal/storage/postgres/integration_test.go`

**Interfaces:**
- Produces:
  - `type ProductOnecRef struct { OneCGUID uuid.NullUUID; SKU string }`
  - `func (s *Storage) GetProductOnecRefs(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]ProductOnecRef, error)`
  - `type CounterpartyOnecRef struct { OneCGUID uuid.NullUUID; INN string; Name string }`
  - `func (s *Storage) GetCounterpartyOnecRef(ctx context.Context, counterpartyID uuid.UUID) (CounterpartyOnecRef, error)`

- [ ] **Step 1: Написать падающий тест**

Добавить в `back/orders/internal/storage/postgres/integration_test.go` (файл уже использует `os.Getenv("TEST_DATABASE_URL")` + `t.Skip`, см. существующие тесты в этом файле — повторить тот же паттерн подключения):

```go
func TestGetProductOnecRefs(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	storage := New(pool)

	guid := uuid.New()
	var productID uuid.UUID
	if err = pool.QueryRow(ctx, `
		INSERT INTO products (one_c_guid, sku, name, is_active) VALUES ($1, $2, 'Товар', TRUE) RETURNING id
	`, guid, "REF-TEST-"+guid.String()[:8]).Scan(&productID); err != nil {
		t.Fatalf("insert product: %v", err)
	}
	var productIDNoGUID uuid.UUID
	if err = pool.QueryRow(ctx, `
		INSERT INTO products (sku, name, is_active) VALUES ($1, 'Товар без GUID', TRUE) RETURNING id
	`, "REF-TEST-NOGUID-"+uuid.New().String()[:8]).Scan(&productIDNoGUID); err != nil {
		t.Fatalf("insert product without guid: %v", err)
	}

	refs, err := storage.GetProductOnecRefs(ctx, []uuid.UUID{productID, productIDNoGUID})
	if err != nil {
		t.Fatalf("GetProductOnecRefs: %v", err)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d: %+v", len(refs), refs)
	}
	if !refs[productID].OneCGUID.Valid || refs[productID].OneCGUID.UUID != guid {
		t.Fatalf("expected productID to have onec guid %s, got %+v", guid, refs[productID])
	}
	if refs[productIDNoGUID].OneCGUID.Valid {
		t.Fatalf("expected productIDNoGUID to have no onec guid, got %+v", refs[productIDNoGUID])
	}
}

func TestGetCounterpartyOnecRef(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	storage := New(pool)

	guid := uuid.New()
	var counterpartyID uuid.UUID
	if err = pool.QueryRow(ctx, `
		INSERT INTO counterparties (one_c_guid, inn, name) VALUES ($1, '7701234567', 'ООО Рефтест') RETURNING id
	`, guid).Scan(&counterpartyID); err != nil {
		t.Fatalf("insert counterparty: %v", err)
	}

	ref, err := storage.GetCounterpartyOnecRef(ctx, counterpartyID)
	if err != nil {
		t.Fatalf("GetCounterpartyOnecRef: %v", err)
	}
	if !ref.OneCGUID.Valid || ref.OneCGUID.UUID != guid || ref.INN != "7701234567" || ref.Name != "ООО Рефтест" {
		t.Fatalf("unexpected ref: %+v", ref)
	}
}
```

- [ ] **Step 2: Прогнать и убедиться, что падает**

```bash
cd back/orders && TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run TestGetProductOnecRefs -v
```

Ожидаемо: `FAIL` — `undefined: (*Storage).GetProductOnecRefs` (компиляция).

- [ ] **Step 3: Реализовать**

Добавить в `back/orders/internal/storage/postgres/orders.go`:

```go
type ProductOnecRef struct {
	OneCGUID uuid.NullUUID
	SKU      string
}

func (s *Storage) GetProductOnecRefs(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]ProductOnecRef, error) {
	result := make(map[uuid.UUID]ProductOnecRef, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, one_c_guid, sku FROM products WHERE id = ANY($1)
	`, productIDs)
	if err != nil {
		return nil, fmt.Errorf("get product onec refs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var ref ProductOnecRef
		if err = rows.Scan(&id, &ref.OneCGUID, &ref.SKU); err != nil {
			return nil, fmt.Errorf("get product onec refs: scan: %w", err)
		}
		result[id] = ref
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("get product onec refs: %w", err)
	}
	return result, nil
}

type CounterpartyOnecRef struct {
	OneCGUID uuid.NullUUID
	INN      string
	Name     string
}

func (s *Storage) GetCounterpartyOnecRef(ctx context.Context, counterpartyID uuid.UUID) (CounterpartyOnecRef, error) {
	var ref CounterpartyOnecRef
	err := s.pool.QueryRow(ctx, `
		SELECT one_c_guid, COALESCE(inn, ''), COALESCE(name, '') FROM counterparties WHERE id = $1
	`, counterpartyID).Scan(&ref.OneCGUID, &ref.INN, &ref.Name)
	if err != nil {
		return CounterpartyOnecRef{}, fmt.Errorf("get counterparty onec ref: %w", err)
	}
	return ref, nil
}
```

`counterparty.go` уже импортирует `context`, `errors`, `fmt`, `uuid`, `pgx` — новые методы кладём в `orders.go`, у которого те же импорты уже есть (проверить перед вставкой, что `fmt` импортирован — используется на соседних методах файла).

- [ ] **Step 4: Прогнать снова**

```bash
cd back/orders && TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run 'TestGetProductOnecRefs|TestGetCounterpartyOnecRef' -v
```

Ожидаемо: `PASS` для обоих тестов.

- [ ] **Step 5: Commit**

```bash
git add back/orders/internal/storage/postgres/orders.go back/orders/internal/storage/postgres/integration_test.go
git commit -m "feat(orders): read product/counterparty one_c_guid refs"
```

---

### Task 3: orders — `GetOrderStatus` (узкая admin-ручка для вебхука)

**Files:**
- Modify: `back/orders/internal/storage/postgres/orders.go`
- Modify: `back/orders/internal/service/service.go`
- Modify: `back/orders/pkg/interfaces/externalapi/interface.go`
- Modify: `back/orders/pkg/models/responses.go`
- Modify: `back/orders/internal/transport/custom-handlers/orders.go`
- Test: `back/orders/internal/storage/postgres/integration_test.go`
- Test: новый `back/orders/internal/service/get_order_status_test.go`

**Interfaces:**
- Consumes: `postgres.ErrOrderNotFound` (уже определён в `orders.go:278`), `customErrors.NotFoundError()`/`ForbiddenError()`/`InternalServerError()` (`back/orders/internal/errors/common.go`), `s.checkAdminAccess(ctx, userID)` (`service.go`).
- Produces:
  - `func (s *Storage) GetOrderStatus(ctx context.Context, orderID uuid.UUID) (string, error)` — возвращает `postgres.ErrOrderNotFound`, если заказа нет.
  - `func (s *service) GetOrderStatus(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (response models.GetOrderStatusResponse, err error)` — часть `externalapi.OrdersAPI`.
  - `models.GetOrderStatusResponse{ Status string }`.
  - HTTP: `GET /api/v1/orders/{orderID}/status`, заголовок `X-User-Id`, только admin (переиспользует существующий `checkAdminAccess`, как `UpdateOrderStatus`).

- [ ] **Step 1: Написать падающий тест на storage**

В `back/orders/internal/storage/postgres/integration_test.go` (тот же файл, что Task 2):

```go
func TestGetOrderStatus(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	storage := New(pool)

	var orderID uuid.UUID
	if err = pool.QueryRow(ctx, `
		INSERT INTO orders (number, status, payment_status) VALUES ($1, 'processing', 'not_paid') RETURNING id
	`, "TEST-STATUS-"+uuid.New().String()[:8]).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}

	status, err := storage.GetOrderStatus(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrderStatus: %v", err)
	}
	if status != "processing" {
		t.Fatalf("expected status processing, got %q", status)
	}

	_, err = storage.GetOrderStatus(ctx, uuid.New())
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound for unknown order, got %v", err)
	}
}
```

- [ ] **Step 2: Прогнать, убедиться что падает (компиляция)**

```bash
cd back/orders && TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run TestGetOrderStatus -v
```

- [ ] **Step 3: Реализовать storage-метод**

В `back/orders/internal/storage/postgres/orders.go`:

```go
func (s *Storage) GetOrderStatus(ctx context.Context, orderID uuid.UUID) (string, error) {
	var status string
	err := s.pool.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1`, orderID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrOrderNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get order status: %w", err)
	}
	return status, nil
}
```

- [ ] **Step 4: Прогнать, убедиться что storage-тест проходит**

```bash
cd back/orders && TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run TestGetOrderStatus -v
```

Ожидаемо: `PASS`.

- [ ] **Step 5: Написать падающий тест на service-слой**

Создать `back/orders/internal/service/get_order_status_test.go`. Использовать существующий паттерн фейкового `Storage`/`AccessClient` из `client_resolution_test.go` (тип `clientResolutionStorage` в этом же пакете уже реализует `Storage` — если он не покрывает `GetOrderStatus`, добавить метод туда же, т.к. это тот же файл/тип, используемый несколькими тестами пакета `service`):

```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/orders/internal/errors"
	"github.com/mbatimel/AMC/orders/internal/storage/postgres"
)

func TestGetOrderStatus_AdminAllowed(t *testing.T) {
	storage := &clientResolutionStorage{orderStatus: "processing"}
	access := &fakeAccessClient{allowed: true}
	svc := NewOrdersApiService(zerolog.Nop(), storage, access, 20)

	resp, err := svc.GetOrderStatus(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "processing" {
		t.Fatalf("expected status processing, got %q", resp.Status)
	}
}

func TestGetOrderStatus_NonAdminForbidden(t *testing.T) {
	storage := &clientResolutionStorage{orderStatus: "processing"}
	access := &fakeAccessClient{allowed: false}
	svc := NewOrdersApiService(zerolog.Nop(), storage, access, 20)

	_, err := svc.GetOrderStatus(context.Background(), uuid.New(), uuid.New())
	var custErr *customErrors.Error
	if !errors.As(err, &custErr) || custErr.GetStatusCode() != 403 {
		t.Fatalf("expected 403 forbidden, got %v", err)
	}
}

func TestGetOrderStatus_NotFound(t *testing.T) {
	storage := &clientResolutionStorage{orderStatusErr: postgres.ErrOrderNotFound}
	access := &fakeAccessClient{allowed: true}
	svc := NewOrdersApiService(zerolog.Nop(), storage, access, 20)

	_, err := svc.GetOrderStatus(context.Background(), uuid.New(), uuid.New())
	var custErr *customErrors.Error
	if !errors.As(err, &custErr) || custErr.GetStatusCode() != 404 {
		t.Fatalf("expected 404 not found, got %v", err)
	}
}
```

Проверить перед написанием: как в `client_resolution_test.go` называется фейк `AccessClient` (например `fakeAccessClient` с полем `allowed bool`) — использовать существующее имя типа, не заводить второй одноимённый фейк в том же пакете. Если имя другое — использовать его. Добавить в `clientResolutionStorage` два новых поля и метод:

```go
// добавить в структуру clientResolutionStorage
orderStatus    string
orderStatusErr error

func (s *clientResolutionStorage) GetOrderStatus(ctx context.Context, orderID uuid.UUID) (string, error) {
	if s.orderStatusErr != nil {
		return "", s.orderStatusErr
	}
	return s.orderStatus, nil
}
```

- [ ] **Step 6: Прогнать, убедиться что падает (компиляция — `GetOrderStatus` ещё не в `externalapi.OrdersAPI`/`service`)**

```bash
cd back/orders && go test ./internal/service/... -run TestGetOrderStatus -v
```

- [ ] **Step 7: Добавить в `externalapi.OrdersAPI`**

В `back/orders/pkg/interfaces/externalapi/interface.go`, после `UpdateOrderStatus`:

```go
	// GetOrderStatus returns just the order's current status.
	// @tg http-method=GET
	// @tg http-path=/v1/admin/orders/status
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=orderID|orderID
	// @tg http-response=github.com/mbatimel/AMC/orders/internal/transport/custom-handlers:GetOrderStatus
	// @tg summary=`Статус заказа`
	// @tg desc=`Узкая admin-ручка: только текущий статус заказа, используется системными вызовами (например вебхуком 1С) для проверки допустимости перехода перед PATCH`
	// @tg uuidPackage=github.com/google/uuid
	GetOrderStatus(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (response models.GetOrderStatusResponse, err error)
```

Также добавить строку `//go:generate tg client --services . --outPath ../../client/transport -go` сразу после существующей `//go:generate tg transport ...` строки (по образцу `back/access/pkg/interfaces/internalAPI/interface.go:7-8`) — нужен для Task 9/10 (integrations вызывает `orders` как Go-клиент).

- [ ] **Step 8: Добавить `models.GetOrderStatusResponse`**

В `back/orders/pkg/models/responses.go`, после `UpdateOrderStatusResponse`:

```go
type GetOrderStatusResponse struct {
	Status string `json:"status"`
}
```

- [ ] **Step 9: Добавить custom-handler**

В `back/orders/internal/transport/custom-handlers/orders.go`, после `UpdateOrderStatus`:

```go
func GetOrderStatus(ctx *fiber.Ctx, svc externalapi.OrdersAPI, userID uuid.UUID, orderID uuid.UUID) error {
	return handle(ctx, "get", "/v1/admin/orders/{orderID}/status", "GetOrderStatus", map[string]interface{}{
		"orderID": orderID,
		"userID":  userID,
	}, func() (interface{}, error) {
		return svc.GetOrderStatus(ctx.UserContext(), userID, orderID)
	})
}
```

- [ ] **Step 10: Реализовать service-метод**

В `back/orders/internal/service/service.go`, после `UpdateOrderStatus`:

```go
func (s *service) GetOrderStatus(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) (response models.GetOrderStatusResponse, err error) {
	if err = s.checkAdminAccess(ctx, userID); err != nil {
		return response, err
	}
	status, err := s.storage.GetOrderStatus(ctx, orderID)
	if err != nil {
		if errors.Is(err, postgres.ErrOrderNotFound) {
			return response, customErrors.NotFoundError()
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}
	return models.GetOrderStatusResponse{Status: status}, nil
}
```

Добавить `GetOrderStatus(ctx context.Context, orderID uuid.UUID) (string, error)` в интерфейс `Storage` этого же файла (`back/orders/internal/service/service.go`, рядом с `UpdateOrderStatus` в списке методов интерфейса).

- [ ] **Step 11: Прогнать service-тесты**

```bash
cd back/orders && go test ./internal/service/... -run TestGetOrderStatus -v
```

Ожидаемо: все три `PASS`.

- [ ] **Step 12: ⚠ Кодогенерация (выполняет пользователь, не эта задача)**

После этого шага нужно перегенерировать транспорт и клиент:
```bash
cd back/orders/pkg/interfaces/externalapi && go generate ./...
```
Это обновит `back/orders/internal/transport/jsonRPC/externalapi/*`, `back/orders/swaggers/externalapi/swagger.yaml` и создаст `back/orders/pkg/client/transport/*`. Без этого шага `go build ./...` в `back/orders` не пройдёт (новый метод есть в интерфейсе, но не в сгенерированном `Server`/wiring). **Следующий шаг плана (`go build ./...`) можно выполнять только после этой генерации.**

- [ ] **Step 13: Собрать весь модуль и убедиться, что всё стыкуется**

```bash
cd back/orders && go build ./... && go vet ./... && go test ./...
```

Ожидаемо: чисто (при условии, что Step 12 выполнен).

- [ ] **Step 14: Commit**

```bash
git add back/orders
git commit -m "feat(orders): add GetOrderStatus admin endpoint + tg client codegen"
```

---

### Task 4: orders — `CreateOrder` держит транзакцию через колбэк

**Files:**
- Modify: `back/orders/internal/storage/postgres/orders.go`
- Test: `back/orders/internal/storage/postgres/integration_test.go`

**Interfaces:**
- Produces: новая сигнатура
  ```go
  func (s *Storage) CreateOrder(
      ctx context.Context,
      params CreateOrderParams,
      pushToOnec func(ctx context.Context, orderID uuid.UUID, orderNumber string) (onecGUID uuid.UUID, onecNumber string, err error),
  ) (CreatedOrder, error)
  ```
  `CreatedOrder` получает новое поле `Status string`.
- Consumes: существующий `CreateOrderParams`, существующие таблицы `orders`/`order_items`/`order_status_history`/`cart_items`.

Это меняет сигнатуру существующего метода — единственный текущий вызыватель (`service.go:510`) будет сломан до Task 5; это ожидаемо, чинится в следующей задаче. Не коммитить эту задачу отдельно, если хочется зелёного `go build` на каждом коммите — тогда объединить Steps этой и следующей задачи в один коммит (см. итоговый Step 5 ниже, который это учитывает).

- [ ] **Step 1: Написать падающие тесты (успех и откат)**

В `back/orders/internal/storage/postgres/integration_test.go`:

```go
func TestCreateOrder_PushSucceeds_StatusProcessingAndOnecFieldsSet(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	storage := New(pool)

	cartID := mustInsertEmptyCart(t, ctx, pool) // см. Step 1a ниже
	onecGUID := uuid.New()

	created, err := storage.CreateOrder(ctx, CreateOrderParams{
		CartID: cartID,
		Items:  nil,
	}, func(ctx context.Context, orderID uuid.UUID, orderNumber string) (uuid.UUID, string, error) {
		if orderNumber == "" {
			t.Fatalf("expected non-empty order number in callback")
		}
		return onecGUID, "УТ-00001", nil
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if created.Status != "processing" {
		t.Fatalf("expected status processing, got %q", created.Status)
	}

	var status string
	var storedGUID uuid.UUID
	if err = pool.QueryRow(ctx, `SELECT status, one_c_guid FROM orders WHERE id = $1`, created.ID).Scan(&status, &storedGUID); err != nil {
		t.Fatalf("read back order: %v", err)
	}
	if status != "processing" || storedGUID != onecGUID {
		t.Fatalf("expected processing/%s, got %s/%s", onecGUID, status, storedGUID)
	}

	var historyCount int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM order_status_history WHERE order_id = $1`, created.ID).Scan(&historyCount); err != nil {
		t.Fatalf("count history: %v", err)
	}
	if historyCount != 2 {
		t.Fatalf("expected 2 history rows (new, processing), got %d", historyCount)
	}
}

func TestCreateOrder_PushFails_WholeTransactionRolledBack(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	storage := New(pool)

	cartID := mustInsertEmptyCart(t, ctx, pool)
	var cartItemsBefore int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM cart_items WHERE cart_id = $1`, cartID).Scan(&cartItemsBefore); err != nil {
		t.Fatalf("count cart items before: %v", err)
	}

	var ordersCountBefore int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM orders`).Scan(&ordersCountBefore); err != nil {
		t.Fatalf("count orders before: %v", err)
	}

	pushErr := errors.New("onec unavailable")
	_, err = storage.CreateOrder(ctx, CreateOrderParams{CartID: cartID}, func(ctx context.Context, orderID uuid.UUID, orderNumber string) (uuid.UUID, string, error) {
		return uuid.Nil, "", pushErr
	})
	if !errors.Is(err, pushErr) {
		t.Fatalf("expected wrapped pushErr, got %v", err)
	}

	var ordersCountAfter int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM orders`).Scan(&ordersCountAfter); err != nil {
		t.Fatalf("count orders after: %v", err)
	}
	if ordersCountAfter != ordersCountBefore {
		t.Fatalf("expected no new order row after rollback, before=%d after=%d", ordersCountBefore, ordersCountAfter)
	}

	var cartItemsAfter int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM cart_items WHERE cart_id = $1`, cartID).Scan(&cartItemsAfter); err != nil {
		t.Fatalf("count cart items after: %v", err)
	}
	if cartItemsAfter != cartItemsBefore {
		t.Fatalf("expected cart_items untouched after rollback, before=%d after=%d", cartItemsBefore, cartItemsAfter)
	}
}

// mustInsertEmptyCart вставляет пустую корзину без привязки к пользователю/контрагенту
// (CreateOrderParams.UserID/CounterpartyID в этих тестах нулевые — проверяется только
// транзакционное поведение вокруг pushToOnec, не бизнес-валидация полей).
func mustInsertEmptyCart(t *testing.T, ctx context.Context, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	var cartID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO carts DEFAULT VALUES RETURNING id`).Scan(&cartID); err != nil {
		t.Fatalf("insert cart: %v", err)
	}
	return cartID
}
```

Перед вставкой сверить реальную схему `carts` (`back/migrations/pkg/migrations/data/20260705171942_cart.sql`) — если `INSERT INTO carts DEFAULT VALUES` не проходит (например есть `NOT NULL` без дефолта), заменить на `INSERT INTO carts (user_id) VALUES (NULL) RETURNING id` или что там реально требуется; посмотреть на существующий тест создания корзины в этом же файле (`TestStorageIntegration`-подобные в orders) для рабочего примера INSERT.

- [ ] **Step 2: Прогнать, убедиться что падает (компиляция — старая сигнатура)**

```bash
cd back/orders && TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run TestCreateOrder_Push -v
```

- [ ] **Step 3: Переписать `CreateOrder`**

В `back/orders/internal/storage/postgres/orders.go`, заменить существующую функцию (строки ~45-96, см. текущий код) на:

```go
type CreatedOrder struct {
	ID        uuid.UUID
	Number    string
	Status    string
	CreatedAt time.Time
}

func (s *Storage) CreateOrder(
	ctx context.Context,
	params CreateOrderParams,
	pushToOnec func(ctx context.Context, orderID uuid.UUID, orderNumber string) (onecGUID uuid.UUID, onecNumber string, err error),
) (CreatedOrder, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var order CreatedOrder
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (
			number, counterparty_id, user_id, status, payment_status,
			delivery_method, delivery_address_id, contact_id, comment,
			subtotal, discount_total, vat_total, total
		) VALUES (
			'AMC-' || to_char(now(), 'YYYYMMDD') || '-' || lpad(nextval('orders_number_seq')::text, 5, '0'),
			$1, $2, 'new', 'not_paid', $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id, number, created_at
	`,
		params.CounterpartyID, params.UserID, params.DeliveryMethod, params.DeliveryAddressID, params.ContactID, params.Comment,
		params.Subtotal, params.DiscountTotal, params.VATTotal, params.Total,
	).Scan(&order.ID, &order.Number, &order.CreatedAt)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("insert order: %w", err)
	}

	for _, item := range params.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, sku, name, quantity, unit_price, discount_percent, vat_rate, line_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, order.ID, item.ProductID, item.SKU, item.Name, item.Quantity, item.UnitPrice, item.DiscountPercent, item.VATRate, item.LineTotal)
		if err != nil {
			return CreatedOrder{}, fmt.Errorf("insert order item: %w", err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, payment_status, changed_by, comment)
		VALUES ($1, NULL, 'new', 'not_paid', $2, $3)
	`, order.ID, params.UserID, params.Comment)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("insert order status history: %w", err)
	}

	if _, err = tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, params.CartID); err != nil {
		return CreatedOrder{}, fmt.Errorf("clear cart after order: %w", err)
	}

	onecGUID, onecNumber, pushErr := pushToOnec(ctx, order.ID, order.Number)
	if pushErr != nil {
		return CreatedOrder{}, fmt.Errorf("push order to onec: %w", pushErr)
	}

	if _, err = tx.Exec(ctx, `
		UPDATE orders SET status = 'processing', one_c_guid = $1, synced_to_1c_at = now() WHERE id = $2
	`, onecGUID, order.ID); err != nil {
		return CreatedOrder{}, fmt.Errorf("mark order synced: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO order_status_history (order_id, old_status, new_status, payment_status, changed_by, comment)
		VALUES ($1, 'new', 'processing', 'not_paid', NULL, $2)
	`, order.ID, "Отправлен в 1С, документ "+onecNumber); err != nil {
		return CreatedOrder{}, fmt.Errorf("insert processing history: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return CreatedOrder{}, fmt.Errorf("commit tx: %w", err)
	}
	order.Status = "processing"
	return order, nil
}
```

Сверить, что старая версия функции не возвращала `order` без `Commit` где-то ниже (посмотреть текущий хвост файла после строки 95 — там раньше было `return order, nil` после commit, теперь заменяется на `order.Status = "processing"; return order, nil`, как выше).

- [ ] **Step 4: Прогнать тесты**

```bash
cd back/orders && TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run 'TestCreateOrder_Push' -v
```

Ожидаемо: оба `PASS`. (`go build ./...` всего модуля в этот момент ещё красный — `service.go` вызывает `CreateOrder` со старой сигнатурой; чинится в Task 5.)

- [ ] **Step 5: Commit — только вместе с Task 5**

Не коммитить здесь отдельно (модуль не собирается). Коммит — в конце Task 5.

---

### Task 5: orders — сервис собирает payload и вызывает колбэк

**Files:**
- Create: `back/orders/internal/service/onecpusher.go`
- Modify: `back/orders/internal/service/service.go`
- Modify: `back/orders/internal/config/config.go`
- Modify: `back/orders/cmd/main.go`
- Test: `back/orders/internal/service/create_order_onec_test.go`

**Interfaces:**
- Consumes: `postgres.GetProductOnecRefs`, `postgres.GetCounterpartyOnecRef` (Task 2), `postgres.CreateOrder` новая сигнатура (Task 4).
- Produces:
  ```go
  // back/orders/internal/service/onecpusher.go
  type OnecOrderItem struct {
      OneCGUID uuid.NullUUID
      SKU      string
      Name     string
      Qty      int
      Price    float64
      VATRate  float64
  }
  type OnecPushOrder struct {
      ClientOrderID     uuid.UUID
      OrderNumber       string
      CounterpartyGUID  uuid.NullUUID
      CounterpartyINN   string
      CounterpartyName  string
      DeliveryType      string
      DeliveryAddress   string
      ContactName       string
      Phone             string
      Email             string
      Comment           string
      Items             []OnecOrderItem
  }
  // OnecPusher — implemented by internal/onecclient.Client (Task 15).
  type OnecPusher interface {
      PushOrder(ctx context.Context, order OnecPushOrder) (onecGUID uuid.UUID, onecNumber string, err error)
  }
  ```
  `service` получает новое поле `onecPusher OnecPusher`, `NewOrdersApiService` — новый параметр.

- [ ] **Step 1: Написать `onecpusher.go`**

```go
package service

import (
	"context"

	"github.com/google/uuid"
)

type OnecOrderItem struct {
	OneCGUID uuid.NullUUID
	SKU      string
	Name     string
	Qty      int
	Price    float64
	VATRate  float64
}

type OnecPushOrder struct {
	ClientOrderID    uuid.UUID
	OrderNumber      string
	CounterpartyGUID uuid.NullUUID
	CounterpartyINN  string
	CounterpartyName string
	DeliveryType     string
	DeliveryAddress  string
	ContactName      string
	Phone            string
	Email            string
	Comment          string
	Items            []OnecOrderItem
}

// OnecPusher is implemented by internal/onecclient.Client.
type OnecPusher interface {
	PushOrder(ctx context.Context, order OnecPushOrder) (onecGUID uuid.UUID, onecNumber string, err error)
}
```

- [ ] **Step 2: Написать падающий тест сервиса**

Создать `back/orders/internal/service/create_order_onec_test.go`. Использовать существующий фейковый `Storage` (`clientResolutionStorage`) — добавить в него минимальную поддержку `CreateOrder`/`GetProductOnecRefs`/`GetCounterpartyOnecRef`/`GetCartItems`/`GetOrCreateCart` (проверить, какие из них там уже есть — файл уже используется тестами `TestCartAndOrdersReachRepositoryWithValidClient`, значит `CreateOrder` там уже как-то замокан под старую сигнатуру; обновить мок под новую сигнатуру с колбэком):

```go
type fakeOnecPusher struct {
	called    bool
	lastOrder OnecPushOrder
	guid      uuid.UUID
	number    string
	err       error
}

func (f *fakeOnecPusher) PushOrder(ctx context.Context, order OnecPushOrder) (uuid.UUID, string, error) {
	f.called = true
	f.lastOrder = order
	return f.guid, f.number, f.err
}

func TestCreateOrder_CallsOnecPusherWithItemsAndReturnsProcessingStatus(t *testing.T) {
	storage := &clientResolutionStorage{
		hasActiveClient: true,
		cartItems: []postgres.CartItemRow{
			{ProductID: uuid.MustParse("11111111-1111-1111-1111-111111111111"), SKU: "SKU-1", ProductName: "Товар 1", Qty: 2, Price: 100},
		},
		productOnecRefs: map[uuid.UUID]postgres.ProductOnecRef{
			uuid.MustParse("11111111-1111-1111-1111-111111111111"): {SKU: "SKU-1"},
		},
	}
	pusher := &fakeOnecPusher{guid: uuid.New(), number: "УТ-00099"}
	svc := NewOrdersApiService(zerolog.Nop(), storage, &fakeAccessClient{allowed: true}, 20, pusher)

	resp, err := svc.CreateOrder(context.Background(), uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", "коммент")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !pusher.called {
		t.Fatal("expected OnecPusher.PushOrder to be called")
	}
	if len(pusher.lastOrder.Items) != 1 || pusher.lastOrder.Items[0].SKU != "SKU-1" || pusher.lastOrder.Items[0].Qty != 2 {
		t.Fatalf("unexpected items passed to pusher: %+v", pusher.lastOrder.Items)
	}
	if resp.Order.Status != "processing" {
		t.Fatalf("expected order status processing, got %q", resp.Order.Status)
	}
}

func TestCreateOrder_OnecPushFails_ReturnsError(t *testing.T) {
	storage := &clientResolutionStorage{
		hasActiveClient: true,
		cartItems: []postgres.CartItemRow{
			{ProductID: uuid.New(), SKU: "SKU-2", ProductName: "Товар 2", Qty: 1, Price: 50},
		},
	}
	pusher := &fakeOnecPusher{err: errors.New("onec down")}
	svc := NewOrdersApiService(zerolog.Nop(), storage, &fakeAccessClient{allowed: true}, 20, pusher)

	_, err := svc.CreateOrder(context.Background(), uuid.New(), "", "delivery", "Адрес", "Иван", "+7900", "a@b.c", "")
	if err == nil {
		t.Fatal("expected error when onec push fails")
	}
}
```

Поля `hasActiveClient`, `cartItems`, `productOnecRefs` — свериться с реальными именами полей `clientResolutionStorage` (файл `client_resolution_test.go`) перед написанием, использовать то, что там уже заведено для похожих тестов (`TestCartAndOrdersReachRepositoryWithValidClient` уже гоняет `CreateOrder` через этот фейк — скопировать оттуда способ настройки корзины/клиента). Добавить в `clientResolutionStorage`:

```go
productOnecRefs map[uuid.UUID]postgres.ProductOnecRef
counterpartyOnecRef postgres.CounterpartyOnecRef

func (s *clientResolutionStorage) GetProductOnecRefs(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]postgres.ProductOnecRef, error) {
	return s.productOnecRefs, nil
}
func (s *clientResolutionStorage) GetCounterpartyOnecRef(ctx context.Context, counterpartyID uuid.UUID) (postgres.CounterpartyOnecRef, error) {
	return s.counterpartyOnecRef, nil
}
```

И обновить существующий фейковый `CreateOrder` в том же файле под новую сигнатуру — он должен вызвать переданный колбэк с каким-нибудь orderID/number и учитывать возвращённую ошибку:

```go
func (s *clientResolutionStorage) CreateOrder(ctx context.Context, params postgres.CreateOrderParams, pushToOnec func(context.Context, uuid.UUID, string) (uuid.UUID, string, error)) (postgres.CreatedOrder, error) {
	orderID := uuid.New()
	_, _, err := pushToOnec(ctx, orderID, "TEST-0001")
	if err != nil {
		return postgres.CreatedOrder{}, err
	}
	return postgres.CreatedOrder{ID: orderID, Number: "TEST-0001", Status: "processing"}, nil
}
```

- [ ] **Step 3: Прогнать, убедиться что падает (компиляция)**

```bash
cd back/orders && go test ./internal/service/... -run TestCreateOrder_ -v
```

- [ ] **Step 4: Обновить `service.go`**

Добавить поле и параметр конструктора:

```go
type service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
	vatRate      float64
	onecPusher   OnecPusher
}

func NewOrdersApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient, vatRate float64, onecPusher OnecPusher) externalapi.OrdersAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
		vatRate:      vatRate,
		onecPusher:   onecPusher,
	}
}
```

Добавить в интерфейс `Storage` этого файла:
```go
	GetProductOnecRefs(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]postgres.ProductOnecRef, error)
	GetCounterpartyOnecRef(ctx context.Context, counterpartyID uuid.UUID) (postgres.CounterpartyOnecRef, error)
```
и поменять сигнатуру `CreateOrder` в этом интерфейсе на новую (с колбэком).

В теле `CreateOrder` (строки ~486-527 сегодня) — после построения `orderItems`/`responseItems` и ПЕРЕД вызовом `s.storage.CreateOrder`, подготовить payload и заменить сам вызов:

```go
	productIDs := make([]uuid.UUID, 0, len(cartRows))
	for _, row := range cartRows {
		productIDs = append(productIDs, row.ProductID)
	}
	productRefs, err := s.storage.GetProductOnecRefs(ctx, productIDs)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}
	var counterpartyRef postgres.CounterpartyOnecRef
	if counterpartyID.Valid {
		counterpartyRef, err = s.storage.GetCounterpartyOnecRef(ctx, counterpartyID.UUID)
		if err != nil {
			return response, customErrors.InternalServerError().SetOuterError(err)
		}
	}

	pushItems := make([]OnecOrderItem, 0, len(cartRows))
	for _, row := range cartRows {
		ref := productRefs[row.ProductID]
		pushItems = append(pushItems, OnecOrderItem{
			OneCGUID: ref.OneCGUID,
			SKU:      row.SKU,
			Name:     row.ProductName,
			Qty:      row.Qty,
			Price:    row.Price,
			VATRate:  s.vatRate,
		})
	}

	created, err := s.storage.CreateOrder(ctx, postgres.CreateOrderParams{
		UserID:            userID,
		CounterpartyID:    counterpartyID,
		CartID:            cartID,
		DeliveryMethod:    deliveryType,
		DeliveryAddressID: addressID,
		ContactID:         contactID,
		Comment:           comment,
		Subtotal:          totals.Subtotal,
		DiscountTotal:     totals.DiscountTotal,
		VATTotal:          totals.VATTotal,
		Total:             totals.Total,
		Items:             orderItems,
	}, func(ctx context.Context, orderID uuid.UUID, orderNumber string) (uuid.UUID, string, error) {
		return s.onecPusher.PushOrder(ctx, OnecPushOrder{
			ClientOrderID:    orderID,
			OrderNumber:      orderNumber,
			CounterpartyGUID: counterpartyRef.OneCGUID,
			CounterpartyINN:  counterpartyRef.INN,
			CounterpartyName: counterpartyRef.Name,
			DeliveryType:     deliveryType,
			DeliveryAddress:  deliveryAddress,
			ContactName:      contactName,
			Phone:            phone,
			Email:            email,
			Comment:          comment,
			Items:            pushItems,
		})
	})
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}
```

И поменять построение `order := models.Order{...}` ниже — заменить захардкоженное `Status: "new"` на `Status: created.Status`.

- [ ] **Step 5: Обновить `cmd/main.go`**

Добавить в `back/orders/internal/config/config.go` (в `Config` и `LoadConfig`):
```go
	IntegrationsURL string
```
```go
	IntegrationsURL: os.Getenv("INTEGRATIONS_URL"),
```
(без fatal — если не задан, `internal/onecclient` вернёт ошибку соединения при первом реальном вызове; поле не обязательное на этапе Task 5, обязательным делаем деплоем в Task 17).

В `back/orders/cmd/main.go`, добавить создание `onecPusher` перед `ordersService.NewOrdersApiService(...)` и передать его четвёртым аргументом. Конкретный тип (`onecclient.New(cfg.IntegrationsURL)`) появится в Task 15 — до тех пор временно передать `nil` явно с комментарием `// TODO(Task 15): заменить на internal/onecclient.New(cfg.IntegrationsURL)`, чтобы модуль собирался. Если это неприемлемо (не хочется коммитить `nil`), выполнить Task 15 сразу вслед за этой задачей до коммита — на усмотрение исполнителя, `go build` должен быть зелёным на каждом коммите.

- [ ] **Step 6: Прогнать тесты**

```bash
cd back/orders && go test ./internal/service/... -v
cd back/orders && go build ./... && go vet ./...
```

Ожидаемо: `PASS` на новых тестах, `go build`/`go vet` чисто (если Step 12 из Task 3 уже выполнена пользователем — иначе `go build` упадёт на отсутствующем `GetOrderStatus` в сгенерированном транспорте; это ожидаемо и не относится к этой задаче).

- [ ] **Step 7: Commit (объединяет Task 4 + Task 5)**

```bash
git add back/orders
git commit -m "feat(orders): push order to 1С synchronously inside CreateOrder tx"
```

---

### Task 6: integrations — errors-пакет

**Files:**
- Create: `back/integrations/internal/errors/errors.go`
- Create: `back/integrations/internal/errors/common.go`

**Interfaces:**
- Produces: `customErrors.NotFoundError()`, `BadRequestError()`, `ConflictError()`, `UnauthorizedError()`, `InternalServerError()` — все `*Error` с методом `Code() int`, по образцу `back/orders/internal/errors`.

- [ ] **Step 1: Скопировать и адаптировать `errors.go`**

Взять `back/orders/internal/errors/errors.go` целиком, заменить только пакетные пути в комментариях, если есть (сам файл не ссылается на другие пакеты orders — чистый). Положить в `back/integrations/internal/errors/errors.go` без изменений содержимого (тип `Error`, `New`, `Code()` и т.д. — универсальны).

- [ ] **Step 2: Написать `common.go` под integrations**

```go
package errors

import (
	"github.com/valyala/fasthttp"
)

var (
	BadRequestError = func() *Error { return New("bad request", fasthttp.StatusBadRequest, ErrBadRequest) }
	UnauthorizedError = func() *Error { return New("unauthorized", fasthttp.StatusUnauthorized, ErrUnauthorized) }
	NotFoundError = func() *Error { return New("not found", fasthttp.StatusNotFound, ErrNotFound) }
	ConflictError = func() *Error { return New("conflict", fasthttp.StatusConflict, ErrConflict) }
	InternalServerError = func() *Error {
		return New("internal server error", fasthttp.StatusInternalServerError, ErrInternal)
	}
)

const (
	ErrInternal    = "integrations.errors.internalError"
	ErrBadRequest  = "integrations.errors.badRequest"
	ErrUnauthorized = "integrations.errors.unauthorized"
	ErrNotFound    = "integrations.errors.notFound"
	ErrConflict    = "integrations.errors.conflict"
)
```

- [ ] **Step 3: Проверить компиляцию**

```bash
cd back/integrations && go build ./internal/errors/... && go vet ./internal/errors/...
```

Ожидаемо: чисто.

- [ ] **Step 4: Commit**

```bash
git add back/integrations/internal/errors
git commit -m "feat(integrations): add errors package for onec-orders-api"
```

---

### Task 7: integrations — интерфейс вебхука (только `@tg`-аннотации)

**Files:**
- Create: `back/integrations/pkg/interfaces/internalAPI/interface.go`
- Create: `back/integrations/pkg/models/requests.go` (если понадобится доп. модель ответа — иначе не создавать, см. Step 1)

**Interfaces:**
- Produces: интерфейс `OnecOrdersAPI` с одним методом `OnecOrderStatusWebhook`. Custom-handler и service-реализация — Task 8. **Кодогенерацию (`tg transport`) для этого файла запускает пользователь** — не входит в эту задачу.

- [ ] **Step 1: Написать интерфейс**

```go
// Package internalapi describes the public 1С→AMC order-status webhook.
// @tg version=0.0.1
// @tg backend=integrations
// @tg title=`onec-orders-api`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/internalapi --outSwagger ../../../swaggers/internalapi/swagger.yaml
package internalAPI

import (
	"context"

	"github.com/google/uuid"
)

// OnecOrdersAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err401
// @tg 404=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err404
// @tg 409=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err409
// @tg 500=github.com/mbatimel/AMC/integrations/swaggers/internalapi/models:Err500
type OnecOrdersAPI interface {
	// OnecOrderStatusWebhook is called by 1С when an order reaches a new
	// stage it needs to report (v1: only "delivered").
	// @tg http-method=POST
	// @tg http-path=/v1/onec/orders/status
	// @tg http-headers=apiKey|X-Onec-Api-Key
	// @tg http-args=clientOrderID|clientOrderID
	// @tg http-args=status|status
	// @tg http-args=onecDocumentNumber|onecDocumentNumber
	// @tg http-args=comment|comment
	// @tg http-response=github.com/mbatimel/AMC/integrations/internal/transport/custom-handlers:OnecOrderStatusWebhook
	// @tg summary=`Статус заказа от 1С`
	// @tg desc=`Приём вебхука 1С о смене стадии заказа; в v1 допустим только status=delivered`
	// @tg uuidPackage=github.com/google/uuid
	OnecOrderStatusWebhook(ctx context.Context, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) (ok bool, err error)
}
```

Сверить точный вид `@tg NNN=...:Errors/Resp` строк и путь для `@tg http-response=...` с реальным существующим примером — скопировать блок `@tg 200=.../400=.../401=.../403=.../500=...` из `back/orders/pkg/interfaces/externalapi/interface.go:19-23` и адаптировать пакетные пути на `integrations`; модели ошибок (`swaggers/internalapi/models`) в этом репо генерируются/пишутся так же, как `back/orders/swaggers/externalapi/models/errors.go` — если у пользователя `tg swagger`/`tg transport` сам создаёт эти модели, отдельная задача на их ручное написание не нужна; если нет — завести `back/integrations/swaggers/internalapi/models/errors.go` и `responses.go` по образцу `back/orders/swaggers/externalapi/models/{errors,responses}.go` (скопировать и адаптировать пакетные имена) до запуска кодогенерации. Явно спросить пользователя, если после первого `tg transport` компиляция падает на отсутствии этих моделей — не изобретать формат самостоятельно, сверить с фактической структурой `orders/swaggers/externalapi/models/errors.go`.

- [ ] **Step 2: Проверить, что файл валиден по Go-синтаксису (без кодогенерации)**

```bash
cd back/integrations && gofmt -l pkg/interfaces/internalAPI/interface.go
```

Ожидаемо: пусто (нет ошибок форматирования). Полная компиляция пакета в этот момент невозможна — интерфейс ничего не реализует, это нормально.

- [ ] **Step 3: Commit**

```bash
git add back/integrations/pkg/interfaces/internalAPI/interface.go
git commit -m "feat(integrations): declare OnecOrdersAPI webhook interface"
```

---

### Task 8: integrations — сервис вебхука (allow-list, идемпотентность, 409)

**Files:**
- Create: `back/integrations/internal/service/onecorders.go`
- Test: `back/integrations/internal/service/onecorders_test.go`
- Create: `back/integrations/internal/transport/custom-handlers/onecorders.go`
- Modify: `back/integrations/go.mod` (require + replace на `../orders`, чтобы дёргать `orders.GetOrderStatus`/`UpdateOrderStatus` сгенерированным клиентом из Task 3)

**Interfaces:**
- Consumes: `orders/pkg/client/transport.NewClientOrdersAPI(endpoint string) *ClientOrdersAPI` (сгенерирован в Task 3, метод `GetOrderStatus(ctx, userID, orderID) (status string, err error)`, `UpdateOrderStatus(ctx, userID, orderID, status, paymentStatus, comment, changedBy string) (order ..., err error)` — **проверить фактическую сигнатуру сгенерированного клиента после Task 3, она зеркалит `externalapi.OrdersAPI`, но возвращаемый тип у `UpdateOrderStatus` — не `models.Order`, а сгенерированный ответ; посмотреть реальный файл `orders/pkg/client/transport/ordersapi-http-client.go` после генерации и подставить точную сигнатуру**).
- Produces:
  ```go
  type OrdersClient interface {
      GetOrderStatus(ctx context.Context, userID, orderID uuid.UUID) (status string, err error)
      UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status, paymentStatus, comment, changedBy string) error
  }
  type OnecOrdersService struct { ... }
  func NewOnecOrdersService(logger zerolog.Logger, ordersClient OrdersClient, systemUserID uuid.UUID, apiKey string, storage OnecOrdersStorage) *OnecOrdersService
  func (s *OnecOrdersService) OnecOrderStatusWebhook(ctx context.Context, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) (ok bool, err error)
  ```

Обёртка `UpdateOrderStatus(...) error` в `OrdersClient` (без возврата заказа — вебхуку тело ответа не нужно) — реализация в Task 14 адаптирует под реальную сгенерированную сигнатуру.

- [ ] **Step 1: Написать падающие тесты**

```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type fakeOrdersClient struct {
	status        string
	statusErr     error
	updateErr     error
	updateCalled  bool
	lastUserID    uuid.UUID
}

func (f *fakeOrdersClient) GetOrderStatus(ctx context.Context, userID, orderID uuid.UUID) (string, error) {
	return f.status, f.statusErr
}
func (f *fakeOrdersClient) UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status, paymentStatus, comment, changedBy string) error {
	f.updateCalled = true
	f.lastUserID = userID
	return f.updateErr
}

const testSystemUserID = "00000000-0000-0000-0000-0000000a0ec1"
const testAPIKey = "test-api-key"

func TestOnecOrderStatusWebhook_WrongAPIKey_Unauthorized(t *testing.T) {
	client := &fakeOrdersClient{status: "processing"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), "wrong-key", uuid.New(), "delivered", "УТ-1", "")
	if err == nil {
		t.Fatal("expected error for wrong api key")
	}
}

func TestOnecOrderStatusWebhook_InvalidStatus_BadRequest(t *testing.T) {
	client := &fakeOrdersClient{status: "processing"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "cancelled", "УТ-1", "")
	if err == nil {
		t.Fatal("expected error for disallowed status")
	}
	if client.updateCalled {
		t.Fatal("expected UpdateOrderStatus not to be called for invalid status")
	}
}

func TestOnecOrderStatusWebhook_OrderCancelled_Conflict(t *testing.T) {
	client := &fakeOrdersClient{status: "cancelled"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "")
	if err == nil {
		t.Fatal("expected conflict error for cancelled order")
	}
	if client.updateCalled {
		t.Fatal("expected UpdateOrderStatus not to be called for cancelled order")
	}
}

func TestOnecOrderStatusWebhook_AlreadyDelivered_IdempotentNoOp(t *testing.T) {
	client := &fakeOrdersClient{status: "delivered"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	ok, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if client.updateCalled {
		t.Fatal("expected UpdateOrderStatus not to be called when already delivered")
	}
}

func TestOnecOrderStatusWebhook_Processing_AppliesTransition(t *testing.T) {
	client := &fakeOrdersClient{status: "processing"}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	ok, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "Вручено")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if !client.updateCalled {
		t.Fatal("expected UpdateOrderStatus to be called")
	}
	if client.lastUserID.String() != testSystemUserID {
		t.Fatalf("expected system user id %s, got %s", testSystemUserID, client.lastUserID)
	}
}

func TestOnecOrderStatusWebhook_OrderNotFound(t *testing.T) {
	client := &fakeOrdersClient{statusErr: errors.New("order not found: 404")}
	svc := NewOnecOrdersService(zerolog.Nop(), client, uuid.MustParse(testSystemUserID), testAPIKey)

	_, err := svc.OnecOrderStatusWebhook(context.Background(), testAPIKey, uuid.New(), "delivered", "УТ-1", "")
	if err == nil {
		t.Fatal("expected error when order status lookup fails")
	}
}
```

- [ ] **Step 2: Прогнать, убедиться что падает**

```bash
cd back/integrations && go test ./internal/service/... -run TestOnecOrderStatusWebhook -v
```

- [ ] **Step 3: Реализовать**

```go
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/integrations/internal/errors"
)

type OrdersClient interface {
	GetOrderStatus(ctx context.Context, userID, orderID uuid.UUID) (status string, err error)
	UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status, paymentStatus, comment, changedBy string) error
}

// allowedWebhookStatuses — единственный допустимый статус от 1С в v1.
// Расширять только через этот список + docs/superpowers/specs/2026-08-26-onec-orders-integration-1c-contract.md.
var allowedWebhookStatuses = map[string]bool{
	"delivered": true,
}

type OnecOrdersService struct {
	logger       zerolog.Logger
	ordersClient OrdersClient
	systemUserID uuid.UUID
	apiKey       string
}

func NewOnecOrdersService(logger zerolog.Logger, ordersClient OrdersClient, systemUserID uuid.UUID, apiKey string) *OnecOrdersService {
	return &OnecOrdersService{logger: logger, ordersClient: ordersClient, systemUserID: systemUserID, apiKey: apiKey}
}

func (s *OnecOrdersService) OnecOrderStatusWebhook(ctx context.Context, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) (ok bool, err error) {
	if apiKey != s.apiKey {
		return false, customErrors.UnauthorizedError()
	}
	if !allowedWebhookStatuses[status] {
		return false, customErrors.BadRequestError().AddCause("field", "status")
	}

	currentStatus, err := s.ordersClient.GetOrderStatus(ctx, s.systemUserID, clientOrderID)
	if err != nil {
		return false, err
	}

	switch currentStatus {
	case status:
		return true, nil
	case "cancelled":
		return false, customErrors.ConflictError().AddCause("field", "status")
	}

	fullComment := comment
	if onecDocumentNumber != "" {
		fullComment = "1С документ " + onecDocumentNumber + ": " + comment
	}
	if err = s.ordersClient.UpdateOrderStatus(ctx, s.systemUserID, clientOrderID, status, "", fullComment, ""); err != nil {
		return false, err
	}
	return true, nil
}

var _ = errors.New // избежать неиспользуемого импорта, если он не понадобится после правок выше
```

Убрать последнюю строку (`var _ = errors.New`) и импорт `"errors"`, если он не используется — оставлена как явное напоминание проверить импорты перед коммитом, не как код для продакшена.

- [ ] **Step 4: Прогнать тесты**

```bash
cd back/integrations && go test ./internal/service/... -run TestOnecOrderStatusWebhook -v
```

Ожидаемо: все `PASS`.

- [ ] **Step 5: Написать custom-handler**

```go
package custom_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	internalapi "github.com/mbatimel/AMC/integrations/pkg/interfaces/internalAPI"
)

func OnecOrderStatusWebhook(ctx *fiber.Ctx, svc internalapi.OnecOrdersAPI, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) error {
	ok, err := svc.OnecOrderStatusWebhook(ctx.UserContext(), apiKey, clientOrderID, status, onecDocumentNumber, comment)
	if err != nil {
		return sendError(ctx, err)
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"ok": ok})
}
```

Сверить точное имя/сигнатуру существующего хелпера отправки ошибок (`sendResponse`/`sendError` — см. `back/orders/internal/transport/custom-handlers/response.go`) — этого файла в `integrations` ещё нет, создать `back/integrations/internal/transport/custom-handlers/response.go` по образцу `back/orders/internal/transport/custom-handlers/response.go` (скопировать, адаптировать пакетные пути на `integrations/internal/errors`).

- [ ] **Step 6: `go.mod` — добавить зависимость на orders**

```bash
cd back/integrations
go mod edit -require=github.com/mbatimel/AMC/orders@v0.0.0-00010101000000-000000000000
go mod edit -replace=github.com/mbatimel/AMC/orders=../orders
go mod tidy
```

⚠ Это сработает только после того, как в Task 3 Step 12 сгенерирован `back/orders/pkg/client/transport` (иначе `go mod tidy` не найдёт нужный пакет при первой сборке, использующей его — сама зависимость в `go.mod` добавится, ошибка появится только там, где реально импортируется несуществующий пакет; отложить фактический импорт в service до Task 14, где он появляется).

- [ ] **Step 7: Собрать пакет**

```bash
cd back/integrations && go build ./internal/service/... ./internal/transport/custom-handlers/... && go vet ./internal/service/... && go test ./internal/service/... -v
```

- [ ] **Step 8: Commit**

```bash
git add back/integrations/internal/service/onecorders.go back/integrations/internal/service/onecorders_test.go back/integrations/internal/transport/custom-handlers back/integrations/go.mod back/integrations/go.sum
git commit -m "feat(integrations): webhook service — status allow-list, idempotency, 409 on cancelled"
```

---

### Task 9: integrations — исходящий клиент к 1С (`PushOrder` target)

**Files:**
- Create: `back/integrations/internal/onecorders/models.go`
- Create: `back/integrations/internal/onecorders/client.go`
- Test: `back/integrations/internal/onecorders/client_test.go`

**Interfaces:**
- Produces:
  ```go
  type PushOrderRequest struct { ClientOrderID uuid.UUID; OrderNumber string; Counterparty CounterpartyDTO; Delivery DeliveryDTO; Comment string; Items []ItemDTO }
  type PushOrderResult struct { OnecDocumentGUID uuid.UUID; OnecDocumentNumber string }
  func New(baseURL, user, password string, timeout time.Duration, logger zerolog.Logger) *Client
  func (c *Client) PushOrder(ctx context.Context, req PushOrderRequest) (PushOrderResult, error)
  ```
  Формат JSON — 1:1 с `docs/superpowers/specs/2026-08-26-onec-orders-integration-1c-contract.md` (раздел "Ручка №1").

- [ ] **Step 1: Написать модели**

```go
package onecorders

import "github.com/google/uuid"

type CounterpartyDTO struct {
	OneCGUID string `json:"onec_guid"`
	INN      string `json:"inn"`
	Name     string `json:"name"`
}

type DeliveryDTO struct {
	Type        string `json:"type"`
	Address     string `json:"address"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}

type ItemDTO struct {
	OneCGUID string  `json:"onec_guid"`
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Qty      int     `json:"qty"`
	Price    float64 `json:"price"`
	VATRate  float64 `json:"vat_rate"`
}

type PushOrderRequest struct {
	ClientOrderID uuid.UUID   `json:"client_order_id"`
	OrderNumber   string      `json:"order_number"`
	Counterparty  CounterpartyDTO `json:"counterparty"`
	Delivery      DeliveryDTO `json:"delivery"`
	Comment       string      `json:"comment"`
	Items         []ItemDTO   `json:"items"`
}

type pushOrderSuccessResponse struct {
	OnecDocumentGUID   uuid.UUID `json:"onec_document_guid"`
	OnecDocumentNumber string    `json:"onec_document_number"`
}

type PushOrderResult struct {
	OnecDocumentGUID   uuid.UUID
	OnecDocumentNumber string
}
```

- [ ] **Step 2: Написать падающий тест клиента**

```go
package onecorders

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestPushOrder_Success(t *testing.T) {
	guid := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hs/amc-integration/orders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var req PushOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(req.Items) != 1 || req.Items[0].SKU != "SKU-1" {
			t.Fatalf("unexpected items: %+v", req.Items)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pushOrderSuccessResponse{OnecDocumentGUID: guid, OnecDocumentNumber: "УТ-00042"})
	}))
	defer srv.Close()

	client := New(srv.URL, "user", "pass", 5*time.Second, zerolog.Nop())
	result, err := client.PushOrder(context.Background(), PushOrderRequest{
		ClientOrderID: uuid.New(),
		OrderNumber:   "AMC-1",
		Items:         []ItemDTO{{SKU: "SKU-1", Qty: 1, Price: 100}},
	})
	if err != nil {
		t.Fatalf("PushOrder: %v", err)
	}
	if result.OnecDocumentGUID != guid || result.OnecDocumentNumber != "УТ-00042" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestPushOrder_NonOKStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad counterparty"}`))
	}))
	defer srv.Close()

	client := New(srv.URL, "user", "pass", 5*time.Second, zerolog.Nop())
	_, err := client.PushOrder(context.Background(), PushOrderRequest{ClientOrderID: uuid.New(), OrderNumber: "AMC-2"})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
```

- [ ] **Step 3: Прогнать, убедиться что падает (компиляция)**

```bash
cd back/integrations && go test ./internal/onecorders/... -v
```

- [ ] **Step 4: Реализовать клиент**

По образцу `back/integrations/internal/onec/client.go` (тот же стиль: `fasthttp.Client`, Basic Auth, `DoTimeout`):

```go
package onecorders

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const pushOrderPath = "/hs/amc-integration/orders"

type Client struct {
	baseURL  string
	user     string
	password string
	timeout  time.Duration
	http     *fasthttp.Client
	logger   zerolog.Logger
}

func New(baseURL, user, password string, timeout time.Duration, logger zerolog.Logger) *Client {
	return &Client{baseURL: baseURL, user: user, password: password, timeout: timeout, http: &fasthttp.Client{}, logger: logger}
}

func (c *Client) PushOrder(ctx context.Context, in PushOrderRequest) (PushOrderResult, error) {
	_ = ctx

	body, err := json.Marshal(in)
	if err != nil {
		return PushOrderResult{}, fmt.Errorf("push order: marshal request: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(c.baseURL + pushOrderPath)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	auth := base64.StdEncoding.EncodeToString([]byte(c.user + ":" + c.password))
	req.Header.Set("Authorization", "Basic "+auth)
	req.SetBody(body)

	if err = c.http.DoTimeout(req, resp, c.timeout); err != nil {
		c.logger.Error().Str("clientOrderID", in.ClientOrderID.String()).Err(err).Msg("onec push order request failed")
		return PushOrderResult{}, fmt.Errorf("push order request: %w", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		c.logger.Error().Str("clientOrderID", in.ClientOrderID.String()).Int("status", resp.StatusCode()).Str("response", string(resp.Body())).Msg("onec push order unexpected status")
		return PushOrderResult{}, fmt.Errorf("push order: unexpected status %d: %s", resp.StatusCode(), resp.Body())
	}

	var out pushOrderSuccessResponse
	if err = json.Unmarshal(resp.Body(), &out); err != nil {
		return PushOrderResult{}, fmt.Errorf("push order: decode response: %w", err)
	}
	return PushOrderResult{OnecDocumentGUID: out.OnecDocumentGUID, OnecDocumentNumber: out.OnecDocumentNumber}, nil
}
```

`httptest.Server` слушает на `http://127.0.0.1:PORT` без Basic Auth проверки — тест из Step 2 не проверяет заголовок `Authorization` явно; это осознанно (клиент уже покрыт тем же паттерном, что `internal/onec/client_test.go` — свериться с ним, если там auth-заголовок всё же проверяется, повторить этот же чек и здесь).

- [ ] **Step 5: Прогнать тесты**

```bash
cd back/integrations && go test ./internal/onecorders/... -v
```

Ожидаемо: `PASS`.

- [ ] **Step 6: Commit**

```bash
git add back/integrations/internal/onecorders
git commit -m "feat(integrations): add 1С order-push HTTP client"
```

---

### Task 10: integrations — ручной fiber-хендлер `PushOrder`

**Files:**
- Create: `back/integrations/internal/transport/http/onecorders.go`
- Test: `back/integrations/internal/transport/http/onecorders_test.go`

**Interfaces:**
- Consumes: `onecorders.Client.PushOrder` (Task 9).
- Produces: `func RegisterPushOrderRoute(app *fiber.App, pusher OnecPusher, logger zerolog.Logger)`, где
  ```go
  type OnecPusher interface {
      PushOrder(ctx context.Context, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error)
  }
  ```
  Wire-формат запроса/ответа на этом хендлере — **тот же JSON**, что `onecorders.PushOrderRequest`/`PushOrderResult` (проксируется почти без изменений; хендлер — тонкий слой, который парсит тело, вызывает клиент, отдаёт результат или ошибку).

- [ ] **Step 1: Написать падающий тест**

```go
package http

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/mbatimel/AMC/integrations/internal/onecorders"
)

type fakePusher struct {
	result onecorders.PushOrderResult
	err    error
	gotReq onecorders.PushOrderRequest
}

func (f *fakePusher) PushOrder(ctx context.Context, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error) {
	f.gotReq = req
	return f.result, f.err
}

func TestPushOrderRoute_Success(t *testing.T) {
	app := fiber.New()
	pusher := &fakePusher{result: onecorders.PushOrderResult{OnecDocumentGUID: uuid.New(), OnecDocumentNumber: "УТ-1"}}
	RegisterPushOrderRoute(app, pusher, zerolog.Nop())

	body, _ := json.Marshal(onecorders.PushOrderRequest{ClientOrderID: uuid.New(), OrderNumber: "AMC-1", Items: []onecorders.ItemDTO{{SKU: "SKU-1", Qty: 1}}})
	req := httptest.NewRequest("POST", "/api/v1/onec-orders/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if pusher.gotReq.OrderNumber != "AMC-1" || len(pusher.gotReq.Items) != 1 {
		t.Fatalf("unexpected request passed to pusher: %+v", pusher.gotReq)
	}
}

func TestPushOrderRoute_PusherFails_Returns502(t *testing.T) {
	app := fiber.New()
	pusher := &fakePusher{err: errors.New("onec down")}
	RegisterPushOrderRoute(app, pusher, zerolog.Nop())

	body, _ := json.Marshal(onecorders.PushOrderRequest{ClientOrderID: uuid.New(), OrderNumber: "AMC-1"})
	req := httptest.NewRequest("POST", "/api/v1/onec-orders/push", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 502 {
		t.Fatalf("expected 502, got %d", resp.StatusCode)
	}
}

func TestPushOrderRoute_MalformedBody_Returns400(t *testing.T) {
	app := fiber.New()
	pusher := &fakePusher{}
	RegisterPushOrderRoute(app, pusher, zerolog.Nop())

	req := httptest.NewRequest("POST", "/api/v1/onec-orders/push", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
```

Добавить недостающие импорты (`context`, `errors`, `zerolog`) при вставке — тест выше опускает их для краткости блока, в реальном файле дописать полный `import (...)`.

- [ ] **Step 2: Прогнать, убедиться что падает**

```bash
cd back/integrations && go test ./internal/transport/http/... -run TestPushOrderRoute -v
```

- [ ] **Step 3: Реализовать**

```go
package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/integrations/internal/onecorders"
)

type OnecPusher interface {
	PushOrder(ctx context.Context, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error)
}

func RegisterPushOrderRoute(app *fiber.App, pusher OnecPusher, logger zerolog.Logger) {
	app.Post("/api/v1/onec-orders/push", func(ctx *fiber.Ctx) error {
		var req onecorders.PushOrderRequest
		if err := ctx.BodyParser(&req); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "malformed request body"})
		}

		result, err := pusher.PushOrder(ctx.UserContext(), req)
		if err != nil {
			logger.Error().Str("clientOrderID", req.ClientOrderID.String()).Err(err).Msg("push order to onec failed")
			return ctx.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
		}

		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"onec_document_guid":   result.OnecDocumentGUID,
			"onec_document_number": result.OnecDocumentNumber,
		})
	})
}
```

- [ ] **Step 4: Прогнать тесты**

```bash
cd back/integrations && go test ./internal/transport/http/... -v
```

Ожидаемо: все `PASS`.

- [ ] **Step 5: Commit**

```bash
git add back/integrations/internal/transport/http/onecorders.go back/integrations/internal/transport/http/onecorders_test.go
git commit -m "feat(integrations): hand-written PushOrder fiber route (tg can't express item arrays)"
```

---

### Task 11: integrations — учёт попыток в `sync_jobs`/`sync_logs`

**Files:**
- Modify: `back/integrations/internal/storage/postgres/postgres.go`
- Create: `back/integrations/internal/storage/postgres/sql/createIntegrationJob.sql`
- Test: `back/integrations/internal/storage/postgres/integration_test.go`

**Interfaces:**
- Produces: `func (s *Storage) CreateIntegrationJob(ctx context.Context, systemID uuid.UUID, direction, entityType string) (uuid.UUID, error)` — не трогает существующий `CreateSyncJob` (используется только `onec-sync`).

- [ ] **Step 1: Написать падающий тест**

Добавить в `back/integrations/internal/storage/postgres/integration_test.go` (в существующий `TestStorageIntegration` или отдельной функцией — свериться со стилем файла, там один большой тест):

```go
func TestCreateIntegrationJob(t *testing.T) {
	dsn := os.Getenv("INTEGRATIONS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("INTEGRATIONS_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	storage := New(pool)

	systemID, err := storage.UpsertIntegrationSystem(ctx, "onec_orders_test", "1С заказы тест")
	if err != nil {
		t.Fatalf("upsert system: %v", err)
	}

	jobID, err := storage.CreateIntegrationJob(ctx, systemID, "outbound", "order_create")
	if err != nil {
		t.Fatalf("CreateIntegrationJob: %v", err)
	}

	var direction, entityType, status string
	if err = pool.QueryRow(ctx, `SELECT direction, entity_type, status FROM sync_jobs WHERE id = $1`, jobID).Scan(&direction, &entityType, &status); err != nil {
		t.Fatalf("read back job: %v", err)
	}
	if direction != "outbound" || entityType != "order_create" || status != "running" {
		t.Fatalf("unexpected job row: direction=%s entityType=%s status=%s", direction, entityType, status)
	}
}
```

- [ ] **Step 2: Прогнать, убедиться что падает**

```bash
cd back/integrations && INTEGRATIONS_TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run TestCreateIntegrationJob -v
```

- [ ] **Step 3: Реализовать**

`back/integrations/internal/storage/postgres/sql/createIntegrationJob.sql`:
```sql
INSERT INTO sync_jobs (system_id, direction, entity_type, status, attempts)
VALUES ($1, $2, $3, 'running', 1)
RETURNING id;
```

В `postgres.go`, добавить к списку `var (...)` в начале файла `sqlCreateIntegrationJob = query("createIntegrationJob.sql")`, и метод:
```go
func (s *Storage) CreateIntegrationJob(ctx context.Context, systemID uuid.UUID, direction, entityType string) (uuid.UUID, error) {
	var id uuid.UUID
	if err := s.pool.QueryRow(ctx, sqlCreateIntegrationJob, systemID, direction, entityType).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("create integration job: %w", err)
	}
	return id, nil
}
```

- [ ] **Step 4: Прогнать тесты**

```bash
cd back/integrations && INTEGRATIONS_TEST_DATABASE_URL="postgres://..." go test ./internal/storage/postgres/... -run TestCreateIntegrationJob -v
```

- [ ] **Step 5: Commit**

```bash
git add back/integrations/internal/storage/postgres
git commit -m "feat(integrations): add CreateIntegrationJob for order push/webhook audit trail"
```

Логирование конкретных попыток (`AddSyncLog` при неудачном `PushOrder`/вебхуке) — подключить в Task 14 при финальной сборке `cmd/onec-orders-api`, самостоятельной задачей не выделяется (несколько строк на месте вызова).

---

### Task 12: integrations — конфиг `onec-orders-api`

**Files:**
- Modify: `back/integrations/internal/config/config.go`
- Test: `back/integrations/internal/config/config_test.go`

**Interfaces:**
- Produces: новые поля `Config`: `OnecOrdersBindAddr`, `OnecOrdersBaseURL`, `OnecOrdersUser`, `OnecOrdersPassword`, `OnecOrdersRequestTimeout`, `OnecWebhookAPIKey`, `OrdersURL`, `OrdersSystemUserID uuid.UUID`.

- [ ] **Step 1: Посмотреть существующий `config_test.go`**

```bash
cat back/integrations/internal/config/config_test.go
```

Понять текущий стиль теста (какие env переменные ставятся/очищаются) — новый тест должен следовать тому же паттерну (скорее всего `t.Setenv` + вызов `LoadConfig()`).

- [ ] **Step 2: Написать падающий тест**

Добавить в `config_test.go` (используя обнаруженный в Step 1 паттерн `t.Setenv` для обязательных полей вроде `PG_DB`/`PG_USER`/`PG_PASSWORD`/`ONEC_BASE_URL`/`ONEC_USER`/`ONEC_PASSWORD`, чтобы `LoadConfig` не упал раньше времени):

```go
func TestLoadConfig_OnecOrdersFields(t *testing.T) {
	// повторить существующие t.Setenv для обязательных PG_*/ONEC_* полей, см. другие тесты этого файла
	t.Setenv("ONEC_ORDERS_BASE_URL", "http://onec-host/hs/amc-integration")
	t.Setenv("ONEC_ORDERS_USER", "pushuser")
	t.Setenv("ONEC_ORDERS_PASSWORD", "pushpass")
	t.Setenv("ONEC_WEBHOOK_API_KEY", "webhook-secret")
	t.Setenv("ORDERS_URL", "http://orders:8082")
	t.Setenv("ORDERS_SYSTEM_USER_ID", "00000000-0000-0000-0000-0000000a0ec1")

	cfg := LoadConfig()

	if cfg.OnecOrdersBaseURL != "http://onec-host/hs/amc-integration" {
		t.Fatalf("unexpected OnecOrdersBaseURL: %s", cfg.OnecOrdersBaseURL)
	}
	if cfg.OnecWebhookAPIKey != "webhook-secret" {
		t.Fatalf("unexpected OnecWebhookAPIKey: %s", cfg.OnecWebhookAPIKey)
	}
	if cfg.OrdersSystemUserID.String() != "00000000-0000-0000-0000-0000000a0ec1" {
		t.Fatalf("unexpected OrdersSystemUserID: %s", cfg.OrdersSystemUserID)
	}
	if cfg.OnecOrdersRequestTimeout != 15*time.Second {
		t.Fatalf("expected default timeout 15s, got %s", cfg.OnecOrdersRequestTimeout)
	}
}
```

- [ ] **Step 3: Прогнать, убедиться что падает (компиляция — новых полей ещё нет)**

```bash
cd back/integrations && go test ./internal/config/... -run TestLoadConfig_OnecOrdersFields -v
```

- [ ] **Step 4: Реализовать**

В `Config`:
```go
	OnecOrdersBindAddr       string
	OnecOrdersBaseURL        string
	OnecOrdersUser           string
	OnecOrdersPassword       string
	OnecOrdersRequestTimeout time.Duration
	OnecWebhookAPIKey        string
	OrdersURL                string
	OrdersSystemUserID       uuid.UUID
```

В `LoadConfig`, после существующих присвоений — добавить чтение (без `log.Fatal` на этом этапе, если хочется чтобы `onec-sync` бинарник, который тоже использует этот же `Config`, не требовал этих переменных; `log.Fatal` на них добавить только в `cmd/onec-orders-api/main.go`, Task 13, не в общем `LoadConfig`, т.к. `LoadConfig` общий для обоих бинарников модуля):

```go
	cfg.OnecOrdersBindAddr = GetEnv("ONEC_ORDERS_BIND_ADDR", ":8090")
	cfg.OnecOrdersBaseURL = os.Getenv("ONEC_ORDERS_BASE_URL")
	cfg.OnecOrdersUser = os.Getenv("ONEC_ORDERS_USER")
	cfg.OnecOrdersPassword = os.Getenv("ONEC_ORDERS_PASSWORD")
	cfg.OnecOrdersRequestTimeout = getEnvDuration("ONEC_ORDERS_REQUEST_TIMEOUT", 15*time.Second)
	cfg.OnecWebhookAPIKey = os.Getenv("ONEC_WEBHOOK_API_KEY")
	cfg.OrdersURL = os.Getenv("ORDERS_URL")
	if raw := os.Getenv("ORDERS_SYSTEM_USER_ID"); raw != "" {
		parsed, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			log.Fatal().Err(parseErr).Msg("invalid ORDERS_SYSTEM_USER_ID")
		}
		cfg.OrdersSystemUserID = parsed
	}
```

Добавить импорт `"github.com/google/uuid"` в `config.go`.

- [ ] **Step 5: Прогнать тесты**

```bash
cd back/integrations && go test ./internal/config/... -v
```

Ожидаемо: всё `PASS`, включая уже существовавшие тесты (не сломать `onec-sync`-related поля).

- [ ] **Step 6: Commit**

```bash
git add back/integrations/internal/config/config.go back/integrations/internal/config/config_test.go
git commit -m "feat(integrations): add onec-orders-api config fields"
```

---

### Task 13: integrations — `cmd/onec-orders-api/main.go` + Dockerfile

**Files:**
- Create: `back/integrations/cmd/onec-orders-api/main.go`
- Create: `back/integrations/Dockerfile.onec-orders-api`

**Interfaces:**
- Consumes: всё из Task 6-12 (`errors`, `internalAPI.OnecOrdersAPI`, `service.OnecOrdersService`, `onecorders.Client`, `transport/http.RegisterPushOrderRoute`, `storage.CreateIntegrationJob`, `config.Config`).
- Требует: ⚠ Task 7's `tg transport` уже сгенерирован пользователем (`internal/transport/jsonRPC/internalapi` существует) — без этого файл не соберётся (импортирует сгенерированный пакет).

- [ ] **Step 1: Написать `main.go`**

По образцу `back/orders/cmd/main.go` (fasthttp.Server поверх `app.Fiber().Handler()`) + существующего `back/integrations/cmd/main.go` (health server, env-файл, graceful shutdown):

```go
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	"github.com/mbatimel/AMC/integrations/internal/config"
	"github.com/mbatimel/AMC/integrations/internal/onecorders"
	"github.com/mbatimel/AMC/integrations/internal/service"
	"github.com/mbatimel/AMC/integrations/internal/storage/postgres"
	transportHTTP "github.com/mbatimel/AMC/integrations/internal/transport/http"
	internalapi "github.com/mbatimel/AMC/integrations/internal/transport/jsonRPC/internalapi"
	ordersTransport "github.com/mbatimel/AMC/orders/pkg/client/transport"
)

const serviceName = "onec-orders-api"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Str("serviceName", serviceName).Logger()

	if err := config.LoadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}
	cfg := config.LoadConfig()
	if cfg.OnecOrdersBaseURL == "" || cfg.OnecOrdersUser == "" || cfg.OnecOrdersPassword == "" {
		log.Fatal().Msg("ONEC_ORDERS_BASE_URL, ONEC_ORDERS_USER and ONEC_ORDERS_PASSWORD must be specified")
	}
	if cfg.OnecWebhookAPIKey == "" || cfg.OrdersURL == "" || cfg.OrdersSystemUserID.String() == "00000000-0000-0000-0000-000000000000" {
		log.Fatal().Msg("ONEC_WEBHOOK_API_KEY, ORDERS_URL and ORDERS_SYSTEM_USER_ID must be specified")
	}

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()
	storageImpl := postgres.New(pool)

	pushClient := onecorders.New(cfg.OnecOrdersBaseURL, cfg.OnecOrdersUser, cfg.OnecOrdersPassword, cfg.OnecOrdersRequestTimeout, log.Logger)
	ordersClient := ordersTransport.NewClientOrdersAPI(cfg.OrdersURL)

	webhookSvc := service.NewOnecOrdersService(log.Logger, ordersClient, cfg.OrdersSystemUserID, cfg.OnecWebhookAPIKey)

	app := internalapi.New(log.Logger, internalapi.OnecOrdersAPI(internalapi.NewOnecOrdersAPI(webhookSvc))).WithLog().WithMetrics()
	transportHTTP.RegisterPushOrderRoute(app.Fiber(), pushOrderAdapter{client: pushClient, storage: storageImpl}, log.Logger)

	server := &fasthttp.Server{Handler: app.Fiber().Handler()}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.OnecOrdersBindAddr).Msg("onec-orders-api started")
		if serveErr := server.ListenAndServe(cfg.OnecOrdersBindAddr); serveErr != nil {
			log.Fatal().Err(serveErr).Msg("failed to listen and serve onec-orders-api")
		}
	}()

	<-shutdown
	if err = server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown server")
	}
}
```

**Проверить перед вставкой:**
- Точное имя конструктора сгенерированного сервера — `internalapi.New(...)` и `internalapi.NewOnecOrdersAPI(svc)` — сверить с реальными именами в `back/orders/internal/transport/jsonRPC/externalapi/server.go`/`ordersapi-http.go` после того, как аналогичная генерация пройдёт у `orders` (Task 3) — паттерн там `externalapi.New(log, externalapi.OrdersAPI(externalapi.NewOrdersAPI(svc)))`; у `internalAPI`-пакетов (см. `access`) конструктор называется по аналогии `internalapi.New(...)`/`internalapi.NewAccessAPI(svc)` — **имя конкретно для `OnecOrdersAPI` появится только после Task 7 кодогенерации**, подставить фактическое.
- Точное имя сгенерированного клиента `orders` — `NewClientOrdersAPI` — это ДОГАДКА по аналогии с `access.NewClientAccessAPI`; сверить с реальным именем в `back/orders/pkg/client/transport/*.go` после Task 3 Step 12.
- `pushOrderAdapter` — обёртка, реализующая `transportHTTP.OnecPusher` (`PushOrder(ctx, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error)`), которая вызывает `pushClient.PushOrder` и по пути логирует попытку через `storage.CreateIntegrationJob`/`AddSyncLog` (Task 11). Дописать прямо в этом файле или в `internal/onecorders/adapter.go`:

```go
type pushOrderAdapter struct {
	client  *onecorders.Client
	storage *postgres.Storage
}

func (a pushOrderAdapter) PushOrder(ctx context.Context, req onecorders.PushOrderRequest) (onecorders.PushOrderResult, error) {
	systemID, err := a.storage.UpsertIntegrationSystem(ctx, "onec_orders", "1С:УТ 10.3 (заказы)")
	if err != nil {
		log.Error().Err(err).Msg("upsert integration system for push order failed")
	}
	jobID, jobErr := a.storage.CreateIntegrationJob(ctx, systemID, "outbound", "order_create")
	if jobErr != nil {
		log.Error().Err(jobErr).Msg("create integration job for push order failed")
	}

	result, pushErr := a.client.PushOrder(ctx, req)
	status := "success"
	lastError := ""
	if pushErr != nil {
		status = "failed"
		lastError = pushErr.Error()
	}
	if finishErr := a.storage.FinishSyncJob(ctx, jobID, status, lastError); finishErr != nil {
		log.Error().Err(finishErr).Msg("finish integration job for push order failed")
	}
	return result, pushErr
}
```

Добавить недостающие импорты (`context`, `log`) в файл при вставке.

- [ ] **Step 2: Написать `Dockerfile.onec-orders-api`**

```dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY back/orders/go.mod back/orders/go.sum ./orders/
COPY back/access/go.mod back/access/go.sum ./access/
COPY back/integrations/go.mod back/integrations/go.sum ./integrations/
WORKDIR /src/integrations
RUN go mod download

COPY back/orders/ /src/orders/
COPY back/access/ /src/access/
COPY back/integrations/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/onec-orders-api ./cmd/onec-orders-api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates && \
    adduser -D -H -u 10001 app

COPY --from=builder /out/onec-orders-api /usr/local/bin/onec-orders-api

USER app
EXPOSE 8090

ENTRYPOINT ["/usr/local/bin/onec-orders-api"]
```

`COPY back/access/` нужен транзитивно — `back/orders`'у для сборки требуется `back/access` (см. его `replace ... => ../access`); проверить фактический relative-path после `go mod edit -replace` в Task 8/9 — если путь получился другим (например `../../orders` из-за структуры `WORKDIR`), поправить `COPY`/`replace` пути соответственно. Свериться с тем, как `back/orders/Dockerfile` уже решает эту же проблему (он тоже зависит от `access`) — скопировать оттуда рабочий паттерн `COPY`, а не изобретать заново.

- [ ] **Step 3: Проверить сборку локально (после того как Task 7/3 кодогенерация выполнена пользователем)**

```bash
cd back/integrations && go build ./cmd/onec-orders-api/...
```

Ожидаемо: чисто.

- [ ] **Step 4: Commit**

```bash
git add back/integrations/cmd/onec-orders-api back/integrations/Dockerfile.onec-orders-api
git commit -m "feat(integrations): wire onec-orders-api binary"
```

---

### Task 14: orders — исходящий клиент к `onec-orders-api` (`internal/onecclient`)

**Files:**
- Create: `back/orders/internal/onecclient/client.go`
- Test: `back/orders/internal/onecclient/client_test.go`
- Modify: `back/orders/cmd/main.go` (заменить `nil`-заглушку из Task 5 Step 5 на реальный клиент)

**Interfaces:**
- Produces: `func New(baseURL string, timeout time.Duration) *Client`, реализует `service.OnecPusher` (Task 5) — `PushOrder(ctx context.Context, order service.OnecPushOrder) (onecGUID uuid.UUID, onecNumber string, err error)`.
- Consumes: `service.OnecPushOrder`/`service.OnecOrderItem` (Task 5). Wire-формат JSON — тот же, что `onecorders.PushOrderRequest` на стороне `integrations` (Task 9) — дублируется как собственные локальные типы (разные Go-модули, нет общего пакета), поля/JSON-теги должны совпадать 1:1.

- [ ] **Step 1: Написать падающий тест**

```go
package onecclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/orders/internal/service"
)

func TestPushOrder_Success(t *testing.T) {
	guid := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req wireRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(req.Items) != 1 || req.Items[0].SKU != "SKU-1" {
			t.Fatalf("unexpected items: %+v", req.Items)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wireResponse{OnecDocumentGUID: guid, OnecDocumentNumber: "УТ-1"})
	}))
	defer srv.Close()

	client := New(srv.URL, 5*time.Second)
	onecGUID, onecNumber, err := client.PushOrder(context.Background(), service.OnecPushOrder{
		ClientOrderID: uuid.New(),
		OrderNumber:   "AMC-1",
		Items:         []service.OnecOrderItem{{SKU: "SKU-1", Qty: 1, Price: 100}},
	})
	if err != nil {
		t.Fatalf("PushOrder: %v", err)
	}
	if onecGUID != guid || onecNumber != "УТ-1" {
		t.Fatalf("unexpected result: %s / %s", onecGUID, onecNumber)
	}
}

func TestPushOrder_NonOKStatus_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	client := New(srv.URL, 5*time.Second)
	_, _, err := client.PushOrder(context.Background(), service.OnecPushOrder{ClientOrderID: uuid.New(), OrderNumber: "AMC-1"})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}
```

- [ ] **Step 2: Прогнать, убедиться что падает**

```bash
cd back/orders && go test ./internal/onecclient/... -v
```

- [ ] **Step 3: Реализовать**

```go
package onecclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mbatimel/AMC/orders/internal/service"
)

const pushPath = "/api/v1/onec-orders/push"

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

type wireCounterparty struct {
	OneCGUID string `json:"onec_guid"`
	INN      string `json:"inn"`
	Name     string `json:"name"`
}
type wireDelivery struct {
	Type        string `json:"type"`
	Address     string `json:"address"`
	ContactName string `json:"contact_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}
type wireItem struct {
	OneCGUID string  `json:"onec_guid"`
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Qty      int     `json:"qty"`
	Price    float64 `json:"price"`
	VATRate  float64 `json:"vat_rate"`
}
type wireRequest struct {
	ClientOrderID uuid.UUID        `json:"client_order_id"`
	OrderNumber   string           `json:"order_number"`
	Counterparty  wireCounterparty `json:"counterparty"`
	Delivery      wireDelivery     `json:"delivery"`
	Comment       string           `json:"comment"`
	Items         []wireItem       `json:"items"`
}
type wireResponse struct {
	OnecDocumentGUID   uuid.UUID `json:"onec_document_guid"`
	OnecDocumentNumber string    `json:"onec_document_number"`
}

func nullUUIDString(v uuid.NullUUID) string {
	if !v.Valid {
		return ""
	}
	return v.UUID.String()
}

func (c *Client) PushOrder(ctx context.Context, order service.OnecPushOrder) (uuid.UUID, string, error) {
	items := make([]wireItem, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, wireItem{
			OneCGUID: nullUUIDString(item.OneCGUID),
			SKU:      item.SKU,
			Name:     item.Name,
			Qty:      item.Qty,
			Price:    item.Price,
			VATRate:  item.VATRate,
		})
	}
	body, err := json.Marshal(wireRequest{
		ClientOrderID: order.ClientOrderID,
		OrderNumber:   order.OrderNumber,
		Counterparty: wireCounterparty{
			OneCGUID: nullUUIDString(order.CounterpartyGUID),
			INN:      order.CounterpartyINN,
			Name:     order.CounterpartyName,
		},
		Delivery: wireDelivery{
			Type:        order.DeliveryType,
			Address:     order.DeliveryAddress,
			ContactName: order.ContactName,
			Phone:       order.Phone,
			Email:       order.Email,
		},
		Comment: order.Comment,
		Items:   items,
	})
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+pushPath, bytes.NewReader(body))
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, "", fmt.Errorf("push order: unexpected status %d", resp.StatusCode)
	}

	var out wireResponse
	if err = json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return uuid.Nil, "", fmt.Errorf("push order: decode response: %w", err)
	}
	return out.OnecDocumentGUID, out.OnecDocumentNumber, nil
}
```

- [ ] **Step 4: Прогнать тесты**

```bash
cd back/orders && go test ./internal/onecclient/... -v
```

Ожидаемо: `PASS`.

- [ ] **Step 5: Заменить `nil`-заглушку в `cmd/main.go`**

В `back/orders/cmd/main.go`, заменить временный `nil` (Task 5 Step 5) на:

```go
onecPusher := onecclient.New(cfg.IntegrationsURL, 15*time.Second)
```

и передать `onecPusher` в `ordersService.NewOrdersApiService(...)` четвёртым аргументом. Добавить импорт `"github.com/mbatimel/AMC/orders/internal/onecclient"` и `"time"`. Сделать `IntegrationsURL` обязательным (`log.Fatal`, если пусто) в `config.go` рядом с проверкой `AccessURL`.

- [ ] **Step 6: Собрать весь модуль**

```bash
cd back/orders && go build ./... && go vet ./... && go test ./...
```

Ожидаемо: чисто (при условии, что Task 3 Step 12 кодогенерация уже выполнена).

- [ ] **Step 7: Commit**

```bash
git add back/orders
git commit -m "feat(orders): wire onecclient into CreateOrder, remove nil placeholder"
```

---

### Task 15: Деплой — docker-compose, nginx, .env.example, CI

**Files:**
- Modify: `deploy/docker-compose.yml`
- Modify: `deploy/nginx/conf.d/wk.amctechgroup.ru.conf`
- Modify: `deploy/.env.example` (если файл существует под этим именем — проверить `ls deploy/*.env*`)
- Modify: `.github/workflows/build.yml`

**Interfaces:**
- Не имеет тестируемого поведения (конфигурация) — верификация: `docker compose config` (синтаксис) + `nginx -t` недоступен локально без контейнера, визуальная сверка.

- [ ] **Step 1: docker-compose — новый сервис + env у `orders`**

В `deploy/docker-compose.yml`, после блока `onec-sync:` (см. текущее содержимое файла), добавить:

```yaml
  onec-orders-api:
    image: ghcr.io/mbatimel/amc-onec-orders-api:${IMAGE_TAG}
    restart: unless-stopped
    environment:
      PG_HOST: postgres
      PG_PORT: "5432"
      PG_DB: ${PG_DB}
      PG_USER: ${PG_USER}
      PG_PASSWORD: ${PG_PASSWORD}
      ONEC_ORDERS_BIND_ADDR: ":8090"
      ONEC_ORDERS_BASE_URL: ${ONEC_ORDERS_BASE_URL}
      ONEC_ORDERS_USER: ${ONEC_ORDERS_USER}
      ONEC_ORDERS_PASSWORD: ${ONEC_ORDERS_PASSWORD}
      ONEC_ORDERS_REQUEST_TIMEOUT: "15s"
      ONEC_WEBHOOK_API_KEY: ${ONEC_WEBHOOK_API_KEY}
      ORDERS_URL: "http://orders:8082"
      ORDERS_SYSTEM_USER_ID: "00000000-0000-0000-0000-0000000a0ec1"
    depends_on:
      migrations:
        condition: service_completed_successfully
      access:
        condition: service_started
    networks: [amc_net]
```

В блоке `orders:` (существующий, см. текущий файл) добавить в `environment`:
```yaml
      INTEGRATIONS_URL: "http://onec-orders-api:8090"
```
и в `depends_on`:
```yaml
      onec-orders-api:
        condition: service_started
```

- [ ] **Step 2: nginx — публичный маршрут вебхука**

В `deploy/nginx/conf.d/wk.amctechgroup.ru.conf`, после блока `location /api/v1/orders/`, добавить:

```nginx
    location /api/v1/onec/ {
        proxy_pass http://onec-orders-api:8090/api/v1/onec/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $http_x_forwarded_proto;
    }
```

`/api/v1/onec-orders/push` **не добавлять** в nginx — недостижим снаружи намеренно (см. спеку).

- [ ] **Step 3: `.env.example`**

```bash
find /Users/macbook/Desktop/AMC/AMC/deploy -iname "*.env*"
```

В найденном файле добавить (по аналогии с уже существующими `ONEC_BASE_URL`/`ONEC_USER`/`ONEC_PASSWORD`):
```
ONEC_ORDERS_BASE_URL=
ONEC_ORDERS_USER=
ONEC_ORDERS_PASSWORD=
ONEC_WEBHOOK_API_KEY=
```

- [ ] **Step 4: CI — новый образ в матрице сборки**

В `.github/workflows/build.yml`, после записи `- name: integrations / dockerfile: ./back/integrations/Dockerfile / image: ghcr.io/mbatimel/amc-onec-sync` добавить:

```yaml
          - name: onec-orders-api
            dockerfile: ./back/integrations/Dockerfile.onec-orders-api
            image: ghcr.io/mbatimel/amc-onec-orders-api
```

`.github/workflows/go.yml` менять не нужно — `back/integrations` и `back/orders` уже в матрице тестов, новые пакеты подхватятся автоматически через `go test ./...`.

- [ ] **Step 5: Проверить синтаксис compose**

```bash
cd deploy && docker compose config > /dev/null && echo OK
```

Ожидаемо: `OK` (переменные типа `${ONEC_ORDERS_BASE_URL}` без значения в окружении — это нормально для `config`, docker compose не требует их быть заданными для синтаксической проверки, если только нет `:?required` — не использовать этот синтаксис здесь, свериться, что остальные переменные в файле его тоже не используют).

- [ ] **Step 6: Commit**

```bash
git add deploy .github/workflows/build.yml
git commit -m "chore(deploy): wire onec-orders-api into compose, nginx, CI"
```

---

## Self-Review

**Spec coverage:**
- Стейт-машина `new→processing→delivered`, `cancelled` вне контроля 1С — Task 4 (транзакция), Task 8 (webhook allow-list + 409). ✅
- Одна транзакция вместо компенсирующего отката — Task 4. ✅
- `PushOrder` без tg (структуры/слайсы) — Task 9/10/14, задокументировано в Global Constraints. ✅
- `OnecOrderStatusWebhook` через tg — Task 7/8. ✅
- Системный admin-пользователь для вызовов `orders` — Task 1 (сид), Task 8/13 (использование). ✅
- `GetOrderStatus` вместо buyer-ориентированного `GetOrder` — Task 3. ✅
- `sync_jobs`/`sync_logs` учёт для push/webhook — Task 11, подключение в Task 13. ✅
- Деплой (compose/nginx/.env/CI) — Task 15. ✅
- Открытые вопросы спеки (уведомление 1С при отмене AMC-стороной, правила мэтчинга контрагента/товара в 1С) — намеренно вне плана, отмечены в спеке как TODO/ответственность 1С-разработчика, не требуют кода с нашей стороны сейчас. ✅

**Placeholder scan:** нет `TBD`/`TODO` кроме одного явного, специально оставленного (Task 5 Step 5, `nil`-заглушка `onecPusher` до Task 14) — это осознанное временное состояние между двумя задачами одного плана, не открытый вопрос; закрывается тем же планом в Task 14 Step 5. Остальные шаги — конкретный код, конкретные команды.

**Type consistency:** `service.OnecPushOrder`/`OnecOrderItem` (Task 5) используются 1:1 в Task 14 (`onecclient`); `onecorders.PushOrderRequest`/`ItemDTO`/`PushOrderResult` (Task 9) используются 1:1 в Task 10 (fiber-хендлер) и в Task 13 (`pushOrderAdapter`); `postgres.ProductOnecRef`/`CounterpartyOnecRef` (Task 2) используются в Task 5 без переименований. `OrdersClient` в Task 8 — сигнатуры (`GetOrderStatus`, `UpdateOrderStatus`) явно помечены как "сверить с реальным сгенерированным клиентом после Task 3" — это не incosistency, это честно обозначенная зависимость от кодогенерации, которую выполняет пользователь вне этого плана.

**Что осознанно не автоматизировано в этом плане:** сам запуск `tg transport`/`tg client` (Tasks 3/7) и первый `go build` всего репо целиком после них — выполняет пользователь между задачами, как и было явно попрошено.
