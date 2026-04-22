# Phase 4.6 — Events + RSVP

**Status:** Not started

**Goal:** event management with musician RSVPs; custom fields for open-form events.

---

## Prerequisites

Depends on: Phase 4.3 (accounts must exist for RSVP seeding)

---

## Specs

- `specs/functional-specs/05-events-and-rsvp.md`
- `specs/functional-adrs/001-rsvp-records-at-anonymization.md`
- `specs/functional-adrs/002-active-account-rsvp-eligibility.md`
- `specs/technical-specs/05-implementation-notes.md` § Bulk RSVP Creation, § Event Type
  Change, § Field Edit/Delete Guard, § RSVP Field Response Clearing
- `specs/technical-specs/04-routing.md` § Authenticated, § Admin — Events
- `specs/technical-specs/00-data-model.md` § events, § rsvps, § event_fields,
  § event_field_choices, § rsvp_field_responses

---

## Design Reference

Wireframe screens: `EventListScreen`, `EventDetailScreen` (concert / rehearsal / other
variants), `PupitreTable`, `AdminEventsScreen`, `AdminEventEditScreen`,
`AdminRetentionScreen` (`musician.jsx`, `admin-other.jsx`).
Source: `specs/plans/v1/wireframes/wireframes/project/`

**Alpine.js usage in this phase (most Alpine-intensive phase):**

| Behavior | Pattern |
|----------|---------|
| RSVP name search | `x-data` on the RSVP card; `x-show` on each row driven by `searchQ` string |
| RSVP filter chips (all / yes / no / maybe / unanswered) | `filterState` in same `x-data`; chip `@click` sets it |
| Pupitre click-to-filter (concerts only) | `filterInstrument` in same `x-data`; row `x-show` checks it |
| RSVP sidebar radio pills (own response) | `x-data="{ rsvp: '{{current_state}}' }"` seeded from server-rendered value |
| Admin inline RSVP edit | `x-data` holds per-row state seeded from server HTML; button `@click` fires `fetch()` `PATCH /evenements/:id/rsvp/:musician_id`; reverts local state on error |
| Event type change warning (admin edit form) | `x-data="{ type: '{{original_type}}' }"` on the form; `x-show` on the warning alert when current type differs from original |

**RSVP PATCH endpoint:** per ADR 008, this is the only non-HTML JSON endpoint in V1.
Returns `{ "ok": true }` on success. It is session-authenticated and CSRF-protected
identically to form POSTs; it is not a public API.

---

## Architecture

See `webapp/architecture.md` for full detail. Key points:

**New handler:** `EventsHandler` in `actions/events.go`
- Depends on: `EventService`
- Handles both authenticated (`/evenements/…`) and admin (`/admin/evenements/…`) routes

**New service:** `EventService` in `services/event.go`
- Owns both events and RSVPs — no separate `RSVPService`
- Depends on `EventRepository` + `RSVPRepository`

| Method | Notes |
|--------|-------|
| `ListForMember` | Past 30 days + upcoming; includes viewer's own RSVP state |
| `GetDetail` | Full RSVP list + custom fields |
| `AdminList` | All events for admin management view |
| `Create` | Bulk RSVP seed for all active accounts on save |
| `Update` | Type-change RSVP effects applied atomically (see implementation notes) |
| `Delete` | DB FK cascades to RSVPs |
| `UpdateRSVP` | Validates instrument for concerts; clears field responses on state change |
| `AddField` | `other` events only |
| `UpdateField` | Blocked if responses exist |
| `DeleteField` | Blocked if responses exist |
| `SeedRSVPsForAccount` | Called by `AccountService.CompleteInvite` (see below) |

**New repository interfaces** in `services/repositories.go`:
- `EventRepository` — CRUD + `ListActive` (for RSVP seeding) + list queries
- `RSVPRepository` — `Update`, `SeedForEvent`, `SeedForAccount`, `GetState`,
  `DeleteByAccount`, `ClearFieldResponses`
- Implemented by: `models.EventStore`, `models.RSVPStore`

**New DTOs** in `services/event.go`:
`EventSummaryDTO`, `EventDetailDTO`, `RSVPRowDTO`, `EventFieldDTO`

**Wire `EventRepository` into existing services (both require update):**

