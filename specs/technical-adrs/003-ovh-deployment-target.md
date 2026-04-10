# ADR 003 — OVH Deployment Target

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-11 |

## Context

The vision requires deployment to OVH and lists the deployment target as a prerequisite for
completing technical specs. OVH offers several products with fundamentally different operational
models. The choice constrains the runtime language, database engine, deployment mechanism, and
operational overhead — and it directly affects whether the technical decisions already recorded
(ADR 001: sessions, ADR 002: UTC/TIMESTAMPTZ) need workarounds.

The system is small: one association, ≲200 users, single-region (France), low traffic, and
maintained by a small team over the long term. The vision calls for the simplest viable solution.

---

## Constraints from the Vision

- Must be straightforward to deploy to OVH
- Must be testable locally
- DB schema must be versioned
- Simplest viable solution preferred
- Developer tooling must allow bootstrapping the first admin account
- No system email infrastructure

---

## Alternatives

### A — OVH Web Hosting (Shared Hosting)

OVH's shared hosting product ("Hébergement Web"). The runtime is PHP; the database is MariaDB
(MySQL-compatible). OVH manages the OS, web server, and PHP runtime. Deployment via FTP/SFTP
or a read-only Git integration on some plans.

**Cost:** ~€2–5/month.

**Pros:**
- Zero server administration: no OS patches, no nginx config, no process supervision
- Cheapest option; SSL managed by OVH
- Familiar model for developers who know PHP

**Cons:**
- **Stack locked to PHP.** No Python, Node.js, or Go. Framework options: Symfony 6/7 or
  Laravel 10/11.
- **MariaDB, not PostgreSQL.** Two consequences for decisions already made:
  - ADR 002 (UTC/TIMESTAMPTZ): MariaDB has no native `TIMESTAMPTZ`; UTC convention must be
    enforced entirely in application code, with no DB-level guarantee.
  - Email partial unique index: MariaDB does not support partial indexes; the
    `UNIQUE(email) WHERE status != 'anonymized'` constraint requires an application-level
    check instead of a DB constraint, introducing a race condition window.
- FTP deployment: not reproducible, no atomic rollback, no audit trail.
- Shared environment: opportunistic exploits are historically more common on shared PHP hosting.
- No custom background processes or daemons.
- **Later milestone — sheets database:** an integrated sheet music experience with search
  requires either PostgreSQL FTS or a separate search process (MeiliSearch, Elasticsearch).
  PostgreSQL FTS is unavailable on shared hosting; a separate search process is impossible.
  This Later milestone would force a full platform migration if option A were chosen.

**Reversibility: Low.** Moving away means rewriting the application layer in a different
language. Data migration is feasible but the application code does not survive the move.

---

### B — OVH VPS (Direct, No Container)

A virtual private server with full root access. The developer manages the OS (Ubuntu/Debian),
a reverse proxy (nginx or Caddy), process supervision (systemd), and security patches. Any
runtime and any database can be installed.

**Cost:** ~€6–10/month (starter VPS, 2–4 GB RAM).

**Pros:**
- Full stack freedom: Python, Node.js, Go, PHP — whatever fits the team
- PostgreSQL: TIMESTAMPTZ and partial unique indexes work as specified in ADR 001/002, no
  workarounds
- Git-based deployment with scripted restart: reproducible and auditable
- systemd for process supervision and automatic restart on crash
- Rollback by checking out a previous Git tag and restarting

**Cons:**
- Developer owns the server: OS security patches, SSH hardening, firewall config (ufw), nginx/
  Caddy TLS config, log rotation
- Single point of failure (no HA, no auto-scaling — acceptable at this scale)
- Local dev environment is set up separately; parity with production depends on discipline

**Reversibility: Medium.** The application code is portable. Moving to a different host or
containerizing later is feasible without a rewrite, but requires effort.

---

### C — OVH VPS with Docker Compose

Same VPS as B, but the application (and its dependencies: database, reverse proxy) runs inside
Docker containers orchestrated by Docker Compose. The same Compose file drives both local
development and production.

**Cost:** Same as B; requires a VPS with ≥2 GB RAM for comfortable Docker operation. An image
registry (GitHub Container Registry, free tier) is optional but recommended.

**Pros:**
- **Local/prod parity:** `docker compose up` gives a running system identical to production.
  The vision requires local testability; this satisfies it with zero configuration drift.
- **Pinned dependencies:** runtime version, DB version, OS libraries — all locked in the image.
  Eliminates "works on my machine" debugging.
- **Reproducible deployment:** `docker compose pull && docker compose up -d` is atomic and
  scriptable. Rollback is pulling the previous image tag.
- **Onboarding:** a new developer gets a working environment with one command.
- The Compose file is machine-readable deployment documentation.
- Portable: images run on any Docker host, Kubernetes cluster, or managed container platform
  without code changes.

