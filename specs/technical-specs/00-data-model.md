# OHM Website — Technical Data Model

> **Depends on:** [Functional Information Model](../functional-specs/00-information-model.md)
> **ADRs:** [002 — Datetime Storage: UTC](../technical-adrs/002-datetime-storage.md),
> [004 — Framework and Language](../technical-adrs/004-framework-and-language.md)
>
> **Note:** Column types use PostgreSQL notation. Migrations are written in Fizz DSL (Pop's
> migration format); see [`03-stack.md`](03-stack.md). Exact DDL lives in `migrations/`.

---

## accounts

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| first_name | TEXT | Yes | Null when anonymized |
| last_name | TEXT | Yes | Null when anonymized |
| email | TEXT | Yes | Null when anonymized. See uniqueness constraint below. |
| password_hash | TEXT | Yes | Null when anonymized. Argon2id. |
| main_instrument_id | BIGINT FK → instruments | No | Retained on anonymization |
| birth_date | DATE | Yes | Null when anonymized or not provided |
| parental_consent_uri | TEXT | Yes | Null when anonymized or not applicable |
| phone | TEXT | Yes | Null when anonymized, consent withdrawn, or not provided |
| address | TEXT | Yes | Null when anonymized, consent withdrawn, or not provided |
| phone_address_consent | BOOLEAN | No | Default false. False when anonymized or withdrawn. |
| status | TEXT | No | `pending` \| `active` \| `anonymized` |
| processing_restricted | BOOLEAN | No | Default false. False when anonymized. |
| anonymization_token | TEXT | Yes | Set at anonymization time. Null otherwise. |

**Constraints:**
- Partial unique index on `email WHERE status != 'anonymized'`. Multiple anonymized rows may
  have null email; this does not conflict with the uniqueness constraint.
- Application invariant (not DB-enforced): at all times, at least one `active` account holds
  the `admin` role. Enforced by last-admin protection checks before any operation that could
  reduce this count to zero. See ADR 005.

---

## roles

A controlled list of permission roles. Seeded by migration; no admin UI in V1.

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| name | TEXT UNIQUE | No | e.g. `admin` |

Seeded with a single entry: `admin`. See ADR 005.

---

## account_roles

Association table granting a role to an account.

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| account_id | BIGINT FK → accounts | No | ON DELETE CASCADE |
| role_id | BIGINT FK → roles | No | — |

**Constraint:** `PRIMARY KEY (account_id, role_id)` — one assignment per (account, role) pair.

`ON DELETE CASCADE` on `account_id` covers hard deletion of `pending` accounts. For
anonymization (account row is retained), roles are removed explicitly:
`DELETE FROM account_roles WHERE account_id = $1`.

---

## instruments

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| name | TEXT UNIQUE | No | — |

Populated by seed/migration data (OHM Agenda instrument list + "Chef d'orchestre"). No admin UI
to add or remove instruments in V1.

---

## seasons

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| label | TEXT | No | e.g. "2025-2026" |
| start_date | DATE | No | — |
| end_date | DATE | No | — |
| is_current | BOOLEAN | No | Default false |

**Constraints:**
- Partial unique index on `is_current WHERE is_current = true`. Ensures at most one season is
  designated current at the DB level.
- Application invariant: exactly one season has `is_current = true` once any season has been
  designated current. The invariant does not apply in the empty state (before any season is
  created). Designation transfers are executed in a single transaction.
- Seasons are immutable after creation (label, start_date, end_date). Enforced by the
  application; no DB trigger required.

---

## fee_payments

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| account_id | BIGINT FK → accounts | No | FK is retained after account anonymization. The anonymization token is read from the account record. |
| season_id | BIGINT FK → seasons | No | — |
| amount | NUMERIC(10, 2) | No | — |
| payment_date | DATE | No | The date the payment was made |
| payment_type | TEXT | No | `chèque` \| `espèces` \| `virement bancaire` |
| comment | TEXT | Yes | — |

**Constraints:**
- `UNIQUE(account_id, season_id)` — at most one payment per account per season.

