# OHM Webapp — Component Architecture

Living document. Update this file when any structural change is made (new service, repository
interface, DTO, or handler file). See `CLAUDE.md` for the guardrail that requires it.

---

## Three-Layer Rule

```
actions/    HTTP orchestration only. Extract request data, call service(s), render.
            No business rules. Never imports models/.

services/   Domain logic. Business rules, invariant enforcement, domain operations.
            Depends on repository interfaces defined in this package. Never calls other services.
            No direct SQL.

models/     DB access only. Pop model structs, repository interface implementations.
            No business rules.
```

Dependency direction: `actions` → `services` → repository interfaces ← `models`.

---

## Bounded Contexts

Four conceptual sub-domains within the single deployment. Each context owns its vocabulary
and its service. See `specs/architecture/context-map.md` for full detail.

| Context | Service | DTO(s) | Owns |
|---------|---------|--------|------|
| Identity & Access | `AccountService` | `AccountDTO` | auth, credentials, tokens, roles, invite/reset flows |
| Membership | `MembershipService` | `MusicianProfile`, `MusicianProfileSummary`, `RetentionEntryDTO` | musician profile, seasons, fee payments |
| Event Coordination | `EventService` | `EventSummaryDTO`, `EventDetailDTO`, `RSVPRowDTO`, `EventFieldDTO` | events, RSVPs, custom fields |
| Compliance | `ComplianceService` | — | anonymization, consent withdrawal, processing restriction, retention review |

`SeasonService` and `FeePaymentService` belong to the Membership context.

---

## Services

### AccountService (`services/account.go`)

**Context:** Identity & Access

**Status:** Phase 2 complete. Phase 4.2 adds token operations.

| Method | Phase | Notes |
|--------|-------|-------|
| `Authenticate` | 2 | Argon2id verify; active-only; no account enumeration |
| `GetByID` | 2 | Returns `AccountDTO` |
| `IsAdmin` | 2 | Role check via `RoleRepository` |
| `CreateAdmin` | 2 | Grift use only; refuses if active admin exists |
| `ResetPassword` | 2 | Force-reset for grift use |
| `CreatePending` | 4.3 | Admin creates musician; status=pending; no password |
| `GenerateInviteToken` | 4.2 | CSPRNG token stored in `invite_tokens`; invalidates existing |
| `ValidateInviteToken` | 4.2 | Checks unused+unexpired+pending account |
| `CompleteInvite` | 4.2 | Activates account, sets password + consent, marks token used; gains EventRepository dep in 4.6 for RSVP seeding |
| `GeneratePasswordResetToken` | 4.2 | Active accounts only; invalidates existing |
| `ValidatePasswordResetToken` | 4.2 | Checks unused+unexpired+active account |
| `CompletePasswordReset` | 4.2 | Updates password, marks token used |

**Repository deps:** `AccountRepository`, `RoleRepository`, `InviteTokenRepository` (4.2),
`PasswordResetTokenRepository` (4.2), `EventRepository` (4.6 only, for RSVP seeding).

---

### MembershipService (`services/musician.go`)

**Context:** Membership

**Status:** Stub. Implement in Phase 4.3.

| Method | Phase | Notes |
|--------|-------|-------|
| `GetProfile` | 4.3 | |
| `SetInitialProfile` | 4.3 | Called by handler on musician creation (after AccountService.CreatePending) |
| `UpdateProfile` | 4.3 | Validates under-15 rule |
| `ListActive` | 4.3 | For musician list page |
| `ConsentWithdrawal` | 4.3 | Clears phone + address + consent flag |
| `ToggleProcessingRestriction` | 4.3 | |

**Repository deps:** `MembershipRepository`

---

### ComplianceService (`services/compliance.go`)

**Context:** Compliance (cross-cutting)

**Status:** Stub. Implement in Phase 4.3.

| Method | Phase | Notes |
|--------|-------|-------|
| `Anonymize` | 4.3 | Atomic: clear I&A fields + Membership fields + roles + tokens + sessions + RSVPs (4.6) |
| `RetentionReviewList` | 4.3 | Accounts past 5-year retention period |

**Repository deps:** `AccountRepository`, `MembershipRepository`, `RoleRepository`,
`InviteTokenRepository`, `PasswordResetTokenRepository`, `SessionRepository`,
`EventRepository` (4.6 only, for RSVP deletion).

**Rule:** ComplianceService is the only service that holds simultaneous dependencies on
both `AccountRepository` and `MembershipRepository`. No other service may do this.
See `specs/technical-adrs/007-account-musician-dtos.md`.