**Cons:**
- Adds Docker as a required concept: image builds, layer caching, Compose networking,
  volume mounts for DB persistence.
- Slightly more operational surface than B in steady state (container runtime, image lifecycle).
- Docker on a 1 GB VPS is tight; a 2 GB plan is the practical minimum.
- Image build step adds ~1–3 minutes to the deployment pipeline.

**Reversibility: High.** Container images are portable across hosts and orchestrators. Dropping
Docker later (if it becomes a burden) means running the same application binary directly on the
OS — the application code is unchanged.

---

### D — OVH Managed PaaS (Web Cloud / Cloud Web)

OVH's managed application hosting ("Web Cloud" or "Cloud Web") supports PHP, Python, and
Node.js runtimes on managed infrastructure. Closer to Heroku than to a VPS: push code, the
platform handles the runtime, scaling, and SSL.

**Cost:** ~€10–20/month depending on plan.

**Pros:**
- No server administration
- Supports Python and Node.js (not PHP-locked like shared hosting)
- Managed SSL, scaling knobs

**Cons:**
- OVH's managed PaaS product line has historically been less stable and less documented than
  their VPS or shared hosting offerings; vendor lock-in is higher.
- Constrained runtime environment: custom background processes, fine-grained DB config, and
  migration tooling may require workarounds.
- More expensive than a VPS for equivalent capability at this scale.
- Local dev parity is harder than with Docker (production environment is managed and opaque).

**Verdict: Eliminated.** Combines the constraints of shared hosting (limited control) with the
cost of a VPS, without a clear advantage at this scale. Not considered further.

---

## Comparison

| Criterion | A (Shared hosting) | B (VPS, no Docker) | C (VPS + Docker) |
|-----------|:------------------:|:------------------:|:----------------:|
| Stack freedom | None (PHP only) | Full | Full |
| Database | MariaDB | PostgreSQL | PostgreSQL |
| TIMESTAMPTZ (ADR 002) | Workaround | Native | Native |
| Partial unique index (ADR, email) | Workaround | Native | Native |
| Server admin burden | None | Medium | Medium |
| Local/prod parity | Low | Medium | High |
| Deployment reproducibility | Low | Medium | High |
| Rollback | Manual FTP | Git checkout | Image tag |
| Cost | €2–5/mo | €6–10/mo | €6–10/mo |
| Reversibility | Low | Medium | High |
| Conceptual overhead | Low | Medium | Medium–High |

---

## Impact on Remaining Technical Decisions

The deployment target unlocks the following decisions, which cascade into all remaining
technical specs:

| Decision | A (Shared hosting) | B / C (VPS) |
|----------|--------------------|-------------|
| Language | PHP | Open |
| Framework | Symfony or Laravel | Open |
| Database engine | MariaDB | PostgreSQL (strongly preferred) |
| Schema migration tool | Doctrine Migrations / Laravel Migrations | Alembic, Flyway, Flyway, golang-migrate, etc. |
| Session store | File or DB | DB-backed (per ADR 001) |
| Process management | OVH-managed (no control) | systemd or Docker Compose |
| Bootstrap CLI | Artisan / Symfony Console | Framework CLI or standalone script |
| Local dev setup | Manual PHP environment | Docker Compose (option C) or manual (option B) |

---

## Decision

**Option C — VPS with Docker Compose.**

1. **PostgreSQL without workarounds.** Two decisions already recorded (TIMESTAMPTZ, partial
   unique index) assume a capable relational database. PostgreSQL delivers both natively.
   Implementing workarounds in MariaDB would mean carrying technical debt from day one against
   decisions that were made on purpose.

2. **Local/prod parity is not optional at this scale.** The association has no dedicated ops
   team. The developer maintaining this system in two years may not be the one who built it.
   A `docker compose up` that produces a running system identical to production is the most
   reliable way to keep the system maintainable over time.

3. **The operational overhead of Docker is front-loaded.** Writing the Compose file is a
   one-time cost. The ongoing benefit — reproducible deploys, rollback, onboarding — accumulates
   over the lifetime of the system.

4. **Reversibility is highest.** If OVH changes their offering, if the association moves hosts,
   or if a future developer wants to containerize differently, the application code is unchanged.

**On the framework:** Option C opens the choice to any language. A separate ADR should record
this decision once the deployment target is confirmed. Python + Django is a strong candidate:
it ships with session auth, migrations, CSRF, and password hashing aligned with every decision
already made — but this is a proposal for the next ADR, not a decision here.

**On cost:** the delta between shared hosting and a starter VPS is ~€5/month. For an
association that is already paying OVH for hosting, this is likely within tolerance. Confirm
with the stakeholder.
