# Adj-1 — Dashboard Redesign + Event Description

**Status:** Not started

**Goal:**
1. Give `/tableau-de-bord` its own card-based template (currently shares `events/index.plush.html` with `/evenements`).
2. Wire the existing `description` column (already in DB and `EventDetailDTO`) through the full stack: service `Create`/`Update` signatures, admin forms, dashboard cards, and event detail page. Render description as markdown.

---

## Prerequisites

**D1 — Markdown rendering library must be decided before implementation begins.**
See decision section below.

---

## Open Decision: D1 — Markdown Rendering Library

### Options

| Option | Pros | Cons |
|--------|------|------|
| **goldmark** | CommonMark-compliant; XSS-safe by default (unsafe HTML stripped unless explicitly enabled); actively maintained; used by Hugo, Go playground, pkg.go.dev | New dependency |
| blackfriday v2 | Established in Go ecosystem | Not CommonMark; maintenance uncertain since 2021 |

### Recommendation

**goldmark.** Safe by default (`html.WithXSSFilter()` or omitting `WithUnsafe()` strips
raw HTML from user input), actively maintained, industry-standard.

### Impact

- Add `github.com/yuin/goldmark` to `go.mod`.
- Register a template helper (e.g. `markdownToHTML(s string) template.HTML`) in
  `actions/render.go` or a new `actions/helpers.go`. The helper converts markdown to
  HTML using goldmark and returns a `template.HTML` value (so Buffalo/Plush won't
  double-escape it).
- Helper must call goldmark with **XSS-safe settings** — do not enable `WithUnsafe()`.
- Warrants **ADR-009** (first rich-content rendering dependency; XSS implications).

### Decision

**[ goldmark — accepted 2026-04-30. See ADR-009. ]**

---

## Current State (pre-implementation)

| Layer | State |
|-------|-------|
| DB migration | `description TEXT` column already exists (`20260426120000_create_events.up.sql`) |
| `EventDetailRow` model | `Description nulls.String` already present and queried |
| `EventDetailDTO` | `Description string` already present and populated in `GetDetail` |
| `EventSummaryDTO` | `Description` field **missing** |
| `EventListRow` model | `description` **not selected** in `ListUpcoming` / `ListAll` SQL queries |
| `EventStore.Create` | does **not** insert description |
| `EventStore.Update` | does **not** update description |
| `EventService.Create` / `Update` | signatures do **not** accept description |
| `EventManager` interface | `Create` / `Update` do **not** accept description |
| `EventsHandler.Create` / `Update` | do **not** read description from form POST |
| Admin forms (`new.plush.html`, `edit.plush.html`) | no description field |
| `events/show.plush.html` | description **not rendered** |
| `events/index.plush.html` | description **not shown** (list view — acceptable) |
| Dashboard template | does **not exist** — Dashboard handler renders `events/index.plush.html` |

---

## Tasks

### T1 — ADR-009: markdown rendering library ✅

Write `specs/technical-adrs/009-markdown-rendering.md` documenting the chosen library,
rationale, XSS safety stance, and the template helper contract.

**DoD:** ADR written

### T1.5 — Add markdown rendering library

Register the `markdownToHTML` helper in `actions/render.go` (or `actions/helpers.go` if
one exists). Return type must be `template.HTML` to prevent double-escaping in Plush.

**DoD:** follows `specs/technical-adrs/009-markdown-rendering.md`; helper compiles; `webapp/architecture.md` updated if a new file is
introduced;

### T2 — Model + repository: wire description through Create/Update/List

Touches: `webapp/models/event.go`, `webapp/services/repositories.go` (interface),
`webapp/services/event.go` (interface + implementation).

- `EventListRow`: add `Description nulls.String \`db:"description"\`` field.
- `EventStore.ListUpcoming` and `EventStore.ListAll`: include `e.description` in the
  `SELECT` clause.
- `EventStore.Create`: extend to accept `description string` (empty string = NULL in DB;
  use `nulls.NewString` / `nulls.String{}` accordingly); update the `INSERT`.
- `EventStore.Update`: extend to accept `description string`; update the `SET` clause.
- `EventRepository` interface in `services/repositories.go`: update `Create` and `Update`
  signatures to include `description string`.

**DoD:** model changes compile; SQL queries include description; interface signatures
updated; existing tests pass.

### T3 — Service + DTO: propagate description

Touches: `webapp/services/event.go`.

- `EventSummaryDTO`: add `Description string` field.
- `EventService.ListForMember` and `EventService.ListAll`: populate `Description` from
  `EventListRow.Description` (unwrap `nulls.String`; empty string if null).
