# Cart / Checkout / Order History (orders service) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Реализовать в сервисе `back/orders` бизнес-логику трёх ручек — корзина (`GetCart` + мутации), оформление заказа (`CreateOrder`), история заказов (`ListOrders`) — вместо текущих заглушек `NotImplementedError`.

**Architecture:** Транспортный слой и интерфейс `OrdersAPI` уже сгенерированы и не меняются. Меняются только: сервисный слой (`internal/service`), слой хранения (`internal/storage/postgres`), одна новая миграция БД, конфиг (ставка НДС) и модели ответа (добавляются два поля).

**Tech Stack:** Go 1.25, pgx/v4 (pgxpool), goose-миграции (`back/migrations`), стандартный `testing` (без testify — в репозитории такой зависимости нет и не добавляем).

## Global Constraints

- Интерфейс `back/orders/pkg/interfaces/externalapi/interface.go` НЕ меняется — все нужные методы уже объявлены.
- Транспортные файлы `internal/transport/jsonRPC/externalapi/*`, `internal/transport/custom-handlers/orders.go`, `swaggers/externalapi/swagger.yaml` НЕ трогаем — они сгенерированы из интерфейса и уже соответствуют нужным сигнатурам.
- Реализуются: `GetCart`, `AddCartItem`, `UpdateCartItem`, `DeleteCartItem`, `ClearCart`, `CreateOrder`, `ListOrders`.
- НЕ реализуются (остаются `NotImplementedError`): `GetOrder`, `CancelOrder`, `RepeatOrder`, `GetOrderDocuments`, `GetOrderHistory`, `UpdateOrderStatus`.
- Ставка НДС — из переменной окружения `NDS_VALUE` (не хардкод), дефолт `22`. НДС добавляется СВЕРХУ суммы после скидки.
- Скидка считается по `volume_discounts` (объёмная скидка от суммы корзины/заказа), приоритет — персональная строка контрагента, затем строка по `price_group_id`.
- Генерация документов (счёт/накладная) — вне скоупа. Поле "документ" в истории просто отражает то, что реально есть в таблице `documents` (для новых заказов — пусто).
- Новых внешних зависимостей (go.mod) не добавляем.
- Полный дизайн и обоснования — `docs/superpowers/specs/2026-07-28-orders-cart-checkout-design.md`.

## Локальная база для проверки заданий

Часть задач проверяется через реальный Postgres в Docker (он уже установлен и запущен на машине). Поднять его один раз перед началом:

```bash
cd back/migrations
docker compose --env-file .env up -d postgres
until PGPASSWORD=mbatimel psql -h localhost -p 5432 -U mbatimel -d AMC -c 'select 1' >/dev/null 2>&1; do sleep 1; done
go run ./cmd
```

Последняя команда прогоняет ВСЕ миграции (включая уже существующие и новую из Задачи 1) на контейнерном Postgres. После добавления новой миграции в Задаче 1 нужно перезапустить `go run ./cmd` ещё раз, чтобы применить её.

---

### Task 1: Миграция БД — payment_status и номер заказа

**Files:**
- Create: `back/migrations/pkg/migrations/data/20260728120000_orders_payment.sql`

**Interfaces:**
- Produces: колонка `orders.payment_status` (`VARCHAR(255) NOT NULL DEFAULT 'not_paid'`), колонка `order_status_history.payment_status` (`VARCHAR(255)`), последовательность `orders_number_seq`. Эти три объекта используются в Задаче 8 (storage/postgres/orders.go) для генерации номера и статуса оплаты.

- [ ] **Step 1: Создать файл миграции**

```sql
-- +goose Up
-- +goose StatementBegin
ALTER TABLE orders ADD COLUMN payment_status VARCHAR(255) NOT NULL DEFAULT 'not_paid';
ALTER TABLE order_status_history ADD COLUMN payment_status VARCHAR(255);

CREATE SEQUENCE orders_number_seq START WITH 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS orders_number_seq;
ALTER TABLE order_status_history DROP COLUMN IF EXISTS payment_status;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_status;
-- +goose StatementEnd
```

- [ ] **Step 2: Собрать миграционный сервис (проверка, что embed FS подхватил новый файл)**

Run: `cd back/migrations && go build ./...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Применить миграцию к локальному Postgres (см. раздел выше про докер)**

Run:
```bash
cd back/migrations
docker compose --env-file .env up -d postgres
until PGPASSWORD=mbatimel psql -h localhost -p 5432 -U mbatimel -d AMC -c 'select 1' >/dev/null 2>&1; do sleep 1; done
go run ./cmd
```
Expected: в логе `Migration applied to node 1`, без ошибок.

- [ ] **Step 4: Проверить колонки и последовательность вручную**

Run: `PGPASSWORD=mbatimel psql -h localhost -p 5432 -U mbatimel -d AMC -c "\d orders" | grep payment_status && PGPASSWORD=mbatimel psql -h localhost -p 5432 -U mbatimel -d AMC -c "\ds orders_number_seq"`
Expected: колонка `payment_status` найдена в `orders`, последовательность `orders_number_seq` существует.

- [ ] **Step 5: Commit**

```bash
git add back/migrations/pkg/migrations/data/20260728120000_orders_payment.sql
git commit -m "feat(migrations): add orders.payment_status and orders_number_seq"
```

---

### Task 2: Конфиг — ставка НДС из NDS_VALUE

**Files:**
- Modify: `back/orders/internal/config/config.go`
- Modify: `back/orders/.env`
- Modify: `deploy/.env.example`
- Modify: `deploy/docker-compose.yml`

**Interfaces:**
- Produces: `config.Config.VATRate float64` — используется в Задаче 13 (передаётся в конструктор сервиса) и Задаче 9 (сервис хранит его в поле `vatRate`).

- [ ] **Step 1: Добавить поле и парсинг в config.go**

В `back/orders/internal/config/config.go` добавить импорт `"strconv"` и поле в структуру:

```go
type Config struct {
	PGHost     string
	PGPort     string
	PGDB       string
	PGUser     string
	PGPassword string
	BindAddr   string
	AccessURL  string
	VATRate    float64
}
```

В `LoadConfig()`, перед `return cfg`:

```go
vatRateStr := GetEnv("NDS_VALUE", "22")
vatRate, err := strconv.ParseFloat(vatRateStr, 64)
if err != nil {
	log.Fatal().Err(err).Str("NDS_VALUE", vatRateStr).Msg("invalid NDS_VALUE")
}
cfg.VATRate = vatRate
```

- [ ] **Step 2: Добавить переменную в .env файлы**

`back/orders/.env` — добавить строку:
```
NDS_VALUE=22
```

`deploy/.env.example` — добавить строку (после блока `PG_*`):
```
NDS_VALUE=22
```

- [ ] **Step 3: Прокинуть переменную в deploy/docker-compose.yml**

В `deploy/docker-compose.yml`, в блоке `orders.environment` (после `ACCESS_INTERNAL_ADDRESS`) добавить:
```yaml
      NDS_VALUE: ${NDS_VALUE}
```

- [ ] **Step 4: Собрать сервис**

Run: `cd back/orders && go build ./...`
Expected: успешная сборка без ошибок.

- [ ] **Step 5: Commit**

```bash
git add back/orders/internal/config/config.go back/orders/.env deploy/.env.example deploy/docker-compose.yml
git commit -m "feat(orders): read VAT rate from NDS_VALUE env var"
```

---

### Task 3: Ошибка BadRequestError

**Files:**
- Modify: `back/orders/internal/errors/common.go`

**Interfaces:**
- Produces: `customErrors.BadRequestError() *Error` — используется во всех сервисных методах из Задач 9-12 для валидации входных данных.

- [ ] **Step 1: Добавить функцию-конструктор ошибки**

В `back/orders/internal/errors/common.go`, в блок `var (...)`, после `ForbiddenError`:

```go
BadRequestError = func() *Error { return New("bad request", fasthttp.StatusBadRequest, ErrBadRequest) }
```

(константа `ErrBadRequest` уже объявлена в этом же пакете, в `errors.go`, просто раньше не использовалась).

- [ ] **Step 2: Собрать пакет**

Run: `cd back/orders && go build ./internal/errors/...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Commit**

