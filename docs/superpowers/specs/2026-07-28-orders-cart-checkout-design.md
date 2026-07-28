
# Корзина, оформление заказа, история заказов (orders service)

Дата: 2026-07-28
Сервис: `back/orders`

## Контекст

В `back/orders/pkg/interfaces/externalapi/interface.go` уже описан весь контракт
`OrdersAPI` (cart CRUD, orders CRUD) и сгенерирован транспортный слой
(`internal/transport/jsonRPC/externalapi`, `internal/transport/custom-handlers`,
`swaggers/externalapi/swagger.yaml`). Все методы сервиса в
`internal/service/service.go` сейчас — заглушки, возвращающие
`customErrors.NotImplementedError()`.

Нужно реализовать бизнес-логику (сервисный слой + storage/postgres) для трёх
пользовательских сценариев:

1. **Корзина** — просмотр текущей корзины с постатейным и итоговым расчётом.
2. **Оформление заказа** — создание заказа из корзины.
3. **История заказов** — список заказов клиента с кратким резюме по каждому.

Интерфейс менять не нужно — все нужные методы там уже объявлены. Меняются
только:

- `back/orders/internal/service/service.go` (реализация методов)
- `back/orders/internal/storage/postgres/*` (новые запросы + `Storage` интерфейс
  в service.go)
- `back/migrations/pkg/migrations/data/*.sql` (новая миграция)
- `back/orders/internal/config/config.go` + `.env`/`deploy/.env.example`/`deploy/docker-compose.yml` (ставка НДС, `NDS_VALUE`)

## Скоуп

Реализуются:

- `GetCart`
- `AddCartItem`, `UpdateCartItem`, `DeleteCartItem`, `ClearCart` — обязательный
  минимум, иначе корзину нечем наполнить и `CreateOrder` не из чего собирать.
- `CreateOrder`
- `ListOrders` (это и есть "история заказов" — список заказов клиента с
  резюме по каждому; НЕ `GetOrderHistory`, который возвращает лог смены
  статусов ОДНОГО заказа по `orderID`)

Не реализуются (остаются `NotImplementedError`, вне запроса пользователя):

- `GetOrder`, `CancelOrder`, `RepeatOrder`, `GetOrderDocuments`,
  `GetOrderHistory`, `UpdateOrderStatus`

Не входит в скоуп:

- Генерация самих документов (счёт/накладная/УПД). Поле "документ" в истории
  заказов просто отражает то, что реально лежит в таблице `documents` —
  для только что созданных заказов это будет пустой список, т.к. ничего их
  сейчас не создаёт.
- Синхронизация с 1С (`one_c_guid`, `synced_to_1c_at`) — поля остаются `NULL`.
- Жёсткая валидация `deliveryType` по enum — принимается любая непустая
  строка (самовывоз/доставка/транспортная компания и т.п.), без справочника.

## Миграция БД

Новый файл `back/migrations/pkg/migrations/data/20260728120000_orders_payment.sql`.

Причина: `orders` не имеет колонки для статуса оплаты, хотя история заказов
должна её возвращать; `order_status_history` тоже не может её трекать.

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

Номер заказа генерируется в SQL при вставке:
`'AMC-' || to_char(now(), 'YYYYMMDD') || '-' || lpad(nextval('orders_number_seq')::text, 5, '0')`,
например `AMC-20260728-00001`. Значение атомарно (nextval), гонок при
параллельном создании заказов нет.

Статусы (строковые константы, без отдельного справочника — по аналогии с
остальными `VARCHAR(255) status` полями в схеме):

- `orders.status`: `new` при создании (жизненный цикл остальных статусов вне
  скоупа — их проставляет `UpdateOrderStatus`, который не реализуется).
- `orders.payment_status`: `not_paid` при создании.

## Конфигурация

`back/orders/internal/config/config.go`: новое поле `VATRate float64`,
читается из переменной окружения `NDS_VALUE` (fallback `"22"`), парсится
`strconv.ParseFloat`. Добавляется в `back/orders/.env` (локальная разработка),
`deploy/.env.example` и в блок `orders.environment` в `deploy/docker-compose.yml`
(там уже прокидываются `PG_*`/`ACCESS_INTERNAL_ADDRESS` для этого сервиса).

