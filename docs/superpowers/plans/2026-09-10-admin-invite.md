# Приглашение администратора портала — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Добавить в `back/admin` реальную ручку `POST /api/v1/admin/portal-users/invite`, которая создаёт пользователя с ролью `admin`, генерирует пароль и шлёт приглашение на email.

**Architecture:** Всё изменение целиком внутри `back/admin` — отдельный сервис `back/auth` не трогаем (см. "Отклонение от спеки" ниже). `back/admin`'s storage слой напрямую создаёт запись в общей таблице `users` и назначает роль `admin` в `user_roles` в одной транзакции (тот же SQL-паттерн, что уже использует `back/auth`'s `assignRole` и `back/access` для ролей). Сервисный слой генерирует пароль (`crypto/rand`), хеширует bcrypt'ом, шлёт письмо через уже подключенный `mailer`. HTTP-слой добавляется по тому же шаблону кодогенератора `tg`, что и остальные `/api/v1/admin/*` ручки (`tg` бинарник недоступен в этой среде — правим сгенерированные файлы вручную, точно повторяя форму существующего метода `CreateCertificate`/`Login`; это уже устоявшаяся практика в репо, см. `back/auth/pkg/client/transport/auth-client.go`, где ровно то же самое явно задокументировано как обходной путь).

**Tech Stack:** Go, pgx/v4 (Postgres), Fiber, `golang.org/x/crypto/bcrypt`, существующий `internal/mailer.SMTPMailer` (SMTP уже настроен в `.env`).

**Spec:** `docs/superpowers/specs/2026-09-10-admin-invite-design.md`

## Отклонение от спеки

Спека (раздел "Архитектура", п.1) описывала отдельный internal RPC `CreateAdminUser` в `back/auth`. При написании плана выяснилось:
- `tg client`-генератор (клиентский кодек для `back/auth`) сломан на опциональных `*string`-параметрах — по этой причине `back/auth/pkg/client/transport/auth-client.go` уже сейчас хендрайтен и прямо документирует это как обходной путь.
- Чтобы добавить RPC в `back/auth`, всё равно пришлось бы вручную править 8+ файлов транспортного слоя `back/auth` (interface.go, server/middleware/logger/metrics/exchange/rest/http.go) — тот же объём работы, что и на стороне `back/admin`, но с риском задеть auth-service, которым пользуется вообще весь логин на платформе.
- `back/admin` уже имеет прямой доступ к тому же Postgres (`internal/storage/postgres`), и паттерн "создать юзера + назначить роль в одной транзакции" уже используется в `back/auth`/`back/access` с идентичным SQL — переиспользуется тот же SQL-паттерн, не тот же код.

Решение: `back/admin` создаёт пользователя и роль напрямую через свой storage-слой, не завися от `back/auth`. Внешний контракт (`POST /api/v1/admin/portal-users/invite`, форма ответа) из спеки не меняется — фронту это невидимо.

## Global Constraints

- Пароль генерируется на бэке: 16 символов, алфавит `A-Za-z0-9`, `crypto/rand` без bias (через `math/big`, а не `%`).
- Хеширование пароля — `bcrypt.DefaultCost`, как везде в `back/auth`/`back/admin`.
- Дубль email → `409` (`admin.errors.conflict`).
- Не-admin caller → `403` (`admin.errors.forbidden`), через существующий `s.requireAdmin`.
- Провал письма не роняет запрос — `emailSent:false`, пароль всё равно возвращается в ответе.
- Роль admin = `RoleCodeAdmin = 0` (см. `back/migrations/pkg/migrations/data/20260705171948_access_roles.sql`).
- Все новые файлы — на том же месте и в том же стиле, что соседние (banners/certificates) в `back/admin`.

---

### Task 1: Storage — `CreateAdminUser`