```bash
git add back/orders/internal/errors/common.go
git commit -m "feat(orders): add BadRequestError helper"
```

---

### Task 4: Модели — Subtotal/DiscountTotal в Cart и Order

**Files:**
- Modify: `back/orders/pkg/models/orders.go`

**Interfaces:**
- Produces: `models.Cart.Subtotal float64`, `models.Cart.DiscountTotal float64`, `models.Order.Subtotal float64`, `models.Order.DiscountTotal float64` — используются в Задачах 9-12 при сборке ответов.

- [ ] **Step 1: Расширить структуру Cart**

В `back/orders/pkg/models/orders.go`, заменить:

```go
type Cart struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ClientID  string     `json:"client_id"`
	Items     []CartItem `json:"items"`
	Total     float64    `json:"total"`
	VAT       float64    `json:"vat"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
```

на:

```go
type Cart struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	ClientID      string     `json:"client_id"`
	Items         []CartItem `json:"items"`
	Subtotal      float64    `json:"subtotal"`
	DiscountTotal float64    `json:"discount_total"`
	Total         float64    `json:"total"`
	VAT           float64    `json:"vat"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
```

- [ ] **Step 2: Расширить структуру Order**

Аналогично, в структуре `Order` (в этом же файле) добавить два поля между `VAT` и `DeliveryType` не нужно — добавить сразу после `VAT`:

```go
type Order struct {
	ID              string             `json:"id"`
	Number          string             `json:"number"`
	UserID          string             `json:"user_id"`
	ClientID        string             `json:"client_id"`
	Items           []OrderItem        `json:"items"`
	Subtotal        float64            `json:"subtotal"`
	DiscountTotal   float64            `json:"discount_total"`
	Total           float64            `json:"total"`
	VAT             float64            `json:"vat"`
	DeliveryType    string             `json:"delivery_type"`
	DeliveryAddress string             `json:"delivery_address"`
	ContactName     string             `json:"contact_name"`
	Phone           string             `json:"phone"`
	Email           string             `json:"email"`
	Comment         string             `json:"comment"`
	Status          string             `json:"status"`
	PaymentStatus   string             `json:"payment_status"`
	Documents       []OrderDocument    `json:"documents"`
	History         []OrderHistoryItem `json:"history"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}
```

- [ ] **Step 3: Собрать пакет**

Run: `cd back/orders && go build ./pkg/models/...`
Expected: успешная сборка без ошибок.

- [ ] **Step 4: Commit**

```bash
git add back/orders/pkg/models/orders.go
git commit -m "feat(orders): add Subtotal/DiscountTotal fields to Cart and Order"
```

---

### Task 5: Storage — резолюция контрагента, адрес/контакт при оформлении

**Files:**
- Create: `back/orders/internal/storage/postgres/counterparty.go`

**Interfaces:**
- Consumes: `s.pool *pgxpool.Pool` (поле `Storage`, уже существует в `postgres.go`).
- Produces:
  - `(*Storage).GetUserCounterpartyID(ctx, userID uuid.UUID) (uuid.UUID, error)` — `uuid.Nil` если у пользователя нет привязанного контрагента.
  - `(*Storage).GetCounterpartyPriceGroupID(ctx, counterpartyID uuid.UUID) (uuid.NullUUID, error)`
  - `(*Storage).InsertDeliveryAddress(ctx, counterpartyID uuid.UUID, addrType string, address string) (uuid.UUID, error)`
  - `(*Storage).InsertContact(ctx, counterpartyID uuid.UUID, fullName string, phone string, email string) (uuid.UUID, error)`

  Используются в Задачах 9-12 (сервисный слой) для резолюции `counterparty_id` из `clientID`/`userID` и для сохранения адреса/контакта при оформлении заказа.

- [ ] **Step 1: Написать файл**

```go
package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *Storage) GetUserCounterpartyID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var counterpartyID uuid.NullUUID
	err := s.pool.QueryRow(ctx, `SELECT counterparty_id FROM users WHERE id = $1`, userID).Scan(&counterpartyID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("get user counterparty id: %w", err)
	}
	if !counterpartyID.Valid {
		return uuid.Nil, nil
	}
	return counterpartyID.UUID, nil
}

func (s *Storage) GetCounterpartyPriceGroupID(ctx context.Context, counterpartyID uuid.UUID) (uuid.NullUUID, error) {
	var priceGroupID uuid.NullUUID
	err := s.pool.QueryRow(ctx, `SELECT price_group_id FROM counterparties WHERE id = $1`, counterpartyID).Scan(&priceGroupID)
	if err != nil {
		return uuid.NullUUID{}, fmt.Errorf("get counterparty price group id: %w", err)
	}
	return priceGroupID, nil
}

func (s *Storage) InsertDeliveryAddress(ctx context.Context, counterpartyID uuid.UUID, addrType string, address string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO counterparty_addresses (counterparty_id, type, address)
		VALUES ($1, $2, $3)
		RETURNING id
	`, counterpartyID, addrType, address).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert delivery address: %w", err)
	}
	return id, nil
}

func (s *Storage) InsertContact(ctx context.Context, counterpartyID uuid.UUID, fullName string, phone string, email string) (uuid.UUID, error) {
	var id uuid.UUID
	err := s.pool.QueryRow(ctx, `
		INSERT INTO counterparty_contacts (counterparty_id, full_name, phone, email)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, counterpartyID, fullName, phone, email).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert contact: %w", err)
	}
	return id, nil
}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/orders && go build ./internal/storage/...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Commit**

```bash
git add back/orders/internal/storage/postgres/counterparty.go
git commit -m "feat(orders): storage layer for counterparty resolution and address/contact insert"
```

---

### Task 6: Storage — корзина (carts.go) + интеграционный тест

**Files:**
- Create: `back/orders/internal/storage/postgres/carts.go`
- Create: `back/orders/internal/storage/postgres/integration_test.go`

**Interfaces:**
- Consumes: `Task 5` не требуется здесь напрямую, но использует то же поле `s.pool`.
- Produces:
  - `type CartItemRow struct { ID, ProductID uuid.UUID; SKU, ProductName string; Qty int; Price float64 }`
  - `(*Storage).GetOrCreateCart(ctx, userID, counterpartyID uuid.UUID) (uuid.UUID, error)`
  - `(*Storage).GetCartItems(ctx, cartID uuid.UUID) ([]CartItemRow, error)`
  - `(*Storage).ResolveProductPrice(ctx, productID uuid.UUID, priceGroupID uuid.NullUUID) (float64, error)`
  - `(*Storage).UpsertCartItem(ctx, cartID, productID uuid.UUID, qty int, price float64) error`
  - `(*Storage).SetCartItemQuantity(ctx, cartItemID, cartID uuid.UUID, qty int) error`
  - `(*Storage).DeleteCartItem(ctx, cartItemID, cartID uuid.UUID) error`
  - `(*Storage).ClearCartItems(ctx, cartID uuid.UUID) error`
  - `var ErrProductPriceNotFound error`, `var ErrCartItemNotFound error` — сентинел-ошибки, проверяются через `errors.Is` в сервисном слое (Задачи 9-10).

  Используются в Задачах 9-12.

- [ ] **Step 1: Написать carts.go**

