# Account / Musician Boundary — Tactical Remediation

**Context:** DDD analysis found that the type system does not yet reflect the Identity & Access /
Membership boundary defined in `architecture/context-map.md`. `AccountDTO` carries Membership
fields (`FirstName`, `LastName`, `MainInstrumentID`, `ProcessingRestricted`), the auth
middleware receives data it has no business knowing, and there is no established pattern for
where Membership types will live once Phase 4.3 begins.

The fix is tactical: no spec changes, no schema changes. The context map already drew the
boundary — this plan makes the code reflect it.

**Dependency:** Stage 1 is a prerequisite for Phase 4.3. Stage 2 is part of Phase 4.3.

---

## Stage 1 — Narrow existing types to I&A (before Phase 4.3)

**Goal:** `AccountDTO` and `AccountAuthenticator` represent Identity & Access only. Membership
fields have no place in the auth layer.

### 1.1 — Role name constant

- [x] Define `const RoleAdmin = "admin"` in `services/` (alongside `AccountStatus`)
- [x] Replace the three `"admin"` string literals in `services/account.go` with the constant

### 1.2 — Narrow `AccountDTO` to Identity & Access fields only

- [x] Remove `FirstName`, `LastName`, `MainInstrumentID`, `ProcessingRestricted` from `AccountDTO`
- [x] Update `toAccountDTO` mapper accordingly
- [x] Callers confirmed clean: middleware uses only `Status` and `ID`; auth handler uses only `ID`
- [x] `TestAuthenticate_Success` updated to reflect narrowed DTO

### 1.3 — Update `AccountAuthenticator` interface

- [x] No signature change required — narrowing `AccountDTO` in 1.2 is sufficient
- [x] Confirmed: no handler downstream of `RequireActiveAccount` reads Membership fields off `current_account`

---

## Stage 2 — Establish Membership types (part of Phase 4.3)

**Goal:** Musician profile data has its own DTO and its own service. The I&A / Membership
boundary survives implementation.

### 2.1 — `MusicianProfile` DTO

- [x] `MusicianProfile` defined in `services/musician.go`
- [ ] `MembershipRepository` interface methods: TODO phase-4.3 — add `GetProfile`,
      `UpdateProfile`, `ListActive`, `ListForRetentionReview`, `ClearProfileFields`
      as features are implemented
- [ ] `models.AccountStore` extended with `MembershipRepository` methods: TODO phase-4.3

### 2.2 — `MembershipService`

- [x] `MembershipService` struct defined in `services/musician.go`
- [x] `ComplianceService` struct defined in `services/compliance.go` — owns anonymization,
      consent withdrawal, and processing restriction flag
- [ ] `MembershipService` methods: TODO phase-4.3 — implement as musician management UI is built
- [ ] `ComplianceService.Anonymize`: TODO phase-4.3 — clear I&A fields via `AccountRepository`
      and Membership fields via `MembershipRepository` in one transaction
- [ ] `ComplianceService` gains `EventRepository` dependency for RSVP deletion: TODO phase-4.6

### 2.3 — Handler wiring

- [ ] Admin musician list + detail pages: use `MembershipService` — TODO phase-4.3
- [ ] Screens showing both auth state and profile: compose at handler level, not via merged DTO — TODO phase-4.3

### 2.4 — Technical ADR

- [x] `specs/technical-adrs/007-account-musician-dtos.md` written

---

## Field ownership reference

| Field | Context | DTO | Service |
|-------|---------|-----|---------|
| `id` | I&A | `AccountDTO` | `AccountService` |
| `email` | I&A | `AccountDTO` | `AccountService` |
| `password_hash` | I&A | — (never in DTO) | `AccountService` |
| `status` | I&A | `AccountDTO` | `AccountService` |
| `anonymization_token` | I&A / Compliance | — (never in DTO) | `AccountService` |
| `first_name`, `last_name` | Membership | `MusicianProfile` | `MembershipService` |
| `main_instrument_id` | Membership | `MusicianProfile` | `MembershipService` |
| `birth_date`, `parental_consent_uri` | Membership | `MusicianProfile` | `MembershipService` |
| `phone`, `address`, `phone_address_consent` | Membership | `MusicianProfile` | `MembershipService` |
| `processing_restricted` | Compliance (on Membership row) | `MusicianProfile` | `MembershipService` |
