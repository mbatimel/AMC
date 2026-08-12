# Проверка ИНН через api-fns.ru в RegisterIP — план реализации

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `RegisterIP` дополнительно проверяет ИНН через внешний сервис api-fns.ru (`GET /api/fl_status`), используя только блок `Корректность` ответа; любой сбой запроса — fail-closed (регистрация отклоняется); полный сырой ответ логируется всегда.

**Architecture:** Новый пакет `internal/client/fns` инкапсулирует HTTP-вызов и парсинг ответа за интерфейсом `service.FnsClient` (`CheckIndividual(ctx, inn) (bool, error)`), внедряемым в `service` через `NewAuthApiService`, по образцу уже существующего `AccessClient`.

**Tech Stack:** Go 1.25, `github.com/valyala/fasthttp` (уже в зависимостях), `github.com/rs/zerolog`, стандартный `net/http/httptest` для тестов.

## Global Constraints

- Адрес/ключ api-fns.ru берутся из env `API_FNS_ADDR` / `API_FNS_KEY` (уже в `back/auth/.env`) — не хардкодить.
- Проверяется только блок JSON-ответа `Корректность` (`КонтрСумма`, `Недействительный`). Остальные блоки (`Самозанятость`, `ИП`, `ДисквЛицо`, `Банкрот`, `ПоддержкаМСП`) не парсятся.
- Fail-closed: любая ошибка запроса/парсинга/`null`-полей → `RegisterIP` возвращает ошибку, регистрация не проходит.
- Внешний вызов — **после** локальной checksum-проверки `validate()` в `common.go`.
- Полный сырой HTTP-ответ логируется всегда (успех и ошибка) через `zerolog`, `Info` level.
- HTTP-таймаут клиента — 5s, фиксированный.
- Никаких моков через codegen — в этом репо мокают вручную (см. `internal/transport/jsonRPC/externalapi/change_password_test.go`).

---

### Task 1: Config — FnsAddr/FnsKey

**Files:**
- Modify: `back/auth/internal/config/config.go`

**Interfaces:**
- Produces: `config.Config.FnsAddr string`, `config.Config.FnsKey string` — читаются из уже существующих env `API_FNS_ADDR`, `API_FNS_KEY`.

- [ ] **Step 1: Добавить поля и чтение env**

В `Config` добавить:

```go
	FnsAddr    string
	FnsKey     string
```

В `LoadConfig()` добавить в конструктор `cfg`:

```go
		FnsAddr:    strings.TrimSpace(os.Getenv("API_FNS_ADDR")),
		FnsKey:     os.Getenv("API_FNS_KEY"),
```

(`strings.TrimSpace` нужен, т.к. в `.env` значение `API_FNS_ADDR= https://...` — с пробелом после `=`, парсер `LoadEnvFile` уже делает `TrimSpace` для всей строки-значения, но подстрахуемся на уровне конфига тоже не будем — `LoadEnvFile` уже вызывает `strings.TrimSpace(strings.Trim(...))`, так что в `os.Getenv` пробела не будет. `strings.TrimSpace` здесь избыточен — **не добавлять**, просто `os.Getenv("API_FNS_ADDR")`.)

Итоговая правка `LoadConfig()`:

