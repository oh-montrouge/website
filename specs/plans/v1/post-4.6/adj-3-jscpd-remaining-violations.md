# jscpd Remaining Violations — Decision Required

After fixing all production-code duplicates (T2a–f), jscpd now exits 0 at **3.77%** total.
The 43 remaining clones are exclusively in two categories: **test files** and **templates**.
Neither is touched by the pre-commit hook in its current state, but the question is whether
they *should* be.

---

## Current state

| Format  | Clones | Dup lines | Dup % |
|---------|--------|-----------|-------|
| Go      | 38     | 448       | 4.60% |
| markup  | 5      | 57        | 2.47% |
| Total   | 43     | 505       | **3.77%** |

All 38 Go clones are in `*_test.go` files. All 5 markup clones are in `templates/`.

---

## Go test-file violations — what is being duplicated

### Pattern A — Service instantiation (compliance_test.go)

`TestAnonymize_*` tests repeat this 6-field struct init with minor variation per test:

```go
// Each test instantiates:
svc := services.ComplianceService{
    Accounts:     accounts,      // stub varies per test
    Membership:   membership,
    Roles:        roles,         // stub varies per test
    InviteTokens: invites,
    ResetTokens:  resets,
    Sessions:     sessions,
}
```

4 clones, ~14 lines each. The duplication is field-by-field struct initialization where the
only variation is which stub has an error injected.

**Refactoring option:** `newComplianceService(accounts, roles)` factory using sensible stubs
for the other four fields. Small and mechanical.

**Counter-argument:** Each test is self-contained; a factory hides what is being injected,
making tests harder to debug when they fail.

---

### Pattern B — HTTP handler test boilerplate (middleware_test.go)

9 clones of this pattern (11–14 lines each):

```go
app := newTestApp(func(a *buffalo.App) {
    a.Use(injectContextValue("current_account", &services.AccountDTO{ID: 42}))
    a.Use(RequireAdmin(svc))
    a.GET("/admin", sentinel)
})

res := httptest.NewRecorder()
app.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/admin", nil))

assert.Equal(t, http.StatusForbidden, res.Code)
```

Every test creates `newTestApp`, wires middleware, makes a request, and asserts the status.
The only variation is the stub configuration and the assert value.

**Refactoring option:** Table-driven tests — each row carries `(svc, expectedStatus)`. The
setup/request/assert loop runs once. Reduces 9 similar test functions to 1 loop.

**Counter-argument:** Each test has a name (`TestRequireAdmin_NotAdmin_Returns403`) that is
lost in table-driven tests. Failure messages become `--- FAIL: TestRequireAdmin/row-2` which
is less searchable.

---

### Pattern C — Handler test setup (events_test.go — 12 clones, musicians_test.go — 5 clones)

Each test function builds a handler, wires it into a test app, and executes an HTTP request:

```go
// events_test.go repeated ~12 times:
h := EventsHandler{
    Events:      &stubEventManager{...},
    Instruments: stubInstrumentRepoE{},
    Membership:  &stubMusicianProfileMgrE{},
}
app := newEventsTestApp(h, defaultAccount(), false, func(a *buffalo.App) {
    a.GET("/evenements/{id}", h.Show)
})
res := httptest.NewRecorder()
req := httptest.NewRequest(http.MethodGet, "/evenements/1", nil)
app.ServeHTTP(res, req)
```

**Refactoring option A — table-driven:** Same as Pattern B. One test function, each row
carries the stub config and expected outcome.

**Refactoring option B — test helper:** `runEventsRequest(h, route, method, path) *httptest.ResponseRecorder`
that hides the app/recorder/request wiring. Keeps per-case function names.

**Counter-argument:** The handler stub configuration (`stubEventManager{...}`) IS the test
setup — it documents which service call returns what. Hiding it in a helper obscures the
cause-effect chain that makes the test readable.

---

### Pattern D — Model test setup (models/season_test.go — 1 clone)

10-line database setup repeated in two adjacent tests. Classical table-driven candidate.

---

### Summary across test files

