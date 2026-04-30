# Adj-2 — E2e Test Framework + CI Coverage

**Status:** In progress

**Goal:**
1. Add a browser-based e2e test suite that covers the human-verified acceptance criteria from
   phases 3 through 4.6 (and adj-1).
2. Integrate e2e tests into CI alongside unit/integration tests.
3. Aggregate unit + e2e coverage into a single percentage and post it as a PR comment.

---

## Prerequisites

**D2 — E2e framework must be decided before implementation begins.**
See decision section below.

Adj-1 should be complete first so all user flows exist before e2e scenarios are written.
Adj-3 (jscpd) should be complete so the pre-commit hook doesn't block e2e scaffolding commits.

---

## Open Decision: D2 — Browser Automation Framework

### Options

| Option | Language | Pros | Cons |
|--------|----------|------|------|
| **Playwright** | Node.js (TS/JS) | Industry standard; exceptional debugging (trace viewer, video); first-class CI support; wide community; parallel test execution; auto-wait | Adds Node.js to toolchain (already partially present via jscpd/npx) |
| Rod | Go | Stays in Go ecosystem; no Node dependency | Smaller community; less mature API; fewer convenience abstractions for server-rendered apps |
| Cypress | Node.js (JS) | Familiar to many | JavaScript-only (no TS native); slower; less suitable for multi-origin or server-rendered flows |

### Recommendation

**Playwright (TypeScript).** Node.js is already present (jscpd uses npx). Playwright's
trace viewer and auto-wait semantics are particularly valuable for server-rendered apps
where page transitions and form submissions trigger full reloads. Warrants **ADR-010**
(first browser automation dependency; explains Node toolchain addition).

### Coverage aggregation approach

Go 1.20+ supports building the application binary with coverage instrumentation:
```
go build -cover -o bin/app-covered ./cmd/app   # (or ./main.go, check entrypoint)
```
When this binary runs and exits cleanly, it writes coverage data to `$GOCOVERDIR`.

Aggregate flow in CI:
1. Run unit/integration tests: `go test -coverprofile=coverage/unit.out ./...`
2. Build covered binary.
3. Start covered binary with `GOCOVERDIR=coverage/e2e/`.
4. Run Playwright tests against it.
5. Stop binary (sends SIGINT so it flushes coverage).
6. Convert e2e coverage: `go tool covdata textfmt -i coverage/e2e -o coverage/e2e.out`
7. Merge profiles: use `gocovmerge` (`github.com/wadey/gocovmerge`) to combine
   `coverage/unit.out` and `coverage/e2e.out` into `coverage/merged.out`.
8. Compute total: `go tool cover -func coverage/merged.out | grep "^total"`.
9. Post as PR comment via `actions/github-script`.

This is the intended approach; agent may surface implementation issues during T2 if the
entrypoint or binary structure differs.

### Decision

**Playwright (TypeScript)** — Accepted (ADR 010, 2026-04-30)

---

## Tasks

### T1 — ADR-010: e2e framework ✅

Write `specs/technical-adrs/010-e2e-framework.md` documenting the chosen framework,
rationale, coverage aggregation strategy, and toolchain additions.

**DoD:** ADR written; `specs/technical-specs/03-stack.md` updated to record the framework
and version.

### T2 — Framework setup + local runner

- Add Playwright (or chosen framework) to the project. For Playwright: `npm init playwright@latest` (or equivalent), configured for TypeScript, tests under `e2e/`.
- Add a `mise` task `e2e` that: starts the local server (`buffalo dev` equivalent or a test binary), waits for it to be ready, runs the e2e suite, tears down.
- Document setup steps in `webapp/TESTING.md` (existing file).
- `.gitignore`: add `e2e/node_modules`, Playwright cache dirs, coverage data dirs.

**DoD:** `mise run e2e` runs locally and tests execute (even if they immediately fail —
the runner works); `TESTING.md` updated.

### T3 — E2e scenarios

Write e2e tests covering all automatable human-verified ACs from phases 3–4.6 and adj-1.
Tests must be independent (no shared state between tests; each test sets up what it needs).