- `EventService.Create` interface method and implementation: add `description string`
  parameter; pass through to `EventStore.Create`.
- `EventService.Update` interface method and implementation: add `description string`
  parameter; pass through to `EventStore.Update`.
- `EventManager` interface: update `Create` and `Update` signatures to match.

**DoD:** service compiles; `EventService` unit tests updated to exercise description
field; all tests pass.

### T4 — Handler: read description from form POST

Touches: `webapp/actions/events.go`.

- `EventsHandler.Create`: read `description` from `c.Param("description")`; pass to
  `h.Events.Create(...)`.
- `EventsHandler.Update`: same for the PUT handler.

**DoD:** handler compiles; action tests updated; all tests pass.

### T5 — Admin forms: description textarea

Touches: `webapp/templates/admin/events/new.plush.html`,
`webapp/templates/admin/events/edit.plush.html`.

- Add a `<textarea name="description">` field to both forms, following existing field
  patterns (label, optional hint about markdown support).
- In `edit.plush.html`, pre-populate with the current event description.
- Follow patterns in `webapp/FRONTEND.md`.

**DoD:** forms render without error; FRONTEND.md pre-ship checklist passes.

### T6 — Dashboard: dedicated card template

Touches: `webapp/templates/events/` (new file), `webapp/actions/events.go`.

- Create `webapp/templates/events/dashboard.plush.html` with a card layout. One card
  per event showing: name, date/time, event type badge, RSVP state, and description
  (markdown-rendered via helper; shown only if non-empty).
- Update `EventsHandler.Dashboard` to render `"events/dashboard.plush.html"` instead of
  `"events/index.plush.html"`.
- Keep `events/index.plush.html` unchanged (used by `/evenements`).
- Follow patterns in `webapp/FRONTEND.md`.

**DoD:** `/tableau-de-bord` renders card layout; `/evenements` renders table layout
unchanged; FRONTEND.md pre-ship checklist passes.

### T7 — Event detail: render description

Touches: `webapp/templates/events/show.plush.html`.

- Render `description` below the event header if non-empty, using the `markdownToHTML`
  helper. No change to layout if description is empty.

**DoD:** detail page renders markdown correctly for events with a description; no
regression for events without one.

---

## Architecture Notes

- Routes are already split at handler level — no `app.go` changes needed.
- `nulls.String` (`github.com/gobuffalo/nulls`) is the existing nullable string pattern
  in this codebase. Use it consistently in model layer; unwrap to `string` at service
  layer boundary.
- The `markdownToHTML` helper produces `template.HTML`; Plush will not escape it again.
  Keep the helper narrow: input is always user-supplied text, so XSS safety is not
  optional.
- `webapp/architecture.md` must be updated after T1 to note the markdown helper
  and after any structural changes introduced by this adjustment.

---

## Test Impact

| Task | Tests |
|------|-------|
| T1 | Update `EventStore` integration tests to assert description is stored/retrieved |
| T2 | Update `EventService` unit tests; add description to stub fixtures |
| T3 | Update action tests for Create/Update to include description param |
| T4 | No new tests — covered by action tests in T3 |
| T5 | No unit tests — covered by FRONTEND checklist (visual) |
| T6 | No unit tests — covered by FRONTEND checklist + AC-H1 |
| T7 | No unit tests — covered by AC-H2 |

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — Description persists through create**
`POST /admin/evenements` with `description=**hello**` → `GET /evenements/{id}` response
body contains `**hello**` (stored) and `<strong>hello</strong>` (rendered).

**AC-M2 — Description persists through update**
`PUT /admin/evenements/{id}` with a changed description → subsequent detail page shows
updated description.

**AC-M3 — Empty description is safe**
`POST /admin/evenements` without a description → event created successfully; detail page
does not error and does not render an empty description block.

**AC-M4 — Dashboard uses dedicated template**
`GET /tableau-de-bord` response body does not contain the table markup used by
`/evenements` (verify by checking for a landmark CSS class or element specific to each
template).

**AC-M5 — Events list unaffected**
`GET /evenements` continues to render correctly after the dashboard template split.

### Human-verified

**AC-H1 — Dashboard shows cards**
`/tableau-de-bord` renders one card per upcoming event. Each card shows name, date/time,
and RSVP state. No table or row layout.

**AC-H2 — Description renders as markdown**
An event with description `**bold** and _italic_` renders `<strong>bold</strong>` and
`<em>italic</em>` on both the dashboard card and the event detail page. Raw markdown
syntax is not shown to the user.

**AC-H3 — Description is editable by admin**
Admin opens event edit form → description textarea is present and pre-populated → saves →
updated description appears on the event detail page.
