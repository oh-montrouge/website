# ADR 004 — Framework and Language

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-11 |

## Context

ADR 003 (Accepted) established VPS + Docker Compose with PostgreSQL as the deployment target.
The language and web framework can now be chosen freely. This decision locks in the development
environment, the ORM and migration tooling, the auth/session implementation, the template
engine, and the long-term maintenance surface.

The application is a server-rendered web application with no public API in V1. It needs:
session-based auth (ADR 001), UTC-aware datetimes (ADR 002), schema migrations, CSRF
protection, password hashing (Argon2id), a management CLI for bootstrap, and a PostgreSQL
driver with support for partial indexes. Traffic is low; performance is not a primary criterion.

---

## Constraints from Prior ADRs

| Requirement | Source |
|-------------|--------|
| Server-side sessions, immediate invalidation | ADR 001 |
| Timezone-aware datetimes (UTC storage, Europe/Paris display) | ADR 002 |
| Argon2id password hashing | Technical spec 01 |
| Partial unique index on email | Technical spec 00 |
| Schema migration tooling | Vision |
| Bootstrap CLI | Vision |

---

## Eliminated Without Full Analysis

### Next.js / Nuxt.js (SSR-capable SPA frameworks)

Raised as an alternative on the basis that TypeScript/Node is the current industry standard and
that these frameworks provide batteries-included SSR.

They are capable but not batteries-included for this use case. Next.js and Nuxt.js are SPA
frameworks with SSR capabilities, not traditional server-rendered frameworks. The mental model
is: write React/Vue components → server renders initial HTML → browser hydrates and takes over
with client-side JavaScript. For OHM's use case (admin-heavy CRUD, simple forms and lists, no
real-time requirements), the React/Vue component model is overhead with no UX return.

Assembly cost against V1 requirements mirrors plain Node/TypeScript (option C above):

| Requirement | Next.js / Nuxt.js |
|---|---|
| Sessions | `iron-session` or Auth.js — neither fits the custom invite flow cleanly; significant config |
| ORM + migrations | Prisma or Drizzle — good, but a separate choice |
| CSRF | Explicit setup; no built-in middleware |
| Argon2id | `argon2` npm package — native bindings, adds Docker build complexity |
| Management CLI | No equivalent to `buffalo task`; custom scripts required |

Familiarity advantage is real: more developers know React/Vue than Buffalo. That is a
maintainability argument, not a technical one — see Decision Revisited below.

---

### Single-Page Application (React / Vue / Svelte)

A SPA frontend paired with a JSON API backend was not evaluated. The exclusion originated in
ADR 001, which established server-side sessions and stated *"the client is the server-rendered
application, and there is no public API surface"* — implicitly committing to SSR before this
ADR was written.

The case against a SPA for this use case:

- **Assembly cost.** A SPA requires a separate build pipeline (Vite/webpack), a JSON API layer,
  CORS configuration, and a second tech stack alongside whatever backend is chosen.
- **Auth complexity.** Sharing session cookies between a SPA and a backend requires careful
  `SameSite` handling; using tokens instead would reintroduce the complexity ADR 001 explicitly
  avoided.
- **No UX benefit.** All interactions in V1 are standard form submissions, list views, and
  detail pages. None require reactivity or client-side state. A SPA adds infrastructure cost
  for zero user-visible gain.
- **Learning goal.** The stated motivation for this project includes learning a backend language.
  A SPA frontend would split the cognitive budget without furthering that goal.

If a future milestone requires rich interactivity (e.g. an in-browser sheet music viewer,
a drag-and-drop event scheduler), the architecture supports adding a targeted JS component via
a `<script>` import without migrating the entire application to a SPA.

### WordPress

A CMS designed for content publishing. Its data model (posts, meta, options, users) is
fundamentally incompatible with the OHM domain (seasons, fee payments, RSVPs, instruments).
Implementing the required features means writing a custom PHP plugin — effectively building
a PHP application inside WordPress while fighting its conventions. Additionally: PHP is excluded
by preference, PostgreSQL requires a third-party plugin, and WordPress is the most-exploited
CMS attack surface on the web. Not a contender.

### Streamlit

A Python library for turning data scripts into shareable dashboards. It is not a web framework.
Blocking gaps for this project:

- No traditional routing: the invite flow, password reset, and privacy notice page (publicly
  reachable without auth) have no clean implementation.
- No session-based auth: workarounds exist but are not production-grade.
- UI is constrained to Streamlit's widget system: the invite form, RSVP list, and admin views
  require HTML layouts that Streamlit cannot express.