| Service | Change |
|---------|--------|
| `AccountService` | Add `Events EventRepository` field; call `Events.SeedRSVPsForAccount` at the end of `CompleteInvite` |
| `ComplianceService` | Add `Events EventRepository` field; call `Events.DeleteByAccount` inside `Anonymize` transaction |

**`app.go` update:** pass `models.EventStore{}` when constructing `AccountService` and
`ComplianceService`.

**New templates:** `templates/events/index.plush.html`, `templates/events/show.plush.html`,
`templates/admin/events/` — index, new, edit

**`webapp/architecture.md` update required:** mark `EventService`, `EventRepository`,
`RSVPRepository`, `EventStore`, `RSVPStore`, `EventsHandler`, all event DTOs as
implemented; update `AccountService` and `ComplianceService` to show `EventRepository`
dependency wired.

---

## Migrations

- `events`
- `rsvps`
- `event_fields`
- `event_field_choices`
- `rsvp_field_responses`

---

## Deliverables

- Event list for authenticated users (`/evenements`): past 30 days + all upcoming, own
  RSVP state shown
- Event detail page: pupitre headcounts; full RSVP list; for concerts, instrument per `yes` RSVP; for
  `other`, field responses per `yes` RSVP
- Admin are able to modify RSVP for other accounts
- RSVP update (`/evenements/{id}/rsvp`): yes/no/maybe; instrument selection for concerts
  (main instrument pre-selected); response clearing on transition away from yes
- Admin event management list (`/admin/evenements`)
- Create event (bulk RSVP creation for all active accounts on save)
- Edit event (name, date/time, type; type-change effects on RSVPs and fields per spec table)
- Delete event (cascades to all RSVPs)
- Custom fields for `other` events: add field (label, type, required, position; choices for
  `choice` type), edit field (blocked if responses exist), delete field (blocked if responses
  exist)
- RSVP on account activation: seed records for future events (inside invite flow transaction)
- Extend `db/dummy-data/` with event and RSVP seed data: one past and one upcoming event
  of each type, with varied RSVP states across accounts

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — Event creation seeds RSVPs for all active accounts**
POST `/admin/evenements` → one `rsvp` row with `state='unanswered'` exists for every
account with `status='active'`. No duplicates (`ON CONFLICT DO NOTHING` guard: creating
the same event twice does not produce double rows).

**AC-M2 — Type change rehearsal → concert resets yes RSVPs**
Edit an event from `rehearsal` to `concert` where some RSVPs have `state='yes'` →
those RSVPs are reset to `state='unanswered'` with `instrument_id=NULL`. RSVPs with
`state='no'` or `state='maybe'` are unchanged.

**AC-M3 — Type change other → concert deletes fields and resets yes RSVPs**
Edit an event from `other` to `concert` → all `event_fields` rows for that event are
deleted (cascades to choices and responses); `yes` RSVPs reset to `unanswered`.

**AC-M4 — RSVP state change from yes clears field responses**
POST `/evenements/{id}/rsvp` changing `state` from `yes` to `no` or `maybe` →
all `rsvp_field_responses` rows for that RSVP are deleted.

**AC-M5 — Field edit and delete blocked when responses exist**
`EventService.UpdateField` and `EventService.DeleteField` on a field that has at least
one `rsvp_field_response` → return an error; field is unchanged.

**AC-M6 — Invite completion seeds RSVPs for future events**
`AccountService.CompleteInvite` on an account where future events exist →
`rsvp` rows created for all those events; no rows for past events.

**AC-M7 — Anonymization deletes RSVPs**
`ComplianceService.Anonymize` → zero `rsvp` rows remain for the anonymized account.

### Human-verified

**AC-H1 — Event list shows own RSVP state**
The `/evenements` page shows each upcoming event with the viewer's current RSVP state
(unanswered / yes / no / maybe). Past events (older than 30 days) are not shown.

**AC-H2 — RSVP form is context-aware**
On a concert event detail page: RSVP form includes an instrument dropdown pre-selected
with the musician's main instrument. On a rehearsal: no instrument dropdown. On an
`other` event with custom fields: fields appear only when selecting `yes`.

**AC-H3 — Custom field lifecycle works end-to-end**
On an `other` event with no responses: add a field, edit it, delete it — all succeed.
After one musician RSVPs `yes` and fills in the field: edit and delete of that field are
blocked with an informative message.
