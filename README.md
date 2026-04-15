# OHM Website

Website for the Orchestre d'Harmonie de Montrouge (OHM), a French concert band.

Replaces a decades-old static site and the internal OHM Agenda tool. Handles musician
accounts, event management (concerts, rehearsals), RSVP, annual fee payments, and GDPR
compliance — for admins and musicians alike.

## Repository layout

```
db/migrations/       SQL migrations (framework-independent, run by Pop via mise tasks)
specs/               Full specification tree — start with specs/INDEX.md
webapp/              Go/Buffalo application (see webapp/README.md)
mise.toml            Task runner and Go version pin
```

## Prerequisites

- [mise](https://mise.jdx.dev/) — manages Go, Buffalo CLI, and all dev tasks
- [Docker](https://docs.docker.com/engine/install/) — runs PostgreSQL locally

Run `mise install` once after cloning to install Go and the Buffalo CLI.

## Quick start

```bash
mise install        # Install Go, Buffalo CLI, Lefthook, golangci-lint — and register git hooks

cp webapp/.env.development.example webapp/.env
# Edit webapp/.env: set SESSION_SECRET to at least 32 random bytes

mise run dev        # Start postgres + buffalo dev server (http://localhost:3000)
mise run migrate    # Apply pending migrations (first run: creates schema + seeds instruments)
```

## Task reference

| Task | Description |
|------|-------------|
| `mise run dev` | Start postgres (Docker) + Buffalo dev server |
| `mise run stop` | Stop postgres + kill the running dev server |
| `mise run migrate` | Apply pending DB migrations |
| `mise run test` | Run test suite (starts postgres, creates test DB, migrates, tests) |
| `mise run lint` | Run golangci-lint (Phase 1 — not yet wired up) |
| `mise run seed-admin` | Create first admin account (Phase 3 — not yet implemented) |
| `mise run seed-dev` | Populate DB with dev dummy data (Phase 3 — not yet implemented) |
| `mise run recover-admin` | Emergency admin password reset (Phase 3 — not yet implemented) |
| `mise run deploy` | Deploy to production VPS (Phase 2 — no infra yet) |

## Specs

`specs/INDEX.md` is the entry point for the full specification tree: vision, functional
specs, technical specs, ADRs, and architecture notes.
