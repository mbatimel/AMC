# Заказ ↔ 1С — дизайн

## Контекст

Сейчас `orders`-сервис создаёт заказ (`CreateOrder`) полностью автономно, статус
всегда `'new'`, и никуда не сообщает. Нужно, чтобы при создании заказ уходил в
1С синхронно, и чтобы 1С мог по вебхуку двигать статус заказа дальше (когда
доставлен).

Схема БД под это уже частично готова (миграция `20260705171941_orders.sql`,
писалась заранее под будущую интеграцию): `orders.one_c_guid UUID UNIQUE` и
`orders.synced_to_1c_at TIMESTAMPTZ` уже существуют и не используются. Также
уже есть `counterparties.one_c_guid`, `products.one_c_guid` — их проставляет
вчерашний воркер `onec-sync` (см.
`docs/superpowers/specs/2026-08-25-onec-sync-worker-design.md`). Новых миграций
на `orders`/`products`/`counterparties` эта работа не требует.

Смежный документ: `docs/superpowers/specs/2026-08-26-onec-orders-integration-1c-contract.md`
— контракт двух HTTP-ручек для 1С-разработчика (что должен поднять 1С, и как
дёргать наш вебхук). Этот файл описывает только сторону AMC.

## Решения (согласованы с пользователем)

- Стейт-машина заказа: `new` (в момент создания, ещё не закоммичен) →
  `processing` (после успешной синхронной отправки в 1С) → `delivered` (по
  вебхуку 1С). `cancelled` — только клиент/админ (`CancelOrder`, как сейчас),
  1С этот статус выставить не может.
- Отправка "заказ создан" в 1С — **синхронно**, внутри запроса `CreateOrder`.
  Если 1С недоступна/ответила ошибкой — **весь заказ не создаётся**, клиент
  видит ошибку и может повторить попытку.
- Вебхук от 1С живёт в `integrations` (том же Go-модуле, что вчерашний воркер
  `onec-sync`), не в `orders` — `integrations` уже держит креды/протокол к 1С.
- Вне объёма (открытый вопрос, не решаем сейчас): нужно ли уведомлять 1С,
  когда заказ отменяет клиент/админ. 1С про `cancelled` явно не спрашивает по
  вебхуку, но не обсуждали, нужно ли слать 1С исходящее уведомление при
  отмене. Оставлено как TODO на будущий спек.

## Одна транзакция вместо компенсирующего отката

Альтернатива, которую сознательно не выбрали: закоммитить заказ в статусе
`'new'`, отдельно дёрнуть 1С, и при неудаче — компенсирующий hard delete
заказа второй транзакцией. Отклонено, потому что `CreateOrder` уже сегодня
чистит корзину (`DELETE FROM cart_items`) в той же транзакции, что вставляет
заказ. Если бы заказ коммитился до похода в 1С, а поход упал — пришлось бы
либо терять содержимое корзины клиента (плохой UX: заново собирать корзину
при ретрае), либо восстанавливать `cart_items` во второй транзакции
(лишняя сложность и гонки, если клиент уже начал менять корзину).

Вместо этого — **одна транзакция держится открытой на весь вызов**:
`storage.CreateOrder` вставляет заказ+items+history+чистит корзину (как
сегодня), затем — **до commit** — вызывает переданный колбэк, который уходит
в `integrations` за пределы транзакции по сети. Колбэк успешен → тем же tx
обновляем `status='processing'`, `one_c_guid`, `synced_to_1c_at`, добавляем
`order_status_history`, коммит. Колбэк упал → транзакция целиком
откатывается (заказ не существует, корзина не тронута, клиент может просто
нажать "оформить" ещё раз).

Осознанный компромисс: транзакция висит на время HTTP-вызова к `integrations`
(таймаут ~15с, конфигурируемый) — держит соединение из пула и блокировки
вставленных строк на этот срок. Для B2B-опта с невысоким QPS оформления
заказов это приемлемо; если в будущем нагрузка вырастет — можно вернуться к
двухфазной схеме с компенсацией.

