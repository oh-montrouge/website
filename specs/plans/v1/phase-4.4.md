# Phase 4.4 — Season Management

**Status:** Not started

**Goal:** admins can create seasons and designate the current one without PhPMyAdmin access.

---

## Prerequisites

Depends on: Phase 2 (independent of 4.2 and 4.3; can be implemented in parallel)

---

## Specs

- `specs/functional-specs/03-season-management.md`
- `specs/technical-specs/05-implementation-notes.md` § Season Designation Transfer
- `specs/technical-specs/04-routing.md` § Admin — Seasons
- `specs/technical-specs/00-data-model.md` § seasons

---

## Design Reference

Wireframe screen: `AdminSeasonsScreen` (`admin-other.jsx`).
Source: `specs/plans/v1/wireframes/wireframes/project/`

**Alpine.js usage in this phase:**
- New season modal: `x-data="{ open: false }"` + `x-show="open"` on a `<dialog>` element;
  triggered by the "+ Nouvelle saison" button

---

## Architecture

See `webapp/architecture.md` for full detail.

**New handler:** `SeasonsHandler` in `actions/seasons.go`
- Depends on: `SeasonService`

**New service:** `SeasonService` in `services/season.go`

| Method | Notes |
|--------|-------|
| `Create` | |
| `List` | |
| `DesignateCurrent` | Atomic swap; enforces exactly-one-current invariant (see implementation notes) |

**New repository interface:** `SeasonRepository` in `services/repositories.go`
- Methods: `Create`, `List`, `DesignateCurrent`
- Implemented by: `models.SeasonStore` (new file `models/season.go`)

**New DTO:** `SeasonDTO` in `services/season.go` — ID, Label, StartDate, EndDate, IsCurrent

**New template:** `templates/admin/seasons/index.plush.html` (list + inline creation form)

**`webapp/architecture.md` update required:** mark `SeasonService`, `SeasonRepository`,
`SeasonStore`, `SeasonsHandler`, `SeasonDTO` as implemented.

---

## Migrations

- `seasons`

---

## Deliverables

- Season list (`/admin/saisons`) with inline creation form
- Create season (label, start date, end date; not automatically designated current)
- Designate current season (atomic swap transaction; exactly-one invariant maintained)
- Extend `db/dummy-data/` with season seed data: one current season and one past season

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — New season is not automatically current**
POST `/admin/saisons` with valid data → season created with `is_current=false`.

**AC-M2 — Designation sets exactly one current season**
POST `/admin/saisons/{id}/courante` → target season has `is_current=true`; a `SELECT
COUNT(*) FROM seasons WHERE is_current=true` returns `1`. All previously current seasons
have `is_current=false`.

**AC-M3 — Designation is idempotent**
POST `/admin/saisons/{id}/courante` twice on different seasons → only the second one is
current; count of current seasons is still `1`.

### Human-verified

**AC-H1 — Current season is visually distinguished**
The season list page clearly marks the current season (e.g. a badge or indicator).
Non-current seasons show a "designate as current" action; the current season does not.

**AC-H2 — Designation updates the UI immediately**
Designating a new current season redirects back to the season list; the previously
current season no longer shows the current indicator.
