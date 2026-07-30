# Admin Panel Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up `back/admin` as a new microservice with staff login, session verification, and an append-only audit log — the foundation every later admin screen (content, catalog, users, reviews) will be built on.

**Architecture:** `back/admin` follows the exact layout already used by `back/orders` and `back/auth`: a hand-written `AdminAPI` interface annotated with `@tg` directives, transport/swagger generated from it by the in-repo `tg` codegen tool, and business logic in `internal/service`. It owns no user/password data — it authenticates staff by calling the existing `auth` service's `LoginUser` (over a newly generated HTTP client, mirroring `access`'s client) and authorizes them by calling the existing `access` service's `CheckAccess` with role code `0` ("admin"), which is already seeded in the shared `roles` table. The only data `back/admin` stores itself is `admin_audit_log`, a new table in the shared `AMC` Postgres database, applied through the existing `back/migrations` service.

**Tech Stack:** Go 1.25, Fiber v2, pgx/v4, zerolog, the in-repo `tg` transport generator, goose migrations (via `back/migrations`).

## Global Constraints

- Module path: `github.com/mbatimel/AMC/admin`, `go 1.25.0` (matches `orders`/`auth`).
- External API bind address defaults to `:8083` (`:8080`=access, `:8081`=auth, `:8082`=orders already taken); health server on `:9093` (`:9091`=auth, `:9092`=orders already taken).
- No server-side session token: the frontend stores the `userID` returned by `Login` and sends it back as the `X-User-Id` header on every subsequent admin request — this is the exact pattern `orders`/`auth`/`access` already use, do not invent a JWT/cookie scheme.
- Admin has no password storage of its own. Every credential check goes through the existing `auth` service; every role check goes through the existing `access` service. Do not create a `staff_users` table.
- All shared-DB migrations live in `back/migrations/pkg/migrations/data/*.sql`, goose format (`-- +goose Up` / `-- +goose Down`, `-- +goose StatementBegin` / `-- +goose StatementEnd`), auto-discovered by `//go:embed */*.sql` — no registration step needed.
- Follow existing per-service error-package convention exactly (`internal/errors/{common.go,errors.go,decoder.go}`, `*Error` funcs with `fasthttp` status codes, `admin.errors.*` translation keys).
- Test DB connection for integration tests: `host=localhost port=5432 dbname=AMC sslmode=disable user=mbatimel password=mbatimel` (override via `TEST_DATABASE_URL`), matching `orders/internal/storage/postgres/integration_test.go`.

---

## File Structure

```
back/auth/pkg/interfaces/externalAPI/interface.go   MODIFY  add tg client generation directive
back/auth/pkg/client/transport/*.go                 GENERATE  new HTTP client for auth's AuthAPI

back/migrations/pkg/migrations/data/20260729220000_admin_audit_log.sql   CREATE

back/admin/go.mod                                                        CREATE
back/admin/internal/doc.go                                               CREATE
back/admin/internal/config/config.go                                     CREATE
back/admin/internal/errors/common.go                                     CREATE
back/admin/internal/errors/errors.go                                     CREATE
back/admin/internal/errors/decoder.go                                    CREATE
back/admin/internal/transport/http/health.go                             CREATE
back/admin/pkg/models/responses.go                                       CREATE
back/admin/pkg/interfaces/externalapi/interface.go                       CREATE
back/admin/swaggers/externalapi/models/errors.go                         CREATE
back/admin/swaggers/externalapi/models/responses.go                      CREATE
back/admin/internal/storage/postgres/postgres.go                         CREATE
back/admin/internal/storage/postgres/connectManager.go                   CREATE
back/admin/internal/storage/postgres/sql/insertAuditLogEntry.sql         CREATE
back/admin/internal/storage/postgres/sql/listAuditLogEntries.sql         CREATE
back/admin/internal/storage/postgres/sql/countAuditLogEntries.sql        CREATE
back/admin/internal/storage/postgres/integration_test.go                 CREATE
back/admin/internal/service/service.go                                  CREATE
back/admin/internal/service/service_test.go                             CREATE
back/admin/internal/transport/custom-handlers/admin.go                  CREATE
back/admin/internal/transport/custom-handlers/response.go               CREATE
back/admin/internal/transport/jsonRPC/externalapi/*.go                  GENERATE (Task 8)
back/admin/swaggers/externalapi/swagger.yaml                            GENERATE (Task 8)
back/admin/cmd/main.go                                                  CREATE (Task 8)

README.md                                                                MODIFY  add back/admin row
```

---

### Task 1: Write an HTTP client for the `auth` service

`back/admin` needs to call `auth`'s `LoginUser` server-to-server (the same way `orders`/`auth` already call `access`'s `CheckAccess` via a generated client).

**This task does NOT use `tg client` codegen — hand-write the client instead.** (Discovered during a prior attempt at this task: `tg client --services .` generates a client for every method in `AuthAPI`, not just `LoginUser`, and the generator — an external binary, `tg` v2.3.95, not vendored in this repo — miscompiles methods with optional `*string` parameters, which `RegisterIP`/`RegisterIndividual` both have; it emits `_req.URI().QueryArgs().Set(key, somePointerString)` instead of dereferencing/`fmt.Sprint`-ing the pointer, a genuine generator bug with no in-repo source to patch. This was confirmed by running the generator and inspecting the compile errors. The human partner chose the hand-written route over patching the generated output in place (which `go generate` would silently clobber again) or redesigning the login flow.) Do **not** add a `//go:generate tg client` line to `back/auth/pkg/interfaces/externalAPI/interface.go` — leave that file untouched.

Reuse `back/access/pkg/client/transport/httpclient` (already a transitive dependency of `back/auth` via its existing `accessTransport` import in `cmd/main.go`) for the low-level fasthttp wrapper — don't duplicate it.

**Important wire-format detail (verified directly against the generated code, not assumed):** `auth`'s `AuthAPI.LoginUser` has **no** `@tg http-response` annotation (see `back/auth/pkg/interfaces/externalAPI/interface.go` — compare to e.g. `orders`' methods, which all carry one). `back/auth/internal/transport/custom-handlers/auth.go` exists but is **dead code** — nothing in `internal/transport/jsonRPC/externalapi/` imports it (confirmed: `grep -rln "custom-handlers" back/auth/internal/transport/jsonRPC/externalapi/` returns nothing). The route that's actually live, `back/auth/internal/transport/jsonRPC/externalapi/authapi-rest.go`'s `serveLoginUser`, calls the *generated* `sendResponse(ctx, response)` in `server.go` (`json.NewEncoder(ctx).Encode(resp)` — no wrapping), where `response` is `responseAuthAPILoginUser{UserID uuid.UUID \`json:"userID,omitempty"\`}` from `authapi-exchange.go`. So the real wire response on success is the **bare** object `{"userID":"<uuid-string>"}`, HTTP 200. On failure, status is set from the error's code (e.g. 401/500) via `ctx.Status(...)` and the body is a JSON-encoded error struct the client doesn't need to parse — treat any non-200 status as a plain error, exactly like `access`'s existing generated `CheckAccess` client does. Getting the success shape wrong silently returns a zero-value UUID with no error — match the shape below exactly.

**Security note:** credentials travel as query args (`?email=...&password=...`), not a JSON body — this is not this client's choice, it's `auth`'s existing, already-shipped wire contract (the generated `serveLoginUser` reads via `ctx.Query("email")`/`ctx.Query("password")`, confirmed above; every password-bearing method in `AuthAPI` — `LoginUser`, `RegisterIP`, `ChangePassword` — uses `@tg http-args=password|password`, i.e. the same query-arg contract). Changing that is a separate, repo-wide fix to `auth`'s API contract and out of scope for this task. What *is* in scope: never let the plaintext password reach a log line or error message from this client — in particular, do not format `_req.URI().String()` (it contains `?password=...`) into an error.