**Files:**
- Create: `back/admin/internal/storage/postgres/sql/insertAdminUser.sql`
- Create: `back/admin/internal/storage/postgres/sql/getRoleByCode.sql`
- Create: `back/admin/internal/storage/postgres/sql/insertUserRole.sql`
- Modify: `back/admin/internal/storage/postgres/postgres.go`

**Interfaces:**
- Produces: `func (s *Storage) CreateAdminUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error)` — возвращает `postgres.ErrEmailTaken`, если email уже занят (уникальный индекс `users.email`).

- [ ] **Step 1: Создать SQL-файлы**

`back/admin/internal/storage/postgres/sql/insertAdminUser.sql`:
```sql
INSERT INTO users (email, password, status)
VALUES ($1, $2, 'active')
RETURNING id
```

`back/admin/internal/storage/postgres/sql/getRoleByCode.sql`:
```sql
SELECT id
FROM roles
WHERE code = $1
```

`back/admin/internal/storage/postgres/sql/insertUserRole.sql`:
```sql
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
```

- [ ] **Step 2: Добавить `CreateAdminUser` в `postgres.go`**

В блок `import` добавить `"github.com/jackc/pgconn"` (уже используется как зависимость в go.mod — то же самое подключение, что и `back/auth`/`back/orders`).

После существующих `//go:embed` добавить:
```go
//go:embed sql/insertAdminUser.sql
var sqlInsertAdminUser string

//go:embed sql/getRoleByCode.sql
var sqlGetRoleByCode string

//go:embed sql/insertUserRole.sql
var sqlInsertUserRole string
```

После `var ErrBannerNotFound = errors.New("banner not found")` добавить:
```go
var ErrEmailTaken = errors.New("email already taken")

const uniqueViolationCode = "23505"

// roleCodeAdmin matches the "admin" role seeded in the shared roles table
// (back/migrations/pkg/migrations/data/20260705171948_access_roles.sql).
const roleCodeAdmin = 0
```

В конец файла добавить метод:
```go
// CreateAdminUser creates a bare user (email + password hash, no counterparty)
// and assigns the admin role, in one transaction.
func (s *Storage) CreateAdminUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("begin create admin user transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID uuid.UUID
	if err = tx.QueryRow(ctx, sqlInsertAdminUser, email, passwordHash).Scan(&userID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return uuid.UUID{}, ErrEmailTaken
		}
		return uuid.UUID{}, fmt.Errorf("insert admin user: %w", err)
	}

	var roleID uuid.UUID
	if err = tx.QueryRow(ctx, sqlGetRoleByCode, roleCodeAdmin).Scan(&roleID); err != nil {
		return uuid.UUID{}, fmt.Errorf("get admin role: %w", err)
	}

	if _, err = tx.Exec(ctx, sqlInsertUserRole, userID, roleID); err != nil {
		return uuid.UUID{}, fmt.Errorf("insert user role: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return uuid.UUID{}, fmt.Errorf("commit create admin user transaction: %w", err)
	}

	return userID, nil
}
```

- [ ] **Step 3: Проверить компиляцию**

Run: `cd back/admin && go build ./...`
Expected: успешная сборка, без ошибок.