**On anonymization:** The FK `account_id` continues to reference the now-anonymized account
record. Display logic reads `accounts.anonymization_token` in place of the account name. No
column in `fee_payments` changes at anonymization time.

---

## events

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| name | TEXT | No | — |
| datetime | TIMESTAMPTZ | No | Stored as UTC. See ADR 002. |
| event_type | TEXT | No | `concert` \| `rehearsal` \| `other` |

---

## rsvps

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| account_id | BIGINT FK → accounts | No | — |
| event_id | BIGINT FK → events | No | Cascade delete when event is deleted |
| state | TEXT | No | `unanswered` \| `yes` \| `no` \| `maybe` |
| instrument_id | BIGINT FK → instruments | Yes | Non-null only when state = `yes` and event type = `concert` |

**Constraints:**
- `UNIQUE(account_id, event_id)` — DB-level guard against duplicate RSVPs (e.g. from concurrent
  event creation and account activation).
- Cascade delete on `event_id`: deleting an event removes all its RSVP records.

---

## event_fields

Custom fields defined by an admin on `other`-type events.

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| event_id | BIGINT FK → events | No | Cascade delete when event is deleted |
| label | TEXT | No | — |
| field_type | TEXT | No | `choice` \| `integer` \| `text` |
| required | BOOLEAN | No | Default false |
| position | INTEGER | No | Display order |

---

## event_field_choices

Selectable options for `choice`-type event fields.

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| event_field_id | BIGINT FK → event_fields | No | Cascade delete when field is deleted |
| label | TEXT | No | — |
| position | INTEGER | No | Display order |

---

## rsvp_field_responses

A musician's response to a single event field, captured as part of a `yes` RSVP.

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| rsvp_id | BIGINT FK → rsvps | No | Cascade delete when RSVP is deleted |
| event_field_id | BIGINT FK → event_fields | No | Cascade delete when field is deleted |
| value | TEXT | No | Stored as text regardless of field type. For `choice` fields, the stored value is the `EventFieldChoice` ID. |

**Constraint:** `UNIQUE(rsvp_id, event_field_id)` — one response per (RSVP, field) pair.

---

## invite_tokens

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| account_id | BIGINT FK → accounts | No | — |
| token | TEXT UNIQUE | No | CSPRNG, 32 bytes, base64url-encoded |
| expires_at | TIMESTAMPTZ | No | UTC. 7 days from generation. |
| used | BOOLEAN | No | Default false. Set true when invite flow completes. |

At most one active (unused, non-expired) token per account. Generating a new token marks any
existing token for that account as used.

---

## password_reset_tokens

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| account_id | BIGINT FK → accounts | No | — |
| token | TEXT UNIQUE | No | CSPRNG, 32 bytes, base64url-encoded |
| expires_at | TIMESTAMPTZ | No | UTC. 7 days from generation. |
| used | BOOLEAN | No | Default false. Set true when password is changed. |

At most one active (unused, non-expired) token per account. Generating a new token marks any
existing token for that account as used.

---

## Infrastructure Tables

Tables required by framework dependencies, not part of the domain model. Migrations must
create these explicitly — they are not auto-created at the right time relative to application
startup.

### http_sessions

Managed by `pgstore` (`gorilla/sessions` PostgreSQL backend). Schema extended with an
`account_id` column beyond pgstore's default to support immediate session revocation on
account anonymization. See `05-implementation-notes.md` § Sessions for the revocation query
and login write pattern.

| Column | Type | Nullable | Notes |
|--------|------|----------|-------|
| id | BIGSERIAL PK | No | — |
| key | TEXT UNIQUE | No | Session ID (opaque) |
| data | BYTEA | No | Serialised session payload |
| created_on | TIMESTAMPTZ | No | — |
| modified_on | TIMESTAMPTZ | No | — |
| expires_on | TIMESTAMPTZ | No | pgstore uses this for expiry |
| account_id | BIGINT FK → accounts | Yes | Null until login completes. `ON DELETE CASCADE`. |

**Index:** `account_id` — required for the `DELETE FROM http_sessions WHERE account_id = $1`
revocation query to be O(sessions-per-account) rather than a full table scan.
