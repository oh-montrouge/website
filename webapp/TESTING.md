# Testing

## Strategy

| Layer | Kind | Tool | DB? |
|---|---|---|---|
| `models/` | Integration | testcontainers-go + plain `testing` | Real PostgreSQL (ephemeral container) |
| `services/` | Unit | plain `testing` + stub repositories | No — stubs only |
| `actions/` | Unit | `net/http/httptest` + plain `testing` | No — stubs only |
| All app | End-to-end | Playwright | Real PostgreSQL (ephemeral container) |

Models are the DB boundary; integration tests there verify SQL against a real engine. Services contain business rules; unit tests there verify logic with stubbed repositories. Actions are HTTP orchestration; unit tests there verify status codes and rendering.

---

## Model integration tests (`models/`) — Target 100% coverage

Each `go test ./models/...` run starts a fresh `postgres:17-alpine` container via testcontainers-go, applies all migrations, runs the tests, then tears the container down. Docker daemon must be available.

`TestMain` in `models/models_test.go` wires the container and overrides the package-level `models.DB`. Tests call `truncateAll(t)` first to reset state; seed data from migrations is wiped on the first call.

```go
func TestSomething(t *testing.T) {
    truncateAll(t) // always first

    // insert fixtures via DB.Create(...)
    // exercise model function
    // assert with if/t.Fatal or testify/assert
}
```

---

## Service unit tests (`services/`) — Target 100% coverage

Services depend on repository interfaces. Tests define a stub struct that implements the interface and returns fixed data — no DB, no container. Tests live in `package services_test`.

```go
type stubAccountRepo struct {
    account *models.Account
    err     error
}
func (s stubAccountRepo) FindByEmail(_ *pop.Connection, _ string) (*models.Account, error) {
    return s.account, s.err
}

func TestSomeService_SomeCase(t *testing.T) {
    svc := services.SomeService{Accounts: stubAccountRepo{err: errors.New("not found")}}
    _, err := svc.SomeMethod(nil, ...)
    assert.ErrorIs(t, err, services.ErrAccountNotFound)
}
```

Pass `nil` for `*pop.Connection` — stubs ignore it.

---

## Action unit tests (`actions/`) — Target 100% coverage

Handlers declare repository dependencies as interfaces (defined in `services/repositories.go`). Tests inject stubs; production wiring in `app.go` passes the real store.

Each test builds a minimal Buffalo app via `newTestApp` (in `actions_test.go`): no DB middleware, i18n registered, `authenticity_token` set to `""`, `"tx"` set to a typed nil `*pop.Connection`.

```go
type stubWidgets struct{ data models.Widgets }
func (s stubWidgets) List(_ *pop.Connection) (models.Widgets, error) { return s.data, nil }

func TestWidgetsIndex_ReturnsWidgets(t *testing.T) {
    h := WidgetsHandler{Widgets: stubWidgets{data: models.Widgets{{ID: 1, Name: "Foo"}}}}
    app := newTestApp(func(a *buffalo.App) {
        a.GET("/widgets", h.Index)
    })

    res := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/widgets", nil)
    app.ServeHTTP(res, req)

    assert.Equal(t, http.StatusOK, res.Code)
    assert.Contains(t, res.Body.String(), "Foo")
}
```

---

## End-to-end tests — Target "human-verified acceptance criterias"

E2e tests live under `e2e/` at the repo root, implemented with **Playwright (TypeScript)**. See [ADR 010](../specs/technical-adrs/010-e2e-framework.md) for the framework choice.

### Prerequisites

- `mise install` (installs Node.js, Playwright, and Chromium)
- Docker running (for local PostgreSQL)
- Admin credentials in environment:
  ```
  export E2E_ADMIN_EMAIL=admin@example.com
  export E2E_ADMIN_PASSWORD=yourpassword
  ```

### Running locally

```bash
mise run test:e2e
```

Playwright's `globalSetup` starts PostgreSQL via Docker Compose, runs migrations, builds a coverage-instrumented binary, and starts it. `globalTeardown` stops the binary and the container.

To run headed or in UI mode:
```bash
cd e2e && npx playwright test --headed
cd e2e && npx playwright test --ui
```

### Test structure

```
e2e/
  fixtures/index.ts       # shared helpers: loginAsAdmin, createMusician, createEvent, …
  tests/
    navigation.spec.ts
    account-lifecycle.spec.ts
    musicians.spec.ts
    seasons.spec.ts
    fee-payments.spec.ts
    events-rsvp.spec.ts
  global-setup.ts         # starts DB + app
  global-teardown.ts      # stops app + DB
  playwright.config.ts
```

Each test creates its own data (no dependency on `db/dummy-data/`) and is independent of other tests. Tests run serially (`workers: 1`) to avoid database race conditions.

### Traces and artifacts

On failure, Playwright records a trace (network + DOM timeline) and a screenshot. Open a trace with:

```bash
npx playwright show-trace path/to/trace.zip
```

---

## Running tests

```bash
mise run test        # Go unit/integration suite (Docker must be running)
mise run test:e2e    # Playwright browser suite (Docker must be running)
```

`mise run test` runs `GO_ENV=test go test -race ./...` from `webapp/`. The race detector is always on.
