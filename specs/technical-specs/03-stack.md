# OHM Website — Stack

> **ADRs:** [003 — OVH Deployment Target](../technical-adrs/003-ovh-deployment-target.md),
> [004 — Framework and Language](../technical-adrs/004-framework-and-language.md)

---

## Runtime Components

| Component | Choice | Notes |
|-----------|--------|-------|
| Language | Go (latest stable) | — |
| Web framework | Buffalo (latest stable) | Full-stack; see ADR 004 |
| ORM | Pop (bundled with Buffalo) | Built on sqlx |
| Migrations | Pop / Fizz DSL | Files in `migrations/`; raw SQL permitted |
| DB driver | pgx (via Pop) | Full `TIMESTAMPTZ` support |
| Template engine | Plush | Buffalo default; extends Go `html/template` |
| Frontend interactivity | Alpine.js (vendored) | Reactive directives on server-rendered HTML; see ADR 008 |
| Session store | gorilla/sessions + pgstore | PostgreSQL-backed; see ADR 001 |
| CSRF | gorilla/csrf | Integrated in Buffalo middleware stack |
| Password hashing | `golang.org/x/crypto/argon2` | First-party; no external dependency |
| Task runner | grift (bundled with Buffalo) | Used for bootstrap and seed tasks |
| Dev tool manager | Mise (`mise.toml` at repo root) | Pins Go version; defines `mise run` tasks |
| Pre-commit hooks | Lefthook (`lefthook.yml` at repo root) | Runs format/vet/lint on commit |
| Test containers | testcontainers-go | Manages PostgreSQL container lifecycle in tests |

---

## Docker

Multi-stage build keeps the final image minimal:

1. **Build stage** — `golang:<version>-alpine`: installs dependencies, runs
   `buffalo build --clean-assets`, produces a single self-contained binary at `./bin/ohm`.
2. **Runtime stage** — `alpine:latest`: copies only the binary. Final image: ~25–30 MB.

The same `docker-compose.yml` drives both local development and production, with environment
differences handled via `.env` files (`.env.development`, `.env.production`).

Buffalo's `app.go` is the single entrypoint; no separate process manager is required inside
the container.

---

## Reverse Proxy

**Caddy** is the recommended reverse proxy for production:
- Automatic HTTPS via Let's Encrypt (zero-config TLS)
- Simple `Caddyfile` format
- Handles TLS termination and forwards to the Buffalo container on `PORT` (default 3000)

nginx is an acceptable alternative if there is existing familiarity.

Static assets in `public/` are served by Buffalo in development. In production, the reverse
proxy can serve them directly for marginally better performance, though this is optional at
OHM's traffic level.

---

## Local Development

```
git clone <repo>
mise install                                   # install Go and tools at pinned versions
cp .env.development.example .env.development   # fill SESSION_SECRET, DATABASE_URL
mise run dev                                   # starts postgres + app with live reload
mise run migrate                               # apply migrations
mise run seed-admin                            # create first admin account
mise run seed-dev                              # (optional) populate with dummy data
```

`buffalo dev` (invoked by `mise run dev`) provides live reload: source changes trigger an
automatic recompile and restart. The Compose file mounts the source directory so `buffalo dev`
can detect changes inside the container.

## Mise Tasks

All developer-facing operations are exposed as `mise run <task>` commands (defined in
`mise.toml`):

| Task | Purpose |
|------|---------|
| `dev` | Start local environment (Docker Compose + live reload) |
| `migrate` | Apply pending DB migrations |
| `seed-admin` | Create first admin account (bootstrap; refuses if one already exists) |
| `recover-admin` | Emergency: force-reset an active account's password directly in the DB (see [Configuration spec](02-configuration.md)) |
| `seed-dev` | Populate dev DB with dummy data (local only) |
| `test` | Run test suite against real PostgreSQL |
| `lint` | Run full linter suite (`golangci-lint`) |
| `deploy` | Deploy to production VPS via SSH (see [CI/CD spec](06-ci-cd.md)) |

---

## Build and Deploy

```
# Build
buffalo build --clean-assets    # produces ./bin/ohm

# Containerise
docker build -t ohm:latest .

# Deploy (from operator workstation)
mise run deploy                 # SSH to VPS: pull, up -d, migrate
```

The `deploy` task handles: `docker compose pull && docker compose up -d && migrate`. See
[CI/CD spec](06-ci-cd.md) for details and the future automated CD option.

Rollback: re-tag the previous image and re-run `mise run deploy`, then roll back migrations
with `mise run migrate --rollback` if the schema changed.

---

## Backup

A daily cron job on the VPS runs `pg_dump`, compresses the output, and uploads it to OVH
Object Storage (30-day rolling retention). The script lives at `scripts/backup.sh` in the
repository. The cron entry and the required credentials (bucket name, S3 endpoint, access
keys) must be configured on the VPS as part of initial production setup.

See [ADR 006](../technical-adrs/006-backup-strategy.md) for the full procedure, restore
instructions, and trade-offs.

---

## Testing

| Layer | Tool |
|-------|------|
| Unit / integration | Go built-in `testing` + `testify/assert` |
| HTTP handlers | `net/http/httptest` + Buffalo test helpers |
| DB (integration) | Real PostgreSQL via **testcontainers-go** |
| Race detection | `go test -race` |

Tests run against a real PostgreSQL instance (not mocks) to ensure migration correctness and
query behaviour match production. testcontainers-go starts and tears down a PostgreSQL
container automatically from within the test code — no manual Docker setup required. A
`TestMain` function spins up the container once per test binary, applies migrations, then
runs all tests in that binary against it.

Running tests locally: `go test -race ./...` (or `mise run test`). Docker must be running.

See [CI/CD spec](06-ci-cd.md) for CI job definitions.