```go
func LoadConfig() Config {
	cfg := Config{
		PGHost:     GetEnv("PG_HOST", "localhost"),
		PGPort:     GetEnv("PG_PORT", "5432"),
		PGDB:       os.Getenv("PG_DB"),
		PGUser:     os.Getenv("PG_USER"),
		PGPassword: os.Getenv("PG_PASSWORD"),
		BindAddr:   GetEnv("BIND_ADDR", ":8081"),
		AccessURL:  os.Getenv("ACCESS_URL"),
		FnsAddr:    os.Getenv("API_FNS_ADDR"),
		FnsKey:     os.Getenv("API_FNS_KEY"),
	}

	if cfg.PGDB == "" || cfg.PGUser == "" || cfg.PGPassword == "" {
		log.Fatal().Msg("PG_DB, PG_USER and PG_PASSWORD must be specified")
	}
	if cfg.AccessURL == "" {
		log.Fatal().Msg("ACCESS_URL must be specified")
	}
	if cfg.FnsAddr == "" || cfg.FnsKey == "" {
		log.Fatal().Msg("API_FNS_ADDR and API_FNS_KEY must be specified")
	}
	return cfg
}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/auth && go build ./...`
Expected: без ошибок компиляции (тестов на config.go в репо нет, отдельный тест не заводим — YAGNI, поведение тривиальное и покрывается end-to-end сборкой сервиса).

- [ ] **Step 3: Commit**

```bash
cd back/auth
git add internal/config/config.go
git commit -m "feat(auth): add FnsAddr/FnsKey to config"
```

---

### Task 2: `internal/client/fns` — HTTP-клиент api-fns.ru

**Files:**
- Create: `back/auth/internal/client/fns/client.go`
- Test: `back/auth/internal/client/fns/client_test.go`

**Interfaces:**
- Consumes: ничего из предыдущих задач напрямую (адрес/ключ передаются как строки при конструировании).
- Produces:
  ```go
  package fns

  func New(addr, key string, logger zerolog.Logger) *Client
  func (c *Client) CheckIndividual(ctx context.Context, inn string) (valid bool, err error)
  ```
  `Client` реализует `service.FnsClient` (интерфейс появится в Task 3) структурно — без явного объявления соответствия.

- [ ] **Step 1: Написать падающий тест**

`back/auth/internal/client/fns/client_test.go`:

```go
package fns

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func testLogger() zerolog.Logger {
	return zerolog.Nop()
}

func TestCheckIndividual_ValidInn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("inn") != "773208978609" || r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		w.Write([]byte(`{"Корректность":{"КонтрСумма":true,"Недействительный":false}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	valid, err := c.CheckIndividual(context.Background(), "773208978609")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !valid {
		t.Fatal("expected valid=true")
	}
}

func TestCheckIndividual_InvalidChecksum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Корректность":{"КонтрСумма":false,"Недействительный":false}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	valid, err := c.CheckIndividual(context.Background(), "773208978609")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Fatal("expected valid=false")
	}
}

func TestCheckIndividual_MarkedInvalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Корректность":{"КонтрСумма":true,"Недействительный":true}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	valid, err := c.CheckIndividual(context.Background(), "773208978609")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Fatal("expected valid=false when Недействительный=true")
	}
}

func TestCheckIndividual_NullFields_FailClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Корректность":{"КонтрСумма":null,"Недействительный":null}}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	_, err := c.CheckIndividual(context.Background(), "773208978609")
	if err == nil {
		t.Fatal("expected error on null fields")
	}
}

func TestCheckIndividual_NonOKStatus_FailClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	_, err := c.CheckIndividual(context.Background(), "773208978609")
	if err == nil {
		t.Fatal("expected error on non-200 status")
	}
}

func TestCheckIndividual_MalformedJSON_FailClosed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := New(server.URL, "test-key", testLogger())
	_, err := c.CheckIndividual(context.Background(), "773208978609")
	if err == nil {
		t.Fatal("expected error on malformed json")
	}
}
```

- [ ] **Step 2: Запустить тест, убедиться что падает**

Run: `cd back/auth && go test ./internal/client/fns/... -v`
Expected: FAIL — `undefined: New` (пакет ещё не существует, кроме файла теста).

- [ ] **Step 3: Реализовать клиент**

`back/auth/internal/client/fns/client.go`:

```go
package fns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
)

const requestTimeout = 5 * time.Second

type Client struct {
	addr   string
	key    string
	http   *fasthttp.Client
	logger zerolog.Logger
}