```go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

var (
	ErrProductPriceNotFound = errors.New("product price not found")
	ErrCartItemNotFound     = errors.New("cart item not found")
)

type CartItemRow struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	SKU         string
	ProductName string
	Qty         int
	Price       float64
}

func (s *Storage) GetOrCreateCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.UUID) (uuid.UUID, error) {
	var cartID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id FROM carts WHERE user_id = $1 AND counterparty_id = $2
	`, userID, counterpartyID).Scan(&cartID)
	if err == nil {
		return cartID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("get cart: %w", err)
	}

	err = s.pool.QueryRow(ctx, `
		INSERT INTO carts (user_id, counterparty_id) VALUES ($1, $2) RETURNING id
	`, userID, counterpartyID).Scan(&cartID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create cart: %w", err)
	}
	return cartID, nil
}

func (s *Storage) GetCartItems(ctx context.Context, cartID uuid.UUID) ([]CartItemRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT ci.id, ci.product_id, p.sku, p.name, ci.quantity, ci.price
		FROM cart_items ci
		JOIN products p ON p.id = ci.product_id
		WHERE ci.cart_id = $1
		ORDER BY ci.id
	`, cartID)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	defer rows.Close()

	items := make([]CartItemRow, 0)
	for rows.Next() {
		var item CartItemRow
		var qty float64
		if err = rows.Scan(&item.ID, &item.ProductID, &item.SKU, &item.ProductName, &qty, &item.Price); err != nil {
			return nil, fmt.Errorf("scan cart item: %w", err)
		}
		item.Qty = int(qty)
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cart items: %w", err)
	}
	return items, nil
}

func (s *Storage) ResolveProductPrice(ctx context.Context, productID uuid.UUID, priceGroupID uuid.NullUUID) (float64, error) {
	var price float64
	err := s.pool.QueryRow(ctx, `
		SELECT price FROM product_prices
		WHERE product_id = $1
		  AND (price_group_id = $2 OR ($2::uuid IS NULL AND price_group_id IS NULL))
		ORDER BY valid_from DESC NULLS LAST
		LIMIT 1
	`, productID, priceGroupID).Scan(&price)
	if err == nil {
		return price, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("resolve product price by group: %w", err)
	}

	err = s.pool.QueryRow(ctx, `
		SELECT price FROM product_prices
		WHERE product_id = $1
		ORDER BY valid_from DESC NULLS LAST
		LIMIT 1
	`, productID).Scan(&price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrProductPriceNotFound
		}
		return 0, fmt.Errorf("resolve product price fallback: %w", err)
	}
	return price, nil
}

func (s *Storage) UpsertCartItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, qty int, price float64) error {
	cmdTag, err := s.pool.Exec(ctx, `
		UPDATE cart_items SET quantity = quantity + $3
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID, qty)
	if err != nil {
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	if cmdTag.RowsAffected() > 0 {
		return nil
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO cart_items (cart_id, product_id, quantity, price) VALUES ($1, $2, $3, $4)
	`, cartID, productID, qty, price)
	if err != nil {
		return fmt.Errorf("insert cart item: %w", err)
	}
	return nil
}

func (s *Storage) SetCartItemQuantity(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID, qty int) error {
	cmdTag, err := s.pool.Exec(ctx, `
		UPDATE cart_items SET quantity = $3 WHERE id = $1 AND cart_id = $2
	`, cartItemID, cartID, qty)
	if err != nil {
		return fmt.Errorf("update cart item quantity: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (s *Storage) DeleteCartItem(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID) error {
	cmdTag, err := s.pool.Exec(ctx, `DELETE FROM cart_items WHERE id = $1 AND cart_id = $2`, cartItemID, cartID)
	if err != nil {
		return fmt.Errorf("delete cart item: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrCartItemNotFound
	}
	return nil
}

func (s *Storage) ClearCartItems(ctx context.Context, cartID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	if err != nil {
		return fmt.Errorf("clear cart items: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/orders && go build ./internal/storage/...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Начать integration_test.go общей "шапкой" файла (build tag + тестовый пул)**

```go
//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 dbname=AMC sslmode=disable user=mbatimel password=mbatimel"
	}
	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		t.Skipf("postgres not reachable, skipping integration test: %v", err)
	}
	return pool
}
```

- [ ] **Step 4: Дописать в тот же файл тест корзины**

```go
func TestCartRoundTrip(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := context.Background()
	storage := New(pool)

	mustScan := func(sql string, args ...interface{}) uuid.UUID {
		var id uuid.UUID
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("fixture insert failed (%s): %v", sql, err)
		}
		return id
	}

	priceGroupID := mustScan(`INSERT INTO price_groups (code, name) VALUES ($1, 'test group') RETURNING id`, uuid.NewString())
	counterpartyID := mustScan(`INSERT INTO counterparties (name, price_group_id) VALUES ('test counterparty', $1) RETURNING id`, priceGroupID)
	userID := mustScan(`INSERT INTO users (email, counterparty_id) VALUES ($1, $2) RETURNING id`, uuid.NewString()+"@test.local", counterpartyID)
	categoryID := mustScan(`INSERT INTO categories (name, slug) VALUES ('test category', $1) RETURNING id`, uuid.NewString())
	unitID := mustScan(`INSERT INTO units (code, name) VALUES ($1, 'test unit') RETURNING id`, uuid.NewString())
	productID := mustScan(`INSERT INTO products (category_id, unit_id, sku, name, slug) VALUES ($1, $2, $3, 'test product', $4) RETURNING id`, categoryID, unitID, uuid.NewString(), uuid.NewString())
	if _, err := pool.Exec(ctx, `INSERT INTO product_prices (product_id, price_group_id, price_type, price) VALUES ($1, $2, 'base', 500)`, productID, priceGroupID); err != nil {
		t.Fatalf("insert product_price: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM cart_items WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM carts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM product_prices WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM units WHERE id = $1`, unitID)
		pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM price_groups WHERE id = $1`, priceGroupID)
	})

	resolvedGroupID, err := storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		t.Fatalf("GetCounterpartyPriceGroupID: %v", err)
	}
	if !resolvedGroupID.Valid || resolvedGroupID.UUID != priceGroupID {
		t.Fatalf("expected price group %s, got %+v", priceGroupID, resolvedGroupID)
	}

	price, err := storage.ResolveProductPrice(ctx, productID, resolvedGroupID)
	if err != nil {
		t.Fatalf("ResolveProductPrice: %v", err)
	}
	if price != 500 {
		t.Fatalf("expected price 500, got %v", price)
	}

	cartID, err := storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		t.Fatalf("GetOrCreateCart: %v", err)
	}

	if err = storage.UpsertCartItem(ctx, cartID, productID, 2, price); err != nil {
		t.Fatalf("UpsertCartItem (insert): %v", err)
	}
	if err = storage.UpsertCartItem(ctx, cartID, productID, 3, price); err != nil {
		t.Fatalf("UpsertCartItem (increment): %v", err)
	}

	items, err := storage.GetCartItems(ctx, cartID)
	if err != nil {
		t.Fatalf("GetCartItems: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 cart item after upsert twice, got %d", len(items))
	}
	if items[0].Qty != 5 {
		t.Fatalf("expected qty 5 (2+3), got %d", items[0].Qty)
	}

	if err = storage.SetCartItemQuantity(ctx, items[0].ID, cartID, 10); err != nil {
		t.Fatalf("SetCartItemQuantity: %v", err)
	}
	items, err = storage.GetCartItems(ctx, cartID)
	if err != nil {
		t.Fatalf("GetCartItems after update: %v", err)
	}
	if items[0].Qty != 10 {
		t.Fatalf("expected qty 10 after update, got %d", items[0].Qty)
	}

	if err = storage.DeleteCartItem(ctx, items[0].ID, cartID); err != nil {
		t.Fatalf("DeleteCartItem: %v", err)
	}
	items, err = storage.GetCartItems(ctx, cartID)
	if err != nil {
		t.Fatalf("GetCartItems after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 cart items after delete, got %d", len(items))
	}
}
```

- [ ] **Step 5: Запустить интеграционный тест против локального Postgres**

Run: `cd back/orders && go test -tags=integration ./internal/storage/postgres/... -run TestCartRoundTrip -v`
Expected: `PASS`. Если Postgres недоступен — тест сам себя скипнет (`SKIP`), это тоже нормальный результат при отсутствии локальной БД.

- [ ] **Step 6: Commit**

```bash
git add back/orders/internal/storage/postgres/carts.go back/orders/internal/storage/postgres/integration_test.go
git commit -m "feat(orders): storage layer for cart items with integration test"
```

---

### Task 7: Storage — объёмная скидка (discounts.go)

**Files:**
- Create: `back/orders/internal/storage/postgres/discounts.go`

**Interfaces:**
- Produces: `(*Storage).GetVolumeDiscountPercent(ctx, counterpartyID uuid.UUID, priceGroupID uuid.NullUUID, subtotal float64) (float64, error)` — `0, nil` если подходящих строк нет. Используется в Задачах 9-11.

- [ ] **Step 1: Написать файл**

```go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

func (s *Storage) GetVolumeDiscountPercent(ctx context.Context, counterpartyID uuid.UUID, priceGroupID uuid.NullUUID, subtotal float64) (float64, error) {
	var discountPercent float64
	err := s.pool.QueryRow(ctx, `
		SELECT discount_percent FROM volume_discounts
		WHERE (counterparty_id = $1 OR (counterparty_id IS NULL AND price_group_id = $2 AND $2::uuid IS NOT NULL))
		  AND min_order_amount <= $3
		ORDER BY (counterparty_id = $1) DESC, min_order_amount DESC
		LIMIT 1
	`, counterpartyID, priceGroupID, subtotal).Scan(&discountPercent)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("get volume discount: %w", err)
	}
	return discountPercent, nil
}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/orders && go build ./internal/storage/...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Commit**

```bash
git add back/orders/internal/storage/postgres/discounts.go
git commit -m "feat(orders): storage layer for volume discount lookup"
```

---

### Task 8: Storage — создание/список заказов (orders.go) + интеграционный тест

**Files:**
- Create: `back/orders/internal/storage/postgres/orders.go`
- Modify: `back/orders/internal/storage/postgres/integration_test.go`

**Interfaces:**
- Consumes: та же миграция из Задачи 1 (`orders.payment_status`, `orders_number_seq`, `order_status_history.payment_status`).
- Produces:
  - `type OrderItemInput struct { ProductID uuid.UUID; SKU, Name string; Quantity int; UnitPrice, DiscountPercent, VATRate, LineTotal float64 }`
  - `type CreateOrderParams struct { UserID, CounterpartyID, CartID, DeliveryAddressID, ContactID uuid.UUID; DeliveryMethod, Comment string; Subtotal, DiscountTotal, VATTotal, Total float64; Items []OrderItemInput }`
  - `type CreatedOrder struct { ID uuid.UUID; Number string; CreatedAt time.Time }`
  - `(*Storage).CreateOrder(ctx, params CreateOrderParams) (CreatedOrder, error)` — одна транзакция: insert `orders` + `order_items` + `order_status_history` + удаление `cart_items` корзины.
  - `type OrderRow struct { ID uuid.UUID; Number, Status, PaymentStatus, DeliveryMethod string; Subtotal, DiscountTotal, VATTotal, Total float64; CreatedAt time.Time }`
  - `type ListOrdersParams struct { CounterpartyID uuid.UUID; Status, PaymentStatus, Sort string; Limit, Offset int }`
  - `(*Storage).ListOrders(ctx, params ListOrdersParams) ([]OrderRow, int, error)` — второе значение — общее количество (для пагинации).
  - `type OrderItemRow struct { ID, ProductID uuid.UUID; SKU, Name string; Quantity int; UnitPrice, LineTotal float64 }`
  - `(*Storage).GetOrderItems(ctx, orderID uuid.UUID) ([]OrderItemRow, error)`
  - `type OrderDocumentRow struct { ID uuid.UUID; Type, Number, URL string; CreatedAt time.Time }`
  - `(*Storage).GetOrderDocumentsByOrderID(ctx, orderID uuid.UUID) ([]OrderDocumentRow, error)`

  Используются в Задачах 11 и 12.

- [ ] **Step 1: Написать orders.go**

```go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID       uuid.UUID
	SKU             string
	Name            string
	Quantity        int
	UnitPrice       float64
	DiscountPercent float64
	VATRate         float64
	LineTotal       float64
}

type CreateOrderParams struct {
	UserID            uuid.UUID
	CounterpartyID    uuid.UUID
	CartID            uuid.UUID
	DeliveryMethod    string
	DeliveryAddressID uuid.UUID
	ContactID         uuid.UUID
	Comment           string
	Subtotal          float64
	DiscountTotal     float64
	VATTotal          float64
	Total             float64
	Items             []OrderItemInput
}

type CreatedOrder struct {
	ID        uuid.UUID
	Number    string
	CreatedAt time.Time
}

func (s *Storage) CreateOrder(ctx context.Context, params CreateOrderParams) (CreatedOrder, error) {
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

	_, err = tx.Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, params.CartID)
	if err != nil {
		return CreatedOrder{}, fmt.Errorf("clear cart after order: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return CreatedOrder{}, fmt.Errorf("commit tx: %w", err)
	}

	return order, nil
}

type OrderRow struct {
	ID             uuid.UUID
	Number         string
	Status         string
	PaymentStatus  string
	DeliveryMethod string
	Subtotal       float64
	DiscountTotal  float64
	VATTotal       float64
	Total          float64
	CreatedAt      time.Time
}

type ListOrdersParams struct {
	CounterpartyID uuid.UUID
	Status         string
	PaymentStatus  string
	Limit          int
	Offset         int
	Sort           string
}

func (s *Storage) ListOrders(ctx context.Context, params ListOrdersParams) ([]OrderRow, int, error) {
	orderBy := "created_at DESC"
	switch params.Sort {
	case "created_at":
		orderBy = "created_at ASC"
	case "created_at desc":
		orderBy = "created_at DESC"
	case "total":
		orderBy = "total ASC"
	case "total desc":
		orderBy = "total DESC"
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT id, number, status, payment_status, delivery_method, subtotal, discount_total, vat_total, total, created_at
		FROM orders
		WHERE counterparty_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR payment_status = $3)
		ORDER BY %s
		LIMIT $4 OFFSET $5
	`, orderBy), params.CounterpartyID, params.Status, params.PaymentStatus, limit, params.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	items := make([]OrderRow, 0)
	for rows.Next() {
		var o OrderRow
		if err = rows.Scan(&o.ID, &o.Number, &o.Status, &o.PaymentStatus, &o.DeliveryMethod, &o.Subtotal, &o.DiscountTotal, &o.VATTotal, &o.Total, &o.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		items = append(items, o)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate orders: %w", err)
	}

	var total int
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM orders
		WHERE counterparty_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR payment_status = $3)
	`, params.CounterpartyID, params.Status, params.PaymentStatus).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	return items, total, nil
}

