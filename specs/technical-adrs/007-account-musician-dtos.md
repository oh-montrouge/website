# ADR 007 — Two-DTO Pattern: AccountDTO and MusicianProfile

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-18 |

## Context

The `accounts` table is shared by two bounded contexts (see `architecture/context-map.md`):

- **Identity & Access** — `email`, `password_hash`, `status`, `anonymization_token`
- **Membership** — `first_name`, `last_name`, `main_instrument_id`, `birth_date`,
  `parental_consent_uri`, `phone`, `address`, `phone_address_consent`, `processing_restricted`

The initial implementation placed all fields in a single `AccountDTO`, causing the auth
middleware to carry Membership data it had no business knowing. It also created gravity toward
a god service: as Membership features are added, there would be pressure to extend
`AccountService` and `AccountDTO` rather than introduce a proper Membership service.

---

## Decision

Two separate DTOs, each scoped to its bounded context:

**`AccountDTO`** — Identity & Access view. Used by auth handlers and session middleware only.

```go
type AccountDTO struct {
    ID     int64
    Email  string
    Status AccountStatus
}
```

**`MusicianProfile`** — Membership view. Used by musician management handlers and the
musician's own profile page.

```go
type MusicianProfile struct {
    AccountID            int64
    FirstName            string
    LastName             string
    MainInstrumentID     int64
    BirthDate            *time.Time
    ParentalConsentURI   string
    Phone                string
    Address              string
    PhoneAddressConsent  bool
    ProcessingRestricted bool
}
```

Service ownership mirrors context ownership:

| Service | Context | Writes to |
|---------|---------|-----------|
| `AccountService` | Identity & Access | I&A fields; returns `AccountDTO` |
| `MembershipService` | Membership | Membership fields; returns `MusicianProfile` |
| `ComplianceService` | Compliance (cross-cutting) | Both contexts, atomically, via their repositories |

Screens that display both auth state and profile data (e.g. admin musician detail) call both
services at the handler level and compose at the template level. No merged DTO exists.

---

## Policy

**`AccountDTO` must not grow Membership fields.** Adding `FirstName`, `LastName`,
`MainInstrumentID`, or `ProcessingRestricted` to `AccountDTO` is a violation of this decision.
If a handler needs profile data, it calls `MembershipService`, not `AccountService`.

---

## Alternatives Considered

**A — Single `AccountDTO` with all fields (rejected)**

Simpler in the short term. Rejected because it collapses the I&A / Membership boundary at
the type level, making it impossible to maintain independent service ownership as features
are added. The auth middleware would receive profile data on every request with no mechanism
to prevent that coupling from spreading.

**B — Two DTOs, composition at handler level (chosen)**

The auth middleware remains narrowly scoped. Each service can evolve independently.
The boundary is enforced by the type system, not by convention.

---

## Consequences

- `RequireActiveAccount` middleware checks `AccountDTO.Status` — no profile data in session
  resolution path.
- `MembershipService` is introduced in Phase 4.3; `MembershipRepository` interface is defined
  as methods become known, not speculatively upfront.
- `ComplianceService` owns anonymization: clears I&A fields via `AccountRepository` and
  Membership fields via `MembershipRepository` in a single transaction. Neither
  `AccountService` nor `MembershipService` reaches into the other.
- When Phase 4.6 (Events) is added, `ComplianceService.Anonymize` gains an
  `EventRepository` dependency to delete RSVP records.