(Отдельного `postgres_test.go` для этого метода не пишем — у соседних storage-методов admin'а (banners/certificates) тоже нет интеграционных тестов на уровне Postgres; реальное покрытие — юнит-тесты сервисного слоя на Task 4 через `fakeStorage`.)

- [ ] **Step 4: Commit**

```bash
git add back/admin/internal/storage/postgres/sql/insertAdminUser.sql \
        back/admin/internal/storage/postgres/sql/getRoleByCode.sql \
        back/admin/internal/storage/postgres/sql/insertUserRole.sql \
        back/admin/internal/storage/postgres/postgres.go
git commit -m "feat(admin): add CreateAdminUser storage method"
```

---

### Task 2: Генератор пароля

**Files:**
- Create: `back/admin/internal/service/password.go`
- Test: `back/admin/internal/service/password_test.go`

**Interfaces:**
- Produces: `func generatePassword() (string, error)` — 16 символов, алфавит `A-Za-z0-9`, без модульного смещения.

- [ ] **Step 1: Написать падающий тест**

`back/admin/internal/service/password_test.go`:
```go
package service

import (
	"strings"
	"testing"
)

func TestGeneratePassword_LengthAndAlphabet(t *testing.T) {
	seen := map[byte]bool{}
	for i := 0; i < 200; i++ {
		password, err := generatePassword()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(password) != passwordLength {
			t.Fatalf("expected length %d, got %d (%q)", passwordLength, len(password), password)
		}
		for _, c := range []byte(password) {
			if !strings.ContainsRune(passwordAlphabet, rune(c)) {
				t.Fatalf("unexpected character %q in password %q", c, password)
			}
			seen[c] = true
		}
	}
	if len(seen) < 20 {
		t.Fatalf("expected reasonable character variety across 200 generations, saw only %d distinct chars", len(seen))
	}
}

func TestGeneratePassword_Unique(t *testing.T) {
	a, err := generatePassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := generatePassword()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == b {
		t.Fatalf("expected two generated passwords to differ, got the same value twice: %q", a)
	}
}
```

- [ ] **Step 2: Убедиться, что тест падает**

Run: `cd back/admin && go test ./internal/service/... -run TestGeneratePassword -v`
Expected: FAIL — `generatePassword`/`passwordLength`/`passwordAlphabet` не определены (ошибка компиляции).

- [ ] **Step 3: Реализация**

`back/admin/internal/service/password.go`:
```go
package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const passwordLength = 16
const passwordAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

// generatePassword returns a cryptographically random password of
// passwordLength characters drawn uniformly from passwordAlphabet (no
// modulo bias — each character comes from its own rand.Int draw).
func generatePassword() (string, error) {
	alphabetLen := big.NewInt(int64(len(passwordAlphabet)))
	result := make([]byte, passwordLength)
	for i := range result {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", fmt.Errorf("generate password: %w", err)
		}
		result[i] = passwordAlphabet[n.Int64()]
	}
	return string(result), nil
}
```

- [ ] **Step 4: Тест проходит**

Run: `cd back/admin && go test ./internal/service/... -run TestGeneratePassword -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add back/admin/internal/service/password.go back/admin/internal/service/password_test.go
git commit -m "feat(admin): add random password generator for admin invites"
```

---

### Task 3: Модель ответа

**Files:**
- Modify: `back/admin/pkg/models/responses.go`

**Interfaces:**
- Produces: `models.InviteAdminResponse{ UserID uuid.UUID; Email string; Password string; EmailSent bool }`

- [ ] **Step 1: Добавить тип**

В `back/admin/pkg/models/responses.go`, рядом с `LoginResponse`/`SessionResponse`, добавить:
```go
type InviteAdminResponse struct {
	UserID    uuid.UUID `json:"userId"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	EmailSent bool      `json:"emailSent"`
}
```

- [ ] **Step 2: Проверить компиляцию**

Run: `cd back/admin && go build ./...`
Expected: успешная сборка.

- [ ] **Step 3: Commit**

```bash
git add back/admin/pkg/models/responses.go
git commit -m "feat(admin): add InviteAdminResponse model"
```

---

### Task 4: Сервисный слой — `InviteAdmin`

**Files:**
- Modify: `back/admin/internal/service/service.go` (добавить `CreateAdminUser` в интерфейс `Storage`)
- Create: `back/admin/internal/service/portal_users.go`
- Modify: `back/admin/internal/service/service_test.go` (добавить `CreateAdminUser` в `fakeStorage`)
- Test: `back/admin/internal/service/portal_users_test.go`

**Interfaces:**
- Consumes: `Storage.CreateAdminUser(ctx, email, passwordHash) (uuid.UUID, error)` (Task 1), `generatePassword() (string, error)` (Task 2), `models.InviteAdminResponse` (Task 3), уже существующие `s.requireAdmin`, `s.mailer.Send`, `s.storage.InsertAuditLogEntry`.
- Produces: `func (s *service) InviteAdmin(ctx context.Context, userID uuid.UUID, email string, name string) (models.InviteAdminResponse, error)`.

- [ ] **Step 1: Добавить `CreateAdminUser` в интерфейс `Storage`**

В `back/admin/internal/service/service.go`, в блок `type Storage interface { ... }`, добавить строку (после `DeleteCertificate`):
```go
	CreateAdminUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error)
