# Phase 4.2 — Account Lifecycle

**Status:** Not started

**Goal:** invite and password reset flows. A musician can follow an invite URL, set their
password and consent, and land on the events page. An admin can generate a reset link for
an active account.

---

## Prerequisites

Depends on: Phase 2

---

## Specs

- `specs/functional-specs/01-account-lifecycle.md` § Invite Flow, § Password Reset
- `specs/technical-specs/05-implementation-notes.md` § Token Validation, § Invite Flow
  Completion, § Token Regeneration
- `specs/technical-specs/04-routing.md` § Public routes (`/invitation/{token}`,
  `/reinitialiser-mot-de-passe/{token}`)
- `specs/technical-specs/00-data-model.md` § invite_tokens, § password_reset_tokens

---

## Design Reference

Wireframe screens: `InviteScreen` (active and expired states), `PasswordResetScreen`
(`public.jsx`).
Source: `specs/plans/v1/wireframes/wireframes/project/`

**Alpine.js usage in this phase:**
- Invite form submit gate: `x-data="{ consentPrivacy: false }"` on the `<form>`; `x-model`
  on the privacy checkbox; `x-bind:disabled="!consentPrivacy"` on the submit button

---

## Architecture

See `webapp/architecture.md` for full detail. Key points:

**New handler:** `TokensHandler` in `actions/tokens.go`
- Depends on: `AccountService` (token operations)
- Routes: `GET/POST /invitation/{token}`, `GET/POST /reinitialiser-mot-de-passe/{token}`

**AccountService additions** (`services/account.go`):

| Method | Notes |
|--------|-------|
| `GenerateInviteToken` | CSPRNG token; invalidates existing token for account; returns `InviteTokenDTO` |
| `ValidateInviteToken` | Checks unused + unexpired + pending account |
| `CompleteInvite` | 2-step transaction: (1) activate account via `AccountRepository.Activate`, (2) mark token used. RSVP seeding deferred — EventRepository dependency wired in Phase 4.6 |
| `GeneratePasswordResetToken` | Active accounts only; invalidates existing; returns `PasswordResetTokenDTO` |
| `ValidatePasswordResetToken` | Checks unused + unexpired + active account |
| `CompletePasswordReset` | Updates password hash; marks token used |

`CreatePending` is Phase 4.3, not part of this phase.

**New DTOs** (add to `services/account.go`):
- `InviteTokenDTO`: `Token`, `URL`, `ExpiresAt`
- `PasswordResetTokenDTO`: `Token`, `URL`, `ExpiresAt`

The invite form GET handler must display the account's name, email, and instrument. These
are not in `AccountDTO`; the handler should obtain them via a dedicated query or by extending
the `ValidateInviteToken` return value. Approach is left to the implementer; the invite form
template must receive them.

**New repository interfaces** (add to `services/repositories.go`):
- `InviteTokenRepository`: `Generate`, `FindByToken`, `MarkUsed`, `InvalidateExisting`
- `PasswordResetTokenRepository`: same shape

**AccountRepository addition (Phase 4.2):**
- `Activate(tx, id, passwordHash string, phoneAddressConsent bool) → error` — single SQL
  UPDATE atomically setting `status='active'`, `password_hash`, `phone_address_consent`,
  and conditionally clearing `phone`/`address`

**New model stores:** `InviteTokenStore`, `PasswordResetTokenStore` in `models/`

**New templates:**
- `templates/tokens/invite.plush.html` — account info card (name, email, instrument),
  password field + confirm field, privacy acknowledgement checkbox (link to
  `/politique-de-confidentialite`), phone/address consent checkbox, Alpine-gated submit button
- `templates/tokens/invite_invalid.plush.html` — expired/used token message
- `templates/tokens/reset.plush.html` — new password field + confirm field
- `templates/tokens/reset_invalid.plush.html` — expired/used token message

**`webapp/architecture.md` update required:** mark `InviteTokenRepository`,
`PasswordResetTokenRepository`, `InviteTokenStore`, `PasswordResetTokenStore`,
`InviteTokenDTO`, `PasswordResetTokenDTO`, and `AccountService` token methods as implemented.

---

## Migrations

- `invite_tokens`
- `password_reset_tokens`

---

## Implementation Units

### Unit 1 — Migrations

- [ ] `invite_tokens` migration (schema per `technical-specs/00-data-model.md`)
- [ ] `password_reset_tokens` migration

### Unit 2 — Service layer

- [ ] `InviteTokenRepository` interface in `services/repositories.go`
- [ ] `PasswordResetTokenRepository` interface in `services/repositories.go`
- [ ] `InviteTokenStore` in `models/` (implements `InviteTokenRepository`)
- [ ] `PasswordResetTokenStore` in `models/` (implements `PasswordResetTokenRepository`)
- [ ] `AccountRepository.Activate` in `models/` (single SQL UPDATE)
- [ ] `InviteTokenDTO`, `PasswordResetTokenDTO` in `services/account.go`
- [ ] `AccountService.GenerateInviteToken`
- [ ] `AccountService.ValidateInviteToken`
- [ ] `AccountService.CompleteInvite` (2-step transaction; no RSVP seeding)
- [ ] `AccountService.GeneratePasswordResetToken`
- [ ] `AccountService.ValidatePasswordResetToken`
- [ ] `AccountService.CompletePasswordReset`
- [ ] Unit tests for all `AccountService` methods (stub repositories)
- [ ] Integration tests for `InviteTokenStore`, `PasswordResetTokenStore`, `AccountRepository.Activate`

### Unit 3 — Invite flow