**Files:**
- Create: `back/auth/pkg/client/transport/auth-client.go`

**Interfaces:**
- Produces: `authTransport.NewClientAuthAPI(endpoint string, opts ...httpclient.Option) *ClientAuthAPI`, with method `LoginUser(ctx context.Context, email string, password string) (userID uuid.UUID, err error)` — package `github.com/mbatimel/AMC/auth/pkg/client/transport`. Consumed by Task 6 and Task 8.

- [ ] **Step 1: Write the client**

```go
// back/auth/pkg/client/transport/auth-client.go

// Package transport is a hand-written HTTP client for auth's LoginUser,
// used for server-to-server calls (back/admin calls this to authenticate
// staff without duplicating password storage/verification).
//
// This is NOT generated by `tg client` — that generator miscompiles
// AuthAPI because two of its other methods (RegisterIP, RegisterIndividual)
// take optional (*string) parameters, which `tg` v2.3.95's client mode
// cannot type-check correctly. Only LoginUser is implemented here; add
// more methods by hand, in this same style, if a future service needs them.
package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/access/pkg/client/transport/httpclient"
	fasthttp "github.com/valyala/fasthttp"
)

type ClientAuthAPI struct {
	httpClient *httpclient.ClientHTTP
}

func NewClientAuthAPI(endpoint string, opts ...httpclient.Option) (client *ClientAuthAPI) {
	httpClient := httpclient.NewClient(endpoint, opts...)
	return &ClientAuthAPI{httpClient: httpClient}
}

// loginUserResponse mirrors the generated responseAuthAPILoginUser wire
// shape (back/auth/internal/transport/jsonRPC/externalapi/authapi-exchange.go):
// LoginUser has no @tg http-response annotation, so its generated handler
// encodes this struct directly on success — no envelope, no wrapping.
type loginUserResponse struct {
	UserID uuid.UUID `json:"userID,omitempty"`
}

// LoginUser performs the LoginUser operation (POST /api/v1/auth/login).
// Credentials travel as query args because that's what the live server
// reads (ctx.Query("email")/ctx.Query("password")) — not this client's
// choice. Error paths deliberately omit the request URI/body from any
// message: the URI contains the plaintext password in its query string.
func (cli *ClientAuthAPI) LoginUser(ctx context.Context, email string, password string) (userID uuid.UUID, err error) {

	var reqBody []byte
	reqBody, err = json.Marshal(struct{}{})
	if err != nil {
		return
	}
	_req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(_req)
	_req.SetRequestURI(fmt.Sprintf("%s/api/v1/auth/login", cli.httpClient.BaseURL))
	_req.Header.SetMethod("POST")
	_req.Header.Set("Content-Type", "application/json")
	_req.SetBody(reqBody)
	_req.URI().QueryArgs().Set("email", email)
	_req.URI().QueryArgs().Set("password", password)
	_resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(_resp)
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		cli.httpClient.SetTimeout(timeout)
	}
	if err = cli.httpClient.Do(ctx, _req, _resp); err != nil {
		return
	}
	if _resp.StatusCode() != fasthttp.StatusOK {
		err = fmt.Errorf("auth login failed: HTTP %d", _resp.StatusCode())
		return
	}
	var response loginUserResponse
	if err = json.Unmarshal(_resp.Body(), &response); err != nil {
		return
	}
	userID = response.UserID
	return
}
```

- [ ] **Step 2: Verify the auth module builds**

```bash
cd back/auth
go build ./...
go vet ./...
```

Expected: exits 0. `back/auth/pkg/interfaces/externalAPI/interface.go` must be unchanged (`git diff --stat back/auth/pkg/interfaces` shows nothing) — only the new `back/auth/pkg/client/transport/auth-client.go` file is added.

- [ ] **Step 3: Write a table-driven unit test proving the response-unmarshaling logic**