- Reruns the entire script on every interaction: wrong execution model for a multi-user CRUD app.

**Note for Later milestone:** Streamlit is a legitimate candidate for the statistics dashboard
(age pyramid, analytics) as a separate internal tool alongside the main application. Not the
main app.

### PHP / Laravel

Excluded by team preference. Laravel is a technically sound framework (batteries-included,
good PostgreSQL support, Argon2id built-in); the exclusion is not a technical judgement.

### Ruby on Rails

Full-featured, similar philosophy to Django. Smaller community than in its peak years; less
common in the French development ecosystem; no technical advantage over Django or Go for this
use case. Eliminated: no differentiating benefit.

---

## Alternatives

### A — Python + Django

Django is a "batteries included" web framework. Its default feature set maps almost entirely
onto V1 requirements without third-party assembly.

| Requirement | Django built-in |
|-------------|----------------|
| Sessions | `django.contrib.sessions` (DB-backed) |
| Auth | `django.contrib.auth` |
| Argon2id | `django[argon2]` — one line in settings |
| UTC datetimes | `USE_TZ = True` — all datetimes are timezone-aware UTC; template filters convert to Europe/Paris for display |
| CSRF | `django.middleware.csrf.CsrfViewMiddleware` |
| ORM + migrations | `django.db.migrations` |
| Partial unique index | `Meta.indexes = [Index(..., condition=~Q(status='anonymized'))]` |
| Management CLI | `manage.py` custom commands |
| Template engine | Django templates (sandboxed) or Jinja2 |
| Admin panel | `django.contrib.admin` (useful for debugging; disable or restrict in production) |
| Testing | `pytest-django` or `django.test` |
| PostgreSQL driver | `psycopg` (v3, async-capable) |

**Pros:**
- Every V1 requirement satisfied by a built-in or one-liner; no framework assembly
- `USE_TZ = True` makes ADR 002 compliance the default, not a discipline: forgetting to handle
  UTC is a configuration error, not a runtime mistake
- 3-year LTS release cycle; predictable upgrade path; extensive documentation
- Python ecosystem is strongest for Later milestones (statistics, potential sheets search)

**Cons:**
- More opinionated than lightweight alternatives; some conventions require learning
- Docker image ~150 MB with a slim base (acceptable; not a constraint at this scale)
- Built-in admin must be disabled or access-restricted in production

**Reversibility:** Medium. Django's ORM and migration format are framework-specific; moving
away requires rewriting the application layer, though the data and migrations are portable.

---

### B — Python + Flask (Lightweight)

Flask is a minimal synchronous Python framework; Jinja2 is its template engine. Everything
else (ORM, migrations, sessions, auth, CSRF) must be assembled from separate libraries:
SQLAlchemy + Alembic + Flask-Session + Flask-WTF + passlib[argon2].

*FastAPI is excluded: it is designed for JSON APIs and OpenAPI, not server-rendered HTML. No
benefit over Flask for this use case.*

**Pros:**
- More explicit control than Django; no hidden conventions
- SQLAlchemy ORM is more portable than Django's (not framework-locked)
- Jinja2 is the same template engine Django optionally uses

**Cons:**
- Assembly tax: each V1 requirement is a separate library selection and wiring decision
- More configuration surface area; more code owned; no built-in admin
- `USE_TZ`-equivalent discipline must be maintained manually (ADR 002 compliance is a
  convention, not a default)

**Reversibility:** Medium.

---

### C — Node.js + TypeScript

TypeScript on Node.js with a router framework (Express, Fastify, or Hono) and Prisma or
Drizzle for ORM + migrations.

No batteries-included server-rendered framework exists in this ecosystem. Typical assembly:
router + Prisma/Drizzle + express-session/comparable + helmet + a CSRF library + a template
engine (Nunjucks, Eta) + the `argon2` npm package (native bindings, adds a compilation step).

**Pros:**
- TypeScript provides compile-time type safety across the full stack
- Prisma is an ergonomic ORM with good migration tooling
- If V2 adds a rich frontend, TypeScript shared between backend and frontend eliminates context
  switching
- Large ecosystem

**Cons:**
- No idiomatic server-rendered framework: the stack is bespoke and must be maintained
- ADR 002 compliance requires explicit convention: JavaScript's `Date` is UTC internally but
  library handling varies; the ORM layer must be configured carefully
- NPM ecosystem churn: dependencies change faster than Python equivalents; more maintenance
- `argon2` native addon requires compilation in the Docker build (slower CI, more complex
  multi-stage image)

**Reversibility:** Medium.

---

### D — Go