---

### SeasonService (`services/season.go`) — new in Phase 4.4

**Context:** Membership

| Method | Notes |
|--------|-------|
| `Create` | |
| `List` | |
| `DesignateCurrent` | Atomic swap; exactly-one invariant |

**Repository deps:** `SeasonRepository`

---

### FeePaymentService (`services/fee_payment.go`) — new in Phase 4.5

**Context:** Membership

| Method | Notes |
|--------|-------|
| `Record` | Duplicate guard: reject if (account, season) pair exists |
| `Update` | |
| `Delete` | |
| `ListByAccount` | |
| `GetFirstInscriptionDate` | Derived: MIN(payment_date) |

**Repository deps:** `FeePaymentRepository`

---

### EventService (`services/event.go`) — new in Phase 4.6

**Context:** Event Coordination. Owns both events and RSVPs (no separate RSVPService).

| Method | Notes |
|--------|-------|
| `ListForMember` | Past 30 days + upcoming, with viewer's own RSVP state |
| `GetDetail` | Full RSVP list + fields |
| `AdminList` | All events for admin view |
| `Create` | Bulk RSVP seed for all active accounts on save |
| `Update` | Type-change RSVP effects applied atomically |
| `Delete` | Cascades to RSVPs via DB FK |
| `UpdateRSVP` | Validates instrument for concerts; clears field responses on state change |
| `AddField` | `other` events only |
| `UpdateField` | Blocked if responses exist |
| `DeleteField` | Blocked if responses exist |
| `SeedRSVPsForAccount` | Called by AccountService.CompleteInvite (Phase 4.6) |

**Repository deps:** `EventRepository`, `RSVPRepository`

---

## Repository Interfaces (`services/repositories.go`)

All repository interfaces are defined in `services/` (consumer side, per DIP). `models/`
provides the production implementations. Tests inject stubs.

### Existing

| Interface | Implemented by |
|-----------|---------------|
| `InstrumentRepository` | `models.InstrumentStore` |
| `AccountRepository` | `models.AccountStore` |
| `SessionRepository` | `models.HTTPSessionStore` |
| `RoleRepository` | `models.AccountRoleStore` |
| `MembershipRepository` | `models.AccountStore` (Phase 4.3; same table, Membership methods) |

**`AccountRepository` additions (Phase 4.2/4.3):**
- `CreatePending(tx, email, instrumentID) → (int64, error)`
- `Activate(tx, id, passwordHash string, phoneAddressConsent bool) → error` — single SQL
  UPDATE that atomically sets `status='active'`, `password_hash`, `phone_address_consent`,
  and conditionally clears `phone`/`address`

**`MembershipRepository` (Phase 4.3):** Replace the `any` placeholder with methods:
`GetProfile`, `UpdateProfile`, `ListActive`, `ListForRetentionReview`, `ClearProfileFields`,
`SetConsent`, `ToggleProcessingRestriction`

### New interfaces

| Interface | Phase | Implemented by |
|-----------|-------|---------------|
| `InviteTokenRepository` | 4.2 | `models.InviteTokenStore` |
| `PasswordResetTokenRepository` | 4.2 | `models.PasswordResetTokenStore` |
| `SeasonRepository` | 4.4 | `models.SeasonStore` |
| `FeePaymentRepository` | 4.5 | `models.FeePaymentStore` |
| `EventRepository` | 4.6 | `models.EventStore` |
| `RSVPRepository` | 4.6 | `models.RSVPStore` |

---

## Handler Files (`actions/`)

One handler struct per file. Struct holds service interface dependencies. Route
registration in `app.go`.

| File | Phase | Handler struct | Service deps |
|------|-------|---------------|--------------|
| `auth.go` | 2 | `AuthHandler` | `AccountAuthenticator`, `SessionRepository` |
| `middleware.go` | 2 | — | `AccountAuthenticator` |
| `home.go` | 2/4.1 | `HomeHandler` | — |
| `tokens.go` | 4.2 | `TokensHandler` | `AccountService` |
| `musicians.go` | 4.3 | `MusiciansHandler` | `AccountService`, `MembershipService`, `ComplianceService`, `InstrumentRepository` |
| `profile.go` | 4.3 | `ProfileHandler` | `MembershipService` |
| `retention.go` | 4.3 | `RetentionHandler` | `ComplianceService` |
| `seasons.go` | 4.4 | `SeasonsHandler` | `SeasonService` |
| `fee_payments.go` | 4.5 | `FeePaymentsHandler` | `FeePaymentService` |
| `events.go` | 4.6 | `EventsHandler` | `EventService` |

