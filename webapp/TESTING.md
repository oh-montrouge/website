# Testing

## Strategy

Tests are split by layer, with a different tool and different isolation level for each:

| Layer | Kind | Tool | DB? |
|---|---|---|---|
| `models/` | Integration | testcontainers-go + plain `testing` | Real PostgreSQL (ephemeral container) |
| `services/` | Unit | plain `testing` + stub repositories | No — stubs only |
| `actions/` | Unit | `net/http/httptest` + plain `testing` | No — stubs only |

The split is deliberate. Models are the DB boundary; integration tests there verify that SQL queries, ordering, and constraints behave correctly against a real engine. Services contain business rules; unit tests there verify logic in isolation with stubbed repositories. Actions are HTTP orchestration; unit tests there verify status codes and rendering without any DB or domain logic.

---

## Model integration tests (`models/`)

### How it works

Each run of `go test ./models/...` starts a fresh `postgres:17-alpine` container via testcontainers-go, applies all migrations, runs the tests, then tears the container down. Docker daemon must be available.

The entry point is `TestMain` in `models/models_test.go`:

1. Starts the container (random host port, ephemeral)
2. Opens a Pop connection to it
3. Overrides the package-level `models.DB` with that connection
4. Runs `pop.NewFileMigrator("../../db/migrations", conn).Up()` — this gives the tests a fully migrated schema
5. Calls `m.Run()` to run all test functions

Tests use `models.DB` directly and call `truncateAll(t)` at the top to reset state between runs. Seed data inserted by migrations is wiped on the first `truncateAll` call; tests insert their own fixtures.

### Why not gobuffalo/suite?

`gobuffalo/suite` creates its own `pop.Connect("test")` call internally, which bypasses the `TestMain`-controlled connection. Plain `testing` with a package-level `DB` variable gives TestMain full control.

### Writing a new model test

```go
func TestSomething(t *testing.T) {
    truncateAll(t) // always first

    // insert your own fixtures via DB.Create(...)
    // exercise your model function
    // assert with standard if/t.Fatal patterns, or import testify/assert
}
```

---

## Service unit tests (`services/`)

### How it works

Services depend on repository interfaces, not on Pop directly. Tests define a stub struct that implements the interface and returns fixed data or errors — no DB, no container.

Tests live in `package services_test` (external test package) to verify the public API.

### Writing a new service test

```go
// 1. Define a stub for the repository the service uses
type stubAccountRepo struct {
    account *models.Account
    err     error
}
func (s stubAccountRepo) FindByEmail(_ *pop.Connection, _ string) (*models.Account, error) {
    return s.account, s.err
}
func (s stubAccountRepo) GetByID(_ *pop.Connection, _ int64) (*models.Account, error) {
    return s.account, s.err
}

// 2. Inject the stub and exercise the service method
func TestSomeService_SomeCase(t *testing.T) {
    svc := services.SomeService{Accounts: stubAccountRepo{err: errors.New("not found")}}
    _, err := svc.SomeMethod(nil, ...)
    assert.ErrorIs(t, err, services.ErrAccountNotFound)
}
```

Pass `nil` for `*pop.Connection` — stubs ignore it, and service logic must not dereference it either.

---

## Action unit tests (`actions/`)

### How it works

Handlers depend on repository interfaces (e.g. `InstrumentRepository`), not on concrete Pop calls. Tests inject stubs that satisfy the interface and return fixed data — no DB involved.

Each test builds a minimal Buffalo app via `newTestApp` (defined in `actions_test.go`):

- No DB middleware, no session store, no CSRF middleware
- i18n translations registered (templates use `t()`)
- `authenticity_token` set to `""` in context (layout requires it)
- `"tx"` set to a typed nil `*pop.Connection` in context so handlers can extract it without panicking — stubs must not dereference it

Requests go through `httptest.NewRecorder` / `httptest.NewRequest` and `app.ServeHTTP`.

### Writing a new handler test

```go
// 1. Define a stub for each repository the handler uses
type stubWidgets struct{ data models.Widgets }
func (s stubWidgets) List(_ *pop.Connection) (models.Widgets, error) { return s.data, nil }

// 2. Build the handler with the stub injected
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

### Repository interfaces

Handler structs declare their dependencies as interfaces defined in `services/repositories.go`. The production wiring in `app.go` passes the real model store; tests pass stubs. See `webapp/README.md` § Architecture conventions for the full rationale.

---

## End-to-end tests

Not implemented in V1. Buffalo has no built-in browser automation, and the cost of
setting up and maintaining e2e tests (flakiness, full-stack CI requirement) outweighs
the benefit at this stage.

The flows that would most warrant e2e coverage — invite flow, password reset, multi-step
RSVP submission — are Phase 3 features that don't exist yet.

If introduced later, **playwright-go** is the most mature Go option. It would live in a
separate `e2e/` directory at the repo root (not inside `webapp/`) and run as an optional
CI job against a staging environment.

## Running tests

```bash
mise run test          # full suite (Docker must be running)
```

Equivalent to `GO_ENV=test go test -race ./...` from `webapp/`.

The race detector is always on. testcontainers pulls `postgres:17-alpine` on first run; subsequent runs reuse the cached image and are faster.