```

- [ ] **Step 2: Добавить `CreateAdminUser` в `fakeStorage`**

В `back/admin/internal/service/service_test.go`, в структуру `fakeStorage` добавить поле:
```go
	createAdminUserFn func(context.Context, string, string) (uuid.UUID, error)
```

И метод (рядом с остальными `func (f *fakeStorage) ...`):
```go
func (f *fakeStorage) CreateAdminUser(ctx context.Context, email, passwordHash string) (uuid.UUID, error) {
	if f.createAdminUserFn != nil {
		return f.createAdminUserFn(ctx, email, passwordHash)
	}
	return uuid.New(), nil
}
```

- [ ] **Step 3: Написать падающие тесты**

`back/admin/internal/service/portal_users_test.go`:
```go
package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
)

func newPortalUsersService(storage *fakeStorage, mail *fakeMailer) *service {
	return NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true}, mail)
}

func TestInviteAdmin_CreatesUserAndSendsEmail(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	response, err := svc.InviteAdmin(context.Background(), uuid.New(), " NewAdmin@Example.com ", "  Иван Иванов  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Email != "newadmin@example.com" {
		t.Fatalf("expected normalized lowercase trimmed email, got %q", response.Email)
	}
	if len(response.Password) != passwordLength {
		t.Fatalf("expected generated password of length %d, got %q", passwordLength, response.Password)
	}
	if !response.EmailSent {
		t.Fatal("expected emailSent=true when mailer succeeds")
	}
	if mail.calls != 1 {
		t.Fatalf("expected mailer.Send called once, got %d", mail.calls)
	}
	if mail.to != "newadmin@example.com" {
		t.Fatalf("expected email sent to newadmin@example.com, got %q", mail.to)
	}
	if !strings.Contains(mail.body, response.Password) {
		t.Fatal("expected email body to contain the generated password")
	}
	if len(storage.inserted) != 1 || storage.inserted[0].action != "Приглашён администратор: newadmin@example.com" {
		t.Fatalf("expected audit log entry to be written, got %+v", storage.inserted)
	}
}

func TestInviteAdmin_NonAdminCaller_Forbidden(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: false}, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "new@example.com", "Имя")
	if err == nil {
		t.Fatal("expected error for non-admin caller")
	}
	if mail.calls != 0 {
		t.Fatal("mailer must not be called when caller is not admin")
	}
}

func TestInviteAdmin_EmailAlreadyTaken_Conflict(t *testing.T) {
	storage := &fakeStorage{createAdminUserFn: func(context.Context, string, string) (uuid.UUID, error) {
		return uuid.UUID{}, postgres.ErrEmailTaken
	}}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "taken@example.com", "Имя")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if mail.calls != 0 {
		t.Fatal("mailer must not be called when user creation fails")
	}
}

func TestInviteAdmin_InvalidEmail_BadRequest(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "not-an-email", "Имя")
	if err == nil {
		t.Fatal("expected validation error for invalid email")
	}
	if mail.calls != 0 {
		t.Fatal("mailer must not be called when email is invalid")
	}
}

func TestInviteAdmin_EmptyName_BadRequest(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{}
	svc := newPortalUsersService(storage, mail)

	_, err := svc.InviteAdmin(context.Background(), uuid.New(), "valid@example.com", "   ")
	if err == nil {
		t.Fatal("expected validation error for empty name")
	}
}

