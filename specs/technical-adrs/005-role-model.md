# ADR 005 — Role Model: Association Table over Boolean Flag

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-11 |

## Context

V1 requires a single permission level beyond "authenticated user": the `admin` role. The
simplest implementation is a boolean `is_admin` column on the `accounts` table.

The vision's Later section identifies at least two features that would require additional
permission levels: a public-facing content management function and the "Commission Artistique" feature.
Both would require new roles that cannot be expressed as booleans without adding more columns.

---

## Decision

Implement roles as a separate `roles` table with an `account_roles` association table, rather
than as boolean flag columns on `accounts`.

```
roles:         id BIGSERIAL PK, name TEXT UNIQUE NOT NULL
account_roles: account_id BIGINT FK → accounts, role_id BIGINT FK → roles,
               PRIMARY KEY (account_id, role_id)
```

The `roles` table is seeded with a single entry (`admin`) at migration time. No admin UI to
manage roles in V1; role assignment is via the existing grant/revoke admin UI routes, which
target the `admin` role specifically.

---

## Alternatives Considered

**A — Boolean `is_admin` column (rejected)**

Simpler for V1. The schema change required when a second role arrives is small in isolation,
but schema migrations on live data carry risk and operational cost disproportionate to their
size. Every additional role would require a new migration, a new column, and updates to all
role-check queries.

**B — Association table (chosen)**

More complex than a boolean for a single role, but the extension cost for new roles is zero
schema work: insert a row in `roles`, add application logic. The schema migrates cleanly to
any permission model.

---

## Trade-offs

| Concern | Impact |
|---------|--------|
| Query complexity | Last-admin check and role-check queries require a JOIN instead of a direct column read. Minor, and the query stays simple. |
| Seed data required | `roles` table must be pre-populated; handled in migration. |
| KISS tension | Vision says "defer decisions that have no present impact." Acknowledged. The schema is the hardest artifact to migrate post-launch; application-layer role logic would be rewritten regardless when a new role arrives. Paying the schema cost now is judged worthwhile. |

---

## Consequences

- `is_admin BOOLEAN` removed from `accounts`.
- `roles` and `account_roles` tables added (see data model).
- Anonymization clears roles via `DELETE FROM account_roles WHERE account_id = $1`.
- Admin role check: `SELECT 1 FROM account_roles ar JOIN roles r ON r.id = ar.role_id WHERE ar.account_id = $1 AND r.name = 'admin'`.
- Last-admin protection uses a COUNT with a JOIN (see implementation notes).
- Future roles require only: a new row in `roles`, new middleware, new UI — no schema migration.