---

## DTOs (`services/`)

Sensitive fields (`PasswordHash`, `AnonymizationToken`) are absent from all DTOs by
construction. See `specs/technical-adrs/007-account-musician-dtos.md`.

| DTO | File | Context | Fields |
|-----|------|---------|--------|
| `AccountDTO` | `account.go` | I&A | ID, Email, Status |
| `InviteTokenDTO` | `account.go` | I&A | Token, URL, ExpiresAt |
| `PasswordResetTokenDTO` | `account.go` | I&A | Token, URL, ExpiresAt |
| `MusicianProfile` | `musician.go` | Membership | AccountID + all profile fields |
| `MusicianProfileSummary` | `musician.go` | Membership | AccountID, Name, Instrument, Status |
| `RetentionEntryDTO` | `musician.go` | Membership | AccountID, Name, Instrument, LastSeasonLabel, LastSeasonEndDate |
| `SeasonDTO` | `season.go` | Membership | ID, Label, StartDate, EndDate, IsCurrent |
| `FeePaymentDTO` | `fee_payment.go` | Membership | ID, AccountID, SeasonID, SeasonLabel, PaymentDate, Amount |
| `EventSummaryDTO` | `event.go` | Event Coordination | ID, Name, EventType, Datetime, OwnRSVPState |
| `EventDetailDTO` | `event.go` | Event Coordination | + full RSVP list, custom fields |
| `RSVPRowDTO` | `event.go` | Event Coordination | AccountID, MusicianName, State, InstrumentName |
| `EventFieldDTO` | `event.go` | Event Coordination | ID, Label, FieldType, Required, Position, Choices |

---

## Template Structure (`templates/`)

Established in Phase 4.1. Convention: `templates/{area}/{action}.plush.html`.

```
templates/
  layouts/
    application.plush.html     ← auth-aware base layout (Phase 4.1)
  home/
    index.plush.html
  privacy/
    index.plush.html
  auth/
    login.plush.html           ← existing
  tokens/
    invite.plush.html
    invite_invalid.plush.html
    reset.plush.html
    reset_invalid.plush.html
  profile/
    show.plush.html
  events/
    index.plush.html
    show.plush.html
  admin/
    index.plush.html           ← existing
    musicians/
      index.plush.html
      new.plush.html
      show.plush.html
      edit.plush.html
    seasons/
      index.plush.html
    events/
      index.plush.html
      new.plush.html
      edit.plush.html
    retention/
      index.plush.html
```

---

## DTO Policy

Introduce a DTO when any of these apply:
- **Sensitive fields must be hidden** — `Account` always needs one (`password_hash` must never reach a template)
- **Multiple queries feed one view** — e.g. musician detail combining profile + fee payments + tokens
- **Display transformation needed** — e.g. UTC datetime → Europe/Paris; computed or formatted labels
- **Post-anonymization display** — `FeePayment` renders a name or an anonymization token depending on account state

Passing a model struct directly to a template is acceptable only when all hold: no sensitive fields, no display transformation, 1:1 with DB columns. `Instrument` (ID + Name, read-only reference data) is the canonical example.

---

## Key Structural Decisions

### 1 — Handler-level composition for multi-context writes

Operations that span I&A + Membership (e.g. musician creation) are sequenced at the
handler level as multiple service calls. All share the same Pop transaction provided by
middleware. Services never call each other.

Example — musician creation:
```
MusiciansHandler:
  1. AccountService.CreatePending(tx, email, instrumentID) → accountID
  2. MembershipService.SetInitialProfile(tx, accountID, ...)
  3. AccountService.GenerateInviteToken(tx, accountID) → token
```

### 2 — AccountRepository.Activate is a single SQL UPDATE

`CompleteInvite` must atomically set `status`, `password_hash`, `phone_address_consent`,
and conditionally clear `phone`/`address`. This lives in `AccountRepository.Activate`
as a single SQL UPDATE. AccountService does not gain a MembershipRepository dependency.
ComplianceService remains the only service that holds both AccountRepository and
MembershipRepository simultaneously (ADR 007).

### 3 — EventService owns events and RSVPs

RSVPs are tightly coupled to event lifecycle (bulk seeding on creation, state effects on
type change, deletion on anonymization). No RSVPService. EventService holds both
`EventRepository` and `RSVPRepository`.

### 4 — Services are the unit of context ownership

A service may only own one bounded context. Cross-context operations belong to either
ComplianceService (GDPR/lifecycle compliance) or are composed at the handler level.
No service-to-service calls.
