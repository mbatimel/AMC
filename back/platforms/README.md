# platforms

Сервис интеграции с маркетплейсами. Отвечает за создание карточек товаров во внешних кабинетах продавца:

- Wildberries — `POST /api/v1/wildberries/cards`
- Ozon — `POST /api/v1/ozon/cards`
- Яндекс Маркет — `POST /api/v1/yandex-market/cards`

Структура повторяет остальные сервисы монорепо (см. `products`, `integrations`):

- `pkg/interfaces/externalapi` — контракт API (`@tg`-аннотации), из него кодогенерацией (`tg`) строится транспорт;
- `pkg/models` — DTO внешнего API;
- `internal/service` — бизнес-логика, валидация, доступ;
- `internal/wildberries`, `internal/ozon`, `internal/yandexmarket` — клиенты внешних API маркетплейсов;
- `internal/transport/custom-handlers` — обработчики HTTP-ответов;
- `internal/transport/jsonRPC/externalapi`, `swaggers/externalapi` — генерируются командой `go generate ./pkg/interfaces/externalapi/...`.

Реализация каждого клиента (`CreateCard`) дорабатывается по мере разбора конкретной ручки с реальным контрактом маркетплейса.

## Wildberries: создание карточки

`CreateWildberriesCard` создаёт отдельную (не объединённую) карточку — один `subjectID` и один вариант — через `POST /content/v2/cards/upload`. Требует `productID` — идентификатор товара в каталоге `products`, чтобы после успешной отправки карточки отметить товар статусом `pending` на маркетплейсе.

Создание в WB асинхронно (до 30 минут на синхронизацию), поэтому успешный ответ означает только то, что запрос принят, а не что карточка опубликована.

Флаг «товар добавлен в маркетплейс» хранится не в `platforms`, а в самом `products` (колонки `wb_status`/`ozon_status`/`ym_status`) — `platforms` выставляет его через внутренний API `products` (`internal/clients/products.go` → `products/pkg/interfaces/internalAPI`, бинарь `products-internal`, порт `:8091`). Простановка статуса — best-effort: если карточка на маркетплейсе уже принята, ошибка простановки флага не проваливает ручку, а только логируется.

## Ozon: создание/обновление карточки

`CreateOzonCard` создаёт/обновляет один товар (один элемент `items[]`) через `POST /v3/product/import`. Как и WB — асинхронно: ответ содержит только `task_id` задания (статус проверяется отдельным методом `/v1/product/import/info`, вне рамок этой ручки).

Обязательные объёмно-весовые характеристики (`depth/width/height/weight`) валидируются на нуль — Ozon отклоняет карточку при нулевых значениях. `price`/`old_price`/`vat` — строки, как того требует Ozon API. Авторизация — заголовки `Client-Id` + `Api-Key`.

Вне scope v1 (можно добавить отдельными ручками при необходимости): объединение карточек по атрибуту 9048, `complex_attributes` (видео/видеообложка), `pdf_list`, `promotions`, батч нескольких товаров за один вызов (наша ручка — всегда один товар за раз).

## Yandex Market: создание/обновление карточки

`CreateYandexMarketCard` создаёт/обновляет один товар (один элемент `offerMappings[]`) через `POST /v2/businesses/{businessId}/offer-mappings/update`. В отличие от WB/Ozon — **синхронно**: статус и предупреждения/ошибки по товару приходят сразу в ответе, отдельного метода опроса статуса не требуется.

Обязательные поля (по документации): `offerId`, `name`, `marketCategoryId`, `pictures`, `vendor`, `description`. Если Маркет вернул ошибку хотя бы по одному товару — весь запрос считается непринятым (`ErrValidation`), даже если формально HTTP-статус 200. Предупреждения (`warnings`) не блокируют приём и пробрасываются в ответ ручки. Авторизация — заголовок `Api-Key` (`businessId` — идентификатор кабинета, конфигурируется через `YANDEX_MARKET_BUSINESS_ID`, не путать с `campaignId` магазина). Нестандартный код `420` — лимит запросов (аналог 429).

Вне scope v1: `mapping.marketSku` (явная привязка к готовой карточке Маркета), устаревшие `params`/`category`/`customsCommodityCode`, `condition` (уценка), `age`/`adult`, документы/сертификаты, мультиязычные `name`/`description` (параметр `language`), батч нескольких товаров за один вызов.
