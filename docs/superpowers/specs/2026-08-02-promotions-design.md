# Акции (Promotions) — дизайн

## Контекст

Сейчас "Акции" в админке — статичная контентная страница (`content.promo`, тип `ListPageContent`: title/text/items), без связи с товарами и ценой. Нужна реальная сущность акции: название, период (дата+время), выбранные товары, процент скидки. Скидка должна реально пересчитывать цену в каталоге и в корзине/заказе для всех пользователей (не только визуальный зачёркнутый ценник).

У товара уже есть собственное постоянное поле `discount_percent` (ручная скидка, задаётся при редактировании товара, хранится в `product_prices` через `computeClientPrice`). Акция должна с ним сосуществовать, а не заменять полностью.

## Решения (согласованы с пользователем)

- Скидка акции **реально меняет цену** в каталоге и в корзине/заказе (не просто маркетинговый показ).
- Если товар попадает в несколько пересекающихся по периоду акций — берётся **максимальный процент** среди активных.
- Взаимодействие с ручной скидкой товара: эффективная скидка = **`GREATEST(discount_percent товара, MAX(discount_percent активных акций по товару))`**.
- Период акции — **дата + время** (не только дата), точность до минуты.
- Сущность `Promotion` хранится в **products-сервисе** (там же, где цена и `discount_percent` товара).
- Название акции **не выводится** отдельно на витрине (front) — только автоматический пересчёт цены. Название используется только в админке.
- Статус акции (scheduled / active / ended) **не хранится** как поле — вычисляется на лету по текущему времени и `starts_at`/`ends_at`.
- Акция действует **от порога количества**, который задаётся **отдельно для каждого товара внутри акции** (напр.: товар A — от 10 шт, товар B — от 50 шт, в рамках одной акции). Порог — не накопительная шкала, а простое условие "если в позиции заказа `qty >= min_qty` — скидка акции применяется целиком, иначе не применяется вовсе".
- Каталог/карточка товара (products-сервис, `getProductByID`/`listProducts`) **не учитывает акции** — показывает обычную цену (как для qty=1), т.к. вне корзины нет контекста количества. Пересчёт с учётом акции происходит **только в корзине/заказе** (orders-сервис), когда известно фактическое `qty` позиции.
- Тесты пишутся по TDD, до реализации.

## Данные

Новая миграция в `back/migrations/pkg/migrations/data/` (products-сервис):

```sql
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    discount_percent NUMERIC NOT NULL CHECK (discount_percent >= 0 AND discount_percent <= 100),
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL CHECK (ends_at > starts_at),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE promotion_products (
    promotion_id UUID NOT NULL REFERENCES promotions(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    min_qty INTEGER NOT NULL DEFAULT 1 CHECK (min_qty >= 1),
    PRIMARY KEY (promotion_id, product_id)
);
CREATE INDEX idx_promotion_products_product_id ON promotion_products(product_id);
CREATE INDEX idx_promotions_period ON promotions(starts_at, ends_at);
```

Активность акции по товару в SQL: `now() BETWEEN starts_at AND ends_at AND :qty >= min_qty`.

## API (products-сервис)

Тот же codegen-паттерн `tg transport` (аннотации `@tg` в `back/products/pkg/interfaces/externalapi/interface.go`), что и у существующих `CreateProduct`/`ListProducts`.

Новый тип запроса `PromotionProductInput{ ProductID uuid.UUID; MinQty int }` — по одному порогу количества на каждый товар акции.

Новые методы:

- `CreatePromotion(ctx, userID, name, discountPercent, startsAt, endsAt, products []models.PromotionProductInput) (Promotion, error)` — `POST /v1/promotions`
- `UpdatePromotion(ctx, userID, promotionID, name, discountPercent, startsAt, endsAt, products []models.PromotionProductInput) (Promotion, error)` — `PATCH /v1/promotions/:promotionID` (список товаров и порогов заменяется целиком)
- `DeletePromotion(ctx, userID, promotionID uuid.UUID) error` — `DELETE /v1/promotions/:promotionID`
- `GetPromotion(ctx, promotionID uuid.UUID) (Promotion, error)` — `GET /v1/promotions/:promotionID`
- `ListPromotions(ctx, limit, offset *int) (ListPromotionsResponse, error)` — `GET /v1/promotions`

Модель ответа `Promotion`: `id, name, discount_percent, starts_at, ends_at, status (scheduled|active|ended), products []{product_id, min_qty}, created_at, updated_at`.

