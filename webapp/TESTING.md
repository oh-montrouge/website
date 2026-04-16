# Testing

## Strategy

Tests are split by layer, with a different tool and different isolation level for each:

| Layer | Kind | Tool | DB? |
|---|---|---|---|
| `models/` | Integration | testcontainers-go + plain `testing` | Real PostgreSQL (ephemeral container) |
| `actions/` | Unit | `net/http/httptest` + plain `testing` | No — stubs only |

The split is deliberate. Models are the DB boundary; integration tests there verify that SQL queries, ordering, and constraints behave correctly against a real engine. Actions are orchestration; unit tests there verify HTTP concerns (status codes, redirects, what gets set on the context) without caring about the DB at all.

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

Every handler struct declares its dependencies as interfaces defined in `actions/repositories.go`. The production wiring in `app.go` passes the real model store; tests pass stubs. See `webapp/README.md` § Architecture conventions for the full rationale.

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
