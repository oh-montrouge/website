# webapp

Go/Buffalo application for the OHM website. See the root `README.md` for project-wide
setup instructions and the full task reference.

See [`architecture.md`](architecture.md) for component structure, service ownership,
layer conventions, and DTO policy.

## Layout

```
actions/         HTTP handlers (Buffalo controllers)
cmd/app/         Application entrypoint
config/          Buffalo configuration (buffalo-app.toml, middleware)
grifts/          CLI tasks (db:seed:admin, db:recover:admin, db:seed:dev)
locales/         i18n strings
models/          Pop models and DB query functions
public/          Static assets
services/        Domain logic, repository interfaces, and DTOs
templates/       Plush HTML templates
docker-compose.yml   PostgreSQL 17 for local development
database.yml     Pop database configuration (reads DATABASE_URL)
.env.development.example   Environment variable template
```

Migrations live at `../db/migrations/` (repo root) — decoupled from the framework so
they can be run by any SQL-capable tool.

## Running locally

From the repo root, use `mise run dev`.
Then, you can seed some dummy data with `mise run seed-dev`.

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
- **Argon2id** — password hashing (`golang.org/x/crypto/argon2`, PHC format)