| File | Clones | Pattern | Refactoring effort |
|------|--------|---------|-------------------|
| `services/compliance_test.go` | 4 | Service factory | Low |
| `actions/middleware_test.go` | 9 | Table-driven tests | Medium |
| `actions/events_test.go` | 12 | Table-driven or request helper | Medium–High |
| `actions/musicians_test.go` | 5 | Same as events | Medium |
| `actions/profile_test.go` | 2 | Request helper | Low |
| `actions/retention_test.go` | 1 | Request helper | Low |
| `actions/home_test.go` | 1 | Request helper | Low |
| `actions/fee_payments_test.go` | 1 | Request helper | Low |
| `models/season_test.go` | 1 | Table-driven | Low |

---

## Template violations — what is being duplicated

5 clones in `templates/`:

| Clone | Lines | Description |
|-------|-------|-------------|
| `invite_invalid.plush.html` vs `reset_invalid.plush.html` | 11 | Full page structure — only the body text differs |
| `invite.plush.html` vs `reset.plush.html` | 13 | Error alert div + `<form>` opening tag |
| `invite.plush.html` vs `reset.plush.html` | 12 | Password input field (hint text, attributes) |
| `invite.plush.html` vs `reset.plush.html` | 12 | Password confirm field |
| `admin/musicians/edit.plush.html` vs `.../new.plush.html` | 10 | Error alert div + `<form>` opening tag |

**Refactoring option:** Plush partial templates. E.g., `tokens/_invalid_link.plush.html`
with a `message` parameter; `tokens/_password_field.plush.html` for the password input.
Plush supports `<%= partial("path/to/_partial.html", {key: value}) %>`.

**Counter-argument:** Creating partials fragments the template structure — a reader must
open 2–3 files to understand one page. The duplication here is short boilerplate (form
structure, error alert) that is stable and context-specific. Templates are not compiled; a
bug in an "extracted" partial silently affects multiple pages.

---

## Options going forward

### Option 1 — Fix test duplicates as production code

Treat `*_test.go` as first-class code. Refactor each category above:
- Compliance tests: factory function (trivial)
- Middleware tests: table-driven (moderate)
- Events/musicians tests: table-driven or request helper (significant effort)

**Pro:** Consistent DRY standard; jscpd Go % drops to near 0.
**Con:** Table-driven tests lose per-case function names and stack traces. Request helper
abstractions require maintenance. Moderate effort.

---

### Option 2 — Configure jscpd to exclude test files

Add `--ignore "**/*_test.go"` to the lefthook command and wherever jscpd is run.
Keeps the hook enforcing DRY on production code only.

```yaml
# lefthook.yml
jscpd:
  run: jscpd webapp --min-lines 10 --gitignore --threshold 4 --ignore "**/*_test.go"
```

Go test files go from 4.60% to 0% in the tool's view. Production Go would be ~0.5%.
**Pro:** Simple one-line change; focus duplication enforcement where it matters most.
**Con:** Test duplication is no longer caught at all, including genuine accidental duplication.

---

### Option 3 — Configure jscpd to exclude templates

Add `--ignore "**/templates/**"`. Markup violations drop to 0%.
Can be combined with Option 2.

---

### Option 4 — Separate thresholds per format

jscpd's `--threshold` is global. There is no per-format threshold in the CLI.
Would require a `.jscpd.json` config file with per-format reporter config.
Not currently supported natively; not recommended.

---

### Option 5 — Keep as-is (status quo)

jscpd exits 0 today (3.77% < 4%). The 43 remaining clones are in known locations.
No action required unless the threshold is breached.

**Pro:** Zero effort; the hook already enforces the 4% cap.
**Con:** New test duplication can push the total back above 4%, triggering false-positive
hook failures that have nothing to do with production code quality.

---

## Recommendation (not yet decided)

If the project's view is that **test code is first-class** (as stated), **Option 1** is the
honest path — but it's real refactoring work, best scoped as a follow-up task.

If the pragmatic view is that **production code is what matters for duplication enforcement**,
**Option 2** (exclude `*_test.go`) is low risk and keeps the hook meaningful without
generating noise.

Options 2 and 3 can be combined (exclude both test files and templates) with minimal
configuration change.
