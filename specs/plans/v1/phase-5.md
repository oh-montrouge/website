# Phase 5 — Data Migration

**Status:** Not started

**Goal:** one-time migration of all data from OHM Agenda + Google Sheet into the new
schema. Run against the full schema once Phase 4.6 is complete.

---

## Prerequisites

Depends on: Phase 4.6 (full schema required)

---

## Specs

- `specs/technical-specs/07-data-migration.md` — full procedure, all 9 steps, verification
  queries, and post-migration checklist

---

## Architecture

Migration script only — no new services, repository interfaces, DTOs, or handler files.
No changes to the application code.

`webapp/architecture.md` does not need updating after this phase.

---

## Deliverables

- Pre-migration: encoding repair (mojibake → UTF-8), Google Sheet XLSX staging,
  referential integrity pre-check
- Migration steps 1–9: seasons, instruments, accounts (OHM Agenda + Google Sheet
  enrichment), roles, fee payments, events (with ID offsets), RSVPs, sequence reset
- Admin role assignments (manual list from stakeholders)
- Verification queries (all 8 checks)
- Post-migration admin checklist: designate current season, review flagged accounts,
  send invite links to all active accounts

---

## Acceptance Criteria

### Machine-verified

Run these after the migration script completes, against the production DB.
The full verification query set is in `specs/technical-specs/07-data-migration.md`.

**AC-M1 — All 8 verification queries pass**
Each query in the spec returns the expected result (counts match source data, no null
references, no orphaned rows).

**AC-M2 — No duplicate fee payments**
`SELECT COUNT(*) FROM fee_payments GROUP BY account_id, season_id HAVING COUNT(*) > 1`
returns zero rows.

**AC-M3 — Referential integrity holds**
All `main_instrument_id` values in `accounts` reference existing rows in `instruments`.
All `season_id` values in `fee_payments` reference existing rows in `seasons`.

**AC-M4 — Sequences are reset**
`INSERT` into `accounts`, `events`, `fee_payments` without specifying `id` → assigned ID
is higher than the highest migrated ID. Confirms sequence reset step ran correctly.

### Human-verified

**AC-H1 — A migrated admin account can log in**
Using credentials set during migration, log in on the production domain. Admin panel is
accessible. Musician list shows expected members from the original data.

**AC-H2 — Fee payment history is present**
Open a migrated musician's detail page. Fee payment records from the original Google
Sheet are present with correct seasons and amounts.

**AC-H3 — Post-migration checklist is complete**
Before declaring Phase 5 done: current season designated, flagged accounts reviewed,
invite links generated and sent to all active accounts.
