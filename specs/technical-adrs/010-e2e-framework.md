# ADR 010 — E2e Framework: Playwright (TypeScript)

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-30 |

## Context

Phases 3 through 4.6 and adj-1 produced a set of human-verified acceptance criteria that have
no machine-verified equivalent. An e2e test suite is needed to automate these checks, detect
regressions before they reach production, and generate coverage data that complements the
Go unit/integration suite.

The framework must:
- Drive a real browser against a running instance of the application.
- Work reliably with server-rendered pages where navigation and form submissions trigger full
  page reloads (not SPA-style DOM patches).
- Integrate into the existing GitHub Actions CI pipeline.
- Support aggregated coverage reporting alongside the Go unit test suite.

---

## Alternatives Considered

### Playwright (TypeScript)

Chromium/Firefox/WebKit automation via the Playwright library. Industry standard as of 2026.

**Strengths:**
- Auto-wait semantics: assertions retry until the condition is met or a timeout fires.
  Eliminates the `sleep()` anti-pattern that causes flakiness on server-rendered page transitions.
- Trace viewer: records a timeline of network, DOM, and screenshot events per test. Failures
  in CI produce an uploadable artifact that can be replayed locally.
- First-class CI support: `playwright install --with-deps` installs browsers and system
  dependencies in a single command.
- Parallel test execution with isolated browser contexts.
- TypeScript support out of the box.
- Node.js is already partially present in the toolchain (jscpd runs via `npx`).

**Weaknesses:**
- Adds a Node.js/npm layer to the Go project.

### Rod

Browser automation library written in Go (Chrome DevTools Protocol).

**Strengths:**
- Stays entirely in the Go ecosystem; no Node.js dependency.

**Weaknesses:**
- Significantly smaller community and ecosystem than Playwright.
- Fewer convenience abstractions for common patterns (file upload, network interception,
  parallel contexts).
- Trace and debugging tooling is immature by comparison.

**Rejected because:** the auto-wait and trace viewer capabilities of Playwright are
particularly valuable for this application's server-rendered flow. The Node.js cost is
already partially paid (jscpd), and the productivity gap is material.

### Cypress

Browser automation with a focus on developer experience.

**Strengths:** familiar to many frontend engineers.

**Weaknesses:**
- JavaScript-only (no native TypeScript compilation pipeline).
- Slower test execution than Playwright.
- Poor support for multi-origin scenarios and non-SPA server-rendered flows.

**Rejected because:** Rod and Playwright are both better fits for this application's
server-rendered architecture.

---

## Decision

**Playwright (TypeScript)**, version pinned in `e2e/package.json` during T2.

---

## Coverage Aggregation Strategy

Go 1.20+ supports instrumenting the application binary for coverage collection at runtime.
The CI pipeline uses this to merge unit and e2e coverage into a single percentage.

**Steps:**

1. Run unit/integration tests with `-coverprofile`:
   ```
   go test -coverprofile=coverage/unit.out ./...
   ```

2. Build a coverage-instrumented binary:
   ```
   go build -cover -o bin/app-covered ./cmd/app
   ```
   (Entrypoint path confirmed during T2.)

3. Start the instrumented binary, directing coverage output to a directory:
   ```
   GOCOVERDIR=coverage/e2e ./bin/app-covered
   ```

4. Run the Playwright suite against the running binary.

5. Stop the binary with `SIGINT` so it flushes coverage data before exiting.

6. Convert the binary coverage format to text profile format:
   ```
   go tool covdata textfmt -i coverage/e2e -o coverage/e2e.out
   ```

7. Merge unit and e2e profiles using `gocovmerge` (`github.com/wadey/gocovmerge`):
   ```
   gocovmerge coverage/unit.out coverage/e2e.out > coverage/merged.out
   ```

8. Compute the aggregated total:
   ```
   go tool cover -func coverage/merged.out | grep "^total"
   ```

9. Post the result as a PR comment via `actions/github-script`. Existing comments from the
   bot are updated rather than duplicated (find-or-create-comment pattern).

---

## Toolchain Additions

| Addition | Purpose | Notes |
|----------|---------|-------|
| Node.js (LTS) | Playwright runtime | Already present via `npx` (jscpd); CI installs explicitly |
| `@playwright/test` | Test runner + browser automation | Pinned in `e2e/package.json` |
| `gocovmerge` | Merge Go coverage profiles | Added as Go tool dependency |

---

## Consequences

- `e2e/` directory added at repo root (parallel to `webapp/`), containing Playwright
  configuration and test files.
- `e2e/package.json` and `e2e/package-lock.json` checked in; `e2e/node_modules/` and
  Playwright cache directories added to `.gitignore`.
- A `mise run e2e` task is added (defined during T2) that starts the application, waits for
  readiness, runs the Playwright suite, and tears down.
- CI gains an `e2e` job (defined during T4) that runs in parallel with `lint` and `test`
  and is required to pass.
- Coverage data directories (`coverage/unit/`, `coverage/e2e/`, `coverage/merged/`) and the
  instrumented binary (`bin/app-covered`) are generated artifacts added to `.gitignore`.
- `webapp/TESTING.md` is updated during T2 to document local e2e test execution.
- `specs/technical-specs/03-stack.md` is updated to record the framework and `mise run e2e` task.