func New(addr, key string, logger zerolog.Logger) *Client {
	return &Client{
		addr:   addr,
		key:    key,
		http:   &fasthttp.Client{},
		logger: logger,
	}
}

type flStatusResponse struct {
	Korrektnost struct {
		KontrSumma     *bool `json:"КонтрСумма"`
		Nedeystvitelny *bool `json:"Недействительный"`
	} `json:"Корректность"`
}

// CheckIndividual queries api-fns.ru fl_status and reports whether inn is
// correct according to the Корректность block only. Any transport, HTTP,
// or parsing failure returns valid=false with a non-nil error (fail-closed) —
// the caller must reject registration rather than assume validity.
func (c *Client) CheckIndividual(ctx context.Context, inn string) (valid bool, err error) {
	reqURL := fmt.Sprintf("%s?inn=%s&key=%s", c.addr, url.QueryEscape(inn), url.QueryEscape(c.key))

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(reqURL)
	req.Header.SetMethod(fasthttp.MethodGet)

	doErr := c.http.DoTimeout(req, resp, requestTimeout)

	statusCode := resp.StatusCode()
	body := string(resp.Body())

	logEvent := c.logger.Info().Str("inn", inn).Int("status", statusCode).Str("response", body)
	if doErr != nil {
		logEvent.Err(doErr).Msg("fns fl_status request failed")
		return false, fmt.Errorf("fns fl_status request: %w", doErr)
	}
	logEvent.Msg("fns fl_status response")

	if statusCode != fasthttp.StatusOK {
		return false, fmt.Errorf("fns fl_status: unexpected status %d", statusCode)
	}

	var parsed flStatusResponse
	if err = json.Unmarshal(resp.Body(), &parsed); err != nil {
		return false, fmt.Errorf("fns fl_status: parse response: %w", err)
	}

	if parsed.Korrektnost.KontrSumma == nil || parsed.Korrektnost.Nedeystvitelny == nil {
		return false, fmt.Errorf("fns fl_status: Корректность fields are null, retry later")
	}

	return *parsed.Korrektnost.KontrSumma && !*parsed.Korrektnost.Nedeystvitelny, nil
}
```

- [ ] **Step 4: Запустить тесты, убедиться что проходят**

Run: `cd back/auth && go test ./internal/client/fns/... -v`
Expected: PASS — все 6 тестов.

- [ ] **Step 5: Commit**

```bash
cd back/auth
git add internal/client/fns/
git commit -m "feat(auth): add fns.Client for api-fns.ru fl_status check"
```

---

### Task 3: `errors.InnInvalidError`

**Files:**
- Modify: `back/auth/internal/errors/common.go`

**Interfaces:**
- Produces: `errors.InnInvalidError(inn string) *Error` — HTTP 400, code `ErrInvalidRequest` (уже существующий).

- [ ] **Step 1: Добавить ошибку**

В `back/auth/internal/errors/common.go`, в блок `var (...)`, после `InnEmptyErr`:

```go
	InnInvalidError = func(inn string) *Error {
		return New("inn is invalid", fasthttp.StatusBadRequest, ErrInvalidRequest).AddCause("inn", inn)
	}
```

- [ ] **Step 2: Собрать пакет**

Run: `cd back/auth && go build ./internal/errors/...`
Expected: без ошибок.

- [ ] **Step 3: Commit**

```bash
cd back/auth
git add internal/errors/common.go
git commit -m "feat(auth): add InnInvalidError"
```

---

### Task 4: `service.FnsClient` + интеграция в `RegisterIP`

**Files:**
- Modify: `back/auth/internal/service/service.go`
- Test: `back/auth/internal/service/register_ip_test.go` (новый файл)

**Interfaces:**
- Consumes: `customErrors.InnInvalidError(inn string) *Error` (Task 3), `fns.Client` структурно реализует то, что определяется здесь (Task 2 уже написан с точно этой сигнатурой).
- Produces: `service.FnsClient` interface — используется в Task 5 (`cmd/main.go`) для передачи `*fns.Client`.

- [ ] **Step 1: Написать падающий тест**

`back/auth/internal/service/register_ip_test.go`:

```go
package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/mbatimel/AMC/auth/internal/storage/postgres"
)

