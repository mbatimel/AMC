# Проверка ИНН через api-fns.ru в RegisterIP — дизайн

## Контекст

`service.RegisterIP` (`back/auth/internal/service/service.go`) уже проверяет ИНН локально по алгоритму контрольной суммы (`validate()` в `common.go`). Нужна дополнительная проверка через внешний сервис [api-fns.ru](https://api-fns.ru/api/fl_status) (`GET /api/fl_status?inn=...&key=...`), который возвращает актуальный статус ИНН из ФНС.

## Решения (согласованы с пользователем)

- Проверяется только блок `Корректность` ответа (`КонтрСумма`, `Недействительный`). Блоки `Самозанятость`, `ИП`, `ДисквЛицо`, `Банкрот`, `ПоддержкаМСП` — не парсятся и не используются.
- ИНН считается невалидным, если `КонтрСумма == false` ИЛИ `Недействительный == true`.
- Если `КонтрСумма` или `Недействительный` в ответе `null` (по документации — "ошибка, повторите запрос позднее"), либо запрос упал (таймаут/сеть/HTTP не 200/битый JSON) — **fail-closed**: `RegisterIP` возвращает ошибку, регистрация не проходит.
- Внешний вызов идёт **после** локальной checksum-проверки (`validate()`) — незачем дёргать внешний API для заведомо невалидного ИНН.
- Адрес и ключ уже есть в `back/auth/.env` (`API_FNS_ADDR`, `API_FNS_KEY`).
- **Полный сырой ответ API логируется всегда** (zerolog, уровень `Info`), независимо от результата — успех, невалидный ИНН, HTTP-ошибка, таймаут, битый JSON. Логируется тело ответа как есть (raw body), плюс ИНН и HTTP-статус.

## Компоненты

**`internal/config/config.go`** — добавить в `Config`: `FnsAddr`, `FnsKey`, читаются из `API_FNS_ADDR` / `API_FNS_KEY`. Fatal, если пустые (как остальные обязательные настройки).

**`internal/client/fns` (новый пакет)** — HTTP-клиент к api-fns.ru:

```go
type Client struct {
    addr   string
    key    string
    http   *fasthttp.Client // timeout 5s
    logger zerolog.Logger
}

func New(addr, key string, logger zerolog.Logger) *Client
func (c *Client) CheckIndividual(ctx context.Context, inn string) (valid bool, err error)
```

Внутри `CheckIndividual`: GET-запрос с query `inn` и `key`. Сразу после получения ответа (успех или ошибка транспорта) — `c.logger.Info().Str("inn", inn).Int("status", statusCode).Str("response", string(rawBody)).Msg("fns fl_status response")` (при сетевой ошибке/таймауте — `rawBody` пуст, логируется факт ошибки отдельным полем). Затем парсинг JSON только в блок `Корректность` (приватная структура с полями `КонтрСумма *bool`, `Недействительный *bool`). Non-200 статус, сетевая ошибка, невалидный JSON, либо `nil` в одном из двух полей — возвращает `err != nil`.

**`service.go`** — новый интерфейс, по образцу `AccessClient`:

```go
type FnsClient interface {
    CheckIndividual(ctx context.Context, inn string) (valid bool, err error)
}
```

Поле `fnsClient FnsClient` в `service`, новый параметр в `NewAuthApiService`.

**`RegisterIP`** — после существующей проверки `validate(*inn)`:

```go
valid, err := s.fnsClient.CheckIndividual(ctx, *inn)
if err != nil {
    return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
}
if !valid {
    return uuid.Nil, customErrors.InnInvalidError(*inn)
}
```

**`errors/common.go`** — новая ошибка:

```go
InnInvalidError = func(inn string) *Error {
    return New("inn is invalid", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("inn", inn)
}
```

(код `ErrInvalidRequest` уже есть, переиспользуется — как у `ValidationError`).

**`cmd/main.go`** — создать `fnsClient := fns.New(cfg.FnsAddr, cfg.FnsKey, log.Logger)`, передать в `authService.NewAuthApiService(...)`.

## Ошибки/edge cases

- Пустой `*inn` — уже обрабатывается раньше (`InnEmptyErr`), до вызова API не доходит.
- ИНН с пробелами/дефисами — `validate()` уже чистит их перед checksum-проверкой; в API уходит очищенная строка.
- Таймаут запроса — 5s, фиксированный (не вынесен в конфиг, YAGNI).
- Логирование полного ответа — включая случаи, когда JSON не парсится: логируется raw body как строка перед попыткой парсинга, поэтому лог не зависит от успеха парсинга.

## Тестирование

- Юнит-тест на `fns.Client.CheckIndividual`: подменный HTTP-сервер (`httptest.Server`) — валидный ответ, `Недействительный=true`, `КонтрСумма=false`, `null`-поля, не-200, битый JSON.
- Юнит-тест на `RegisterIP` с моком `FnsClient` (интерфейс) — valid/invalid/error-ветки.