func TestInviteAdmin_MailerFails_StillReturnsPasswordWithEmailSentFalse(t *testing.T) {
	storage := &fakeStorage{}
	mail := &fakeMailer{sendErr: errors.New("smtp down")}
	svc := newPortalUsersService(storage, mail)

	response, err := svc.InviteAdmin(context.Background(), uuid.New(), "new@example.com", "Имя")
	if err != nil {
		t.Fatalf("mailer failure must not fail the request, got error: %v", err)
	}
	if response.EmailSent {
		t.Fatal("expected emailSent=false when mailer fails")
	}
	if response.Password == "" {
		t.Fatal("expected password to still be returned when mailer fails")
	}
}

func TestInviteAdmin_MailerNil_StillReturnsPasswordWithEmailSentFalse(t *testing.T) {
	storage := &fakeStorage{}
	svc := newPortalUsersService(storage, nil)

	response, err := svc.InviteAdmin(context.Background(), uuid.New(), "new@example.com", "Имя")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.EmailSent {
		t.Fatal("expected emailSent=false when mailer is not configured")
	}
}
```

- [ ] **Step 4: Убедиться, что тесты падают**

Run: `cd back/admin && go test ./internal/service/... -run TestInviteAdmin -v`
Expected: FAIL — компиляция падает (`InviteAdmin`/`normalizeInviteEmail`/`normalizeInviteName` не определены).

- [ ] **Step 5: Реализация**

`back/admin/internal/service/portal_users.go`:
```go
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

const maxPortalUserFieldLen = 255

func portalUserValidation(field string) error {
	return customErrors.BadRequestError().AddCause("field", field)
}

func normalizeInviteEmail(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || utf8.RuneCountInString(value) > maxPortalUserFieldLen || strings.ContainsAny(value, "\r\n\x00") {
		return "", portalUserValidation("email")
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", portalUserValidation("email")
	}
	return value, nil
}

func normalizeInviteName(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > maxPortalUserFieldLen {
		return "", portalUserValidation("name")
	}
	return value, nil
}

func inviteAdminEmailBody(name, email, password string) string {
	return fmt.Sprintf(
		"Здравствуйте, %s!\n\nВас пригласили в качестве администратора портала Volint.\n\nЛогин: %s\nПароль: %s\n\nВход: https://admin.volint.ru",
		name, email, password,
	)
}

// InviteAdmin creates a new portal admin account: generates a random
// password, creates the user with the admin role, and emails the
// credentials. A failed email delivery does not fail the request — the
// caller (an already-authenticated admin) gets the password back in the
// response as a fallback (see docs/superpowers/specs/2026-09-10-admin-invite-design.md).
func (s *service) InviteAdmin(ctx context.Context, userID uuid.UUID, email string, name string) (models.InviteAdminResponse, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return models.InviteAdminResponse{}, err
	}

	email, err := normalizeInviteEmail(email)
	if err != nil {
		return models.InviteAdminResponse{}, err
	}
	name, err = normalizeInviteName(name)
	if err != nil {
		return models.InviteAdminResponse{}, err
	}

	password, err := generatePassword()
	if err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	newUserID, err := s.storage.CreateAdminUser(ctx, email, string(passwordHash))
	if errors.Is(err, postgres.ErrEmailTaken) {
		return models.InviteAdminResponse{}, customErrors.ConflictError().AddCause("field", "email")
	}
	if err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	emailSent := true
	if s.mailer == nil {
		emailSent = false
		s.logger.Error().Str("email", email).Msg("invite admin: mailer is not configured, password not sent")
	} else if sendErr := s.mailer.Send(ctx, email, "Приглашение в панель администратора Volint", inviteAdminEmailBody(name, email, password)); sendErr != nil {
		emailSent = false
		s.logger.Error().Err(sendErr).Str("email", email).Msg("invite admin: failed to send invitation email")
	}

	if err = s.storage.InsertAuditLogEntry(ctx, userID, actorLabelAdmin, "Приглашён администратор: "+email); err != nil {
		return models.InviteAdminResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	return models.InviteAdminResponse{
		UserID:    newUserID,
		Email:     email,
		Password:  password,
		EmailSent: emailSent,
	}, nil
}
```

- [ ] **Step 6: Тесты проходят**

Run: `cd back/admin && go test ./internal/service/... -v`
Expected: PASS (весь пакет, включая уже существующие тесты — не должно быть регрессий).

- [ ] **Step 7: Commit**

```bash
git add back/admin/internal/service/service.go \
        back/admin/internal/service/service_test.go \
        back/admin/internal/service/portal_users.go \
        back/admin/internal/service/portal_users_test.go
