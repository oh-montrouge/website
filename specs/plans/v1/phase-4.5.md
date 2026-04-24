# Phase 4.5 — Fee Payments

**Status:** Complete

**Goal:** admins can record, edit, and delete fee payments per musician per season;
first inscription date is derived automatically.

---

## Prerequisites

Depends on: Phase 4.3 (musician accounts must exist) + Phase 4.4 (seasons must exist)

---

## Specs

- `specs/functional-specs/04-fee-payments.md`
- `specs/technical-specs/05-implementation-notes.md` § First Inscription Date
- `specs/technical-specs/04-routing.md` § Admin — Fee payments
- `specs/technical-specs/00-data-model.md` § fee_payments

---

## Design Reference

The fee payment section is embedded in `AdminMusicianDetailScreen` (`admin-musicians.jsx`),
not a standalone screen. Refer to the Phase 4.3 design reference for the full detail page
context.
Source: `specs/plans/v1/wireframes/wireframes/project/`

No Alpine.js usage beyond what Phase 4.3 introduces on the musician detail page.

---

## Architecture

See `webapp/architecture.md` for full detail.

**New handler:** `FeePaymentsHandler` in `actions/fee_payments.go`
- Depends on: `FeePaymentService`

**New service:** `FeePaymentService` in `services/fee_payment.go`

| Method | Notes |
|--------|-------|
| `Record` | Duplicate guard: reject if (account, season) pair already exists |
| `Update` | |
| `Delete` | |
| `ListByAccount` | Used on musician detail page |
| `GetFirstInscriptionDate` | Derived: `MIN(payment_date)`; returns nil if no payments |

**New repository interface:** `FeePaymentRepository` in `services/repositories.go`
- Methods: `Create`, `Update`, `Delete`, `ListByAccount`, `GetFirstInscriptionDate`
- Implemented by: `models.FeePaymentStore` (new file `models/fee_payment.go`)

**New DTOs** in `services/fee_payment.go`:
- `FeePaymentDTO` — ID, AccountID, SeasonID, SeasonLabel, PaymentDate, Amount

**`webapp/architecture.md` update required:** mark `FeePaymentService`,
`FeePaymentRepository`, `FeePaymentStore`, `FeePaymentsHandler`, `FeePaymentDTO`
as implemented.

---

## Migrations

- `fee_payments`

---

## Deliverables

- Record fee payment on musician detail page (duplicate guard: reject if (account, season)
  pair exists)
- Edit fee payment (`/admin/cotisations/{id}/modifier`)
- Delete fee payment
- First inscription date: derived on demand (`MIN(payment_date)`), shown in admin
  musician detail view
- Extend `db/dummy-data/` with fee payment seed data: payment records spread across
  accounts and seasons, including at least one account with no payments (retention case)

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — Duplicate payment is rejected**
POST `/admin/musiciens/{id}/cotisations` with a (account, season) pair → payment
recorded. POST again with the same pair → request rejected; only one payment row exists
for that (account, season) pair.

**AC-M2 — First inscription date is derived correctly**
`FeePaymentService.GetFirstInscriptionDate` for an account with payments on dates A, B,
C → returns the earliest of the three. For an account with no payments → returns nil.

**AC-M3 — Delete removes only the target payment**
DELETE `/admin/cotisations/{id}` → that payment row is gone; other payments for the same
account are unchanged.

### Human-verified

**AC-H1 — Fee payment section renders on musician detail**
Musician detail page shows the list of fee payments with amount, payment date, and season
label; edit and delete actions are present per row. First inscription date is shown when
at least one payment exists and absent when none exist.

**AC-H2 — First inscription date updates on earlier payment**
Record a payment for season A. Note the first inscription date. Record a second payment
for an earlier season B. Reload the musician detail — first inscription date updates to
the earlier date.
