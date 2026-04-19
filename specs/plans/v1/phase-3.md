# Phase 3 — Production Infra

**Status:** Done

**Goal:** application reachable at the OVH domain over HTTPS, daily backups running.

---

## Prerequisites

Depends on: Phase 2 (any stable commit works; Phase 3 is independent of Phase 4)

---

## Specs

- `specs/technical-adrs/003-ovh-deployment-target.md` — hosting decision
- `specs/technical-adrs/006-backup-strategy.md` — backup approach
- `specs/technical-specs/02-configuration.md` — environment variables
- `specs/technical-specs/06-ci-cd.md` — deployment pipeline

---

## Architecture

This phase is ops/infra only. No new services, repository interfaces, DTOs, or handler
files. No changes to the three-layer architecture.

`webapp/architecture.md` does not need updating after this phase.

---

## Deliverables

- Multi-stage `Dockerfile` (build stage: `golang:alpine`; runtime stage: `alpine`, ~25–30 MB image)
- `docker-compose.yml` extended with `caddy` service; `Caddyfile` with automatic HTTPS
- `.env.production.example`
- `mise run deploy` task: SSH to VPS → `docker compose pull` → `up -d` → `buffalo pop migrate`
- `DEPLOY_HOST` read from operator shell profile (not committed)
- `scripts/backup.sh`: `pg_dump` → gzip → upload to OVH Object Storage (30-day retention)
- VPS cron setup instructions for the backup script
- Production smoke test checklist (HTTPS green, login page loads, backup job runs once manually)

---

## Acceptance Criteria

### Machine-verified

Commands the deployer runs; each has a clear pass/fail output.

**AC-M1 — Image is lean** *(local)*
`docker build -t ohm-webapp .` succeeds. `docker image inspect ohm-webapp --format '{{.Size}}'`
returns ≤ 90 MB. Confirms the multi-stage build is not carrying Go toolchain or dev
dependencies into the runtime image. The ceiling is higher than a pure app image because
`buffalo` and `buffalo-pop` are co-packaged for deployment migrations (see Dockerfile comment).

**AC-M2 — App starts with production config** *(local)*
`docker compose up` with a `.env` containing valid `DATABASE_URL`, `SESSION_SECRET`, and
`GO_ENV=production` → container reaches healthy state; `curl -s -o /dev/null -w "%{http_code}"
-H "X-Forwarded-Proto: https" http://localhost:{port}/connexion` returns `200`.
Note: without the header, the app returns 301 (force-SSL redirect). In production, Caddy
injects `X-Forwarded-Proto: https`, so the app serves 200 normally.

**AC-M3 — Deploy runs migrations** *(production)*
After `mise run deploy` on a VPS with a pending migration,
`docker compose exec app buffalo-pop pop migrate status` shows all migrations as applied.
Confirms the deploy task runs `buffalo-pop pop migrate` after pulling the image.

**AC-M4 — HTTPS is live** *(production)*
`curl -s -o /dev/null -w "%{http_code}" https://{domain}/connexion` returns `200` with no certificate errors.
`curl -I http://{domain}/connexion` returns a redirect to `https://`.

**AC-M5 — Backup produces an artifact** *(production)*
`bash scripts/backup.sh` exits 0. A `.sql.gz` file dated today appears in the OVH Object
Storage bucket. File is non-empty (`du -h` shows > 0).

### Human-verified

**AC-H1 — Login persists across pages in production**
Using the `seed-admin` account on the live domain: log in, navigate to at least two
protected pages, log out. Session is maintained throughout and destroyed on logout.

**AC-H2 — No secrets in the image**
`docker run --rm ohm-webapp env` does not print `SESSION_SECRET`, `DATABASE_URL`, or any
credential. Confirms secrets are injected at runtime only.

**AC-H3 — Backup retention is configured**
In the OVH Object Storage console, confirm the bucket lifecycle policy is set to 30-day
retention (or equivalent). Running the backup script a second time does not fail due to
a duplicate filename.
