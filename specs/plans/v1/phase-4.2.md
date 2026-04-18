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

## Architecture

See `webapp/architecture.md` for full detail. Key points:

**New handler:** `TokensHandler` in `actions/tokens.go`
- Depends on: `AccountService` (token operations)
- Routes: `GET/POST /invitation/{token}`, `GET/POST /reinitialiser-mot-de-passe/{token}`

**AccountService additions** (`services/account.go`):

| Method | Notes |
|--------|-------|
| `CreatePending` | Called here for admin-side token generation display; full musician creation in Phase 4.3 |
| `GenerateInviteToken` | CSPRNG token; invalidates existing token for account |
| `ValidateInviteToken` | Checks unused + unexpired + pending account |
| `CompleteInvite` | Atomically: activates account, sets password, sets consent via `AccountRepository.Activate`; marks token used. RSVP seeding deferred — events table does not exist yet; wired in Phase 4.6 |
| `GeneratePasswordResetToken` | Active accounts only; invalidates existing |
| `ValidatePasswordResetToken` | Checks unused + unexpired + active account |
| `CompletePasswordReset` | Updates password hash; marks token used |

**New repository interfaces** (add to `services/repositories.go`):
- `InviteTokenRepository`: `Generate`, `FindByToken`, `MarkUsed`, `InvalidateExisting`
- `PasswordResetTokenRepository`: same shape

**AccountRepository additions:**
- `CreatePending(tx, email, instrumentID) → (int64, error)`
- `Activate(tx, id, passwordHash string, phoneAddressConsent bool) → error` — single SQL
  UPDATE atomically setting `status='active'`, `password_hash`, `phone_address_consent`,
  and conditionally clearing `phone`/`address`

**New model stores:** `InviteTokenStore`, `PasswordResetTokenStore` in `models/`

**New templates:**
- `templates/tokens/invite.plush.html` — invite form (password + consent checkboxes)
- `templates/tokens/invite_invalid.plush.html` — expired/used token message
- `templates/tokens/reset.plush.html` — new password form
- `templates/tokens/reset_invalid.plush.html` — expired/used token message

**`webapp/architecture.md` update required:** mark `InviteTokenRepository`,
`PasswordResetTokenRepository`, `InviteTokenStore`, `PasswordResetTokenStore`, and
`AccountService` token methods as implemented.

---

## Migrations

- `invite_tokens`
- `password_reset_tokens`

---

## Deliverables

- Invite URL display in admin UI (after account creation — admin copies and sends manually)
- Invite form: password field, privacy acknowledgement checkbox, phone/address consent checkbox
- Invite flow completion (atomic transaction per `technical-specs/05-implementation-notes.md`):
  activate account, set password hash, set consent flag, clear phone/address if no consent,
  mark token used, create session, redirect to `/evenements`
- Token regeneration (admin generates new invite link; invalidates existing token)
- Password reset: token generation + display, reset form, apply new password, mark token used
- Expired/used token handling (generic informative message, no account state change)

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

### Human-verified

**AC-H1 — Invite URL is displayed and copyable after musician creation**
Create a musician via the admin UI (Phase 4.3 dependency — verify together if 4.3 ships
first). The musician detail page shows the invite URL in a copyable field. Following the
URL in a new browser session renders the invite form.

**AC-H2 — Token regeneration updates the displayed URL**
Generate a new invite token for an existing pending account. The URL displayed in the
admin UI changes. The previous URL redirects to the invalid-token page.
