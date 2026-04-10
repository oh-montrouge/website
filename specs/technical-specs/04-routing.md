# OHM Website — Routing

> **Depends on:** [Account Lifecycle](../functional-specs/01-account-lifecycle.md),
> [Events and RSVP](../functional-specs/05-events-and-rsvp.md),
> [Auth and Security](01-auth-and-security.md)

---

## Middleware Stack

Applied globally in `app.go`, in this order:

1. Recovery (panic → HTTP 500, log stack trace)
2. Request logger
3. Session store (pgstore)
4. CSRF (`gorilla/csrf`)
5. Route-level auth middleware (per group — see below)

---

## Route Groups

### Public (no authentication required)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/` | Homepage (static; no DB query) |
| GET | `/connexion` | Login form |
| POST | `/connexion` | Authenticate |
| GET | `/deconnexion` | Destroy session and redirect to `/connexion` |
| GET | `/politique-de-confidentialite` | Privacy notice (static template, no DB query) |
| GET | `/invitation/{token}` | Invite form |
| POST | `/invitation/{token}` | Complete account setup |
| GET | `/reinitialiser-mot-de-passe/{token}` | Password reset form |
| POST | `/reinitialiser-mot-de-passe/{token}` | Apply new password |

`/deconnexion` uses GET for simplicity; CSRF risk is mitigated by SameSite=Strict on the
session cookie. Prefer POST with a form if stricter compliance is required.

### Authenticated (active account required)

Middleware: `RequireActiveAccount` — checks session, loads account, verifies `status = 'active'`;
redirects to `/connexion` otherwise.

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/profil` | Own profile (read-only in V1) |
| GET | `/evenements` | Event list with own RSVP state (last 30 days + all upcoming; see [Events and RSVP](../functional-specs/05-events-and-rsvp.md#event-list-view)) |
| GET | `/evenements/{id}` | Event detail, full RSVP list, own RSVP form |
| POST | `/evenements/{id}/rsvp` | Update own RSVP state (and instrument if concert) |

### Admin (active account + admin role required)

Middleware: `RequireAdmin` — applied after `RequireActiveAccount`; checks for a row in
`account_roles` joining to `roles.name = 'admin'`; returns HTTP 403 otherwise.

All paths are prefixed `/admin`.

**Accounts / musicians:**

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/admin/musiciens` | Musician list |
| GET | `/admin/musiciens/nouveau` | New musician form |
| POST | `/admin/musiciens` | Create musician account |
| GET | `/admin/musiciens/{id}` | Musician detail (profile + fees + tokens) |
| GET | `/admin/musiciens/{id}/modifier` | Edit form |
| PUT | `/admin/musiciens/{id}` | Save edits |
| DELETE | `/admin/musiciens/{id}` | Delete pending account |
| POST | `/admin/musiciens/{id}/anonymiser` | Anonymize account (requires confirmation param) |
| POST | `/admin/musiciens/{id}/invitation` | Generate new invite link |
| POST | `/admin/musiciens/{id}/reinitialisation` | Generate new password reset link |
| POST | `/admin/musiciens/{id}/role-admin` | Grant admin role |
| DELETE | `/admin/musiciens/{id}/role-admin` | Revoke admin role |
| DELETE | `/admin/musiciens/{id}/consentement` | Withdraw phone/address consent |
| POST | `/admin/musiciens/{id}/restriction` | Toggle processing-restricted flag |

**Fee payments** (nested under musician for creation, standalone for edit/delete):

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/admin/musiciens/{account_id}/cotisations` | Record fee payment |
| GET | `/admin/cotisations/{id}/modifier` | Edit fee payment form |
| PUT | `/admin/cotisations/{id}` | Save edit |
| DELETE | `/admin/cotisations/{id}` | Delete fee payment |

**Seasons:**

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/admin/saisons` | Season list |
| POST | `/admin/saisons` | Create season |
| POST | `/admin/saisons/{id}/courante` | Designate as current season |

**Events:**

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/admin/evenements` | Event management list |
| GET | `/admin/evenements/nouveau` | New event form |
| POST | `/admin/evenements` | Create event |
| GET | `/admin/evenements/{id}/modifier` | Edit event form |
| PUT | `/admin/evenements/{id}` | Save edits |
| DELETE | `/admin/evenements/{id}` | Delete event and all RSVPs |

**Event custom fields** (only applicable to `other`-type events):

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/admin/evenements/{id}/champs` | Add a field to the event |
| GET | `/admin/evenements/{event_id}/champs/{field_id}/modifier` | Edit field form |
| PUT | `/admin/evenements/{event_id}/champs/{field_id}` | Save field edits |
| DELETE | `/admin/evenements/{event_id}/champs/{field_id}` | Delete field (and choices) |

**GDPR / retention:**

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/admin/retention` | Accounts whose retention period has elapsed |

---

## Conventions

- **PUT and DELETE from browsers:** sent as POST with a hidden `_method` field
  (`<input type="hidden" name="_method" value="PUT">`). Buffalo's method-override middleware
  handles this transparently.
- **Path parameters:** `{id}`, `{token}`, `{account_id}` — accessed via `c.Param("id")` in
  handlers.
- **Confirmation for destructive actions:** anonymization, account deletion, and event deletion
  require a confirmation step before the final POST. Implemented as a two-step flow: a
  confirmation form (GET or modal) followed by the destructive POST with a `confirmed=true`
  parameter. The handler rejects the POST without the parameter.
- **Admin re-check on every request:** `RequireAdmin` queries `account_roles` on each request
  (not from session state) — the role can be revoked while a session is active. See ADR 005.
