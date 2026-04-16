# webapp

Go/Buffalo application for the OHM website. See the root `README.md` for project-wide
setup instructions and the full task reference.

## Layout

```
actions/         HTTP handlers (Buffalo controllers)
cmd/app/         Application entrypoint
config/          Buffalo configuration (buffalo-app.toml, middleware)
dto/             View models / DTOs (see Architecture conventions below)
grifts/          CLI tasks (db:seed:admin, db:recover:admin, db:seed:dev)
locales/         i18n strings
models/          Pop models and DB query functions
public/          Static assets
templates/       Plush HTML templates
docker-compose.yml   PostgreSQL 17 for local development
database.yml     Pop database configuration (reads DATABASE_URL)
.env.development.example   Environment variable template
```

Migrations live at `../db/migrations/` (repo root) — decoupled from the framework so
they can be run by any SQL-capable tool.

## Architecture conventions

### Request lifecycle

```
Router (actions/app.go)
  → Middleware (auth check, DB transaction, CSRF…)
    → Handler (actions/*.go)       ← orchestrates: fetch, shape, render
        → Model (models/*.go)      ← DB queries only, no display logic
        → DTO (dto/*.go)           ← optional: shapes data for the template
        → Template (templates/)    ← display only, no logic
```

### When to pass the ORM struct directly to the template

Acceptable when all three hold:
- The struct contains no sensitive fields (no `password_hash`, no `anonymization_token`)
- No display transformation is needed (no date formatting, no computed labels)
- The template maps 1-to-1 with the DB columns

`Instrument` (ID + Name, read-only reference data) is the canonical example.

### When to use a DTO

Introduce a DTO when any of these apply:
- **Sensitive fields must be hidden** — e.g. `Account` (never expose `password_hash`)
- **Display logic is needed** — e.g. `Event.datetime` is stored UTC, displayed in Europe/Paris; the conversion belongs here, not in the template
- **Multiple queries feed one view** — e.g. a musician profile page combining account, fee payments, and RSVPs
- **Anonymization changes the display** — e.g. `FeePayment` shows a name or an anonymization token depending on account state

DTOs live in `dto/`. Name them after the template they serve, not the model:
`dto.InstrumentRow`, `dto.AccountListItem`, `dto.EventDetail`. The handler is
responsible for building them from model data before calling `c.Set()`.

## Running locally

From the repo root, use `mise run dev`. To run Buffalo commands directly from this
directory:

```bash
cd webapp
buffalo dev          # Dev server with live reload
buffalo pop migrate --path ../db/migrations   # Run migrations manually
```

## Environment variables

Copy `.env.development.example` to `.env` and fill in:

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | Postgres connection string |
| `SESSION_SECRET` | Yes | At least 32 random bytes |
| `GO_ENV` | No | Defaults to `development` |
| `SHEET_MUSIC_URL` | No | Google Drive folder URL for sheet music |

## Testing

See [`TESTING.md`](TESTING.md) for the full test strategy, how testcontainers and stubs work, and patterns for writing new tests.

## Tech stack

- **Go 1.26** / **Buffalo v1.1.4** — web framework with routing, middleware, Plush templates
- **Pop v6** — ORM and migration runner (migrations are raw SQL)
- **PostgreSQL 17** — primary database
- **pgstore** — gorilla/sessions PostgreSQL backend; 7-day fixed TTL sessions
- **Argon2id** — password hashing
