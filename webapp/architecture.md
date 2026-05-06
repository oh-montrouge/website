# OHM Webapp — Component Architecture

Living document. Update when any structural change is made (new service, repository
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

| Context | Service(s) | Owns |
|---------|-----------|------|
| Identity & Access | `AccountService` | auth, credentials, tokens, roles, invite/reset flows |
| Membership | `MembershipService`, `SeasonService`, `FeePaymentService` | musician profile, seasons, fee payments |
| Event Coordination | `EventService` | events, RSVPs, custom fields |
| Compliance | `ComplianceService` | anonymization, consent withdrawal, processing restriction, retention review |

`SeasonService` and `FeePaymentService` belong to the Membership context.

---

## Services

### AccountService (`services/account.go`) — Context: Identity & Access

**Repository deps:** `AccountRepository`, `RoleRepository`, `InviteTokenRepository`,
`PasswordResetTokenRepository`, `EventRepository` (RSVP seeding on invite completion only).

---

### MembershipService (`services/musician.go`) — Context: Membership

**Repository deps:** `MembershipRepository`

---

### ComplianceService (`services/compliance.go`) — Context: Compliance (cross-cutting)

**Repository deps:** `AccountRepository`, `MembershipRepository`, `RoleRepository`,
`InviteTokenRepository`, `PasswordResetTokenRepository`, `SessionRepository`, `EventRepository`.

**Rule:** ComplianceService is the only service permitted to hold simultaneous dependencies
on both `AccountRepository` and `MembershipRepository`. No other service may do this.
See `specs/technical-adrs/007-account-musician-dtos.md`.

---

### SeasonService (`services/season.go`) — Context: Membership

**Repository deps:** `SeasonRepository`

---

### FeePaymentService (`services/fee_payment.go`) — Context: Membership

**Repository deps:** `FeePaymentRepository`

---

### EventService (`services/event.go`) — Context: Event Coordination

Owns both events and RSVPs — no separate RSVPService. See Key Structural Decision #3.

**Repository deps:** `EventRepository`, `RSVPRepository`

---

## Repository Interfaces (`services/repositories.go`)

All repository interfaces are defined in `services/` (consumer side, per DIP). `models/`
provides the production implementations. Tests inject stubs.

| Interface | Implemented by |
|-----------|---------------|
| `InstrumentRepository` | `models.InstrumentStore` |
| `AccountRepository` | `models.AccountStore` |
| `SessionRepository` | `models.HTTPSessionStore` |
| `RoleRepository` | `models.AccountRoleStore` |
| `MembershipRepository` | `models.AccountStore` |
| `InviteTokenRepository` | `models.InviteTokenStore` |
| `PasswordResetTokenRepository` | `models.PasswordResetTokenStore` |
| `SeasonRepository` | `models.SeasonStore` |
| `FeePaymentRepository` | `models.FeePaymentStore` |
| `EventRepository` | `models.EventStore` |
| `RSVPRepository` | `models.RSVPStore` |

---

## Handler Files (`actions/`)

One handler struct per file. Struct holds service interface dependencies. Route
registration in `app.go`.

| File | Handler struct | Service deps |
|------|---------------|--------------|
| `auth.go` | `AuthHandler` | `AccountAuthenticator`, `SessionRepository` |
| `middleware.go` | — | `AccountAuthenticator` |
| `home.go` | `HomeHandler` | — |
| `tokens.go` | `TokensHandler` | `AccountTokenManager`, `SessionRepository` |
| `musicians.go` | `MusiciansHandler` | `AccountAdminManager`, `MusicianProfileManager`, `ComplianceManager`, `InstrumentRepository`, `FeePaymentManager`, `SeasonManager` |
| `profile.go` | `ProfileHandler` | `MusicianProfileManager` |
| `retention.go` | `RetentionHandler` | `ComplianceManager` |
| `seasons.go` | `SeasonsHandler` | `SeasonService` |
| `fee_payments.go` | `FeePaymentsHandler` | `FeePaymentService` |
| `events.go` | `EventsHandler` | `EventService`, `InstrumentRepository`, `MusicianProfileManager` |

---

## DTO Policy

Introduce a DTO when any of these apply:
- **Sensitive fields must be hidden** — `Account` always needs one (`password_hash` must never reach a template)
- **Multiple queries feed one view** — e.g. musician detail combining profile + fee payments + tokens
- **Display transformation needed** — e.g. UTC datetime → Europe/Paris; computed or formatted labels
- **Post-anonymization display** — `FeePayment` renders a name or an anonymization token depending on account state

Passing a model struct directly to a template is acceptable only when all hold: no sensitive
fields, no display transformation, 1:1 with DB columns. `Instrument` (ID + Name, read-only
reference data) is the canonical example.

Sensitive fields (`PasswordHash`, `AnonymizationToken`) are absent from all DTOs by
construction. See `specs/technical-adrs/007-account-musician-dtos.md`.

---

## Template Structure (`templates/`)

Convention: `templates/{area}/{action}.plush.html`.

```
templates/
  layouts/
    application.plush.html
  home/
  privacy/
  auth/
  tokens/
  profile/
  events/
    dashboard.plush.html    ← /tableau-de-bord card layout
    index.plush.html
    show.plush.html
  admin/
    musicians/
    seasons/
    events/
    retention/
    cotisations/
```

---

## Template Helpers (`actions/render.go`)

Helpers registered alongside the render engine init. Inspect `render.go` for the current list.

**XSS invariant:** `markdownToHTML` uses goldmark, never configured with `WithUnsafe()`.
Raw HTML in user input is stripped. Return type is `template.HTML` — Plush will not
double-escape it. See ADR-009.

---

## Key Structural Decisions

### 1 — Handler-level composition for multi-context writes

Operations spanning I&A + Membership (e.g. musician creation) are sequenced at the handler
level as multiple service calls. All share the same Pop transaction provided by middleware.
Services never call each other.

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