type stubStorage struct {
	createIPUserCalled bool
}

func (s *stubStorage) GetUserByEmail(ctx context.Context, email string) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (postgres.User, error) {
	return postgres.User{}, postgres.ErrUserNotFound
}

func (s *stubStorage) CreateIPUser(
	ctx context.Context,
	email, passwordHash string,
	fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
	directorFullName, directorPosition, phone, additionalPhone, website,
	bankAccount, bankName, bankBik, correspondentAccount *string,
	roleCode int,
) (uuid.UUID, error) {
	s.createIPUserCalled = true
	return uuid.New(), nil
}

func (s *stubStorage) CreateIndividualUser(
	ctx context.Context,
	email, passwordHash, surename, name, middleName, phone, city, deliveryAddress string,
	inn *string,
	roleCode int,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (s *stubStorage) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return nil
}

type stubFnsClient struct {
	valid   bool
	err     error
	calls   int
	lastINN string
}

func (f *stubFnsClient) CheckIndividual(ctx context.Context, inn string) (bool, error) {
	f.calls++
	f.lastINN = inn
	return f.valid, f.err
}

const validINN = "773208978609" // passes local checksum (validate12)

func registerIPArgs(inn *string) []*string {
	// fullName, shortName, inn, kpp, ogrn, okved, taxSystem, legalAddress, actualAddress,
	// directorFullName, directorPosition, phone, additionalPhone, website,
	// bankAccount, bankName, bankBik, correspondentAccount
	empty := ""
	return []*string{
		&empty, &empty, inn, &empty, &empty, &empty, &empty, &empty, &empty,
		&empty, &empty, &empty, &empty, &empty,
		&empty, &empty, &empty, &empty,
	}
}

func TestRegisterIP_FnsValid_CreatesUser(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	inn := validINN
	args := registerIPArgs(&inn)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !storage.createIPUserCalled {
		t.Fatal("expected CreateIPUser to be called")
	}
	if fnsClient.calls != 1 || fnsClient.lastINN != validINN {
		t.Fatalf("expected fns client called once with %q, got calls=%d inn=%q", validINN, fnsClient.calls, fnsClient.lastINN)
	}
}

func TestRegisterIP_FnsInvalid_RejectsRegistration(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: false}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	inn := validINN
	args := registerIPArgs(&inn)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when fns marks inn invalid")
	}
}

func TestRegisterIP_FnsError_FailsClosed(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{err: errors.New("timeout")}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	inn := validINN
	args := registerIPArgs(&inn)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if storage.createIPUserCalled {
		t.Fatal("CreateIPUser must not be called when fns check errors")
	}
}

