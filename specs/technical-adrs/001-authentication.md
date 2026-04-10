# ADR 001 — Authentication: Server-Side Sessions

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-11 |

## Context

The application requires authentication: only `active` accounts may access protected routes. Two
main approaches exist:

1. **Server-side sessions** — a session record is stored server-side; the client holds an opaque
   session ID in a cookie.
2. **Stateless tokens (JWT / PASETO)** — the client holds a signed token containing the session
   claims; the server verifies the signature on each request without a session store.

A constraint from the functional spec: when an account is anonymized, its login access must be
revoked immediately.

## Decision

**Server-side sessions with an HttpOnly + Secure + SameSite=Strict cookie.**

## Rationale

1. **Immediate revocation.** Anonymization must revoke login access on the spot. With server-side
   sessions, invalidating the session record is sufficient. With stateless tokens, a revocation
   list is required — equivalent complexity to a session store, with none of the simplicity
   benefit.

2. **No public API in V1.** Stateless tokens' main advantage is enabling stateless horizontal
   scaling and cross-origin API access. V1 has neither: the client is the server-rendered
   application, and there is no public API surface.

3. **Simpler implementation.** No key management, no token expiry/refresh logic, no clock-skew
   considerations.

## Consequences

- A session store is required. For a single-server OVH deployment, a DB-backed session store
  (same DB as the application) is sufficient. An in-memory store is acceptable for local
  development only.
- Session lifetime and expiry policy (rolling vs. absolute) are defined in the technical specs.
- Cookie flags: HttpOnly (no JS access), Secure (HTTPS only), SameSite=Strict (CSRF
  mitigation layer 1).

## Alternatives Considered

**JWT / PASETO:** Rejected. Adds key management and token refresh complexity. Immediate
revocation requires a revocation list, eliminating the stateless advantage. No V1 use case
justifies the trade-off.
