# Разделение регистрации в auth: ИП и физлицо

Дата: 2026-07-27

## Проблема

Сервис `auth` имеет одну ручку регистрации `SignUpUser` (`POST /v1/auth/signup`, поля: email, password, name, surename). Нужно заменить её двумя ручками:

1. Регистрация ИП/организации — полный набор реквизитов.
2. Регистрация физлица — упрощённый набор.

Старая ручка удаляется полностью (не остаётся как deprecated).

## Контракт

### `RegisterIP` — `POST /v1/auth/register/ip`

```go
RegisterIP(
    ctx context.Context,
    email string,
    password string,
    fullName *string,
    shortName *string,
    inn *string,
    kpp *string,
    ogrn *string,
    okved *string,
    taxSystem *string,
    legalAddress *string,
    actualAddress *string,
    directorFullName *string,
    directorPosition *string,
    phone *string,
    additionalPhone *string,
    website *string,
    bankAccount *string,
    bankName *string,
    bankBik *string,
    correspondentAccount *string,
) (userID uuid.UUID, err error)
```

`email`/`password` обязательны (400 `ValidationError` если пусто после `TrimSpace`). Остальные поля — `*string`, `nil` = не передано (в БД пишется `NULL`, а не пустая строка).

### `RegisterIndividual` — `POST /v1/auth/register/individual`

```go
RegisterIndividual(
    ctx context.Context,
    fio string,
    phone string,
    email string,
    deliveryAddress string,
    password string,
    city string,
    inn *string,
) (userID uuid.UUID, err error)
```

`fio, phone, email, deliveryAddress, password, city` обязательны (400 при пустом после `TrimSpace`). `inn` — `*string`, опционален.

`fio` бьётся по пробелам (`strings.Fields`) → `surename` (1-е слово), `name` (2-е слово), `middle_name` (остальное через пробел). Меньше 2 слов — остаток в `surename`, `name`/`middle_name` пустые.

## Схема БД

Новая миграция `back/migrations/pkg/migrations/data/20260727120000_registration_fields.sql`.

`counterparties` — добавить колонки (переиспользуются существующие `name, inn, kpp, ogrn, legal_address, actual_address, type, status`):
- `short_name VARCHAR(255)`
- `okved VARCHAR(255)`
- `tax_system VARCHAR(255)`
- `website VARCHAR(255)`
- `director_full_name VARCHAR(255)`
- `director_position VARCHAR(255)`
- `phone VARCHAR(255)`
- `additional_phone VARCHAR(255)`
- `email VARCHAR(255)`
- `bank_account VARCHAR(255)`
- `bank_name VARCHAR(255)`
- `bank_bik VARCHAR(255)`
- `correspondent_account VARCHAR(255)`

При регистрации ИП: `type='ip'`, `status='new'`.

`users` — добавить:
- `inn VARCHAR(255)`
- `city VARCHAR(255)`
- `delivery_address TEXT`
- `FOREIGN KEY (counterparty_id) REFERENCES counterparties(id)`
- `INDEX idx_users_counterparty_id ON users(counterparty_id)`

`down`-миграция дропает всё симметрично.

## Логика сервиса

`internal/service/service.go`:
- Удалить `SignUpUser`, `defaultSignUpRoleCode` (если не используется больше нигде — переиспользуется в новых методах как `RoleCodeBuyer`).
- `RegisterIP`: валидация email/password → bcrypt-хеш пароля → `storage.CreateIPUser(...)`.
- `RegisterIndividual`: валидация обязательных полей → split `fio` → bcrypt-хеш пароля → `storage.CreateIndividualUser(...)`.
- Обе роли — `RoleCodeBuyer` (как в старом `SignUpUser`).

`Storage`-интерфейс: убрать `CreateUserWithRole`, добавить:
```go
CreateIPUser(ctx context.Context, email, passwordHash string, fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress, directorFullName, directorPosition, phone, additionalPhone, website, bankAccount, bankName, bankBik, correspondentAccount *string, roleCode int) (uuid.UUID, error)

CreateIndividualUser(ctx context.Context, email, passwordHash, surename, name, middleName, phone, city, deliveryAddress string, inn *string, roleCode int) (uuid.UUID, error)
```

`internal/storage/postgres/postgres.go`:
- Убрать `CreateUserWithRole`, `sqlInsertUser`/`insertUser.sql`.
- `CreateIPUser` — транзакция: `INSERT INTO counterparties(...) RETURNING id` → `INSERT INTO users(email, password, counterparty_id, status) VALUES (..., 'active') RETURNING id` (unique violation на email → `ErrEmailTaken`) → `GetRoleByCode` → `INSERT INTO user_roles` → commit.
- `CreateIndividualUser` — транзакция: `INSERT INTO users(email, password, surename, name, middle_name, phone, city, delivery_address, inn, status) VALUES (..., 'active') RETURNING id` (unique violation → `ErrEmailTaken`) → `GetRoleByCode` → `INSERT INTO user_roles` → commit.
- Новые SQL-файлы: `sql/insertCounterparty.sql`, `sql/insertUserForIP.sql`, `sql/insertIndividualUser.sql`. `*string` параметры передаются в pgx как есть (nil → NULL).

`internal/errors/common.go`:
```go
ValidationError = func(field string) *Error {
    return New("validation failed", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("field", field)
}
```
(`ErrInvalidRequest` уже объявлена константа, ранее не использовалась).

## Транспорт

`pkg/interfaces/externalAPI/interface.go` — обновляется (не генерённый файл, источник для `@tg`). Файлы под `internal/transport/jsonRPC/externalapi/*` и `swaggers/externalapi/swagger.yaml` — **не трогать**, пользователь регенерирует/допилит их сам через `tg` (бинарь недоступен в этом окружении).

`internal/transport/custom-handlers/auth.go` — заменить функцию `SignUpUser` на `RegisterIP`/`RegisterIndividual` по тому же шаблону (defer-логирование, вызов `svc.RegisterIP`/`svc.RegisterIndividual`, `sendResponse`). Это сейчас нигде не вызывается (main.go подключает generated-транспорт напрямую), но приводится в соответствие по аналогии с остальными хендлерами в файле.

## Edge cases

- Email занят → `EmailTakenError` (409).
- Пустое обязательное поле → `ValidationError(field)` (400).
- `RoleCodeBuyer` не найден в `roles` → `ErrRoleNotFound` → 500.
- Ошибка на любом шаге транзакции → rollback, ничего не пишется.
- IP-регистрация со всеми опциональными полями = nil — не падает, создаёт минимальную запись.

## Не в скоупе

- Регенерация generated-transport и swagger — делает пользователь сам.
- Фронтенд — не использует старую ручку, изменений не требует.
- `custom-handlers/auth.go` не подключён к рантайму — правится только по аналогии, без wiring.
