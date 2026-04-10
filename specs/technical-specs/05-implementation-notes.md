# OHM Website — Implementation Notes

> **Depends on:** all functional specs, [Data Model](00-data-model.md),
> [Auth and Security](01-auth-and-security.md), [Routing](04-routing.md)
>
> This spec documents the non-trivial implementation decisions for operations that are
> specified functionally but require algorithmic or transactional precision to implement
> correctly.

---

## Sessions

### Custom sessions table schema

pgstore's default session table has no `account_id` column. ADR 001 requires immediate session
invalidation on anonymization, which means being able to delete sessions by account. The
sessions table must therefore be created manually in a migration (before pgstore's
auto-creation runs, or by disabling auto-creation) with an additional column:

```sql
CREATE TABLE http_sessions (
    id          BIGSERIAL PRIMARY KEY,
    key         TEXT UNIQUE NOT NULL,
    data        BYTEA NOT NULL,
    created_on  TIMESTAMPTZ DEFAULT NOW(),
    modified_on TIMESTAMPTZ DEFAULT NOW(),
    expires_on  TIMESTAMPTZ,
    account_id  BIGINT REFERENCES accounts(id) ON DELETE CASCADE
);
CREATE INDEX ON http_sessions (account_id);
```

When creating a session on login, write the account ID into both the session data and this
column. On anonymization, `DELETE FROM http_sessions WHERE account_id = $1` immediately
revokes access.

---

## Token Validation (Invite and Password Reset)

Shared pattern for both token types:

```sql
SELECT t.id, t.account_id, a.status
FROM [invite_tokens | password_reset_tokens] t
JOIN accounts a ON a.id = t.account_id
WHERE t.token = $1
  AND t.used = false
  AND t.expires_at > NOW()
```

If no row is returned (token not found, already used, or expired), display a generic
informative message — do not distinguish between the three cases in the UI response.

For invite tokens: additionally verify `a.status = 'pending'`.
For password reset tokens: additionally verify `a.status = 'active'`.

---

## Invite Flow Completion

Executed atomically in a single database transaction:

```sql
BEGIN;

-- 1. Activate account, set password and consent
UPDATE accounts
SET status = 'active',
    password_hash = $1,          -- Argon2id hash
    phone_address_consent = $2   -- from checkbox
WHERE id = $3;

-- 2. Mark token used
UPDATE invite_tokens SET used = true WHERE id = $4;

-- 3. Create RSVP records for all future events
INSERT INTO rsvps (account_id, event_id, state)
SELECT $3, id, 'unanswered'
FROM events
WHERE datetime > NOW()           -- UTC comparison; see ADR 002
ON CONFLICT (account_id, event_id) DO NOTHING;  -- guard against race

COMMIT;
```

After the transaction: create a session for the newly active account (including the
`account_id` column write — see Sessions above), then redirect to the home page.

---

## Token Regeneration (Invite and Password Reset)

When the admin generates a new token for an account, invalidate any existing active token
first. Atomic:

```sql
BEGIN;
UPDATE [invite_tokens | password_reset_tokens]
SET used = true
WHERE account_id = $1 AND used = false;

INSERT INTO [invite_tokens | password_reset_tokens]
    (account_id, token, expires_at, used)
VALUES ($1, $2, NOW() + INTERVAL '7 days', false);
COMMIT;
```

---

## Anonymization

Executed atomically in a single database transaction. The anonymization token is generated in
Go before the transaction begins (CSPRNG, 32 bytes, base64url-encoded).

```sql
BEGIN;

-- 1. Clear personal fields, set anonymization token, transition status
UPDATE accounts
SET first_name             = NULL,
    last_name              = NULL,
    email                  = NULL,
    password_hash          = NULL,
    birth_date             = NULL,
    parental_consent_uri   = NULL,
    phone                  = NULL,
    address                = NULL,
    phone_address_consent  = false,
    processing_restricted  = false,
    status                 = 'anonymized',
    anonymization_token    = $2       -- generated before transaction
WHERE id = $1;

-- 2. Remove all role assignments
DELETE FROM account_roles WHERE account_id = $1;

-- 3. Delete RSVP records
DELETE FROM rsvps WHERE account_id = $1;

-- 4. Invalidate any pending invite or reset tokens
UPDATE invite_tokens
SET used = true
WHERE account_id = $1 AND used = false;

UPDATE password_reset_tokens
SET used = true
WHERE account_id = $1 AND used = false;

-- 5. Destroy active session
DELETE FROM http_sessions WHERE account_id = $1;

COMMIT;
```

FeePayment records are intentionally not modified. The anonymization token is now stored on
the account record; display logic reads it from there when rendering payment history.

Last-admin protection is checked before the transaction begins (see below).