Go has two distinct tiers of web tooling:

**D1 — Buffalo (batteries-included)**

Buffalo is a full-stack Go framework explicitly inspired by Rails and Django. It ships with:
Pop (ORM built on sqlx), Fizz (migration DSL), Plush (template engine), session support,
and code generators. Closest to Django in philosophy among Go options.

| Requirement | Buffalo / Go ecosystem |
|-------------|----------------------|
| Sessions | Buffalo sessions (cookie or DB-backed) |
| Auth | `buffalo-auth` plugin (generates scaffolding) |
| Argon2id | `golang.org/x/crypto/argon2` (first-party, no third-party dependency) |
| UTC datetimes | `time.Time` carries timezone; PostgreSQL `pgx` driver returns UTC — **requires explicit attention** but is straightforward |
| CSRF | `gorilla/csrf` |
| ORM + migrations | Pop + Fizz |
| Partial unique index | Raw SQL in Fizz migration |
| Management CLI | `buffalo` CLI + custom tasks |
| Template engine | Plush (Go template extension) |
| PostgreSQL driver | `pgx` — full `TIMESTAMPTZ` support |
| Testing | Go's built-in `testing` package + testify |

Buffalo's main weakness: it is less widely used than Django or Rails and has had periods of
slow maintenance. The community is smaller; onboarding documentation is thinner.

**D2 — Gin / Echo / Chi (router only)**

These are fast HTTP routers with middleware ecosystems. They require assembling: GORM or sqlx
for the ORM, goose or golang-migrate for migrations, gorilla/sessions or similar, gorilla/csrf,
and a template engine. Similar assembly cost to Flask.

**Go language characteristics:**
- Compiled binary: Docker final image is ~20–30 MB with a multi-stage build (smallest of all
  alternatives)
- Fast compilation; excellent built-in tooling (formatter, race detector, test runner)
- Strong typing; goroutines and channels are new concepts but the language is designed to be
  learnable quickly
- Good choice if the goal is to ship the project and learn a modern language: Go is accessible
  and production-ready without a steep curve

**Reversibility:** Medium. The data layer (SQL migrations) is portable; the application code is Go-specific.

---

### E — Rust

Rust has a mature set of async web frameworks. A realistic production stack:
**Axum + SQLx + Tera + tower-sessions**.

| Requirement | Rust ecosystem |
|-------------|---------------|
| Sessions | `tower-sessions` + `tower-sessions-sqlx-store` (SQLx-backed) |
| Auth | Custom (no auth scaffold; implement login handler) |
| Argon2id | `argon2` crate — OWASP-recommended, excellent |
| UTC datetimes | `chrono::DateTime<Utc>` — **timezone is enforced by the type system**: you cannot accidentally store a naive datetime; ADR 002 compliance is structural |
| CSRF | `axum-csrf` |
| ORM + migrations | SQLx (async, compile-time query validation against live DB) or Diesel (sync, compile-time schema) |
| Partial unique index | Raw SQL in `.sql` migration files (SQLx migrations) |
| Management CLI | Custom binary or `clap`-based CLI |
| Template engine | Tera (Jinja2-style, runtime) or Askama (compile-time checked) |
| PostgreSQL driver | `sqlx` with `tokio-postgres` — full `TIMESTAMPTZ` support |
| Testing | Built-in `#[test]` + `tokio::test` for async; `axum::test` helpers |

**Axum** is the most active framework (Tokio ecosystem). **Actix-web** is an alternative:
very mature, very fast, slightly more complex API. **Rocket** is more ergonomic but
historically lagged on async support.

**Rust language characteristics:**
- Compiled binary: Docker final image ~10–20 MB with a multi-stage build
- `chrono::DateTime<Utc>` makes ADR 002 compliance enforced at compile time — the strongest
  guarantee of any alternative
- The borrow checker and ownership model will slow initial development significantly: the
  compiler rejects code that other languages would silently accept; this is the learning, not
  a bug
- Compilation times are slow for full rebuilds; `cargo watch` and incremental builds
  mitigate this in development
- **Steepest learning curve of all alternatives.** If learning the language is a goal,
  this is the most intellectually rewarding choice. If shipping the project in reasonable
  time is also a goal, the investment is higher than Go

**Reversibility:** Medium. SQL migrations are portable `.sql` files; application code is Rust-specific.

---

## Comparison