Валидация (по аналогии с `validateRequired`/`validateLength` в `service.go`):
- `name`: обязателен, ≤255 символов;
- `discountPercent`: 0–100;
- `endsAt` > `startsAt`;
- `products`: непустой массив, у каждого `min_qty >= 1`, все `product_id` существуют (`ErrProductNotFound` при отсутствии);
- доступ — тот же `checkWriteAccess`, что у `CreateProduct`.

## Расчёт цены

### products-сервис (каталог/карточка) — без изменений

`getProductByID.sql` и `listProducts.sql` **не трогаем**. Акции зависят от количества в позиции заказа, которого нет в контексте каталога (там всегда фактически qty=1), поэтому каталог продолжает показывать цену с учётом только ручной `discount_percent` товара, как сейчас.

### orders-сервис — единственное место расчёта акций

`ResolveProductPrice` (`back/orders/internal/storage/postgres/carts.go`) меняет сигнатуру: добавляется параметр `qty int`.

```go
func (s *Storage) ResolveProductPrice(ctx context.Context, productID uuid.UUID, priceGroupID uuid.NullUUID, qty int) (float64, error)
```

Логика: после того как определена базовая цена и ручная скидка (как сейчас), добавляется подзапрос к `promotion_products`/`promotions` (та же БД, что и products-сервис):

```sql
SELECT MAX(pr.discount_percent)
FROM promotion_products pp
JOIN promotions pr ON pr.id = pp.promotion_id
WHERE pp.product_id = $1
  AND pp.min_qty <= $2
  AND now() BETWEEN pr.starts_at AND pr.ends_at
```

Эффективная скидка = `GREATEST(ручная discount_percent, promo_max или 0)`, итоговая цена позиции = `base_price * (1 - effective_discount/100)`, округление — как в существующей `computeClientPrice`. Выносится в переиспользуемую Go-функцию `calcEffectiveDiscount(manualDiscount, promoDiscount *float64) float64`, покрытую тестом.

Оба вызова `ResolveProductPrice` (`service.go:290` — добавление товара в корзину, `service.go:788` — пересчёт корзины по сохранённым позициям) уже имеют доступ к `qty`/`item.Quantity` — просто пробрасываются новым аргументом.

### Тесты (TDD, до реализации)

- Go unit-тест на `calcEffectiveDiscount`: нет акций, одна акция (qty выше/ниже порога), несколько акций с разными `min_qty` и процентами (в т.ч. пересекающиеся периоды — берётся максимум среди тех, где порог пройден).
- Тест на SQL-выборку "активных на дату и достаточных по qty" акций (граничные случаи: `qty == min_qty`, `qty == min_qty - 1`, ровно в `starts_at`/`ends_at`, до/после периода) — интеграционный тест по образцу `back/orders/internal/storage/postgres/integration_test.go`.
- Существующий `cart_calc_test.go` дополняется кейсом, где `ResolveProductPrice` возвращает промо-цену при достижении порога и обычную — при недостижении.

## Админка (admin-front)

Новый пункт меню "Акции" в `admin-front/src/views/Admin/lib/nav.ts` (в группе "Каталог", рядом с "Товары") — по факту заменяет текущую ссылку на контентную страницу `getContentPath('promo')` в блоке "Контент" (статичный контент акций там был плейсхолдером; реальные акции теперь отдельный раздел).

Новая страница `AdminPromotionsPage.tsx` по образцу `AdminBannersPage.tsx`:
- список акций: название, период, %, статус (scheduled/active/ended), количество товаров;
- форма создания/редактирования: название, % скидки, дата+время начала и конца, мультиселект товаров (переиспользуется `fetchAllProductsRequest`/поиск, как в `AdminProductPage.tsx`) — **для каждого выбранного товара отдельное поле "мин. количество"** (числовой инпут рядом с товаром в списке выбранных);
- удаление акции с подтверждением.

Новый файл `admin-front/src/core/shared/api/promotions.ts` с `fetchPromotionsRequest`, `createPromotionRequest`, `updatePromotionRequest`, `deletePromotionRequest` — по паттерну `admin-front/src/core/shared/api/products.ts` (`fetch` на `/api/v1/promotions...`).

## Вне рамок (явно не делаем)

- Бейдж/название акции на витрине (front) — только пересчитанная цена.
- Ручной статус активности (pause/resume) независимо от дат — статус целиком управляется периодом.
- Изменение существующей архитектуры `product_prices`/`ResolveProductPrice` сверх необходимого для GREATEST-подзапроса.
