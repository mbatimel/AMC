# Приглашение администратора портала — дизайн и контракт

Статус: **утверждено** (короткий bounded/architectural-процесс, вопросы
согласованы в чате 2026-09-10). Кнопка "Пригласить администратора" в
admin-front сейчас дёргает мок (`front/src/app/portal-api/portal-users`),
который ничего не создаёт и не шлёт. Этот документ описывает настоящую
реализацию на бэке и контракт для фронта, который должен на неё
переключиться.

## Зачем

Нужно, чтобы приглашение реального человека в администраторы портала:
1. Заводило обычного пользователя (`users`) с ролью `admin`.
2. Генерировало пароль на бэке (16 символов).
3. Отправляло письмо приглашённому с логином (email) и паролем.

Кто может звать ручку: **только пользователь с ролью `admin`** (та же
проверка, что у остальных admin-эндпоинтов — `requireAdmin`).

## Архитектура

**Изменена по ходу реализации** (см. `docs/superpowers/plans/2026-09-10-admin-invite.md`,
раздел "Отклонение от спеки"): `back/auth` не трогаем — `tg client` для
auth-сервиса уже сломан на опциональных `*string`-параметрах (задокументировано
в `back/auth/pkg/client/transport/auth-client.go`), а трогать сервис, от
которого зависит весь логин на платформе, ради нишевой фичи — лишний риск.
Вместо этого всё целиком внутри `back/admin`.

### 1. `back/admin` — прямое создание пользователя в своём Postgres-слое

```go
// back/admin/internal/storage/postgres
func (s *Storage) CreateAdminUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error)
```

Одна транзакция: `INSERT INTO users (email, password, status) VALUES (...)`
→ `SELECT id FROM roles WHERE code = $1` (роль `admin` = `RoleCodeAdmin = 0`)
→ `INSERT INTO user_roles (...)`. Тот же SQL-паттерн, что уже используют
`back/auth`'s `assignRole` и `back/access` для ролей — переиспользуется
паттерн, не код. Конфликт по `users.email UNIQUE` → `postgres.ErrEmailTaken`.

### 2. `back/admin` — новый метод `InviteAdmin`

```go
InviteAdmin(ctx context.Context, callerUserID uuid.UUID, email, name string) (models.InviteAdminResponse, error)
```

Шаги:
1. `s.requireAdmin(ctx, callerUserID)` — переиспользуется как есть.
2. Валидация: `email` непустой и похож на email, `name` непустой.
3. Генерация пароля: 16 символов, алфавит `A-Za-z0-9`, `crypto/rand`
   (без bias — через `math/big`, не `%`; ~95 бит энтропии).
4. `bcrypt.GenerateFromPassword` — хеширование внутри `back/admin`
   (единственное место, где создаётся этот пользователь, так что здесь же
   и хешируем — как это делает `back/auth` для остальных пользователей).
5. `s.storage.CreateAdminUser(ctx, email, passwordHash)`.
   - `postgres.ErrEmailTaken` → `customErrors.ConflictError()` (409).
6. Письмо через уже подключенный `s.mailer` (см. решение по фейлу ниже).
7. `s.storage.InsertAuditLogEntry(ctx, callerUserID, actorLabelAdmin, "Приглашён администратор: "+email)`.
8. Ответ — см. контракт ниже.

**Если письмо не отправилось** (SMTP недоступен/`s.mailer == nil`):
аккаунт уже создан, откатывать не будем (это потребовало бы отдельной
delete-user RPC в auth — лишняя сложность для этой задачи). Вместо этого:
- логируем ошибку отправки (`s.logger.Error()...`), не роняем запрос;
- в ответе `emailSent: false`;
- пароль всё равно возвращается в теле ответа (см. контракт) — админ,
  который приглашал, сможет передать его вручную.

### 3. Email

Тема: `Приглашение в панель администратора Volint`

Тело (plain text, как и остальные письма в этом сервисе):
```
Здравствуйте, {name}!

Вас пригласили в качестве администратора портала Volint.

Логин: {email}
Пароль: {password}

Вход: https://admin.volint.ru
```

## Обработка ошибок

| Условие                                   | Код | errorText                          |
|--------------------------------------------|-----|-------------------------------------|
| Caller не admin                             | 403 | `admin.errors.forbidden`            |
| `email`/`name` пустые или email невалиден   | 400 | `admin.errors.badRequest` (`field`) |
| Email уже зарегистрирован                   | 409 | `admin.errors.conflict`             |
| Внутренняя ошибка (auth недоступен и т.п.)  | 500 | `admin.errors.internalError`        |
| Письмо не ушло                              | 200, `emailSent:false` — не ошибка  |

## Контракт для фронта

```
POST /api/v1/admin/portal-users/invite
Content-Type: application/json
Authorization: <сессия админа, как у остальных /api/v1/admin/* ручек>
```

Запрос:
```json
{
  "email": "new-admin@example.com",
  "name": "Иван Иванов"
}
```

Ответ, успех — `200 OK`:
```json
{
  "data": {
    "userId": "b7e6c2b0-1234-4d9a-9a0e-8f1a2b3c4d5e",
    "email": "new-admin@example.com",
    "password": "aB3dEfGh9kLmNoPq",
    "emailSent": true
  },
  "error": false,
  "errorText": "",
  "additionalErrors": null
}
```

`password` — присутствует всегда (и при `emailSent: true`, и при
`false`), чтобы фронт мог показать/дать скопировать его пригласившему
админу как fallback. Это внутренний admin-only инструмент, не
пользовательский экран — показывать пароль в UI единожды сразу после
создания приемлемо.

Ответ, ошибка — как у остальных `/api/v1/admin/*`:
```json
{
  "data": null,
  "error": true,
  "errorText": "admin.errors.conflict",
  "additionalErrors": { "field": "email" }
}
```
(код конфликта — `409`; полный список кодов см. таблицу выше)

Фронту нужно:
- заменить вызов `invitePortalUser` (`admin-front/src/core/shared/api/portalUsers.ts`,
  сейчас шлёт `POST /portal-api/portal-users`) на
  `POST /api/v1/admin/portal-users/invite` с телом `{email, name}`;
- при успехе показать `password` из ответа (например, в модалке
  "Скопируйте пароль — он больше нигде не отобразится") и, если
  `emailSent: false`, отдельно предупредить, что письмо не ушло.

Фронт-код меняет фронт-разработчик — бэк этого не трогает.

## Тестирование

- `back/admin`: юнит-тесты `InviteAdmin` — успех (`emailSent:true`),
  не-admin caller → 403, дубль email → 409, мейлер вернул ошибку →
  200 с `emailSent:false` и паролем в ответе, пустые email/name → 400.
- Пароль: тест на длину (16) и алфавит (только `A-Za-z0-9`).