```go
// back/orders/internal/storage/postgres/orders.go
func (s *Storage) CreateOrder(
    ctx context.Context,
    params CreateOrderParams,
    pushToOnec func(ctx context.Context, orderID uuid.UUID, orderNumber string) (onecGUID uuid.UUID, onecNumber string, err error),
) (CreatedOrder, error) {
    tx, err := s.pool.Begin(ctx)
    // ... вставка order/order_items/order_status_history('new')/DELETE cart_items — как сегодня ...

    onecGUID, onecNumber, pushErr := pushToOnec(ctx, order.ID, order.Number)
    if pushErr != nil {
        return CreatedOrder{}, fmt.Errorf("push order to onec: %w", pushErr) // defer tx.Rollback уже стоит
    }

    if _, err = tx.Exec(ctx, `
        UPDATE orders SET status = 'processing', one_c_guid = $1, synced_to_1c_at = now() WHERE id = $2
    `, onecGUID, order.ID); err != nil { ... }
    if _, err = tx.Exec(ctx, `
        INSERT INTO order_status_history (order_id, old_status, new_status, payment_status, changed_by, comment)
        VALUES ($1, 'new', 'processing', 'not_paid', NULL, $2)
    `, order.ID, "Отправлен в 1С, документ "+onecNumber); err != nil { ... }

    return order, tx.Commit(ctx)
}
```

`changed_by` — `NULL` для этого перехода (колонка уже nullable сегодня):
это системный переход, а не действие конкретного пользователя, отдельный
"системный" аккаунт заводить не нужно (в отличие от обратного направления —
см. ниже).

## AMC → 1С: `PushOrder`

`service.CreateOrder` (сегодня — `back/orders/internal/service/service.go:~470-550`)
до вызова `storage.CreateOrder` уже резолвит `counterpartyID`, товары корзины
и их цены. Перед вызовом `storage.CreateOrder` строим payload и передаём
колбэком:

- Товары корзины уже даны по `product_id` (внутренний UUID) — нужен доп.
  `SELECT one_c_guid, sku FROM products WHERE id = ANY($1)` (новый метод
  `Storage.GetProductOnecRefs`), т.к. `orders`-сервис уже читает таблицу
  `products` напрямую в других местах (`ResolveProductPrice`) — тот же
  паттерн, не новая связность.
- Контрагент — `SELECT one_c_guid, inn, name FROM counterparties WHERE id = $1`
  (аналогично, `GetCounterpartyPriceGroupID` уже читает эту таблицу).
- **Если у товара/контрагента ещё нет `one_c_guid`** (новый клиент/товар,
  которого 1С ещё не видела) — в payload всё равно уходят естественные ключи
  (SKU у товара; ИНН+название у контрагента), `onec_guid` — пустая строка.
  Решение (явно, не оставляем TBD): 1С обязан смэтчить/завести сущность по
  естественному ключу сама — это часть контракта, см. 1C-спеку. AMC не ждёт
  обратно созданный GUID контрагента/товара — только GUID и номер документа
  заказа.

Клиент `internal/onecorders/client.go` в модуле `back/integrations`
(отдельный от `internal/onec/client.go`, который используется дневным
воркером и говорит по OData — здесь другой протокол, кастомный HTTP-service
1С, см. 1C-контракт), Basic Auth отдельным техпользователем.

## integrations: новый процесс `onec-orders-api`

Рядом с существующим `cmd/main.go` (воркер `onec-sync`, batch, тикер,
без публичных HTTP-ручек) — **второй бинарник**
`back/integrations/cmd/onec-orders-api/main.go`: постоянный HTTP-сервер
(request/response, не batch), отдельный docker-сервис и образ, тот же Go-модуль
(общие `internal/storage/postgres`, `internal/config`). Два процесса вместо
одного, потому что у них разный lifecycle (тикер vs request/response) — как и
остальные сервисы в репо (один `cmd` = один бинарник = один docker-сервис).

**Важное ограничение `tg` (проверено экспериментально: пробный запуск
`tg transport` на интерфейсе с параметром-структурой и с `[]struct` ломает
кодогенерацию — тип параметра теряется, `tg` умеет только скалярные типы и
слайсы скалярных типов).** `PushOrder` несёт массив позиций заказа (SKU/qty/
price на каждую) — это не выражается как `tg`-аргумент. Поэтому только
`OnecOrderStatusWebhook` (все поля — скаляры) идёт через `tg`; `PushOrder`
регистрируется вручную как fiber-роут на ТОМ ЖЕ `*fiber.App`, что отдаёт
`tg`-сгенерированный сервер (`Server.Fiber()` — публичный метод
сгенерированного `Server`, уже используется в `orders/cmd/main.go` через
`app.Fiber().Handler()`) — один процесс, один порт, оба пути.

Интерфейс `back/integrations/pkg/interfaces/internalAPI/interface.go` —
`@tg`-аннотированный, по образцу `back/access/pkg/interfaces/internalAPI`
(codegen транспорта через `go:generate tg transport`; `tg client` для этого
интерфейса не нужен — вызывающая сторона (1С) не Go-клиент):

