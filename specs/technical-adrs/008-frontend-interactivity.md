# ADR 008 — Frontend Interactivity: Alpine.js

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-20 |

## Context

The wireframes (`specs/plans/v1/wireframes/`) define five categories of client-side behavior
across the V1 screens:

| Behavior | Screen(s) |
|----------|-----------|
| Search + filter chips + pupitre click-to-filter | EventDetail |
| Admin inline RSVP state editing | EventDetail (admin only) |
| Modal dialogs (invite link, anonymize confirm, last-admin block) | AdminMusicianDetail |
| Disable submit until checkbox checked | InviteScreen |
| Mobile nav drawer | All authenticated pages |

The stack (ADR 004) is Go + Buffalo with Plush server-side templates. This ADR decides how
client-side interactivity is layered on top of server-rendered HTML.

---

## Constraints

No SPA framework. ADR 004 explicitly rejected React/Vue as frontend frameworks — the
rationale (no hydration overhead for CRUD screens, no second build pipeline, no CORS/token
complexity) holds equally here. The chosen approach must:

- Require no build step change (Buffalo's existing webpack pipeline or a plain `<script>` tag)
- Be invisible to the server: Plush templates remain the source of truth for markup
- Not introduce a JSON API surface (ADR 001: no public API in V1, one exception below)

---

## Alternatives Considered

### Vanilla JS

No dependencies. `<dialog>` handles modals natively. Event listeners and `fetch()` handle
everything else.

**Rejected because:** viable but verbose. Every show/hide, form-state, and modal behavior
requires explicit DOM manipulation. The RSVP filter UI alone (search + four chips + pupitre
click-to-filter) would produce 80–100 lines of boilerplate that Alpine reduces to a handful
of HTML attributes. The long-term maintenance cost is higher for zero technical advantage.

### htmx

Extends HTML with server-side interaction attributes (`hx-get`, `hx-patch`, `hx-target`).
Keeps all state on the server; every interaction is a server roundtrip that returns an HTML
fragment.

**Rejected because:** the behaviors in the wireframes split into two categories:

1. **Server-state changes** (RSVP inline edit) — htmx is a natural fit
2. **Pure UI state** (show/hide modal, disable button, filter visible rows) — htmx requires
   partial-template handlers for behaviors that carry no server state

Requiring a Go handler + Plush partial for "show the anonymize modal" is overhead without
payoff. Vanilla JS handles it in two lines; Alpine in one attribute. htmx adds server
roundtrips and handler surface where neither is needed.

htmx remains a valid option for V2 if real-time behaviors (live RSVP counts, push
notifications) are introduced — it is complementary to Alpine and the two can coexist.

---

## Decision

**Alpine.js** (15 KB min+gzip). Included via a single `<script defer>` tag in
`templates/layouts/application.plush.html`. No changes to the webpack pipeline. The library
is vendored to `public/assets/` rather than loaded from a CDN — this keeps the application
self-contained, consistent with the OVH single-server deployment model, and removes the
external dependency from every page load.

### Behavior mapping

| Behavior | Alpine pattern |
|----------|----------------|
| RSVP search + filter chips | `x-data` on the list container; `x-show` on each row driven by search string + active filter chip |
| Pupitre click-to-filter (concerts) | `x-data` shared with filter state; click sets `filterInstrument` |
| Admin inline RSVP edit | `x-data` holds row state seeded from server HTML; `@click` fires `fetch()` PATCH, updates state on success |
| Modal dialogs | `x-data="{ open: false }"` + `x-show="open"` on `<dialog>` element |
| Invite screen submit gate | `x-bind:disabled="!consentPrivacy"` on the submit button |
| Mobile nav drawer | `x-data="{ menuOpen: false }"` + `x-show="menuOpen"` on the nav drawer |

### RSVP filtering scope

The RSVP list is bounded by active musicians (~40 in the OHM context). Client-side filtering
of a pre-rendered list at this scale is instant and requires no server roundtrip. If the list
grew to hundreds of rows, the approach would revisit server-side filtering (a form GET with
query params). That threshold does not exist in V1.

### Admin RSVP inline edit

The one interaction that touches server state. Pattern:

1. Alpine `x-data` holds the row's current RSVP state, seeded from server-rendered attribute
2. Button `@click` fires `fetch('PATCH /events/:id/rsvp/:musician_id', { state: newState })`
3. On success: Alpine updates local state (button highlight, badge); on error: revert + show inline alert

This requires one Go handler per event for RSVP updates — already required by the functional
spec regardless of the JS approach. No partial-template infrastructure is added.

---

## Consequences

- `templates/layouts/application.plush.html` gains one `<script defer src="/assets/alpine.min.js">` tag
- `alpine.min.js` vendored to `public/assets/` (download from the Alpine.js GitHub release, pin the version)
- `specs/technical-specs/03-stack.md` updated to list Alpine.js
- The RSVP PATCH endpoint (`PATCH /events/:id/rsvp/:musician_id`) returns JSON `{ ok: true }` — the only non-HTML response in V1. It is not a public API; it is session-authenticated and CSRF-protected identically to form POSTs.
- No change to the test strategy: Alpine behavior is UI-layer only. The RSVP PATCH handler is tested as a standard action test.