This is the one piece of hand-written logic in this task worth a real test — an `httptest`-style fake server is overkill for fasthttp, so test the response shape directly via `json.Unmarshal` (mirroring exactly what `LoginUser` does internally, and exactly what `back/auth`'s own generated `responseAuthAPILoginUser` produces) to lock in the wire-format contract:

```go
// back/auth/pkg/client/transport/auth-client_test.go
package transport

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestLoginUserResponse_UnmarshalsUserID(t *testing.T) {
	id := uuid.New()
	body := []byte(`{"userID":"` + id.String() + `"}`)

	var response loginUserResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}
	if response.UserID != id {
		t.Fatalf("response.UserID = %v, want %v", response.UserID, id)
	}
}

func TestLoginUserResponse_EmptyBodyIsZeroUUID(t *testing.T) {
	var response loginUserResponse
	if err := json.Unmarshal([]byte(`{}`), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v, want nil", err)
	}
	if response.UserID != uuid.Nil {
		t.Fatalf("response.UserID = %v, want uuid.Nil", response.UserID)
	}
}
```

Run it:

```bash
cd back/auth
go test ./pkg/client/transport/... -v
```

Expected: both tests `PASS`.

- [ ] **Step 4: Commit**

```bash
git add back/auth/pkg/client
git commit -m "feat(auth): hand-write HTTP client for server-to-server LoginUser calls"
```

---

### Task 2: Add the `admin_audit_log` migration

**Files:**
- Create: `back/migrations/pkg/migrations/data/20260729220000_admin_audit_log.sql`

**Interfaces:**
- Produces: table `admin_audit_log(id, actor_user_id, actor_label, action, created_at)`, consumed by Task 5's storage layer.

- [ ] **Step 1: Write the migration**

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE admin_audit_log (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id UUID NOT NULL,
    actor_label   VARCHAR(255) NOT NULL,
    action        TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_admin_audit_log_created_at ON admin_audit_log (created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_admin_audit_log_created_at;
DROP TABLE IF EXISTS admin_audit_log;
-- +goose StatementEnd
```

- [ ] **Step 2: Apply it against the local dev database and verify it round-trips**

```bash
cd back/migrations
PG_ADDRESS=localhost:5432 PG_DB=AMC PG_USER=mbatimel PG_PASSWORD=mbatimel go run ./cmd
```

Expected: log line `Database migration completed successfully`. Then verify the table exists:

```bash
psql "host=localhost port=5432 dbname=AMC user=mbatimel password=mbatimel" -c "\d admin_audit_log"
```

Expected: shows the 5 columns and the `idx_admin_audit_log_created_at` index. (If postgres isn't running locally, skip this verification — Task 5's integration test will `t.Skip` in that case too, and this migration will simply apply the next time someone runs the migrations service against a real database.)

- [ ] **Step 3: Commit**

```bash
git add back/migrations/pkg/migrations/data/20260729220000_admin_audit_log.sql
git commit -m "feat(migrations): add admin_audit_log table"
```

---

### Task 3: Scaffold the `back/admin` Go module

**Files:**
- Create: `back/admin/go.mod`
- Create: `back/admin/internal/doc.go`
- Create: `back/admin/internal/config/config.go`
- Create: `back/admin/internal/errors/common.go`
- Create: `back/admin/internal/errors/errors.go`
- Create: `back/admin/internal/errors/decoder.go`
- Create: `back/admin/internal/transport/http/health.go`

**Interfaces:**
- Produces: `config.Config{PGHost, PGPort, PGDB, PGUser, PGPassword, BindAddr, AccessURL, AuthURL string}`, `config.LoadConfig() Config`, `config.LoadEnvFile(path string) error`, `config.GetEnv(key, fallback string) string` — consumed by Task 5 (`postgres.NewPool`) and Task 8 (`cmd/main.go`).
- Produces: `errors.InternalServerError()`, `errors.ForbiddenError()`, `errors.NotFoundError()`, `errors.BadRequestError()`, `errors.AccessDeniedError()`, `errors.MethodNotAllowedError()`, each `func() *errors.Error` — consumed by Task 6 (`service.go`).
- Produces: `transporthttp.NewHealthServer() *HealthServer` with `Start(bindURL string) error` / `Stop() error` — consumed by Task 8.

- [ ] **Step 1: Create the module file**

```go
// back/admin/go.mod
module github.com/mbatimel/AMC/admin

go 1.25.0

require (
	github.com/gofiber/adaptor/v2 v2.2.1
	github.com/gofiber/fiber/v2 v2.52.13
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v4 v4.18.3
	github.com/mbatimel/AMC/access v0.0.0-00010101000000-000000000000
	github.com/mbatimel/AMC/auth v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.23.2
	github.com/rs/zerolog v1.35.1
	github.com/valyala/fasthttp v1.72.0
)

replace github.com/mbatimel/AMC/access => ../access

replace github.com/mbatimel/AMC/auth => ../auth
```

- [ ] **Step 2: Create the package doc**

```go
// back/admin/internal/doc.go
// Package internal contains private implementation details for the admin service.
package internal
```

- [ ] **Step 3: Create the config package**

```go
// back/admin/internal/config/config.go
package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

type Config struct {
	PGHost     string
	PGPort     string
	PGDB       string
	PGUser     string
	PGPassword string
	BindAddr   string
	AccessURL  string
	AuthURL    string
}

func LoadConfig() Config {
	cfg := Config{
		PGHost:     GetEnv("PG_HOST", "localhost"),
		PGPort:     GetEnv("PG_PORT", "5432"),
		PGDB:       os.Getenv("PG_DB"),
		PGUser:     os.Getenv("PG_USER"),
		PGPassword: os.Getenv("PG_PASSWORD"),
		BindAddr:   GetEnv("BIND_ADDR", ":8083"),
		AccessURL:  os.Getenv("ACCESS_URL"),
		AuthURL:    os.Getenv("AUTH_URL"),
	}

	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	if cfg.AccessURL == "" {
		cfg.AccessURL = "http://localhost:8080"
		log.Warn().Msg("ACCESS_URL must be specified")
	}
	if cfg.AuthURL == "" {
		cfg.AuthURL = "http://localhost:8081"
		log.Warn().Msg("AUTH_URL must be specified")
	}

	return cfg
}

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func LoadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		value = strings.Trim(strings.TrimSpace(value), `"'`)
		value = os.Expand(value, func(name string) string {
			return os.Getenv(name)
		})
		if err = os.Setenv(key, value); err != nil {
			return err
		}
	}
	return scanner.Err()
}
```

- [ ] **Step 4: Create the errors package**

```go
// back/admin/internal/errors/common.go
package errors

import (
	"errors"
	"fmt"
)

const (
	noneValue = "None"
	Errors    = "errors"

	defaultErrorCode = -32603 // internalError

	CauseErrDescription = "description"
)

type TrParams struct {
	TrKey  string                 `json:"trKey"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type Error struct {
	ErrorText  string
	Cause      map[string]interface{}
	statusCode int
	errorCode  *int

	internalError error
}

func (e *Error) Error() string {
	var cause string
	if e.Cause != nil {
		cause = fmt.Sprintf(". Causes: %v", e.Cause)
	}
	return e.ErrorText + cause
}

func Is(errOne, errTwo error) bool {
	custErrOne, custErrTwo, ok := getCustomError(errOne, errTwo)

	if ok {
		return custErrOne.ErrorText == custErrTwo.ErrorText
	} else {
		return errOne.Error() == errTwo.Error()
	}
}

func HasCause(e error, cause string) bool {
	var custErr *Error
	ok := errors.As(e, &custErr)
	if !ok {
		return false
	}

	for key, value := range custErr.Cause {
		if key == cause || value == cause {
			return true
		}
	}

	return false
}

func New(msg string, statusCode int, outerError string) *Error {
	return &Error{
		ErrorText:     msg,
		statusCode:    statusCode,
		internalError: errors.New(outerError),
	}
}

func getCustomError(err1 error, err2 error) (*Error, *Error, bool) {
	var e1, e2 *Error

	var jsonErr1 *JsonRPCError
	ok := errors.As(err1, &jsonErr1)
	if ok {
		e1 = jsonErr1.Data
	} else {
		ok = errors.As(err1, &e1)
		if !ok {
			return e1, e2, false
		}
	}

	var jsonErr2 *JsonRPCError
	ok = errors.As(err2, &jsonErr2)
	if ok {
		e2 = jsonErr2.Data
	} else {
		ok = errors.As(err2, &e2)
		if !ok {
			return e1, e2, false
		}
	}

	return e1, e2, true
}

func (e *Error) SetStatusCode(code int) *Error {
	e.statusCode = code
	return e
}

func (e *Error) GetStatusCode() int {
	return e.statusCode
}

func (e *Error) SetOuterError(err interface{}) *Error {
	e.internalError = fmt.Errorf("%v", err)
	return e
}

func (e *Error) GetOuterError() error {
	return e.internalError
}

func (e *Error) AddTrErrors(trError TrParams) *Error {
	if e.Cause == nil {
		e.Cause = make(map[string]interface{}, 1)
	}

	errors, ok := e.Cause[Errors].([]TrParams)
	if !ok {
		e.Cause[Errors] = []TrParams{{
			TrKey:  trError.TrKey,
			Params: trError.Params,
		}}

		return e
	}

	e.Cause[Errors] = append(errors, trError)

	return e
}

func (e *Error) AddCause(args ...string) *Error {
	if e.Cause == nil {
		e.Cause = make(map[string]interface{})
	}

	for i := 0; i < len(args); i += 2 {
		strKey := args[i]
		e.Cause[strKey] = noneValue
		if i+1 < len(args) {
			e.Cause[strKey] = args[i+1]
		}
	}

	return e
}

// Code returns the HTTP status code. The generated transport layer
// (jsonRPC/externalapi) type-asserts errors against a `Code() int` method
// to set the response status, so this must stay in sync with statusCode
// rather than the JSON-RPC errorCode.
func (e *Error) Code() int {
	return e.statusCode
}

func (e *Error) SetErrorCode(errorCode int) *Error {
	e.errorCode = &errorCode
	return e
}

type BadRequestTypeError struct {
	StatusCode int
	Body       []byte
}

func (err *BadRequestTypeError) Error() string {
	return fmt.Sprintf("status code %d, data '%s'", err.StatusCode, err.Body)
}

func HTTPStatusCode(err error) (int, bool) {
	var badRequestErr *BadRequestTypeError
	if !errors.As(err, &badRequestErr) {
		return 0, false
	}

	return badRequestErr.StatusCode, true
}

func IsHTTPClientError(err error) bool {
	statusCode, ok := HTTPStatusCode(err)
	return ok && statusCode >= 400 && statusCode < 500
}
func EqualErrorCode(err error, targetError int) bool {
	customError, _, _ := getCustomError(err, nil)
	if customError == nil {
		return false
	}

	if customError.errorCode != nil && *customError.errorCode == targetError {
		return true
	}
	return false
}
```

```go
// back/admin/internal/errors/errors.go
package errors

import (
	"github.com/valyala/fasthttp"
)

var (
	AccessDeniedError     = func() *Error { return New("access denied", fasthttp.StatusForbidden, ErrAccessDenied) }
	ForbiddenError        = func() *Error { return New("forbidden", fasthttp.StatusForbidden, ErrForbidden) }
	BadRequestError       = func() *Error { return New("bad request", fasthttp.StatusBadRequest, ErrBadRequest) }
	MethodNotAllowedError = func() *Error { return New("method not allowed", fasthttp.StatusBadRequest, ErrMethodNotAllowed) }
	InternalServerError   = func() *Error {
		return New("internal server error", fasthttp.StatusInternalServerError, ErrInternal)
	}
	NotFoundError       = func() *Error { return New("not found", fasthttp.StatusNotFound, ErrNotFound) }
	NotImplementedError = func() *Error {
		return New("not implemented", fasthttp.StatusNotImplemented, ErrNotImplemented)
	}
)

const (
	ErrInternal         = "admin.errors.internalError"    // Внутренняя ошибка
	ErrBadRequest       = "admin.errors.badRequest"       // Плохой запрос
	ErrMethodNotAllowed = "admin.errors.methodNotAllowed" // Метод не поддерживается
	ErrForbidden        = "admin.errors.forbidden"        // Доступ запрещен
	ErrInvalidRequest   = "admin.errors.invalidRequest"   // Неправильный запрос
	ErrAccessDenied     = "admin.errors.accessDenied"     // Отказано в доступе
	ErrNotFound         = "admin.errors.notFound"         // Не найдено
	ErrNotImplemented   = "admin.errors.notImplemented"   // Метод не реализован
)
```

```go
// back/admin/internal/errors/decoder.go
package errors

import "encoding/json"

type JsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *Error `json:"data,omitempty"`
}

func (err JsonRPCError) Error() string {
	return err.Message
}

func ErrorDecoder(errData json.RawMessage) (err error) {

	var jsonrpcError JsonRPCError
	if err = json.Unmarshal(errData, &jsonrpcError); err != nil {
		return
	}
	return jsonrpcError
}
```

- [ ] **Step 5: Create the health server**

```go
// back/admin/internal/transport/http/health.go
package http

import (
	"github.com/gofiber/fiber/v2"
)

type HealthServer struct {
	server *fiber.App
}

func NewHealthServer() *HealthServer {
	health := &HealthServer{
		server: fiber.New(fiber.Config{DisableStartupMessage: true}),
	}
	health.server.Get("/liveness", probesHandler)
	health.server.Get("/readiness", probesHandler)

	return health
}

func probesHandler(ctx *fiber.Ctx) error {
	return ctx.SendStatus(fiber.StatusOK)
}

func (h *HealthServer) Start(bindURL string) error {
	return h.server.Listen(bindURL)
}

func (h *HealthServer) Stop() error {
	return h.server.Shutdown()
}
```

- [ ] **Step 6: Resolve dependencies and verify the module builds**

```bash
cd back/admin
go mod tidy
go build ./...
```

Expected: exits 0 (only `config`, `errors`, and `transport/http` packages exist so far).

- [ ] **Step 7: Commit**

```bash
git add back/admin/go.mod back/admin/go.sum back/admin/internal/doc.go back/admin/internal/config back/admin/internal/errors back/admin/internal/transport/http
git commit -m "feat(admin): scaffold service module (config, errors, health)"
```

---

### Task 4: Define the `AdminAPI` contract

**Files:**
- Create: `back/admin/pkg/models/responses.go`
- Create: `back/admin/pkg/interfaces/externalapi/interface.go`
- Create: `back/admin/swaggers/externalapi/models/errors.go`
- Create: `back/admin/swaggers/externalapi/models/responses.go`

**Interfaces:**
- Consumes: none (declarative contract).
- Produces: `models.LoginResponse{UserID uuid.UUID; Role string}`, `models.SessionResponse{UserID uuid.UUID; Role string}`, `models.AuditLogEntry{ID uuid.UUID; CreatedAt time.Time; ActorLabel string; Action string}`, `models.ListAuditLogResponse{Items []models.AuditLogEntry; Total int}` — consumed by Task 6 (`service.go`) and Task 7 (`custom-handlers`).
- Produces: `externalapi.AdminAPI` interface with methods `Login(ctx, email, password string) (models.LoginResponse, error)`, `Logout(ctx, userID uuid.UUID) error`, `GetSession(ctx, userID uuid.UUID) (models.SessionResponse, error)`, `ListAuditLog(ctx, userID uuid.UUID, limit, offset int) (models.ListAuditLogResponse, error)` — consumed by Task 6 (implemented by `service`), Task 7 (`custom-handlers` take `svc externalapi.AdminAPI`), and Task 8 (codegen target).

- [ ] **Step 1: Create the response models**

```go
// back/admin/pkg/models/responses.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginResponse struct {
	UserID uuid.UUID `json:"userID"`
	Role   string    `json:"role"`
}

type SessionResponse struct {
	UserID uuid.UUID `json:"userID"`
	Role   string    `json:"role"`
}

type AuditLogEntry struct {
	ID         uuid.UUID `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	ActorLabel string    `json:"actorLabel"`
	Action     string    `json:"action"`
}

type ListAuditLogResponse struct {
	Items []AuditLogEntry `json:"items"`
	Total int             `json:"total"`
}
```

- [ ] **Step 2: Create the swagger response/error model boilerplate**

These two files are the `@tg 200=...`/`@tg 400=...` etc. targets referenced below; copied verbatim from `back/orders/swaggers/externalapi/models/` (the same boilerplate is byte-identical across every service in this repo).

```go
// back/admin/swaggers/externalapi/models/responses.go
package models

import (
	"go/types"
)

type Resp200 struct {
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=false
	Error bool `json:"error"`
	// @tg example=``
	ErrorText string `json:"errorText"`
	// @tg example=`true`
	Data types.Nil `json:"data"`
	// @tg example=``
	AdditionalErrors types.Nil `json:"additionalErrors"`
}
```

```go
// back/admin/swaggers/externalapi/models/errors.go
package models

import "go/types"

type Err400 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.auth.badRequest`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе`
	AdditionalErrors struct {
		Errors []struct {
			TrKey string `json:"trKey"`
			// @tg example=`{"1": "value one", "2": "value two"}`
			Params map[string]string `json:"params"`
		} `json:"errors"`
	} `json:"additionalErrors"`
}

