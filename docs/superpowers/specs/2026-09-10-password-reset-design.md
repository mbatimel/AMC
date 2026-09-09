# Смена пароля с подтверждением по почте — дизайн и контракт

Статус: **черновик**, контракт для фронта, реализация не начата.

## Зачем

Пользователь нажимает кнопку "Сбросить/сменить пароль" → на почту
приходит письмо со ссылкой → ссылка открывает окно на фронте с формой
(старый пароль + новый пароль) → форма шлёт запрос на ручку, которая
сверяет старый пароль и ставит новый.

Важно: ручка "сверить старый и новый пароль" **уже существует** —
`/api/v1/auth/change-password` (см. `swaggers/externalapi/swagger.yaml:8`,
схема `requestAuthAPIChangePassword` на строке 533). Она принимает
`X-User-Id` в заголовке (ставится шлюзом для залогиненного юзера) и тело
`{oldPassword, newPassword}`. Чего не хватает — шага
"кнопка → письмо → окно по ссылке" перед вызовом этой логики, плюс
защита токеном на случай, если юзер открывает ссылку не в текущей
залогиненной сессии.

Это флоу **не для "забыл пароль"** (там пароль неизвестен и старый
вводить нечем) — это подтверждение по почте смены пароля залогиненным
юзером, который старый пароль помнит.

## Архитектура

### 1. Инициация — кнопка в профиле

`internal/service/service.go`: новый метод, генерирует токен
(`crypto/rand`, как пароль в admin-invite-инвайте — без bias, через
`math/big`), кладёт в БД, шлёт письмо через уже подключенный мейлер
(`s.mailer`, тот же, что у `DeactivateUser`-письма).

### 2. Хранение токена

Новая таблица `password_reset_tokens`:

```sql
CREATE TABLE password_reset_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    token_hash text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz
);
```

Миграция — новый файл в `migrations/pkg/migrations/data/`, по аналогии
с `20260909120000_counterparty_inn_unique.sql`. Хранить `token_hash`
(sha256 от токена), не сам токен — так же, как хэшируются пароли.

TTL токена: 30 минут (значение вынести в конфиг
`internal/config/config.go`, аналогично остальным таймаутам).

### 3. Письмо

Тема: `Смена пароля Volint`

Тело (plain text, как остальные письма в сервисе):
```
Здравствуйте!

Вы запросили смену пароля.

Перейдите по ссылке, чтобы задать новый пароль:
https://app.volint.ru/reset-password?token={token}

Ссылка действует 30 минут. Если вы не запрашивали смену пароля,
проигнорируйте это письмо.
```

### 4. Подтверждение — форма по ссылке

Токен из query фронт достаёт сам и просто прикладывает к запросу формы
(старый пароль, новый пароль, подтверждение нового — сверка
new == confirm остаётся на фронте).

## Обработка ошибок

| Условие | Код | errorText |
|---|---|---|
| Токен/поля пустые | 400 | `auth.errors.invalidRequest` (`field`) |
| Токен не найден / уже использован | 400 | `auth.errors.tokenInvalid` (новая константа) |
| Токен просрочен | 400 | `auth.errors.tokenExpired` (новая константа) |
| `oldPassword` не совпал с хэшем | 401 | `auth.errors.invalidCredentials` (уже есть `InvalidCredentialsError`, `internal/errors/common.go`) |
| Письмо не ушло на шаге инициации | 200, `emailSent:false` — не ошибка (как в admin-invite) |
| Внутренняя ошибка | 500 | `auth.errors.internalError` |

Добавить в `internal/errors/common.go` (рядом с существующими
фабриками):

```go
ErrTokenInvalid = "auth.errors.tokenInvalid"
ErrTokenExpired = "auth.errors.tokenExpired"

TokenInvalidError = func() *Error {
    return New("reset token invalid", fasthttp.StatusBadRequest, ErrTokenInvalid)
}
TokenExpiredError = func() *Error {
    return New("reset token expired", fasthttp.StatusBadRequest, ErrTokenExpired)
}
```

## Контракт для фронта

### Шаг 1 — инициация

```
POST /api/v1/auth/password/reset/request
Headers: X-User-Id: <uuid>
Body: нет
```

Успех — `200 OK`:
```json
{
  "data": { "emailSent": true },
  "error": false,
  "errorText": "",
  "additionalErrors": null
}
```

Ошибка — общий формат `/api/v1/auth/*`:
```json
{
  "data": null,
  "error": true,
  "errorText": "auth.errors.internalError",
  "additionalErrors": null
}
```

### Шаг 2 — окно по ссылке `/reset-password?token=...`

Форма: старый пароль, новый пароль, подтверждение нового
(подтверждение сверяется на фронте, на бэк не шлётся).

```
POST /api/v1/auth/password/reset/confirm
Body:
```
```json
{
  "token": "a1b2c3...",
  "oldPassword": "string",
  "newPassword": "string"
}
```

Успех — `200 OK`:
```json
{
  "data": true,
  "error": false,
  "errorText": "",
  "additionalErrors": null
}
```

Ошибка, пример (токен просрочен):
```json
{
  "data": null,
  "error": true,
  "errorText": "auth.errors.tokenExpired",
  "additionalErrors": null
}
```

После успеха: токен помечается `used_at`, пароль перезаписывается
bcrypt-хэшем нового (как в `RegisterIP`). Обсудить отдельно: инвалидировать
ли текущие сессии юзера после смены (проверить, делается ли это уже для
`DeactivateUser` в `internal/service/service.go`).

## Что нужно фронту

- Кнопка "Сменить пароль" в профиле → `POST /api/v1/auth/password/reset/request`,
  показать тост "Письмо отправлено на почту".
- Страница `/reset-password?token=...`: форма из трёх полей, кнопка
  "Сохранить" → `POST /api/v1/auth/password/reset/confirm`.
- Обработать `auth.errors.tokenInvalid` / `auth.errors.tokenExpired` —
  показать "ссылка недействительна, запросите новую" с кнопкой назад на
  шаг 1.
- Обработать `auth.errors.invalidCredentials` на confirm — "неверный
  текущий пароль".

## Что нужно бэку (auth)

- `internal/errors/common.go`: `ErrTokenInvalid`, `ErrTokenExpired` +
  фабрики.
- Миграция `password_reset_tokens`.
- `internal/service/service.go`: методы инициации и подтверждения.
- Хендлеры в `internal/transport/custom-handlers/` (или через кодоген,
  если процесс для `/auth/*` такой же, как для остальных путей) —
  `POST /api/v1/auth/password/reset/request`,
  `POST /api/v1/auth/password/reset/confirm`.
- `swaggers/externalapi/swagger.yaml`: два новых `paths` + схема
  `requestAuthAPIPasswordResetConfirm` рядом с существующей
  `requestAuthAPIChangePassword` (строка 533).
- Конфиг: TTL токена в `internal/config/config.go`.

## Тестирование

- Успех: request → письмо (`emailSent:true`) → confirm с валидным
  токеном и верным `oldPassword` → пароль сменился, токен помечен
  `used_at`.
- Токен просрочен → 400 `tokenExpired`.
- Токен уже использован (повторный confirm) → 400 `tokenInvalid`.
- `oldPassword` неверный → 401 `invalidCredentials`, токен не
  расходуется (можно повторить попытку до истечения TTL).
- Мейлер вернул ошибку на request → не роняем запрос, `emailSent:false`
  (или 500 — решить отдельно, т.к. тут в отличие от admin-invite нет
  fallback вроде "показать пароль руками").
- Пустые поля → 400 `invalidRequest` с `field`.