See [Scenario Inventory](#scenario-inventory) below for the full list.

**DoD:** all scenarios pass locally against a seeded dev database; scenarios are grouped
by phase in the `e2e/` directory structure.

### T4 — CI integration + coverage comment

Extends `.github/workflows/ci.yml`:

- Add an `e2e` job that:
  1. Installs Node + Playwright dependencies.
  2. Runs Go unit tests with `-coverprofile`.
  3. Builds the covered binary (`go build -cover`).
  4. Starts the covered binary against a test database (PostgreSQL service container,
     same pattern as the existing `test` job).
  5. Runs Playwright tests.
  6. Stops the binary.
  7. Merges coverage profiles (unit + e2e) using `gocovmerge`.
  8. Computes the aggregated total.
  9. Posts the result as a PR comment using `actions/github-script`.
- The `e2e` job runs in parallel with `lint` and `test`; it is required to pass.
- PR comments must be updated (not duplicated) on re-runs of the same PR: use
  `find-or-create-comment` pattern (check for existing bot comment, edit if found).

**DoD:** CI job passes; PR receives a coverage comment showing the aggregated percentage
(e.g. `Coverage: 74.3% (unit + e2e)`); comment is updated on re-run, not duplicated.

---

## Scenario Inventory

Scenarios are grouped by phase. Each maps to one or more human-verified ACs.
Scenarios marked **skip** require production infrastructure or external systems and cannot
be automated in CI.

### Phase 3 — Production Infra

| Scenario | Maps to | Automatable? |
|----------|---------|-------------|
| Session persists across page navigations | AC-H1 (adapted: local, not prod) | Yes — covered by auth scenarios in 4.1 |
| No secrets in Docker image | AC-H2 | Skip (infra check) |
| Backup retention configured | AC-H3 | Skip (external system) |

### Phase 4.1 — Public Pages + Navigation

| Scenario | Maps to | Notes |
|----------|---------|-------|
| Unauthenticated nav shows only login link | AC-H1 | Check nav links visible/hidden |
| Authenticated nav shows full menu | AC-H1 | Login, verify nav items |
| Privacy notice page is not blank | AC-H2 | Visit `/politique-de-confidentialite`, assert non-trivial content |

### Phase 4.2 — Account Lifecycle

| Scenario | Maps to | Notes |
|----------|---------|-------|
| Invite form pre-populates account email | AC-H1 | Admin creates invite; follow invite link; assert email shown |
| Password reset form is reachable | AC-H2 | Visit reset URL; assert form renders |
| Password reset completes successfully | AC-H2 | Fill and submit; assert success state |

### Phase 4.3 — Musician Management

| Scenario | Maps to | Notes |
|----------|---------|-------|
| Invite URL leads to complete-invite form | AC-H1 | Follow invite link, assert form present |
| Musician detail page shows all key sections | AC-H2 | Admin views musician; check name, instrument, status, etc. |
| Anonymization removes identifying information | AC-H3 | Admin anonymizes; verify name/email replaced on musician page |

### Phase 4.4 — Season Management

| Scenario | Maps to | Notes |
|----------|---------|-------|
| Current season is visually distinguished | AC-H1 | Check CSS class or badge on current season row |
| Designating a season updates immediately | AC-H2 | Click designate; verify UI reflects change without full reload |

### Phase 4.5 — Fee Payments

| Scenario | Maps to | Notes |
|----------|---------|-------|
| Fee payment section renders on musician detail | AC-H1 | Admin views musician; fee section present |
| Earlier payment updates first inscription date | AC-H2 | Add payment with earlier date; verify inscription date updated |

### Phase 4.6 — Events + RSVP

| Scenario | Maps to | Notes |
|----------|---------|-------|
| Dashboard shows upcoming events with own RSVP state | AC-H1 | Authenticated user visits `/tableau-de-bord`; verify events + states |
| Concert RSVP form includes instrument dropdown | AC-H2 | Visit concert detail; RSVP form has instrument select |
| Rehearsal RSVP form has no instrument dropdown | AC-H2 | Visit rehearsal detail; no instrument select |
| Custom field lifecycle: add, edit, delete (no responses) | AC-H3 | Admin: add field → edit label → delete; all succeed |
| Edit blocked when responses exist | AC-H3 | Musician RSVPs yes + fills field; admin edit blocked |

### Adj-1 — Dashboard + Description

| Scenario | Maps to | Notes |
|----------|---------|-------|
| Dashboard renders cards, not table | AC-H1 | Assert card elements; assert no table |
| Markdown description renders as HTML | AC-H2 | Create event with `**bold**`; verify `<strong>` in page |
| Admin can set and update description | AC-H3 | Create with description; update; verify changes |

---

## Test Data Strategy

E2e tests must not depend on `db/dummy-data/` (which may or may not be loaded).
Each test is responsible for creating its own state via the admin UI or API.

A shared Playwright fixture should:
1. Seed a fresh admin account before the test suite (or use a known test credential
   injected via environment).
2. Provide helpers: `loginAsAdmin()`, `createMusician()`, `createEvent()`, etc.

Tests must clean up after themselves or rely on database isolation (separate test DB per run).

---

## Architecture Notes

- E2e tests live under `e2e/` at the repo root, parallel to `webapp/`.
- Coverage data directories (`coverage/unit/`, `coverage/e2e/`, `coverage/merged/`) are
  generated artifacts; add to `.gitignore`.
- The covered binary (`bin/app-covered`) is a build artifact; add to `.gitignore`.
- `TESTING.md` must document how to run e2e tests locally, including the covered-binary
  approach if used.

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — CI e2e job passes**
All e2e scenarios in `e2e/` pass in CI against a seeded test database.

**AC-M2 — Coverage comment posted on PR**
Opening a PR triggers a comment from the CI bot with a line of the form:
`Coverage: XX.X% (unit + e2e)`.

**AC-M3 — Coverage comment updated, not duplicated**
A second CI run on the same PR updates the existing coverage comment rather than
posting a new one.

### Human-verified

**AC-H1 — All scenarios pass locally**
`mise run e2e` runs cleanly against a local seeded database. No flaky timeouts or
ordering dependencies between tests.

**AC-H2 — Trace viewer works**
On Playwright failure (simulate by breaking one test), the CI run uploads a trace
artifact that can be opened in the Playwright trace viewer to inspect the failure.