type Err401 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.auth.forbidden`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе, со статус кодом 401, не указывается`
	AdditionalErrors types.Nil `json:"additionalErrors,omitempty"`
}

type Err403 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.auth.accessDenied`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе, со статус кодом 403, не указывается`
	AdditionalErrors types.Nil `json:"additionalErrors,omitempty"`
}

type Err500 struct {
	// @tg example=``
	Data types.Nil `json:"data,omitempty"`
	// @tg desc=`Флаг показывающий, что ответ пришел с ошибкой`
	// @tg example=`true`
	Error bool `json:"error"`
	// @tg desc=`Заголовок ошибки`
	// @tg example=`content.api.errors.auth.internalError`
	ErrorText string `json:"errorText"`
	// @tg desc=`Текст ошибки, при ответе, со статус кодом 403, не указывается`
	AdditionalErrors types.Nil `json:"additionalErrors,omitempty"`
}
```

- [ ] **Step 3: Create the `AdminAPI` interface**

```go
// back/admin/pkg/interfaces/externalapi/interface.go
// Package externalapi describes the public admin panel API contract.
// @tg version=0.0.1
// @tg backend=admin
// @tg title=`admin`
// @tg servers=
//
//go:generate tg transport --services . --out ../../../internal/transport/jsonRPC/externalapi --outSwagger ../../../swaggers/externalapi/swagger.yaml
package externalapi