## Разрешение клиента (counterparty) по запросу

Во всех новых ручках есть header `X-User-Id` (userID) и опциональный query
`clientID` (уже в интерфейсе, тип `string`).

Правило (одно на все ручки): если `clientID` непустой — он используется как
`counterparty_id` напрямую (`uuid.Parse`, ошибка парсинга →
`BadRequestError`). Если пустой — резолвится из `users.counterparty_id` по
`userID` (новый запрос `Storage.GetUserCounterpartyID`). Если у пользователя
`counterparty_id IS NULL` и `clientID` не передан — `BadRequestError`
("пользователь не привязан к контрагенту").

Доступ: перед выполнением бизнес-логики — `accessClient.CheckAccess(ctx,
userID, RoleCodeBuyer)`, по образцу `back/auth/internal/service/service.go`
(`err != nil` → `InternalServerError().SetOuterError(err)`, `!allowed` →
`customErrors.ForbiddenError()`).

## Расчёт корзины (используется в GetCart и CreateOrder)

Для каждой позиции корзины:

- `line_total = qty * price` (`price` — цена, зафиксированная в `cart_items`
  на момент `AddCartItem`, не пересчитывается на лету при просмотре).

Итоги:

- `subtotal = Σ line_total`
- Скидка — `volume_discounts`: строка выбирается по
  `(counterparty_id = X) OR (counterparty_id IS NULL AND price_group_id = <группа контрагента>)`,
  `min_order_amount <= subtotal`, приоритет — сначала персональные строки
  контрагента, затем максимальный подходящий `min_order_amount`. Если строк
  нет — скидка `0`.
  `discount_amount = subtotal * discount_percent / 100`
- НДС считается сверху от суммы после вычета скидки, по ставке из `NDS_VALUE`:
  `vat_amount = (subtotal - discount_amount) * VATRate / 100`
- `total_to_pay = (subtotal - discount_amount) + vat_amount`

`CreateOrder` считает то же самое и замораживает результат в колонках
`orders.subtotal / discount_total / vat_total / total`.

## Резолюция цены при AddCartItem

Цена товара при добавлении в корзину берётся из `product_prices`:

```sql
SELECT price FROM product_prices
WHERE product_id = $1
  AND (price_group_id = $2 OR ($2 IS NULL AND price_group_id IS NULL))
ORDER BY valid_from DESC NULLS LAST
LIMIT 1
```

где `$2` — `price_group_id` контрагента (из `counterparties.price_group_id`
по резолвленному `counterparty_id`). Если строка не найдена — берём любую
строку по `product_id` тем же способом сортировки (без фильтра по группе),
как деградацию. Если и её нет — `NotFoundError` ("цена на товар не найдена").

`AddCartItem`: если для `(cart_id, product_id)` уже есть строка в
`cart_items` — увеличиваем `quantity` на переданный `qty` (не заменяем);
иначе — вставляем новую строку. Корзина (`carts`) находится или создаётся
по `(user_id, counterparty_id)` (get-or-create).

`UpdateCartItem` — прямая замена `quantity` на переданное значение (`qty <= 0`
→ `BadRequestError`, для удаления позиции есть отдельный `DeleteCartItem`).

## Изменения в моделях (pkg/models/orders.go)

Пользователь явно просил в ответе корзины: сумму всех позиций (subtotal) и
скидку — отдельно от итоговой суммы к оплате. Сейчас `models.Cart` имеет
только `Total`/`VAT`, этого не хватает. Добавляю два поля (аддитивно,
JSON-совместимо, сигнатур методов интерфейса это не касается):

```go
type Cart struct {
    ...
    Subtotal      float64 `json:"subtotal"`       // сумма всех позиций до скидки
    DiscountTotal float64 `json:"discount_total"`  // сумма скидки
    Total         float64 `json:"total"`           // к оплате после скидки + НДС
    VAT           float64 `json:"vat"`             // сумма НДС (входит в Total)
    ...
}
```

Для симметрии (те же данные уже есть в колонках `orders.subtotal` /
`orders.discount_total`) те же два поля добавляются и в `models.Order`, чтобы
`CreateOrder`/`ListOrders` тоже могли их отдавать без отдельного решения в
будущем.

