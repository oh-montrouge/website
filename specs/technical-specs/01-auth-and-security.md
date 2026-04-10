# OHM Website — Authentication and Security

> **Depends on:** [Functional Account Lifecycle](../functional-specs/01-account-lifecycle.md)
> **ADRs:** [001 — Authentication: Server-Side Sessions](../technical-adrs/001-authentication.md),
> [004 — Framework and Language](../technical-adrs/004-framework-and-language.md)

---

## Session-Based Authentication

Authentication uses server-side sessions per ADR 001.

- On successful login, a session record is created server-side and a session ID is set in a
  cookie.
- **Cookie flags:** HttpOnly (no JS access), Secure (HTTPS only), SameSite=Strict.
- Only `active` accounts may log in. The application checks account status before creating a
  session.
- On anonymization, the account's session (if any) is invalidated server-side. The cookie
  becomes stale and subsequent requests are treated as unauthenticated.
- Session lifetime: 7 days rolling, via `MaxAge` on the `pgstore` session store.
- Session store: `pgstore` — a PostgreSQL-backed `gorilla/sessions` store. Sessions are kept in
  a `sessions` table managed by pgstore. Required for ADR 001: the session row is deleted on
  account anonymization, immediately revoking access regardless of cookie expiry.

---

## Password Hashing

Passwords are hashed with **Argon2id**.

- Parameters: use the library's recommended defaults unless performance tuning is required.
  Minimum: time cost ≥ 1, memory cost ≥ 64 MiB, parallelism ≥ 1.
- Passwords are never stored in plain text or logged.
- Passwords are never retrievable; only verification is supported.
- On account anonymization, the stored password hash is cleared (set to null).

---

## Token Generation

All application tokens (InviteToken, PasswordResetToken, anonymization token) are generated as:
- **Source:** CSPRNG (cryptographically secure random number generator)
- **Length:** 32 random bytes
- **Encoding:** base64url (URL-safe, no padding)

The anonymization token must not be derived from the account ID or any other stable identifier.

---

## CSRF Protection

All state-changing requests (POST, PUT, PATCH, DELETE) are protected via `gorilla/csrf`,
integrated in the Buffalo middleware stack. Plush templates expose the token via
`<%= authenticity_token %>` (Buffalo convention). Requests missing a valid token are rejected
with HTTP 403.

Combined with SameSite=Strict on the session cookie, this provides defense in depth.

---

## Public Routes

The following routes are accessible without authentication:

| Route | Purpose |
|-------|---------|
| `/politique-de-confidentialite` | Privacy notice (static page) |
| `/invite/:token` | Invite flow (account setup) |
| `/reinitialiser-mot-de-passe/:token` | Password reset |

All other routes require an active-account session. Unauthenticated requests to protected routes
are redirected to the login page.

The login page itself is public (no redirect loop).

---

## Role Checks

Two access levels exist:

| Level | Condition |
|-------|-----------|
| Authenticated | Session exists for an `active` account |
| Admin | Authenticated + a row exists in `account_roles` joining to `roles.name = 'admin'` for this account |

Admin-only routes and operations must verify both conditions. The check must be re-evaluated on
each request by querying `account_roles`; a cached claim from session creation is not sufficient
(roles can be revoked while a session is active). See ADR 005.