```go
// @tg http-prefix=/api
type OnecOrdersAPI interface {
    // OnecOrderStatusWebhook — вызывает 1С снаружи, публикуется в nginx
    // (`/api/v1/onec/`).
    // @tg http-method=POST
    // @tg http-path=/v1/onec/orders/status
    // @tg http-headers=apiKey|X-Onec-Api-Key
    // @tg http-args=clientOrderID|clientOrderID
    // @tg http-args=status|status
    // @tg http-args=onecDocumentNumber|onecDocumentNumber
    // @tg http-args=comment|comment
    OnecOrderStatusWebhook(ctx context.Context, apiKey string, clientOrderID uuid.UUID, status string, onecDocumentNumber string, comment string) (ok bool, err error)
}
```

`PushOrder` (ручной fiber-хендлер, `internal/transport/http/onecorders.go`):
`POST /api/v1/onec-orders/push`, тело — JSON (`encoding/json`, обычный
`fiber.Ctx.BodyParser`), не публикуется в nginx (недостижим снаружи docker-
сети `amc_net`). Как и у `orders→access` (`CheckAccess` без какого-либо
auth-заголовка, см. сгенерированный клиент access) — дополнительной
авторизации поверх сетевой изоляции не добавляем, тот же уровень доверия,
что у остальных внутренних вызовов сервис-к-сервису в этом репо.

На стороне `orders` — свой маленький HTTP-клиент
`back/orders/internal/onecclient/client.go` (fasthttp, JSON, без auth —
симметрично серверной стороне), а не `tg`-клиент — по той же причине
(массив items).

`PushOrder`: строит запрос к 1С через `internal/onecorders/client.go`,
логирует попытку в `sync_jobs`/`sync_logs` (`direction='outbound'`,
`entity_type='order_create'` — переиспользуем таблицы из
`20260705171945_integration_sync.sql`, новый метод
`Storage.CreateIntegrationJob(ctx, systemID, direction, entityType string)`,
не трогая существующий `CreateSyncJob`, у которого эти два поля захардкожены
под дневной синк), возвращает GUID+номер документа 1С.

`OnecOrderStatusWebhook`:
1. Проверяет `apiKey` == `ONEC_WEBHOOK_API_KEY` (статический ключ из конфига,
   без этого — 401).
2. Валидирует `status` по allow-list (сейчас единственное валидное значение —
   `"delivered"`; попытка передать `cancelled`/`new`/`processing`/что угодно
   ещё — 400, запись в `sync_logs` `level=warn`). Список расширяется без
   новых ручек — правкой этой карты и спеки.
3. Читает текущий статус заказа — новая узкая admin-гейтед ручка
   `GetOrderStatus(ctx, userID, orderID uuid.UUID) (status string, err error)`
   в `orders` (не переиспользуем buyer-ориентированный `GetOrder` — он резолвит
   `counterpartyID` по связке "пользователь→клиент", что не имеет смысла для
   системного аккаунта без привязанных клиентов; `GetOrderStatus` гейтится
   тем же `checkAdminAccess`, что и `UpdateOrderStatus`, отдаёт только статус)
   — **до** применения перехода, потому что
   `UpdateOrderStatus` сам по себе не проверяет допустимость перехода (это
   admin-ручка для произвольных ручных правок статуса, не state machine):
   - заказ уже `delivered` — `ok=true`, no-op, повторной записи в историю не
     создаём (1С может ретраить вебхук при сетевых сбоях на своей стороне);
   - заказ `cancelled` — **409**, не применяем (это терминальный статус
     клиента/админа, 1С не может его переопределить — заказ мог быть отменён
     уже после того, как 1С начала его собирать);
   - заказ `processing` — применяем переход (шаг 4);
   - заказ не найден — 404.
4. Вызывает уже существующий `PATCH /api/v1/admin/orders/status`
   (`orders.UpdateOrderStatus`) через сгенерированный `tg`-клиент
   `orders/pkg/client/transport` (добавляем orders'у
   `//go:generate tg client --services . --outPath ../../client/transport -go`,
   как у `access`, и один раз генерируем) — **от имени зарезервированного
   системного admin-пользователя**, не от имени "никого". Такой пользователь
   заводится сид-миграцией (`users`+`access_roles`, роль
   `RoleCodeAdmin`, фиксированный UUID), UUID кладём в конфиг
   `ORDERS_SYSTEM_USER_ID` у `onec-orders-api`. Нужен, потому что
   `UpdateOrderStatus` уже сегодня требует `checkAdminAccess(ctx, userID)` —
   это осознанно НЕ обходим новым бэкдор-эндпоинтом в `orders`, а
   переиспользуем существующую проверку прав с настоящим (хоть и системным)
   аккаунтом.