type OrderItemRow struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	SKU       string
	Name      string
	Quantity  int
	UnitPrice float64
	LineTotal float64
}

func (s *Storage) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]OrderItemRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, product_id, sku, name, quantity, unit_price, line_total
		FROM order_items WHERE order_id = $1 ORDER BY id
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	items := make([]OrderItemRow, 0)
	for rows.Next() {
		var item OrderItemRow
		var qty float64
		if err = rows.Scan(&item.ID, &item.ProductID, &item.SKU, &item.Name, &qty, &item.UnitPrice, &item.LineTotal); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		item.Quantity = int(qty)
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}
	return items, nil
}

type OrderDocumentRow struct {
	ID        uuid.UUID
	Type      string
	Number    string
	URL       string
	CreatedAt time.Time
}

func (s *Storage) GetOrderDocumentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]OrderDocumentRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.id, d.type, d.number, COALESCE(f.storage_key, ''), d.created_at
		FROM documents d
		LEFT JOIN files f ON f.id = d.file_id
		WHERE d.order_id = $1
		ORDER BY d.created_at
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order documents: %w", err)
	}
	defer rows.Close()

	items := make([]OrderDocumentRow, 0)
	for rows.Next() {
		var d OrderDocumentRow
		if err = rows.Scan(&d.ID, &d.Type, &d.Number, &d.URL, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order document: %w", err)
		}
		items = append(items, d)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order documents: %w", err)
	}
	return items, nil
}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/orders && go build ./internal/storage/...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Дописать в integration_test.go тест создания и листинга заказов**