- [ ] `TokensHandler` struct in `actions/tokens.go`; routes wired in `app.go`
- [ ] `GET /invitation/{token}` — validate token, render form with account info card
- [ ] `POST /invitation/{token}` — validate token, validate password (complexity + confirm match),
  call `CompleteInvite`, create session (with `account_id` column write per
  `technical-specs/05-implementation-notes.md` § Sessions), redirect to `/evenements`
- [ ] `templates/tokens/invite.plush.html`
- [ ] `templates/tokens/invite_invalid.plush.html`
- [ ] Handler tests: invite GET (valid token, invalid token) and POST (valid, invalid token,
  password mismatch, missing privacy consent)

### Unit 4 — Password reset flow

- [ ] `GET /reinitialiser-mot-de-passe/{token}` — validate token, render reset form
- [ ] `POST /reinitialiser-mot-de-passe/{token}` — validate token, validate password
  (complexity + confirm match), call `CompletePasswordReset`, redirect to `/connexion`
- [ ] `templates/tokens/reset.plush.html`
- [ ] `templates/tokens/reset_invalid.plush.html`
- [ ] Handler tests: reset GET (valid token, invalid token) and POST (valid, invalid token,
  password mismatch)

---

## Password Validation

Applied server-side to both invite and reset forms:

- Minimum 22 characters
- Must contain at least one uppercase letter, one lowercase letter, one digit, and one
  special character
- Confirmation field must match the password field

On validation failure, re-render the form with an error message. Do not echo the submitted
token into the page.

---

## Deliverables

- Invite form: account info card (name, email, instrument), password + confirm fields,
  privacy acknowledgement checkbox, phone/address consent checkbox, Alpine-gated submit button
- Invite flow completion: activate account, set password hash, set consent flag, clear
  phone/address if no consent, mark token used, create session, redirect to `/evenements`
- Token regeneration service method (admin UI entry point in Phase 4.3)
- Password reset: reset form (password + confirm), apply new password, mark token used,
  redirect to `/connexion`
- Expired/used token handling: generic informative message, no account state change,
  no distinction between expired and used
- `InviteTokenDTO`, `PasswordResetTokenDTO`

---

## Manual Verification

The admin UI that creates musicians and displays invite URLs lands in Phase 4.3. To verify
the invite and reset forms in Phase 4.2, seed state directly via `psql`:

```
docker compose exec postgres psql -U ohm ohm_development
```

**Seed a pending account and invite token:**

```sql
DO $$
DECLARE
  v_account_id    BIGINT;
  v_instrument_id BIGINT;
BEGIN
  SELECT id INTO v_instrument_id FROM instruments ORDER BY id LIMIT 1;
  INSERT INTO accounts (email, first_name, last_name, main_instrument_id,
                        status, phone_address_consent)
    VALUES ('testinvite@localhost', 'Alice', 'Testinvite',
            v_instrument_id, 'pending', false)
    RETURNING id INTO v_account_id;
  INSERT INTO invite_tokens (account_id, token, expires_at, used)
    VALUES (v_account_id, 'dev-invite-token',
            NOW() + INTERVAL '7 days', false);
  RAISE NOTICE 'Invite URL: http://localhost:3000/invitation/dev-invite-token';
END $$;
```

**Test the expired/used state:** `UPDATE invite_tokens SET used = true WHERE token = 'dev-invite-token';`

**Seed a password reset token** (after completing the invite flow above):

```sql
INSERT INTO password_reset_tokens (account_id, token, expires_at, used)
SELECT id, 'dev-reset-token', NOW() + INTERVAL '7 days', false
FROM accounts WHERE email = 'testinvite@localhost';
```

Visit: `http://localhost:3000/reinitialiser-mot-de-passe/dev-reset-token`

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — Expired and used tokens show the same generic message**
`GET /invitation/{expired-token}` and `GET /invitation/{used-token}` both render the
invalid-token page; neither reveals whether the token expired or was already used.
Account `status` remains `pending` in both cases.

**AC-M2 — Invite completion activates the account atomically**
POST `/invitation/{valid-token}` with `consent=false` → in a single transaction:
`status='active'`, `password_hash` set, `phone_address_consent=false`, `phone=NULL`,
`address=NULL`, token marked `used=true`. A session cookie is set on the response.

**AC-M3 — Consent flag preserves existing phone/address when true**
POST `/invitation/{valid-token}` on an account that has pre-populated `phone` and
`address` (e.g. from migration) with `consent=true` → `phone` and `address` are
unchanged after activation.

**AC-M4 — Token regeneration invalidates the previous token**
`AccountService.GenerateInviteToken` called on an account that already has an unused
token → old token row has `used=true`; new token row has `used=false`. Only one unused
token exists for the account at any time.

**AC-M5 — Password reset is unavailable for pending accounts**
`GET /reinitialiser-mot-de-passe/{token}` for a token belonging to a `pending` account
→ renders the invalid-token page; account is unaffected.

**AC-M6 — Invite completion redirects to `/evenements`**
Successful POST `/invitation/{valid-token}` → `303` redirect to `/evenements`.
(The route may not be implemented yet; the redirect itself is what is verified here.)

**AC-M7 — Password validation is enforced server-side**
POST `/invitation/{valid-token}` with a password shorter than 22 characters → form
re-rendered with an error; account remains `pending`, token remains unused.
Same check applies to `/reinitialiser-mot-de-passe/{token}`.

### Human-verified

**AC-H1 — Invite form displays account info**
Using the seed SQL from Manual Verification, follow the invite URL. The form shows an
account info card with the seeded name, email, and instrument. The submit button is
disabled until the privacy checkbox is checked.

**AC-H2 — Password reset form is reachable and functional**
Using the seed SQL from Manual Verification, follow the reset URL after completing the
invite flow. The form accepts a valid password and redirects to `/connexion`.
