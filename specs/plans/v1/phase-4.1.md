# Phase 4.1 — Public Pages + Navigation Shell

**Status:** Complete

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

## Design Reference

Wireframe screens: `HomeScreen`, `LoginScreen`, `PrivacyScreen` (`public.jsx`); `AppBar`,
`Footer` (`shared.jsx`); design tokens (`styles/tokens.css`); component styles
(`styles/components.css`).
Source: `specs/plans/v1/wireframes/wireframes/project/`

**Alpine.js setup — establish here, used in all subsequent phases:**

This phase creates `templates/layouts/application.plush.html`. Vendor Alpine.js as part of
this phase:
1. Download `alpine.min.js` from the Alpine.js GitHub releases page
2. Place at `public/assets/alpine.min.js`
3. Add to the layout `<head>`:
   ```html
   <!-- alpine.js vX.Y.Z — update specs/technical-specs/03-stack.md when upgrading -->
   <script defer src="/assets/alpine.min.js"></script>
   ```
   Replace `vX.Y.Z` with the exact version downloaded. Update `03-stack.md` to record it.

**Alpine.js usage in this phase:**
- Mobile nav drawer: `x-data="{ menuOpen: false }"` on the `<header>`; `x-show="menuOpen"`
  on the drawer panel; `@click="menuOpen = !menuOpen"` on the `☰` button

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
- Alpine.js vendored to `public/assets/alpine.min.js`; version pinned in layout `<head>` comment
- Mobile nav drawer (hamburger toggles nav on small screens via Alpine)

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