## GetCart

Вход: `userID` (header), `clientID` (query, опционально).
Логика: resolve counterparty → get-or-create cart → список `cart_items` (join
`products` для `name`/`sku`) → расчёт по формуле выше → `models.Cart` с
заполненными `Items`, `Subtotal`, `DiscountTotal`, `VAT`, `Total` (=
`total_to_pay`).

## CreateOrder

Вход: `userID` (header), `clientID` (query), `deliveryType`, `deliveryAddress`,
`contactName`, `phone`, `email`, `comment` (query/body, как уже объявлено в
интерфейсе).

Шаги (одна транзакция):

1. `CheckAccess` (buyer).
2. Resolve counterparty.
3. Прочитать `cart_items` корзины; если пусто — `BadRequestError` ("корзина
   пуста").
4. Посчитать subtotal/discount/vat/total (формула выше).
5. Вставить `counterparty_addresses` (`type = deliveryType`, `address =
   deliveryAddress`, `counterparty_id`) → получить `delivery_address_id`.
6. Вставить `counterparty_contacts` (`full_name = contactName`, `phone`,
   `email`, `counterparty_id`) → получить `contact_id`.
7. Сгенерировать `number` (`nextval('orders_number_seq')`, см. миграцию).
8. Вставить `orders` (`status='new'`, `payment_status='not_paid'`,
   `delivery_method=deliveryType`, `delivery_address_id`, `contact_id`,
   `comment`, `subtotal`, `discount_total`, `vat_total`, `total`, `user_id`,
   `counterparty_id`).
9. Скопировать `cart_items` → `order_items` (product_id, sku, name, quantity,
   unit_price=price, discount_percent=применённый % скидки, vat_rate=VATRate,
   line_total).
10. Вставить начальную строку `order_status_history` (`old_status=NULL,
    new_status='new', payment_status='not_paid', changed_by=userID`).
11. Удалить все `cart_items` корзины (очистка после оформления).
12. Вернуть `models.CreateOrderResponse{Order: ...}` с `Number` и остальными
    полями (включая `Subtotal`/`DiscountTotal`/`VAT`/`Total`), заполненными
    из вставленных данных.

## ListOrders (история заказов)

Вход: `userID` (header), `clientID` (query), `status`, `paymentStatus`,
`limit`, `offset`, `sort` (уже объявлены в интерфейсе).

Логика: resolve counterparty → выборка `orders` по `counterparty_id` (+
фильтры `status`/`paymentStatus`, если непустые) с `LIMIT`/`OFFSET` и
`ORDER BY` (белый список `sort`: `created_at` / `created_at desc` /
`total` / `total desc`, всё остальное → `created_at desc` по умолчанию) +
отдельный `COUNT(*)` для `Pagination.Total`.

Для каждого найденного заказа — подзапросы: `order_items` (полностью, чтобы
`len(Items)` дал "количество позиций" без изменения модели `Order`) и
`documents` (join по `order_id`, обычно пусто для новых заказов).

`Order.PaymentStatus`, `Order.DeliveryType` (= `delivery_method`),
`Order.Status`, `Order.Total`, `Order.CreatedAt`, `Order.Number` — прямые
поля из таблицы `orders`.

## Обработка ошибок

Используются существующие `customErrors` (`back/orders/internal/errors`):
`BadRequestError`-эквивалент (в `common.go` его сейчас нет — нужно добавить
`BadRequestError = func() *Error { return New("bad request",
fasthttp.StatusBadRequest, ErrBadRequest) }`, код `ErrBadRequest` уже объявлен
в `errors.go`, просто не использовался), `ForbiddenError`, `NotFoundError`,
`InternalServerError` — по аналогии с существующим `GetCities`.

## Тестирование

Given/When/Then вручную не гоняем (нет тестовой БД в текущем окружении) —
но пишутся unit-тесты на чистую расчётную функцию (subtotal → discount →
vat → total), т.к. она не требует БД и содержит основную бизнес-логику,
которую проще всего сломать. Остальное (SQL-запросы, транзакции) проверяется
компиляцией + `go vet`; ручной прогон через реальный Postgres — по
возможности, если локальная БД поднята.
