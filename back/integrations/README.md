# 1C Integration

Сервис синхронизации данных с 1С (УТ 10.3) в БД `products`. Периодически загружает номенклатуру, цены и остатки.

## Компоненты

**onec-sync** — воркер синхронизации (отдельный контейнер в `deploy/docker-compose.yml`, сервис `onec-sync`).

## Переменные окружения

Обязательные:

- `ONEC_BASE_URL` — базовый URL OData-сервиса 1С (например, `http://PVISERVER/UT/odata/standard.odata`)
- `ONEC_USER` — техпользователь для доступа к OData
- `ONEC_PASSWORD` — пароль техпользователя
- `PG_DB` — имя БД `products`
- `PG_USER` — пользователь PostgreSQL
- `PG_PASSWORD` — пароль PostgreSQL

Опциональные:

- `PG_HOST` (по умолчанию `localhost`)
- `PG_PORT` (по умолчанию `5432`)
- `SYNC_INTERVAL` (по умолчанию `24h`)

## Запуск

Через Docker Compose (из корня репозитория, после создания `deploy/.env`):

```sh
docker compose --env-file deploy/.env -f deploy/docker-compose.yml up onec-sync
```

Или локально:

```sh
cd back/integrations
ONEC_BASE_URL=... ONEC_USER=... ONEC_PASSWORD=... PG_DB=amc PG_USER=amc PG_PASSWORD=secret \
go run ./cmd
```

## Мониторинг

Логи синхронизации сохраняются в таблицу `sync_logs` (БД `products`). Каждый запуск создаёт запись в `sync_jobs` с результатом (`success`/`partial_failure`/`failure`).

## Рекомендуемый путь (исследование доступа)

Возможные каналы доступа к 1С:

1. **OData standard interface** (рекомендуется)
   - Дозволяется в конфиге 1С: Конфигуратор → Стандартные интерфейсы → OData (v4)
   - Требует: IIS на сервере, галка "Стандартный интерфейс OData" в Конфигураторе, техпользователь на чтение
   - Адрес: `http://<server>/<ib>/odata/standard.odata`
   - Аутентификация: Basic Auth (user:password в header)
   - Преимущества: стандартный протокол, не зависит от схемы БД, простая фильтрация через `$filter`

2. **HTTP-сервис** (если OData недоступен)
   - Пишется отдельно в 1С, медленнее, требует синхронизации разработки
   - Не рекомендуется

3. **Прямое подключение к БД 1С** (не делаем)
   - Риск для продакшена, сложное именование таблиц, бинарные UUID — не подходит

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