import (
	"context"

	"github.com/google/uuid"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

// AdminAPI
// @tg http-server metrics log
// @tg http-prefix=/api
// @tg 200=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Resp200
// @tg 400=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err400
// @tg 401=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err401
// @tg 403=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err403
// @tg 500=github.com/mbatimel/AMC/admin/swaggers/externalapi/models:Err500
type AdminAPI interface {
	// Login ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/auth/login
	// @tg http-args=email|email
	// @tg http-args=password|password
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:Login
	// @tg summary=`Вход в админ-панель`
	// @tg desc=`Авторизация сотрудника портала по email и паролю`
	// @tg uuidPackage=github.com/google/uuid
	Login(ctx context.Context, email string, password string) (response models.LoginResponse, err error)

	// Logout ...
	// @tg http-method=POST
	// @tg http-path=/v1/admin/auth/logout
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:Logout
	// @tg summary=`Выход из админ-панели`
	// @tg desc=`Завершение сессии сотрудника портала`
	// @tg uuidPackage=github.com/google/uuid
	Logout(ctx context.Context, userID uuid.UUID) (err error)

	// GetSession ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/auth/session
	// @tg http-headers=userID|X-User-Id
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:GetSession
	// @tg summary=`Проверка сессии`
	// @tg desc=`Проверка, что пользователь авторизован и имеет права администратора`
	// @tg uuidPackage=github.com/google/uuid
	GetSession(ctx context.Context, userID uuid.UUID) (response models.SessionResponse, err error)

	// ListAuditLog ...
	// @tg http-method=GET
	// @tg http-path=/v1/admin/audit-log
	// @tg http-headers=userID|X-User-Id
	// @tg http-args=limit|limit
	// @tg http-args=offset|offset
	// @tg http-response=github.com/mbatimel/AMC/admin/internal/transport/custom-handlers:ListAuditLog
	// @tg summary=`Журнал действий`
	// @tg desc=`Список действий администраторов портала без возможности удаления`
	// @tg uuidPackage=github.com/google/uuid
	ListAuditLog(ctx context.Context, userID uuid.UUID, limit int, offset int) (response models.ListAuditLogResponse, err error)
}
```

- [ ] **Step 4: Verify it compiles**

```bash
cd back/admin
go build ./pkg/...
go vet ./pkg/...
```

Expected: exits 0. (Do not run `go generate` yet — the `@tg http-response` targets point at `internal/transport/custom-handlers`, which doesn't exist until Task 7; generating now would produce code that fails to build until then.)

- [ ] **Step 5: Commit**

```bash
git add back/admin/pkg back/admin/swaggers
git commit -m "feat(admin): define AdminAPI contract"
```

---

### Task 5: Implement audit log storage

**Files:**
- Create: `back/admin/internal/storage/postgres/connectManager.go`
- Create: `back/admin/internal/storage/postgres/postgres.go`
- Create: `back/admin/internal/storage/postgres/sql/insertAuditLogEntry.sql`
- Create: `back/admin/internal/storage/postgres/sql/listAuditLogEntries.sql`
- Create: `back/admin/internal/storage/postgres/sql/countAuditLogEntries.sql`
- Test: `back/admin/internal/storage/postgres/integration_test.go`

**Interfaces:**
- Consumes: `config.Config` (Task 3), table `admin_audit_log` (Task 2).
- Produces: `postgres.NewPool(cfg config.Config) (*pgxpool.Pool, error)`, `postgres.New(pool *pgxpool.Pool) *Storage`, `postgres.AuditLogEntry{ID, ActorUserID uuid.UUID; ActorLabel, Action string; CreatedAt time.Time}`, `(*Storage) InsertAuditLogEntry(ctx, actorUserID uuid.UUID, actorLabel, action string) error`, `(*Storage) ListAuditLogEntries(ctx, limit, offset int) ([]AuditLogEntry, error)`, `(*Storage) CountAuditLogEntries(ctx) (int, error)` — consumed by Task 6 and Task 8.

- [ ] **Step 1: Write the failing integration test**

```go
// back/admin/internal/storage/postgres/integration_test.go
//go:build integration

package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost port=5432 dbname=AMC sslmode=disable user=mbatimel password=mbatimel"
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		t.Skipf("postgres not reachable, skipping integration test: %v", err)
	}
	return pool
}

func TestAuditLogRoundTrip(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()

	ctx := context.Background()
	storage := New(pool)

	actorID := uuid.New()

	if err := storage.InsertAuditLogEntry(ctx, actorID, "Админ портала", "Выполнен вход в систему"); err != nil {
		t.Fatalf("InsertAuditLogEntry (login) failed: %v", err)
	}
	if err := storage.InsertAuditLogEntry(ctx, actorID, "Админ портала", "Выполнен выход из системы"); err != nil {
		t.Fatalf("InsertAuditLogEntry (logout) failed: %v", err)
	}

	entries, err := storage.ListAuditLogEntries(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListAuditLogEntries failed: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("ListAuditLogEntries returned %d entries, want at least 2", len(entries))
	}
	if entries[0].Action != "Выполнен выход из системы" {
		t.Fatalf("ListAuditLogEntries[0].Action = %q, want the most recently inserted entry first", entries[0].Action)
	}
	if entries[0].ActorUserID != actorID {
		t.Fatalf("ListAuditLogEntries[0].ActorUserID = %v, want %v", entries[0].ActorUserID, actorID)
	}

	total, err := storage.CountAuditLogEntries(ctx)
	if err != nil {
		t.Fatalf("CountAuditLogEntries failed: %v", err)
	}
	if total < 2 {
		t.Fatalf("CountAuditLogEntries = %d, want at least 2", total)
	}
}
```

- [ ] **Step 2: Run it and confirm it fails to compile (Storage/New don't exist yet)**

```bash
cd back/admin
go test -tags=integration ./internal/storage/postgres/... -run TestAuditLogRoundTrip -v
```

Expected: build failure — `undefined: New`.

- [ ] **Step 3: Write the SQL queries**

```sql
-- back/admin/internal/storage/postgres/sql/insertAuditLogEntry.sql
INSERT INTO admin_audit_log (actor_user_id, actor_label, action)
VALUES ($1, $2, $3);
```

```sql
-- back/admin/internal/storage/postgres/sql/listAuditLogEntries.sql
SELECT id, actor_user_id, actor_label, action, created_at
FROM admin_audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
```

```sql
-- back/admin/internal/storage/postgres/sql/countAuditLogEntries.sql
SELECT count(*) FROM admin_audit_log;
```

- [ ] **Step 4: Implement the storage layer**

```go
// back/admin/internal/storage/postgres/postgres.go
package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed sql/insertAuditLogEntry.sql
var sqlInsertAuditLogEntry string

//go:embed sql/listAuditLogEntries.sql
var sqlListAuditLogEntries string

//go:embed sql/countAuditLogEntries.sql
var sqlCountAuditLogEntries string

type AuditLogEntry struct {
	ID          uuid.UUID
	ActorUserID uuid.UUID
	ActorLabel  string
	Action      string
	CreatedAt   time.Time
}

type Storage struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Storage {
	return &Storage{pool: pool}
}

func (s *Storage) InsertAuditLogEntry(ctx context.Context, actorUserID uuid.UUID, actorLabel string, action string) error {
	if _, err := s.pool.Exec(ctx, sqlInsertAuditLogEntry, actorUserID, actorLabel, action); err != nil {
		return fmt.Errorf("insert audit log entry: %w", err)
	}
	return nil
}

func (s *Storage) ListAuditLogEntries(ctx context.Context, limit int, offset int) ([]AuditLogEntry, error) {
	rows, err := s.pool.Query(ctx, sqlListAuditLogEntries, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list audit log entries: %w", err)
	}
	defer rows.Close()

	entries := make([]AuditLogEntry, 0)
	for rows.Next() {
		var e AuditLogEntry
		if err = rows.Scan(&e.ID, &e.ActorUserID, &e.ActorLabel, &e.Action, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit log entry: %w", err)
		}
		entries = append(entries, e)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit log entries: %w", err)
	}

	return entries, nil
}

func (s *Storage) CountAuditLogEntries(ctx context.Context) (int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, sqlCountAuditLogEntries).Scan(&total); err != nil {
		return 0, fmt.Errorf("count audit log entries: %w", err)
	}
	return total, nil
}
```

```go
// back/admin/internal/storage/postgres/connectManager.go
package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/mbatimel/AMC/admin/internal/config"
)