Добавить в конец `back/orders/internal/storage/postgres/integration_test.go`:

```go
func TestCreateOrderAndListOrders(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := context.Background()
	storage := New(pool)

	mustScan := func(sql string, args ...interface{}) uuid.UUID {
		var id uuid.UUID
		if err := pool.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
			t.Fatalf("fixture insert failed (%s): %v", sql, err)
		}
		return id
	}

	priceGroupID := mustScan(`INSERT INTO price_groups (code, name) VALUES ($1, 'test group 2') RETURNING id`, uuid.NewString())
	counterpartyID := mustScan(`INSERT INTO counterparties (name, price_group_id) VALUES ('test counterparty 2', $1) RETURNING id`, priceGroupID)
	userID := mustScan(`INSERT INTO users (email, counterparty_id) VALUES ($1, $2) RETURNING id`, uuid.NewString()+"@test.local", counterpartyID)
	categoryID := mustScan(`INSERT INTO categories (name, slug) VALUES ('test category 2', $1) RETURNING id`, uuid.NewString())
	unitID := mustScan(`INSERT INTO units (code, name) VALUES ($1, 'test unit 2') RETURNING id`, uuid.NewString())
	productID := mustScan(`INSERT INTO products (category_id, unit_id, sku, name, slug) VALUES ($1, $2, $3, 'test product 2', $4) RETURNING id`, categoryID, unitID, uuid.NewString(), uuid.NewString())
	if _, err := pool.Exec(ctx, `INSERT INTO product_prices (product_id, price_group_id, price_type, price) VALUES ($1, $2, 'base', 1000)`, productID, priceGroupID); err != nil {
		t.Fatalf("insert product_price: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO volume_discounts (counterparty_id, price_group_id, min_order_amount, discount_percent) VALUES ($1, $2, 1000, 10)`, counterpartyID, priceGroupID); err != nil {
		t.Fatalf("insert volume_discount: %v", err)
	}

	var orderID uuid.UUID
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM order_status_history WHERE order_id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM order_items WHERE order_id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM orders WHERE id = $1`, orderID)
		pool.Exec(ctx, `DELETE FROM volume_discounts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM cart_items WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM carts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM product_prices WHERE product_id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, productID)
		pool.Exec(ctx, `DELETE FROM units WHERE id = $1`, unitID)
		pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, categoryID)
		pool.Exec(ctx, `DELETE FROM counterparty_addresses WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM counterparty_contacts WHERE counterparty_id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
		pool.Exec(ctx, `DELETE FROM counterparties WHERE id = $1`, counterpartyID)
		pool.Exec(ctx, `DELETE FROM price_groups WHERE id = $1`, priceGroupID)
	})

	discountPercent, err := storage.GetVolumeDiscountPercent(ctx, counterpartyID, uuid.NullUUID{UUID: priceGroupID, Valid: true}, 2000)
	if err != nil {
		t.Fatalf("GetVolumeDiscountPercent: %v", err)
	}
	if discountPercent != 10 {
		t.Fatalf("expected discount 10, got %v", discountPercent)
	}

	addressID, err := storage.InsertDeliveryAddress(ctx, counterpartyID, "delivery", "Test City, Test Street 1")
	if err != nil {
		t.Fatalf("InsertDeliveryAddress: %v", err)
	}
	contactID, err := storage.InsertContact(ctx, counterpartyID, "Test Contact", "+70000000000", "test@test.local")
	if err != nil {
		t.Fatalf("InsertContact: %v", err)
	}

	cartID, err := storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		t.Fatalf("GetOrCreateCart: %v", err)
	}

	created, err := storage.CreateOrder(ctx, CreateOrderParams{
		UserID:            userID,
		CounterpartyID:    counterpartyID,
		CartID:            cartID,
		DeliveryMethod:    "delivery",
		DeliveryAddressID: addressID,
		ContactID:         contactID,
		Comment:           "test order",
		Subtotal:          2000,
		DiscountTotal:     200,
		VATTotal:          396,
		Total:             2196,
		Items: []OrderItemInput{
			{
				ProductID:       productID,
				SKU:             "test-sku",
				Name:            "test product 2",
				Quantity:        2,
				UnitPrice:       1000,
				DiscountPercent: 10,
				VATRate:         22,
				LineTotal:       2000,
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	orderID = created.ID

	if created.Number == "" {
		t.Fatal("expected non-empty order number")
	}

	orderItems, err := storage.GetOrderItems(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrderItems: %v", err)
	}
	if len(orderItems) != 1 || orderItems[0].Quantity != 2 {
		t.Fatalf("expected 1 order item with qty 2, got %+v", orderItems)
	}

	docs, err := storage.GetOrderDocumentsByOrderID(ctx, orderID)
	if err != nil {
		t.Fatalf("GetOrderDocumentsByOrderID: %v", err)
	}
	if len(docs) != 0 {
		t.Fatalf("expected no documents for freshly created order, got %d", len(docs))
	}

	rows, total, err := storage.ListOrders(ctx, ListOrdersParams{CounterpartyID: counterpartyID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListOrders: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("expected 1 order in list, got total=%d rows=%d", total, len(rows))
	}
	if rows[0].ID != orderID || rows[0].Status != "new" || rows[0].PaymentStatus != "not_paid" {
		t.Fatalf("unexpected order row: %+v", rows[0])
	}
}
```

- [ ] **Step 4: Запустить оба интеграционных теста**

Run: `cd back/orders && go test -tags=integration ./internal/storage/postgres/... -v`
Expected: `PASS` для `TestCartRoundTrip` и `TestCreateOrderAndListOrders` (или `SKIP`, если Postgres недоступен).

- [ ] **Step 5: Commit**

```bash
git add back/orders/internal/storage/postgres/orders.go back/orders/internal/storage/postgres/integration_test.go
git commit -m "feat(orders): storage layer for order creation and listing with integration tests"
```

---

### Task 9: Service — чистая функция расчёта корзины + unit-тесты

**Files:**
- Create: `back/orders/internal/service/cart_calc.go`
- Create: `back/orders/internal/service/cart_calc_test.go`

**Interfaces:**
- Produces:
  - `type CartCalcItem struct { Qty int; Price float64 }`
  - `type CartTotals struct { Subtotal, DiscountTotal, VATTotal, Total float64 }`
  - `func sumLineTotals(items []CartCalcItem) float64`
  - `func calcCartTotals(subtotal float64, discountPercent float64, vatRatePercent float64) CartTotals`
  - `func round2(v float64) float64`

  Используются в Задачах 10-12.

- [ ] **Step 1: Написать failing-тесты для round2/sumLineTotals/calcCartTotals**

`back/orders/internal/service/cart_calc_test.go`:

```go
package service

import "testing"

func TestSumLineTotals(t *testing.T) {
	items := []CartCalcItem{
		{Qty: 2, Price: 100},
		{Qty: 3, Price: 50},
	}
	got := sumLineTotals(items)
	want := 350.0
	if got != want {
		t.Fatalf("sumLineTotals() = %v, want %v", got, want)
	}
}

func TestCalcCartTotals_NoDiscountNoVAT(t *testing.T) {
	totals := calcCartTotals(200, 0, 0)
	if totals.Subtotal != 200 || totals.DiscountTotal != 0 || totals.VATTotal != 0 || totals.Total != 200 {
		t.Fatalf("unexpected totals: %+v", totals)
	}
}

func TestCalcCartTotals_DiscountAndVAT(t *testing.T) {
	// subtotal=1000, скидка 10% -> 900, НДС 22% сверху -> 198, итого 1098
	totals := calcCartTotals(1000, 10, 22)
	if totals.Subtotal != 1000 {
		t.Fatalf("Subtotal = %v, want 1000", totals.Subtotal)
	}
	if totals.DiscountTotal != 100 {
		t.Fatalf("DiscountTotal = %v, want 100", totals.DiscountTotal)
	}
	if totals.VATTotal != 198 {
		t.Fatalf("VATTotal = %v, want 198", totals.VATTotal)
	}
	if totals.Total != 1098 {
		t.Fatalf("Total = %v, want 1098", totals.Total)
	}
}

func TestCalcCartTotals_RoundingToTwoDecimals(t *testing.T) {
	// subtotal=99.99, без скидки, НДС 22%: 99.99*0.22=21.9978 -> округление до 22.0
	totals := calcCartTotals(99.99, 0, 22)
	if totals.VATTotal != 22.0 {
		t.Fatalf("VATTotal = %v, want 22.0", totals.VATTotal)
	}
	if totals.Total != 121.99 {
		t.Fatalf("Total = %v, want 121.99", totals.Total)
	}
}
```

- [ ] **Step 2: Запустить тесты и убедиться, что падают из-за отсутствия реализации**

Run: `cd back/orders && go test ./internal/service/... -run TestCalcCartTotals -v`
Expected: FAIL (`undefined: calcCartTotals` / `undefined: CartCalcItem`).

- [ ] **Step 3: Написать реализацию**

`back/orders/internal/service/cart_calc.go`:

```go
package service

import "math"

type CartCalcItem struct {
	Qty   int
	Price float64
}

type CartTotals struct {
	Subtotal      float64
	DiscountTotal float64
	VATTotal      float64
	Total         float64
}

func sumLineTotals(items []CartCalcItem) float64 {
	var subtotal float64
	for _, item := range items {
		subtotal += float64(item.Qty) * item.Price
	}
	return round2(subtotal)
}

func calcCartTotals(subtotal float64, discountPercent float64, vatRatePercent float64) CartTotals {
	discountTotal := round2(subtotal * discountPercent / 100)
	afterDiscount := subtotal - discountTotal
	vatTotal := round2(afterDiscount * vatRatePercent / 100)
	total := round2(afterDiscount + vatTotal)

	return CartTotals{
		Subtotal:      round2(subtotal),
		DiscountTotal: discountTotal,
		VATTotal:      vatTotal,
		Total:         total,
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
```

- [ ] **Step 4: Запустить тесты, убедиться, что проходят**

Run: `cd back/orders && go test ./internal/service/... -run 'TestSumLineTotals|TestCalcCartTotals' -v`
Expected: `PASS` для всех четырёх тестов.

- [ ] **Step 5: Commit**

```bash
git add back/orders/internal/service/cart_calc.go back/orders/internal/service/cart_calc_test.go
git commit -m "feat(orders): pure cart totals calculation with unit tests"
```

---

### Task 10: Service — Storage-интерфейс, конструктор, корзина (GetCart/AddCartItem/UpdateCartItem/DeleteCartItem/ClearCart)

**Files:**
- Modify: `back/orders/internal/service/service.go`

**Interfaces:**
- Consumes: все методы `Storage` из Задач 5-8 (`postgres.CartItemRow`, `postgres.ErrProductPriceNotFound`, `postgres.ErrCartItemNotFound` и т.д.), `CartCalcItem`/`calcCartTotals`/`sumLineTotals` из Задачи 9, `config.Config.VATRate` из Задачи 2, `customErrors.BadRequestError` из Задачи 3, `models.Cart`/`CartItem` (с новыми полями) из Задачи 4.
- Produces: `NewOrdersApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient, vatRate float64) externalapi.OrdersAPI` — сигнатура меняется (добавлен `vatRate`), используется в Задаче 13 (`cmd/main.go`).

- [ ] **Step 1: Расширить интерфейс Storage и структуру service**

Заменить текущий блок в `back/orders/internal/service/service.go`:

```go
// Storage is implemented by internal/storage/postgres.Storage.
type Storage interface {
	GetCities(ctx context.Context) ([]postgres.City, error)
}
```

на:

```go
// Storage is implemented by internal/storage/postgres.Storage.
type Storage interface {
	GetCities(ctx context.Context) ([]postgres.City, error)

	GetUserCounterpartyID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetCounterpartyPriceGroupID(ctx context.Context, counterpartyID uuid.UUID) (uuid.NullUUID, error)
	InsertDeliveryAddress(ctx context.Context, counterpartyID uuid.UUID, addrType string, address string) (uuid.UUID, error)
	InsertContact(ctx context.Context, counterpartyID uuid.UUID, fullName string, phone string, email string) (uuid.UUID, error)

	GetOrCreateCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.UUID) (uuid.UUID, error)
	GetCartItems(ctx context.Context, cartID uuid.UUID) ([]postgres.CartItemRow, error)
	ResolveProductPrice(ctx context.Context, productID uuid.UUID, priceGroupID uuid.NullUUID) (float64, error)
	UpsertCartItem(ctx context.Context, cartID uuid.UUID, productID uuid.UUID, qty int, price float64) error
	SetCartItemQuantity(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID, qty int) error
	DeleteCartItem(ctx context.Context, cartItemID uuid.UUID, cartID uuid.UUID) error
	ClearCartItems(ctx context.Context, cartID uuid.UUID) error

	GetVolumeDiscountPercent(ctx context.Context, counterpartyID uuid.UUID, priceGroupID uuid.NullUUID, subtotal float64) (float64, error)

	CreateOrder(ctx context.Context, params postgres.CreateOrderParams) (postgres.CreatedOrder, error)
	ListOrders(ctx context.Context, params postgres.ListOrdersParams) ([]postgres.OrderRow, int, error)
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]postgres.OrderItemRow, error)
	GetOrderDocumentsByOrderID(ctx context.Context, orderID uuid.UUID) ([]postgres.OrderDocumentRow, error)
}
```

Заменить:

```go
type service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
}

func NewOrdersApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient) externalapi.OrdersAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
	}
}
```

на:

```go
type service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
	vatRate      float64
}

func NewOrdersApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient, vatRate float64) externalapi.OrdersAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
		vatRate:      vatRate,
	}
}
```

Добавить импорт `"errors"` (стандартная библиотека, для `errors.Is`) в блок `import`.

- [ ] **Step 2: Добавить общие хелперы**

Добавить в `service.go` (после конструктора):

```go
func (s *service) checkBuyerAccess(ctx context.Context, userID uuid.UUID) error {
	allowed, err := s.accessClient.CheckAccess(ctx, userID, RoleCodeBuyer)
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}
	if !allowed {
		return customErrors.ForbiddenError()
	}
	return nil
}

func (s *service) resolveCounterpartyID(ctx context.Context, userID uuid.UUID, clientID string) (uuid.UUID, error) {
	if clientID != "" {
		id, err := uuid.Parse(clientID)
		if err != nil {
			return uuid.Nil, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "clientID")
		}
		return id, nil
	}

	id, err := s.storage.GetUserCounterpartyID(ctx, userID)
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	if id == uuid.Nil {
		return uuid.Nil, customErrors.BadRequestError().AddCause("field", "clientID")
	}
	return id, nil
}

func (s *service) buildCart(ctx context.Context, userID uuid.UUID, counterpartyID uuid.UUID) (models.Cart, error) {
	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}

	rows, err := s.storage.GetCartItems(ctx, cartID)
	if err != nil {
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}

	priceGroupID, err := s.storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}

	calcItems := make([]CartCalcItem, 0, len(rows))
	items := make([]models.CartItem, 0, len(rows))
	for _, row := range rows {
		lineTotal := round2(float64(row.Qty) * row.Price)
		calcItems = append(calcItems, CartCalcItem{Qty: row.Qty, Price: row.Price})
		items = append(items, models.CartItem{
			ID:          row.ID.String(),
			CartID:      cartID.String(),
			ProductID:   row.ProductID.String(),
			SKU:         row.SKU,
			ProductName: row.ProductName,
			Qty:         row.Qty,
			Price:       row.Price,
			Total:       lineTotal,
		})
	}

	subtotal := sumLineTotals(calcItems)

	discountPercent, err := s.storage.GetVolumeDiscountPercent(ctx, counterpartyID, priceGroupID, subtotal)
	if err != nil {
		return models.Cart{}, customErrors.InternalServerError().SetOuterError(err)
	}

	totals := calcCartTotals(subtotal, discountPercent, s.vatRate)

	return models.Cart{
		ID:            cartID.String(),
		UserID:        userID.String(),
		ClientID:      counterpartyID.String(),
		Items:         items,
		Subtotal:      totals.Subtotal,
		DiscountTotal: totals.DiscountTotal,
		VAT:           totals.VATTotal,
		Total:         totals.Total,
	}, nil
}
```

- [ ] **Step 3: Заменить заглушки GetCart/AddCartItem/UpdateCartItem/DeleteCartItem/ClearCart**

Заменить:

```go
func (s *service) GetCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.GetCartResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) AddCartItem(ctx context.Context, userID uuid.UUID, clientID string, productID string, qty int) (response models.AddCartItemResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) UpdateCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string, qty int) (response models.UpdateCartItemResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) DeleteCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string) (response models.DeleteCartItemResponse, err error) {
	return response, customErrors.NotImplementedError()
}

func (s *service) ClearCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.ClearCartResponse, err error) {
	return response, customErrors.NotImplementedError()
}
```

на:

```go
func (s *service) GetCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.GetCartResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.GetCartResponse{Cart: cart}, nil
}

func (s *service) AddCartItem(ctx context.Context, userID uuid.UUID, clientID string, productID string, qty int) (response models.AddCartItemResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if qty <= 0 {
		return response, customErrors.BadRequestError().AddCause("field", "qty")
	}

	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return response, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "productID")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	priceGroupID, err := s.storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	price, err := s.storage.ResolveProductPrice(ctx, productUUID, priceGroupID)
	if err != nil {
		if errors.Is(err, postgres.ErrProductPriceNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "productID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.UpsertCartItem(ctx, cartID, productUUID, qty, price); err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.AddCartItemResponse{Cart: cart}, nil
}

func (s *service) UpdateCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string, qty int) (response models.UpdateCartItemResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if qty <= 0 {
		return response, customErrors.BadRequestError().AddCause("field", "qty")
	}

	itemUUID, err := uuid.Parse(cartItemID)
	if err != nil {
		return response, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "cartItemID")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.SetCartItemQuantity(ctx, itemUUID, cartID, qty); err != nil {
		if errors.Is(err, postgres.ErrCartItemNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "cartItemID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.UpdateCartItemResponse{Cart: cart}, nil
}

func (s *service) DeleteCartItem(ctx context.Context, userID uuid.UUID, clientID string, cartItemID string) (response models.DeleteCartItemResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	itemUUID, err := uuid.Parse(cartItemID)
	if err != nil {
		return response, customErrors.BadRequestError().SetOuterError(err).AddCause("field", "cartItemID")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.DeleteCartItem(ctx, itemUUID, cartID); err != nil {
		if errors.Is(err, postgres.ErrCartItemNotFound) {
			return response, customErrors.NotFoundError().AddCause("field", "cartItemID")
		}
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.DeleteCartItemResponse{Deleted: true, Cart: cart}, nil
}

func (s *service) ClearCart(ctx context.Context, userID uuid.UUID, clientID string) (response models.ClearCartResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.storage.ClearCartItems(ctx, cartID); err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cart, err := s.buildCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, err
	}

	return models.ClearCartResponse{Cleared: true, Cart: cart}, nil
}
```

- [ ] **Step 4: Собрать пакет (CreateOrder/ListOrders пока остаются заглушками — это нормально до Задач 11-12)**

Run: `cd back/orders && go build ./...`
Expected: успешная сборка без ошибок (заглушки `CreateOrder`/`ListOrders` всё ещё соответствуют интерфейсу `Storage`, поэтому пакет собирается).

- [ ] **Step 5: Commit**

```bash
git add back/orders/internal/service/service.go
git commit -m "feat(orders): implement cart handlers (get/add/update/delete/clear)"
```

---

### Task 11: Service — CreateOrder

**Files:**
- Modify: `back/orders/internal/service/service.go`

**Interfaces:**
- Consumes: всё из Задачи 10 (`resolveCounterpartyID`, `checkBuyerAccess`, `sumLineTotals`, `calcCartTotals`, `round2`), `postgres.CreateOrderParams`/`postgres.OrderItemInput`/`postgres.CreatedOrder` из Задачи 8.
- Produces: рабочий `CreateOrder`, используется транспортным слоем (уже сгенерирован, без изменений).

- [ ] **Step 1: Заменить заглушку CreateOrder**

Заменить:

```go
func (s *service) CreateOrder(ctx context.Context, userID uuid.UUID, clientID string, deliveryType string, deliveryAddress string, contactName string, phone string, email string, comment string) (response models.CreateOrderResponse, err error) {
	return response, customErrors.NotImplementedError()
}
```

на:

```go
func (s *service) CreateOrder(ctx context.Context, userID uuid.UUID, clientID string, deliveryType string, deliveryAddress string, contactName string, phone string, email string, comment string) (response models.CreateOrderResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}
	if deliveryType == "" {
		return response, customErrors.BadRequestError().AddCause("field", "deliveryType")
	}
	if deliveryAddress == "" {
		return response, customErrors.BadRequestError().AddCause("field", "deliveryAddress")
	}
	if contactName == "" {
		return response, customErrors.BadRequestError().AddCause("field", "contactName")
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	cartID, err := s.storage.GetOrCreateCart(ctx, userID, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	cartRows, err := s.storage.GetCartItems(ctx, cartID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}
	if len(cartRows) == 0 {
		return response, customErrors.BadRequestError().AddCause("field", "cart")
	}

	priceGroupID, err := s.storage.GetCounterpartyPriceGroupID(ctx, counterpartyID)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	calcItems := make([]CartCalcItem, 0, len(cartRows))
	for _, row := range cartRows {
		calcItems = append(calcItems, CartCalcItem{Qty: row.Qty, Price: row.Price})
	}
	subtotal := sumLineTotals(calcItems)

	discountPercent, err := s.storage.GetVolumeDiscountPercent(ctx, counterpartyID, priceGroupID, subtotal)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	totals := calcCartTotals(subtotal, discountPercent, s.vatRate)

	addressID, err := s.storage.InsertDeliveryAddress(ctx, counterpartyID, deliveryType, deliveryAddress)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	contactID, err := s.storage.InsertContact(ctx, counterpartyID, contactName, phone, email)
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	orderItems := make([]postgres.OrderItemInput, 0, len(cartRows))
	responseItems := make([]models.OrderItem, 0, len(cartRows))
	for _, row := range cartRows {
		lineTotal := round2(float64(row.Qty) * row.Price)
		orderItems = append(orderItems, postgres.OrderItemInput{
			ProductID:       row.ProductID,
			SKU:             row.SKU,
			Name:            row.ProductName,
			Quantity:        row.Qty,
			UnitPrice:       row.Price,
			DiscountPercent: discountPercent,
			VATRate:         s.vatRate,
			LineTotal:       lineTotal,
		})
		responseItems = append(responseItems, models.OrderItem{
			ProductID:   row.ProductID.String(),
			SKU:         row.SKU,
			ProductName: row.ProductName,
			Qty:         row.Qty,
			Price:       row.Price,
			Total:       lineTotal,
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
	})
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	order := models.Order{
		ID:              created.ID.String(),
		Number:          created.Number,
		UserID:          userID.String(),
		ClientID:        counterpartyID.String(),
		Items:           responseItems,
		Subtotal:        totals.Subtotal,
		DiscountTotal:   totals.DiscountTotal,
		VAT:             totals.VATTotal,
		Total:           totals.Total,
		DeliveryType:    deliveryType,
		DeliveryAddress: deliveryAddress,
		ContactName:     contactName,
		Phone:           phone,
		Email:           email,
		Comment:         comment,
		Status:          "new",
		PaymentStatus:   "not_paid",
		CreatedAt:       created.CreatedAt,
	}

	return models.CreateOrderResponse{Order: order}, nil
}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/orders && go build ./...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Commit**

```bash
git add back/orders/internal/service/service.go
git commit -m "feat(orders): implement CreateOrder"
```

---

### Task 12: Service — ListOrders

**Files:**
- Modify: `back/orders/internal/service/service.go`

**Interfaces:**
- Consumes: `postgres.ListOrdersParams`/`postgres.OrderRow`/`postgres.OrderItemRow`/`postgres.OrderDocumentRow` из Задачи 8, `resolveCounterpartyID`/`checkBuyerAccess` из Задачи 10.
- Produces: рабочий `ListOrders` (это и есть "история заказов" из исходного запроса).

- [ ] **Step 1: Заменить заглушку ListOrders**

Заменить:

```go
func (s *service) ListOrders(ctx context.Context, userID uuid.UUID, clientID string, status string, paymentStatus string, limit int, offset int, sort string) (response models.ListOrdersResponse, err error) {
	return response, customErrors.NotImplementedError()
}
```

на:

```go
func (s *service) ListOrders(ctx context.Context, userID uuid.UUID, clientID string, status string, paymentStatus string, limit int, offset int, sort string) (response models.ListOrdersResponse, err error) {
	if err = s.checkBuyerAccess(ctx, userID); err != nil {
		return response, err
	}

	counterpartyID, err := s.resolveCounterpartyID(ctx, userID, clientID)
	if err != nil {
		return response, err
	}

	rows, total, err := s.storage.ListOrders(ctx, postgres.ListOrdersParams{
		CounterpartyID: counterpartyID,
		Status:         status,
		PaymentStatus:  paymentStatus,
		Limit:          limit,
		Offset:         offset,
		Sort:           sort,
	})
	if err != nil {
		return response, customErrors.InternalServerError().SetOuterError(err)
	}

	orders := make([]models.Order, 0, len(rows))
	for _, row := range rows {
		itemRows, itemsErr := s.storage.GetOrderItems(ctx, row.ID)
		if itemsErr != nil {
			return response, customErrors.InternalServerError().SetOuterError(itemsErr)
		}
		docRows, docsErr := s.storage.GetOrderDocumentsByOrderID(ctx, row.ID)
		if docsErr != nil {
			return response, customErrors.InternalServerError().SetOuterError(docsErr)
		}

		items := make([]models.OrderItem, 0, len(itemRows))
		for _, item := range itemRows {
			items = append(items, models.OrderItem{
				ID:          item.ID.String(),
				OrderID:     row.ID.String(),
				ProductID:   item.ProductID.String(),
				SKU:         item.SKU,
				ProductName: item.Name,
				Qty:         item.Quantity,
				Price:       item.UnitPrice,
				Total:       item.LineTotal,
			})
		}

		docs := make([]models.OrderDocument, 0, len(docRows))
		for _, doc := range docRows {
			docs = append(docs, models.OrderDocument{
				ID:        doc.ID.String(),
				OrderID:   row.ID.String(),
				Type:      doc.Type,
				Name:      doc.Number,
				URL:       doc.URL,
				CreatedAt: doc.CreatedAt,
			})
		}

		orders = append(orders, models.Order{
			ID:            row.ID.String(),
			Number:        row.Number,
			ClientID:      counterpartyID.String(),
			Items:         items,
			Subtotal:      row.Subtotal,
			DiscountTotal: row.DiscountTotal,
			VAT:           row.VATTotal,
			Total:         row.Total,
			DeliveryType:  row.DeliveryMethod,
			Status:        row.Status,
			PaymentStatus: row.PaymentStatus,
			Documents:     docs,
			CreatedAt:     row.CreatedAt,
		})
	}

	return models.ListOrdersResponse{
		Items: orders,
		Pagination: models.Pagination{
			Limit:  limit,
			Offset: offset,
			Total:  total,
		},
	}, nil
}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/orders && go build ./...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: Commit**

```bash
git add back/orders/internal/service/service.go
git commit -m "feat(orders): implement ListOrders (order history)"
```

---

### Task 13: Обновить cmd/main.go и финальная проверка

**Files:**
- Modify: `back/orders/cmd/main.go`

**Interfaces:**
- Consumes: `NewOrdersApiService(logger, storage, accessClient, vatRate float64)` из Задачи 10, `cfg.VATRate` из Задачи 2.

- [ ] **Step 1: Передать VATRate в конструктор сервиса**

Найти в `back/orders/cmd/main.go` строку:

```go
svc := ordersService.NewOrdersApiService(log.Logger, postgresStorage, access)
```

Заменить на:

```go
svc := ordersService.NewOrdersApiService(log.Logger, postgresStorage, access, cfg.VATRate)
```

- [ ] **Step 2: Собрать весь сервис**

Run: `cd back/orders && go build ./...`
Expected: успешная сборка без ошибок.

- [ ] **Step 3: go vet по всему сервису**

Run: `cd back/orders && go vet ./...`
Expected: без предупреждений.

- [ ] **Step 4: Прогнать все unit- и integration-тесты**

Run:
```bash
cd back/orders
go test ./...
go test -tags=integration ./internal/storage/postgres/... -v
```
Expected: все `PASS` (интеграционные — `PASS` при доступном Postgres, иначе корректный `SKIP`).

- [ ] **Step 5: Собрать migrations-сервис ещё раз (после всех изменений в схеме)**

Run: `cd back/migrations && go build ./...`
Expected: успешная сборка без ошибок.

- [ ] **Step 6: Commit**

```bash
git add back/orders/cmd/main.go
git commit -m "feat(orders): wire VAT rate into service constructor"
```
