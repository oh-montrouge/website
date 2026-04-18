# Phase 4.1 — Public Pages + Navigation Shell

**Status:** Not started

**Goal:** static pages and the navigation structure that wraps all subsequent features.

---

## Prerequisites

Depends on: Phase 2

---

## Specs

- `specs/functional-specs/07-homepage.md`
- `specs/goals/gdpr.md` § 5 (privacy notice content)
- `specs/technical-specs/04-routing.md` § Public routes
- `specs/technical-specs/02-configuration.md` (`SHEET_MUSIC_URL`)

---

## Architecture

No new services, repository interfaces, or DTOs.

**Template structure to establish (see `webapp/architecture.md` § Template Structure):**
- `templates/layouts/application.plush.html` — auth-aware base layout used by all
  subsequent phases; nav shows login link when unauthenticated, full menu when authenticated
- `templates/home/index.plush.html`
- `templates/privacy/index.plush.html`

**`app.go` changes:**
- Add `GET /politique-de-confidentialite` to public routes
- Wire `SHEET_MUSIC_URL` env var; pass to render context for nav show/hide logic

`webapp/architecture.md` does not need updating after this phase.

---

## Deliverables

- Homepage template (placeholder copy; real copy provided by association before launch)
- Privacy notice template (content from `specs/goals/gdpr.md` § 5)
- Layout template with auth-aware navigation:
  - Unauthenticated: login link
  - Authenticated: events, profile, sheet music (if configured), logout
- Footer with persistent privacy notice link (all pages)
- `SHEET_MUSIC_URL` env var wired to show/hide "Partitions" menu item

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — Public routes require no authentication**
`GET /` and `GET /politique-de-confidentialite` return `200` without a session cookie.

**AC-M2 — Sheet music link is env-gated**
With `SHEET_MUSIC_URL` unset: authenticated response body does not contain "Partitions".
With `SHEET_MUSIC_URL` set to any URL: authenticated response body contains "Partitions"
and the configured URL.

**AC-M3 — Privacy notice link is present on all pages**
Response bodies for `/`, `/connexion`, and any authenticated page all contain a link to
`/politique-de-confidentialite`.

### Human-verified

**AC-H1 — Nav reflects auth state**
Unauthenticated: nav shows only the login link. Authenticated: nav shows events, profile,
logout (and sheet music link if configured). No authenticated-only items leak to
unauthenticated views.

**AC-H2 — Privacy notice is not blank**
`/politique-de-confidentialite` renders the GDPR content from `specs/goals/gdpr.md` § 5,
not a placeholder or empty template.