func NewPool(cfg config.Config) (*pgxpool.Pool, error) {
	port, err := strconv.Atoi(cfg.PGPort)
	if err != nil {
		return nil, fmt.Errorf("parse PG_PORT: %w", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d dbname=%s sslmode=disable user=%s password=%s",
		cfg.PGHost,
		port,
		cfg.PGDB,
		cfg.PGUser,
		cfg.PGPassword,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}
```

- [ ] **Step 5: Run the test again and verify it passes (or skips cleanly without a local postgres)**

```bash
cd back/admin
go test -tags=integration ./internal/storage/postgres/... -run TestAuditLogRoundTrip -v
```

Expected: `PASS` if `admin_audit_log` exists locally (Task 2, Step 2 applied it); otherwise `--- SKIP` with "postgres not reachable". Either way, exit code 0.

- [ ] **Step 6: Commit**

```bash
git add back/admin/internal/storage
git commit -m "feat(admin): implement audit log storage"
```

---

### Task 6: Implement the service layer

**Files:**
- Create: `back/admin/internal/service/service.go`
- Test: `back/admin/internal/service/service_test.go`

**Interfaces:**
- Consumes: `externalapi.AdminAPI` (Task 4), `models.{LoginResponse,SessionResponse,AuditLogEntry,ListAuditLogResponse}` (Task 4), `postgres.AuditLogEntry` (Task 5), `errors.{InternalServerError,ForbiddenError}` (Task 3), `authTransport.ClientAuthAPI.LoginUser` shape (Task 1), `accessTransport.ClientAccessAPI.CheckAccess` shape (existing).
- Produces: `service.NewAdminApiService(logger zerolog.Logger, storage Storage, authClient AuthClient, accessClient AccessClient) externalapi.AdminAPI` — consumed by Task 8 (`cmd/main.go`). Also `service.Storage`, `service.AuthClient`, `service.AccessClient` interfaces, and the exported const `service.RoleCodeAdmin = 0`.

- [ ] **Step 1: Write the failing unit tests**

```go
// back/admin/internal/service/service_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
)

type fakeStorage struct {
	inserted []auditCall
	entries  []postgres.AuditLogEntry
	total    int
	listErr  error
}

type auditCall struct {
	actorUserID uuid.UUID
	actorLabel  string
	action      string
}

func (f *fakeStorage) InsertAuditLogEntry(_ context.Context, actorUserID uuid.UUID, actorLabel string, action string) error {
	f.inserted = append(f.inserted, auditCall{actorUserID, actorLabel, action})
	return nil
}

func (f *fakeStorage) ListAuditLogEntries(_ context.Context, _ int, _ int) ([]postgres.AuditLogEntry, error) {
	return f.entries, f.listErr
}

func (f *fakeStorage) CountAuditLogEntries(_ context.Context) (int, error) {
	return f.total, nil
}

type fakeAuthClient struct {
	userID uuid.UUID
	err    error
}

func (f *fakeAuthClient) LoginUser(_ context.Context, _ string, _ string) (uuid.UUID, error) {
	return f.userID, f.err
}

type fakeAccessClient struct {
	allowed bool
	err     error
}

func (f *fakeAccessClient) CheckAccess(_ context.Context, _ uuid.UUID, _ int) (bool, error) {
	return f.allowed, f.err
}

func TestLogin_Success(t *testing.T) {
	userID := uuid.New()
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{userID: userID}, &fakeAccessClient{allowed: true})

	resp, err := svc.Login(context.Background(), "admin@volzhsky-instrument.ru", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v, want nil", err)
	}
	if resp.UserID != userID {
		t.Fatalf("Login() UserID = %v, want %v", resp.UserID, userID)
	}
	if resp.Role != "admin" {
		t.Fatalf("Login() Role = %q, want %q", resp.Role, "admin")
	}
	if len(storage.inserted) != 1 || storage.inserted[0].action != "Выполнен вход в систему" {
		t.Fatalf("Login() did not write an audit log entry, got %+v", storage.inserted)
	}
}

func TestLogin_NotAdmin_Forbidden(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{userID: uuid.New()}, &fakeAccessClient{allowed: false})

	_, err := svc.Login(context.Background(), "buyer@example.com", "secret")
	if err == nil {
		t.Fatal("Login() error = nil, want forbidden error")
	}
	if len(storage.inserted) != 0 {
		t.Fatalf("Login() wrote an audit log entry for a rejected login: %+v", storage.inserted)
	}
}

func TestLogin_AuthClientError(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{err: errors.New("boom")}, &fakeAccessClient{allowed: true})

	_, err := svc.Login(context.Background(), "admin@volzhsky-instrument.ru", "wrong")
	if err == nil {
		t.Fatal("Login() error = nil, want an error when the auth client fails")
	}
}

func TestGetSession_RequiresAdminRole(t *testing.T) {
	svc := NewAdminApiService(zerolog.Nop(), &fakeStorage{}, &fakeAuthClient{}, &fakeAccessClient{allowed: false})

	_, err := svc.GetSession(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("GetSession() error = nil, want forbidden error for a non-admin user")
	}
}

func TestListAuditLog_ReturnsItemsAndTotal(t *testing.T) {
	entry := postgres.AuditLogEntry{ID: uuid.New(), ActorUserID: uuid.New(), ActorLabel: "Админ портала", Action: "Портал запущен"}
	storage := &fakeStorage{entries: []postgres.AuditLogEntry{entry}, total: 1}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true})

	resp, err := svc.ListAuditLog(context.Background(), uuid.New(), 10, 0)
	if err != nil {
		t.Fatalf("ListAuditLog() error = %v, want nil", err)
	}
	if resp.Total != 1 {
		t.Fatalf("ListAuditLog() Total = %d, want 1", resp.Total)
	}
	if len(resp.Items) != 1 || resp.Items[0].Action != "Портал запущен" {
		t.Fatalf("ListAuditLog() Items = %+v, want the entry from storage", resp.Items)
	}
}

func TestListAuditLog_ClampsLimit(t *testing.T) {
	storage := &fakeStorage{}
	svc := NewAdminApiService(zerolog.Nop(), storage, &fakeAuthClient{}, &fakeAccessClient{allowed: true})

	if _, err := svc.ListAuditLog(context.Background(), uuid.New(), 0, -5); err != nil {
		t.Fatalf("ListAuditLog() error = %v, want nil", err)
	}
}
```

- [ ] **Step 2: Run the tests and confirm they fail to compile (service.go doesn't exist yet)**

```bash
cd back/admin
go test ./internal/service/... -v
```

Expected: build failure — `undefined: NewAdminApiService`.

- [ ] **Step 3: Implement the service**

```go
// back/admin/internal/service/service.go
package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	customErrors "github.com/mbatimel/AMC/admin/internal/errors"
	"github.com/mbatimel/AMC/admin/internal/storage/postgres"
	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
	"github.com/mbatimel/AMC/admin/pkg/models"
)

// RoleCodeAdmin matches the "admin" role seeded in the shared roles table
// (back/migrations/pkg/migrations/data/20260705171948_access_roles.sql).
const RoleCodeAdmin = 0

const actorLabelAdmin = "Админ портала"

const (
	defaultAuditLogLimit = 50
	maxAuditLogLimit     = 100
)

// Storage is implemented by internal/storage/postgres.Storage.
type Storage interface {
	InsertAuditLogEntry(ctx context.Context, actorUserID uuid.UUID, actorLabel string, action string) error
	ListAuditLogEntries(ctx context.Context, limit int, offset int) ([]postgres.AuditLogEntry, error)
	CountAuditLogEntries(ctx context.Context) (int, error)
}

// AuthClient is implemented by auth/pkg/client/transport.ClientAuthAPI.
type AuthClient interface {
	LoginUser(ctx context.Context, email string, password string) (userID uuid.UUID, err error)
}