---

## Last-Admin Protection

Before any operation that could reduce the active admin count (revoke admin role, anonymize an
admin account, delete a pending admin account), run:

```sql
SELECT COUNT(*)
FROM account_roles ar
JOIN roles r ON r.id = ar.role_id
JOIN accounts a ON a.id = ar.account_id
WHERE r.name = 'admin' AND a.status = 'active';
```

If the result is `1` and the target account is that admin, reject the operation with an
informative message. No transaction needed — this is a pre-flight check; the atomic mutation
follows separately.

Note: this check is subject to a TOCTOU race (two admins revoking each other simultaneously).
At OHM's scale, the risk is negligible. If stricter enforcement is needed, run the check
inside the mutation transaction with a `SELECT ... FOR UPDATE`.

---

## Bulk RSVP Creation on Event Creation

When an admin creates an event, RSVP records are created for all active accounts in one
statement:

```sql
INSERT INTO rsvps (account_id, event_id, state)
SELECT id, $1, 'unanswered'
FROM accounts
WHERE status = 'active'
ON CONFLICT (account_id, event_id) DO NOTHING;
```

The `ON CONFLICT` clause is a safety net against a race where an account is activated
concurrently with event creation.

---

## Event Type Change: Effects on RSVPs and Fields

When an admin saves an event with a changed type, the following run within the same
transaction as the event UPDATE:

**To `concert` (from `rehearsal`):** Reset all `yes` RSVPs to `unanswered` (instrument
selection was not collected):

```sql
UPDATE rsvps
SET state = 'unanswered', instrument_id = NULL
WHERE event_id = $1 AND state = 'yes';
```

**To `concert` (from `other`):** Delete all custom fields (cascades to choices and responses),
then reset `yes` RSVPs:

```sql
DELETE FROM event_fields WHERE event_id = $1;  -- cascades to choices and responses

UPDATE rsvps
SET state = 'unanswered', instrument_id = NULL
WHERE event_id = $1 AND state = 'yes';
```

**To `rehearsal` (from `other`):** Delete all custom fields (cascades to choices and
responses); RSVP states are retained:

```sql
DELETE FROM event_fields WHERE event_id = $1;  -- cascades to choices and responses
```

**From `concert` (to `rehearsal` or `other`):** Keep all RSVP states; clear instrument
selections:

```sql
UPDATE rsvps SET instrument_id = NULL
WHERE event_id = $1 AND instrument_id IS NOT NULL;
```

**No type change, or same type:** No RSVP or field modification.

---

## Field Edit/Delete Guard

Before allowing an admin to edit or delete a custom field, check whether any responses exist:

```sql
SELECT COUNT(*) FROM rsvp_field_responses WHERE event_field_id = $1;
```

If the count is non-zero, reject the operation with an informative message.

---

## RSVP Field Response Clearing

When a `yes` RSVP is changed to `no` or `maybe`, delete all field responses for that RSVP:

```sql
DELETE FROM rsvp_field_responses WHERE rsvp_id = $1;
```

This runs within the same transaction as the RSVP state UPDATE.

---

## Season Designation Transfer

Executed in a single transaction to maintain the exactly-one-current invariant:

```sql
BEGIN;
UPDATE seasons SET is_current = false WHERE is_current = true;
UPDATE seasons SET is_current = true  WHERE id = $1;
COMMIT;
```

The partial unique index on `is_current WHERE is_current = true` (defined in the data model)
provides DB-level defense against concurrent designation.

---

## First Inscription Date

Computed on demand; not stored. Used in the admin musician detail view:

```sql
SELECT MIN(payment_date) AS first_inscription_date
FROM fee_payments
WHERE account_id = $1;
```

Returns `NULL` if the account has no fee payment records. The UI omits the field when `NULL`.

---

## Retention Review List

Accounts whose retention period has elapsed: the end date of the season of their most recent
fee payment is more than 5 years ago, and they have not been anonymized.

```sql
SELECT
    a.id,
    a.first_name,
    a.last_name,
    i.name  AS instrument,
    s.label AS last_season_label,
    s.end_date
FROM accounts a
JOIN LATERAL (
    SELECT season_id
    FROM fee_payments
    WHERE account_id = a.id
    ORDER BY payment_date DESC
    LIMIT 1
) last_fp ON true
JOIN seasons s ON s.id = last_fp.season_id
JOIN instruments i ON i.id = a.main_instrument_id
WHERE a.status != 'anonymized'
  AND s.end_date < NOW() - INTERVAL '5 years'
ORDER BY s.end_date ASC;
```

`LATERAL` is a PostgreSQL feature; this query does not need to be portable. Accounts with no
fee payment records are excluded by the `JOIN LATERAL` (inner join semantics).