git commit -m "feat(admin): add InviteAdmin service method"
```

---

### Task 5: HTTP/RPC-слой

Пользователь подтвердил: у него есть рабочий `tg` (в этой среде бинарника нет). Поэтому вручную правим только источник истины (`interface.go`) и `custom-handlers` (на который указывает `@tg http-response`) — всё остальное (`adminapi-exchange.go`, `adminapi-rest.go`, `adminapi-http.go`, `adminapi-server.go`, `adminapi-middleware.go`, `adminapi-metrics.go`, `adminapi-logger.go`) генерируется командой `go generate`, которая уже объявлена в шапке `interface.go` (`//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml`).

**Files:**
- Modify: `back/admin/pkg/interfaces/externalapi/interface.go`
- Create: `back/admin/internal/transport/custom-handlers/portal_users.go`
- Generated (пользователь, через `tg`): `adminapi-exchange.go`, `adminapi-rest.go`, `adminapi-http.go`, `adminapi-server.go`, `adminapi-middleware.go`, `adminapi-metrics.go`, `adminapi-logger.go`, `swaggers/externalapi/swagger.yaml`

**Interfaces:**
- Consumes: `svc.InviteAdmin` (Task 4), `models.InviteAdminResponse` (Task 3).
- Produces: `POST /api/v1/admin/portal-users/invite`, заголовок `X-User-Id`, JSON-тело `{email, name}` (см. контракт в спеке).

- [ ] **Step 1: `interface.go` — добавить метод в контракт**

В `back/admin/pkg/interfaces/externalapi/interface.go`, после блока `DeleteCertificate`, добавить:
```go
	// InviteAdmin ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/portal-users/invite
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:InviteAdmin
	// @tg summary=`Приглашение администратора`
	// @tg desc=`Создаёт учётную запись администратора портала, генерирует пароль и отправляет приглашение на email`
	// @tg uuidPackage=github.com/google/uuid
	InviteAdmin(ctx context.Context, userID uuid.UUID, email string, name string) (response models.InviteAdminResponse, err error)
```

- [ ] **Step 2: `custom-handlers/portal_users.go` — новый файл**

```go
// back/admin/internal/transport/custom-handlers/portal_users.go
package custom_handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
)

func InviteAdmin(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, email string, name string) error {
	return handle(ctx, "post", "/v1/admin/portal-users/invite", "InviteAdmin", map[string]interface{}{
		"userID": userID,
		"email":  email,
		"name":   name,
	}, func() (interface{}, error) {
		return svc.InviteAdmin(ctx.UserContext(), userID, email, name)
	})
}
```

- [ ] **Step 3: Собрать проект — убедиться, что падает ровно там, где ожидается**

Run: `cd back/admin && go build ./...`
Expected: FAIL. `loggerAdminAPI`, `metricsAdminAPI`, `serverAdminAPI`, `httpAdminAPI` больше не удовлетворяют `externalapi.AdminAPI` — в них нет метода `InviteAdmin` (мы добавили его только в интерфейс). Это ожидаемо и нормально — генератор ещё не запускался.

- [ ] **Step 4: СТОП — прогнать `tg` (это делает пользователь)**