| Criterion | A (Django) | B (Flask) | C (Node/TS) | D (Go/Buffalo) | E (Rust/Axum) |
|-----------|:----------:|:---------:|:-----------:|:--------------:|:-------------:|
| Batteries included | ✅ | ❌ | ❌ | ✅ (Buffalo) / ❌ (Gin) | ❌ |
| ADR 001 sessions | Built-in | Assembled | Assembled | Built-in (Buffalo) | Assembled |
| ADR 002 UTC | Automatic (`USE_TZ`) | Manual | Manual | Explicit | Type-enforced |
| Argon2id | 1-line config | Library | Native addon | stdlib | Crate |
| Assembly cost | Low | High | High | Medium (Buffalo) / High (Gin) | High |
| Learning curve | Low–Medium | Low | Medium | Low–Medium | High |
| Docker image | ~150 MB | ~120 MB | ~180 MB | ~25 MB | ~15 MB |
| Statistics (Later) | Python ecosystem ✅ | Python ecosystem ✅ | JS ecosystem | Go ecosystem | Rust ecosystem |
| Type safety | Runtime | Runtime | Compile-time | Compile-time | Compile-time + borrow checker |
| Ecosystem maturity | High | High | High | Medium | Medium (growing fast) |
| "Learning for fun" | ➖ | ➖ | ➖ | ✅ Accessible | ✅✅ Deepest |

---

## Impact on Remaining Technical Specs

Once this ADR is resolved, the following can be completed:

- `00-data-model.md`: replace `[STACK TBD]` with actual DDL conventions and column types
- `01-auth-and-security.md`: replace `[STACK TBD]` with framework-specific session config and
  CSRF middleware
- `02-configuration.md`: replace `[STACK TBD]` with env var handling and CLI command syntax
- New spec: `03-stack.md` covering runtime version, Docker image structure, migration tool,
  static file serving, and reverse proxy config

---

## Decision

**Go + Buffalo.**

Go was chosen over Django to satisfy the goal of learning a modern language while shipping a
real project. Buffalo is the Go framework closest to "batteries included": Pop (ORM + Fizz
migrations), Plush (templates), gorilla/sessions, and gorilla/csrf cover every V1 requirement
without custom assembly.

Key alignment with prior ADRs:

- **ADR 001 (sessions):** Buffalo uses `gorilla/sessions` with a configurable backend. A
  PostgreSQL-backed store (`pgstore`) provides server-side persistence and immediate session
  invalidation on account anonymization (by deleting the session row).
- **ADR 002 (UTC):** Go's `time.Time` carries explicit timezone information; the pgx driver
  returns `TIMESTAMPTZ` columns as `time.Time` in UTC. ADR 002 compliance requires explicit
  attention at the application layer but is straightforward to enforce.
- **Argon2id:** `golang.org/x/crypto/argon2` is a first-party Go package; no external
  dependency.

Flask and Node/TypeScript were not chosen: they offer neither the assembly-free convenience of
Buffalo nor the learning payoff of Rust. Rust (Axum) remains the strongest alternative for a
future project where learning the language itself is the primary goal.

---

## Decision Revisited

**Date:** 2026-04-16. **Outcome: reaffirmed.** Walking skeleton to proceed with Go + Buffalo.

### Challenge raised

Two concerns were raised by a stakeholder:

1. **Familiarity:** Buffalo is not widely known; TypeScript/Node (specifically Next.js or
   Nuxt.js) is the current industry standard for web development. Long-term maintainability
   could suffer if future maintainers are TypeScript-native.

2. **Architectural detection:** Working in an unfamiliar language might make it harder to
   detect architectural issues early.

### Response

**On familiarity:** Next.js and Nuxt.js were evaluated (see Eliminated section above). They
are not more batteries-included than Buffalo for this use case — sessions, ORM, CSRF, and CLI
tooling all require separate assembly. The familiarity argument is real but applies equally to
the entire Go ecosystem, not specifically to Buffalo. It is a valid long-term maintenance
concern, not a technical disqualifier.

**On architectural detection:** Architectural issues are structural, not linguistic. Schema
coupling, middleware ordering, session lifecycle, missing invariants, and GDPR compliance gaps
surface from the design, not from which language implements it. The specs themselves — written
before a single line of code — are the primary guard. Language familiarity affects
implementation speed, not the ability to reason about correctness.

### Resolution

A walking skeleton will be built with Go + Buffalo. This is a deliberate, low-risk trial:
the skeleton exercises the full stack (DB, sessions, auth, CSRF, templates) with minimal
domain logic. If it reveals fundamental friction — Buffalo's conventions fighting the design,
rough edges around Pop/pgstore/gorilla that cannot be smoothed without significant effort —
the decision reopens at that point with evidence. If the skeleton ships cleanly, the decision
is confirmed for V1.