// AccessClient is implemented by access/pkg/client/transport.ClientAccessAPI.
type AccessClient interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, role int) (allowed bool, err error)
}

type service struct {
	logger       zerolog.Logger
	storage      Storage
	authClient   AuthClient
	accessClient AccessClient
}

func NewAdminApiService(logger zerolog.Logger, storage Storage, authClient AuthClient, accessClient AccessClient) externalapi.AdminAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		authClient:   authClient,
		accessClient: accessClient,
	}
}

// requireAdmin checks that userID currently holds the admin role. Every
// AdminAPI method other than Login must call this before doing anything else.
func (s *service) requireAdmin(ctx context.Context, userID uuid.UUID) error {
	allowed, err := s.accessClient.CheckAccess(ctx, userID, RoleCodeAdmin)
	if err != nil {
		return customErrors.InternalServerError().SetOuterError(err)
	}
	if !allowed {
		return customErrors.ForbiddenError()
	}
	return nil
}

func (s *service) Login(ctx context.Context, email string, password string) (response models.LoginResponse, err error) {
	userID, err := s.authClient.LoginUser(ctx, email, password)
	if err != nil {
		return models.LoginResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	if err = s.requireAdmin(ctx, userID); err != nil {
		return models.LoginResponse{}, err
	}

	if err = s.storage.InsertAuditLogEntry(ctx, userID, actorLabelAdmin, "Выполнен вход в систему"); err != nil {
		s.logger.Error().Err(err).Msg("failed to write audit log entry")
	}

	return models.LoginResponse{UserID: userID, Role: "admin"}, nil
}

func (s *service) Logout(ctx context.Context, userID uuid.UUID) (err error) {
	if err = s.requireAdmin(ctx, userID); err != nil {
		return err
	}

	if err = s.storage.InsertAuditLogEntry(ctx, userID, actorLabelAdmin, "Выполнен выход из системы"); err != nil {
		s.logger.Error().Err(err).Msg("failed to write audit log entry")
	}

	return nil
}

func (s *service) GetSession(ctx context.Context, userID uuid.UUID) (response models.SessionResponse, err error) {
	if err = s.requireAdmin(ctx, userID); err != nil {
		return models.SessionResponse{}, err
	}
	return models.SessionResponse{UserID: userID, Role: "admin"}, nil
}

func (s *service) ListAuditLog(ctx context.Context, userID uuid.UUID, limit int, offset int) (response models.ListAuditLogResponse, err error) {
	if err = s.requireAdmin(ctx, userID); err != nil {
		return models.ListAuditLogResponse{}, err
	}

	if limit <= 0 || limit > maxAuditLogLimit {
		limit = defaultAuditLogLimit
	}
	if offset < 0 {
		offset = 0
	}

	entries, err := s.storage.ListAuditLogEntries(ctx, limit, offset)
	if err != nil {
		return models.ListAuditLogResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	total, err := s.storage.CountAuditLogEntries(ctx)
	if err != nil {
		return models.ListAuditLogResponse{}, customErrors.InternalServerError().SetOuterError(err)
	}

	items := make([]models.AuditLogEntry, 0, len(entries))
	for _, e := range entries {
		items = append(items, models.AuditLogEntry{
			ID:         e.ID,
			CreatedAt:  e.CreatedAt,
			ActorLabel: e.ActorLabel,
			Action:     e.Action,
		})
	}

	return models.ListAuditLogResponse{Items: items, Total: total}, nil
}
```

- [ ] **Step 4: Run the tests and verify they pass**

```bash
cd back/admin
go test ./internal/service/... -v
```

Expected: all 6 tests `PASS`.

- [ ] **Step 5: Commit**

```bash
git add back/admin/internal/service
git commit -m "feat(admin): implement login, session and audit log service"
```

---

### Task 7: Implement the custom HTTP handlers

**Files:**
- Create: `back/admin/internal/transport/custom-handlers/admin.go`
- Create: `back/admin/internal/transport/custom-handlers/response.go`

**Interfaces:**
- Consumes: `externalapi.AdminAPI` (Task 4).
- Produces: `custom_handlers.Login(ctx *fiber.Ctx, svc externalapi.AdminAPI, email, password string) error`, `custom_handlers.Logout(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error`, `custom_handlers.GetSession(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error`, `custom_handlers.ListAuditLog(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, limit, offset int) error` — consumed by Task 8's generated transport layer (referenced by the `@tg http-response=...` annotations from Task 4).

This layer is thin request/response glue with no branching logic of its own (identical in shape to `back/orders/internal/transport/custom-handlers/orders.go`); the codebase does not unit-test this layer anywhere (`orders`/`auth` don't either) — it's verified by `go build` here and by the manual smoke test in Task 8.

- [ ] **Step 1: Create the response helper**

```go
// back/admin/internal/transport/custom-handlers/response.go
package custom_handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

const errInternal = "admin.errors.internalError"

type RestResponse struct {
	Data             interface{}            `json:"data"`
	Error            bool                   `json:"error"`
	ErrorText        string                 `json:"errorText"`
	AdditionalErrors map[string]interface{} `json:"additionalErrors"`
}

type statusCoder interface {
	GetStatusCode() int
}

func sendResponse(ctx *fiber.Ctx, log zerolog.Logger, data interface{}, respError error) {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(http.StatusOK)

	response := &RestResponse{
		Data:  data,
		Error: respError != nil,
	}

	if response.Error {
		ctx.Response().SetStatusCode(http.StatusInternalServerError)
		response.ErrorText = errInternal
		response.AdditionalErrors = map[string]interface{}{
			"reason": respError.Error(),
		}

		if customErr, ok := respError.(statusCoder); ok && customErr.GetStatusCode() != 0 {
			ctx.Response().SetStatusCode(customErr.GetStatusCode())
		}
	}

	respBody, err := json.Marshal(response)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal response")
		return
	}

	if _, err = ctx.Write(respBody); err != nil {
		log.Error().Err(err).Msg("failed to send response")
		return
	}
}
```

- [ ] **Step 2: Create the handlers**

```go
// back/admin/internal/transport/custom-handlers/admin.go
package custom_handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	externalapi "github.com/mbatimel/AMC/admin/pkg/interfaces/externalapi"
	"github.com/rs/zerolog/log"
)

const ServiceName = "admin"

func Login(ctx *fiber.Ctx, svc externalapi.AdminAPI, email string, password string) error {
	return handle(ctx, "post", "/v1/admin/auth/login", "Login", map[string]interface{}{
		"email": email,
	}, func() (interface{}, error) {
		return svc.Login(ctx.UserContext(), email, password)
	})
}

func Logout(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error {
	return handle(ctx, "post", "/v1/admin/auth/logout", "Logout", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return nil, svc.Logout(ctx.UserContext(), userID)
	})
}

func GetSession(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID) error {
	return handle(ctx, "get", "/v1/admin/auth/session", "GetSession", map[string]interface{}{
		"userID": userID,
	}, func() (interface{}, error) {
		return svc.GetSession(ctx.UserContext(), userID)
	})
}

func ListAuditLog(ctx *fiber.Ctx, svc externalapi.AdminAPI, userID uuid.UUID, limit int, offset int) error {
	return handle(ctx, "get", "/v1/admin/audit-log", "ListAuditLog", map[string]interface{}{
		"userID": userID,
		"limit":  limit,
		"offset": offset,
	}, func() (interface{}, error) {
		return svc.ListAuditLog(ctx.UserContext(), userID, limit, offset)
	})
}

func handle(
	ctx *fiber.Ctx,
	method string,
	path string,
	methodName string,
	fields map[string]interface{},
	call func() (interface{}, error),
) error {
	var err error

	defer func(begin time.Time) {
		fields["method"] = method
		fields["path"] = path
		fields["methodName"] = methodName
		fields["serviceName"] = ServiceName
		fields["took"] = time.Since(begin).String()

		l := log.Info()
		if err != nil {
			l = log.Error().Err(err)
		}

		l.Fields(fields).Msg("call")
	}(time.Now())

	data, err := call()
	if err != nil {
		sendResponse(ctx, log.Logger, nil, err)
		return nil
	}

	sendResponse(ctx, log.Logger, data, nil)
	return nil
}
```

Note `Login`'s log fields deliberately omit `password` (unlike `auth`'s equivalent handler, which logs it in plaintext) — don't copy that part of the pattern.

- [ ] **Step 3: Verify it compiles**

```bash
cd back/admin
go build ./internal/transport/custom-handlers/...
```

Expected: exits 0.

- [ ] **Step 4: Commit**

```bash
git add back/admin/internal/transport/custom-handlers
git commit -m "feat(admin): implement HTTP handlers for AdminAPI"
```

---

### Task 8: Generate the transport layer, wire up `main.go`, and verify end-to-end

**Files:**
- Generate: `back/admin/internal/transport/jsonRPC/externalapi/*.go`
- Generate: `back/admin/swaggers/externalapi/swagger.yaml`
- Create: `back/admin/cmd/main.go`
- Modify: `README.md`

**Interfaces:**
- Consumes: everything from Tasks 1–7.
- Produces: a runnable `admin-api` binary listening on `:8083` (health `:9093`).

- [ ] **Step 1: Run the transport generator**

```bash
cd back/admin
go generate ./pkg/interfaces/externalapi/...
```

Expected: `back/admin/internal/transport/jsonRPC/externalapi/` is populated (mirroring `back/orders/internal/transport/jsonRPC/externalapi/`: `adminapi-server.go`, `adminapi-rest.go`, `adminapi-http.go`, `fiber.go`, `options.go`, `metrics.go`, `context.go`, `header.go`, `errors.go`, `version.go`, `viewer/`, etc., exposing `externalapi.New(logger zerolog.Logger, api externalapi.AdminAPI) *Server` with `.WithLog()` / `.WithMetrics()` and `.Fiber() *fiber.App`), and `back/admin/swaggers/externalapi/swagger.yaml` is created.

- [ ] **Step 2: Write `cmd/main.go`**

```go
// back/admin/cmd/main.go
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"

	accessTransport "github.com/mbatimel/AMC/access/pkg/client/transport"
	authTransport "github.com/mbatimel/AMC/auth/pkg/client/transport"

	"github.com/mbatimel/AMC/admin/internal/config"
	adminService "github.com/mbatimel/AMC/admin/internal/service"
	postgres "github.com/mbatimel/AMC/admin/internal/storage/postgres"
	transportHttp "github.com/mbatimel/AMC/admin/internal/transport/http"
	"github.com/mbatimel/AMC/admin/internal/transport/jsonRPC/externalapi"
)

const serviceName = "admin-api"

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Str("serviceName", serviceName).Logger()

	if err := config.LoadEnvFile(".env"); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("load env file")
	}

	cfg := config.LoadConfig()

	pool, err := postgres.NewPool(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer pool.Close()

	postgresStorage := postgres.New(pool)
	access := accessTransport.NewClientAccessAPI(cfg.AccessURL)
	auth := authTransport.NewClientAuthAPI(cfg.AuthURL)
	svc := adminService.NewAdminApiService(log.Logger, postgresStorage, auth, access)

	app := externalapi.New(log.Logger, externalapi.AdminAPI(externalapi.NewAdminAPI(svc))).WithLog().WithMetrics()
	server := &fasthttp.Server{
		Handler: app.Fiber().Handler(),
	}

	healthServer := transportHttp.NewHealthServer()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info().Str("address", cfg.BindAddr).Msg("admin external api server started")
		if serveErr := server.ListenAndServe(cfg.BindAddr); serveErr != nil {
			log.Fatal().Err(serveErr).Msg("failed to listen and serve admin server")
		}
	}()

	go func() {
		if healthErr := healthServer.Start(":9093"); healthErr != nil {
			log.Error().Err(healthErr).Msg("failed to start health server")
		}
	}()

	<-shutdown

	if err = healthServer.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to stop health server")
	}

	if err = server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown server")
	}
}
```

- [ ] **Step 3: Full module build and test run**

```bash
cd back/admin
go mod tidy
go build ./...
go vet ./...
go test ./...
```

Expected: all exit 0.

- [ ] **Step 4: Manual smoke test**

With Postgres, `access` (`:8080`), and `auth` (`:8081`) already running locally, and an existing user that has the `admin` role (assign one via `access`'s `AddRole` if needed):

```bash
cd back/admin
PG_DB=AMC PG_USER=mbatimel PG_PASSWORD=mbatimel ACCESS_URL=http://localhost:8080 AUTH_URL=http://localhost:8081 go run ./cmd
```

In another terminal:

```bash
curl -s -X POST 'http://localhost:8083/api/v1/admin/auth/login' \
  -H 'Content-Type: application/json' \
  --data-urlencode 'email=admin@volzhsky-instrument.ru' \
  -G --data-urlencode 'password=<the account password>'
```

Expected: `{"data":{"userID":"...","role":"admin"},"error":false,...}`. Copy the returned `userID`, then:

```bash
curl -s 'http://localhost:8083/api/v1/admin/audit-log?limit=10&offset=0' -H "X-User-Id: <userID>"
```

Expected: `{"data":{"items":[{"...","action":"Выполнен вход в систему",...}],"total":1},"error":false,...}`.

- [ ] **Step 5: Document the new service in the root README**

In `README.md`, add a row to the "Backend-сервисы" table (after the `back/warehouses` row):

```markdown
| `back/admin` | - | Админ-панель: вход сотрудников, журнал действий; отдельный backend для будущих разделов (контент, каталог, пользователи, отзывы). |
```

- [ ] **Step 6: Commit**

```bash
git add back/admin/internal/transport/jsonRPC back/admin/swaggers/externalapi/swagger.yaml back/admin/cmd back/admin/go.mod back/admin/go.sum README.md
git commit -m "feat(admin): wire transport layer and main entrypoint"
```

---

## Self-Review

**Spec coverage:**
- Вход в админку (login) → Task 4 `Login` + Task 6 + Task 8 smoke test. ✓
- Отдельная аутентификация для сотрудников без дублирования данных → Task 1 (auth client) + Task 6 `requireAdmin` (access client, role code 0). ✓
- Журнал действий (audit log, no delete) → Task 2 (table, no delete statement anywhere), Task 4 `ListAuditLog`, Task 5 storage, Task 6 service, Task 7 handler. ✓
- Заголовок "Роль: Администратор портала" shown on every admin screen → `GetSession`/`Login` both return `Role: "admin"` for the frontend to render. ✓
- Everything else from the 19 Figma screens (banners, pages, catalog proxy, clients, invites, reviews, dashboard) is explicitly out of scope for this plan per the user's choice of "Фундамент" — left for follow-up plans.

**Placeholder scan:** no TBD/TODO, no "add error handling" hand-waves, no "similar to Task N" — every step has literal file content or a literal runnable command.

**Type consistency:** `models.LoginResponse`/`SessionResponse`/`AuditLogEntry`/`ListAuditLogResponse` (Task 4) are used with identical field names in Task 6 (service) and referenced identically in Task 7's handler return values. `postgres.AuditLogEntry` (Task 5) fields (`ID, ActorUserID, ActorLabel, Action, CreatedAt`) match exactly what Task 6's `ListAuditLog` reads. `service.Storage`/`AuthClient`/`AccessClient` method signatures (Task 6) match `postgres.Storage` (Task 5), `authTransport.ClientAuthAPI` (Task 1), and the existing `accessTransport.ClientAccessAPI` exactly.