Дальше сборку не чинить руками. Попроси пользователя выполнить у себя (там, где есть бинарник `tg`):
```bash
cd back/admin/pkg/interfaces/externalapi && go generate ./...
```
(Это ровно команда из `//go:generate` в шапке `interface.go` — перегенерит `adminapi-exchange.go`, `adminapi-rest.go`, `adminapi-http.go`, `adminapi-server.go`, `adminapi-middleware.go`, `adminapi-metrics.go`, `adminapi-logger.go` и `swaggers/externalapi/swagger.yaml`.)

Дождаться, когда пользователь пришлёт результат — либо diff перегенерированных файлов, либо вывод `go build ./... && go test ./...` из `back/admin` после генерации.

- [ ] **Step 5: Проверить результат генерации**

По присланному диффу/выводу убедиться:
- в каждом из 7 файлов появился метод/поле `InviteAdmin` (по аналогии с `CreateCertificate`/`Login`);
- маршрут `route.Post("/api/v1/admin/portal-users/invite", http.serveInviteAdmin)` появился в `adminapi-http.go`;
- `go build ./...` и `go test ./...` в `back/admin` — зелёные.

Если генератор кладёт пароль в лог открытым текстом в `adminapi-logger.go` (это вероятно — генератор не знает, что `Password` — секрет) — это единственное ручное послабление, которое стоит внести поверх сгенерированного: в методе `InviteAdmin` у `loggerAdminAPI` заменить логируемый `response` на копию с `Password: "REDACTED"`, по образцу:
```go
redactedResponse := response
redactedResponse.Password = "REDACTED"
// ...использовать redactedResponse вместо response в viewer.Sprintf
```
Это единственная ручная правка внутри generated-файла в этом плане — остальное всё из генератора as-is.

- [ ] **Step 6: Commit**

```bash
git add back/admin/pkg/interfaces/externalapi/interface.go \
        back/admin/internal/transport/custom-handlers/portal_users.go \
        back/admin/internal/transport/jsonRPC/externalapi/ \
        back/admin/swaggers/externalapi/swagger.yaml
git commit -m "feat(admin): wire POST /api/v1/admin/portal-users/invite"
```

---

### Task 6: go.mod, финальная сборка, обновление спеки

**Files:**
- Modify: `back/admin/go.mod`, `back/admin/go.sum`
- Modify: `docs/superpowers/specs/2026-09-10-admin-invite-design.md`

- [ ] **Step 1: Промоутнуть bcrypt из indirect в прямую зависимость**

Run:
```bash
cd back/admin && go mod tidy
```
Expected: `golang.org/x/crypto` в `go.mod` теряет комментарий `// indirect` (он теперь напрямую импортируется в `portal_users.go`). `go.sum` может обновиться незначительно.

- [ ] **Step 2: Полная сборка и тесты всего сервиса**

Run: `cd back/admin && go build ./... && go test ./...`
Expected: всё зелёное.

- [ ] **Step 3: Обновить спеку под финальную архитектуру**

В `docs/superpowers/specs/2026-09-10-admin-invite-design.md`, раздел "Архитектура", пункт 1 ("back/auth — новый internal RPC CreateAdminUser") заменить на факт: `back/admin` создаёт пользователя и роль напрямую через свой storage-слой (без нового RPC в `back/auth`) — переиспользуется тот же SQL-паттерн создания роли, что уже есть в `back/auth`/`back/access`. Обоснование — см. раздел "Отклонение от спеки" в `docs/superpowers/plans/2026-09-10-admin-invite.md`.

- [ ] **Step 4: Commit**

```bash
git add back/admin/go.mod back/admin/go.sum docs/superpowers/specs/2026-09-10-admin-invite-design.md
git commit -m "chore(admin): tidy deps and reconcile spec with final architecture"
```

---

## Что дальше (не входит в этот план)

- Фронт (`admin-front/src/core/shared/api/portalUsers.ts`) должен переключить `invitePortalUser` на `POST /api/v1/admin/portal-users/invite` — контракт см. в спеке. Это делает фронт-разработчик, не входит в этот бэк-план.
- Деплой `admin`-сервиса на volint после мержа.