func TestRegisterIP_LocalChecksumInvalid_NeverCallsFns(t *testing.T) {
	storage := &stubStorage{}
	fnsClient := &stubFnsClient{valid: true}
	svc := &service{logger: zerolog.Nop(), storage: storage, fnsClient: fnsClient}

	badINN := "773208978608" // validINN with last digit flipped — fails validate12 checksum
	args := registerIPArgs(&badINN)
	_, err := svc.RegisterIP(context.Background(), "user@example.com", "password",
		args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8],
		args[9], args[10], args[11], args[12], args[13],
		args[14], args[15], args[16], args[17],
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if fnsClient.calls != 0 {
		t.Fatal("fns client must not be called when local checksum is already invalid")
	}
}
```

- [ ] **Step 2: Запустить тесты, убедиться что падают**

Run: `cd back/auth && go test ./internal/service/... -run TestRegisterIP -v`
Expected: FAIL — `service` struct has no field `fnsClient` / `AccessClient` required in struct literal missing is fine (Go zero value), реальная ошибка компиляции: `unknown field fnsClient in struct literal`.

- [ ] **Step 3: Добавить интерфейс, поле и интеграцию в RegisterIP**

В `back/auth/internal/service/service.go`, после блока `AccessClient`:

```go
// FnsClient is implemented by internal/client/fns.Client.
type FnsClient interface {
	CheckIndividual(ctx context.Context, inn string) (valid bool, err error)
}
```

В `service` struct добавить поле:

```go
type service struct {
	logger       zerolog.Logger
	storage      Storage
	accessClient AccessClient
	fnsClient    FnsClient
}
```

`NewAuthApiService` — новый параметр:

```go
func NewAuthApiService(logger zerolog.Logger, storage Storage, accessClient AccessClient, fnsClient FnsClient) externalAPI.AuthAPI {
	return &service{
		logger:       logger,
		storage:      storage,
		accessClient: accessClient,
		fnsClient:    fnsClient,
	}
}
```

В `RegisterIP`, сразу после существующего блока локальной checksum-проверки (`if !valid{ return uuid.Nil,fmt.Errorf(...) }`) и до `passwordHash, err := bcrypt...`, вставить:

```go
	fnsValid, err := s.fnsClient.CheckIndividual(ctx, *inn)
	if err != nil {
		return uuid.Nil, customErrors.InternalServerError().SetOuterError(err)
	}
	if !fnsValid {
		return uuid.Nil, customErrors.InnInvalidError(*inn)
	}
```

- [ ] **Step 4: Запустить тесты, убедиться что проходят**

Run: `cd back/auth && go test ./internal/service/... -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd back/auth
git add internal/service/service.go internal/service/register_ip_test.go
git commit -m "feat(auth): check inn via api-fns.ru in RegisterIP"
```

---

### Task 5: Wiring в `cmd/main.go`

**Files:**
- Modify: `back/auth/cmd/main.go`

**Interfaces:**
- Consumes: `fns.New(addr, key string, logger zerolog.Logger) *fns.Client` (Task 2), `authService.NewAuthApiService(logger, storage, accessClient, fnsClient)` (Task 4, новая сигнатура).

- [ ] **Step 1: Добавить импорт и создание клиента**

В `back/auth/cmd/main.go` добавить импорт:

```go
	fnsClientPkg "github.com/mbatimel/AMC/auth/internal/client/fns"
```

Изменить строку создания `svc`:

```go
	postgresStorage := postgres.New(pool)
	access := accessTransport.NewClientAccessAPI(cfg.AccessURL)
	fnsClient := fnsClientPkg.New(cfg.FnsAddr, cfg.FnsKey, log.Logger)
	svc := authService.NewAuthApiService(log.Logger, postgresStorage, access, fnsClient)
```

- [ ] **Step 2: Собрать весь сервис**

Run: `cd back/auth && go build ./...`
Expected: без ошибок.

- [ ] **Step 3: Прогнать все тесты пакета auth**

Run: `cd back/auth && go test ./...`
Expected: PASS по всем пакетам.

- [ ] **Step 4: Commit**

```bash
cd back/auth
git add cmd/main.go
git commit -m "feat(auth): wire fns.Client into auth service"
```

---

## Self-Review Notes

- Спецификация покрыта: config (Task 1), HTTP-клиент + логирование полного ответа + fail-closed (Task 2), новая ошибка (Task 3), интеграция в `RegisterIP` + порядок после checksum-проверки (Task 4), wiring (Task 5).
- Банкрот-блок нигде не парсится и не упомянут в коде плана — соответствует последнему решению пользователя.
- Тесты для `RegisterIP` используют существующий `service` struct напрямую (unexported), т.к. `register_ip_test.go` лежит в том же пакете `service` — mock-подход зеркалит `changePasswordRecorder` из `change_password_test.go` (ручные стабы, без генераторов).