Опознаём заказ по `client_order_id` = наш `orders.id`, который мы сами же
передали 1С в `PushOrder` — 1С обязан ЭХОм вернуть его в каждом вебхуке.
Обратного lookup'а по `one_c_guid` не нужно (не заводим для этого отдельную
публичную ручку в `orders`).

## Деплой

- `docker-compose.yml`: новый сервис `onec-orders-api` (порт, `depends_on:
  migrations, access`), env: `ONEC_ORDERS_BASE_URL`, `ONEC_ORDERS_USER`,
  `ONEC_ORDERS_PASSWORD`, `ONEC_ORDERS_REQUEST_TIMEOUT` (default `15s`),
  `ONEC_WEBHOOK_API_KEY`, `ORDERS_URL`, `ORDERS_SYSTEM_USER_ID`. `orders`
  получает `INTEGRATIONS_URL` (адрес `onec-orders-api` по `amc_net`).
- `deploy/nginx/conf.d/wk.amctechgroup.ru.conf`: новый
  `location /api/v1/onec/ { proxy_pass http://onec-orders-api:PORT/api/v1/onec/; }`.
  `/v1/onec-orders/push` **не публикуется** — недостижим снаружи.
- `deploy/.env.example`: добавить новые переменные.
- Сид-миграция системного пользователя — в `back/migrations`.

## Ошибки / edge cases

- 1С недоступна/таймаут/5xx на `PushOrder` — вся транзакция создания заказа
  откатывается (см. выше), клиент видит ошибку, может повторить.
- 1С ответила успехом, но AMC не успела получить ответ (таймаут именно на
  чтении ответа, не на самом создании документа в 1С) — редкий, но реальный
  случай двойного заказа при ретрае клиента. Смягчается требованием к 1С:
  `PushOrder` идемпотентен по `client_order_id` (см. 1C-контракт) — повторный
  запрос с тем же `client_order_id` не создаёт второй документ, возвращает
  тот же GUID/номер. Полностью не устраняет (у AMC при таймауте всё равно
  откатится СВОЙ заказ и уйдёт новый `client_order_id` при ретрае клиента) —
  отмечено как известный, принятый риск, не блокирующий v1.
- Вебхук с неизвестным `client_order_id` (заказа с таким `orders.id` нет) —
  404 от `UpdateOrderStatus` (уже возвращается сегодня), `onec-orders-api`
  логирует в `sync_logs` `level=error`, отдаёт 1С 404.
- Вебхук с недопустимым `status` — 400, `sync_logs` `level=warn`, заказ не
  трогается.
- Повторный вебхук с тем же статусом (заказ уже `delivered`) — 200 `ok=true`,
  без повторной записи в `order_status_history`.
- Вебхук `delivered` для заказа в статусе `cancelled` — 409, заказ не
  трогается, `sync_logs` `level=warn`.

## Тестирование

- `internal/onecorders/client_test.go` — `httptest.Server`, мок ответа 1С
  (успех/timeout/5xx), проверка формирования payload.
- `internal/service` (или где будет лежать логика вебхука) — unit-тест
  allow-list статусов, идемпотентности повторного вебхука, мок
  `orders`-клиента.
- `back/orders/internal/storage/postgres/integration_test.go` — новый кейс:
  `CreateOrder` с колбэком, который возвращает ошибку → заказ, items,
  history, `cart_items` не сохраняются (реальный Postgres, откат проверяется
  через прямой `SELECT count(*)`).
- `back/orders/internal/service` — unit-тест: колбэк подставной, проверка что
  при успехе статус в ответе — `processing`, при ошибке — заказ не создан,
  ошибка проброшена клиенту.

## Что не входит в этот спек

- Уведомление 1С при отмене заказа AMC-стороной (клиент/админ) — открытый
  вопрос, см. "Решения" выше.
- Расширение статусной модели дальше `delivered` (например промежуточные
  "собран"/"отгружен") — добавляется расширением allow-list в
  `OnecOrderStatusWebhook`, без новых ручек; пока не запрошено бизнесом.
- Правила мэтчинга контрагента/товара по естественному ключу на стороне
  1С (по ИНН/названию/SKU) — сама логика мэтчинга/создания — ответственность
  1С-разработчика, AMC только передаёт данные (см. 1C-контракт).
