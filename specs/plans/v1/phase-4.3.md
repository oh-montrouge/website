# Phase 4.3 — Musician Management

**Status:** Not started

**Goal:** full admin control over musician accounts, plus the musician's own profile view.

---

## Prerequisites

Depends on: Phase 4.2

---

## Specs

- `specs/functional-specs/02-musician-management.md`
- `specs/functional-specs/06-privacy-and-consent.md`
- `specs/functional-specs/01-account-lifecycle.md` § Last-Admin Protection
- `specs/technical-specs/05-implementation-notes.md` § Anonymization, § Last-Admin
  Protection, § Retention Review List
- `specs/technical-specs/04-routing.md` § Admin — Accounts/musicians
- `specs/technical-specs/00-data-model.md` § accounts (full field set)
- `specs/architecture/context-map.md` § Membership, § Compliance

---

## Architecture

See `webapp/architecture.md` for full detail. Key points:

**`MusicianProfile` DTO, `MembershipService`, `ComplianceService`, and
`MembershipRepository` are pre-defined stub skeletons from the Phase 2 cleanup.
Implement their methods here; do not create new types.**

**New handlers:**
- `MusiciansHandler` in `actions/musicians.go` — depends on `AccountService`,
  `MembershipService`, `ComplianceService`, `InstrumentRepository`
- `ProfileHandler` in `actions/profile.go` — depends on `MembershipService`
- `RetentionHandler` in `actions/retention.go` — depends on `ComplianceService`

**Musician creation — handler-level composition (see `webapp/architecture.md`
§ Handler-level composition):**
```
MusiciansHandler.Create:
  1. AccountService.CreatePending(tx, email, instrumentID)  → accountID
  2. MembershipService.SetInitialProfile(tx, accountID, …)
  3. AccountService.GenerateInviteToken(tx, accountID)      → token (display URL)
```
All three calls share the same Pop transaction provided by middleware. No service calls
another service.

**MembershipService methods to implement** (`services/musician.go`):

| Method | Notes |
|--------|-------|
| `GetProfile` | |
| `SetInitialProfile` | Called on musician creation |
| `UpdateProfile` | Validates under-15 rule (birth date + parental consent URI) |
| `ListActive` | For musician list page |
| `ConsentWithdrawal` | Clears phone + address + consent flag |
| `ToggleProcessingRestriction` | |

**MembershipRepository methods to define** (replace `any` placeholder in
`services/repositories.go`; implemented by `models.AccountStore`):
`GetProfile`, `SetProfile`, `UpdateProfile`, `ListActive`, `ListForRetentionReview`,
`ClearProfileFields`, `SetConsent`, `ToggleProcessingRestriction`

**ComplianceService methods to implement** (`services/compliance.go`):

| Method | Notes |
|--------|-------|
| `Anonymize` | Atomic cross-context transaction: clear I&A + Membership fields, delete roles + tokens + sessions. Check last-admin protection before transaction. RSVP deletion wired in Phase 4.6. |
| `RetentionReviewList` | Accounts past 5-year retention period |

**AccountService additions:**
- `GrantAdmin(tx, accountID) → error` — last-admin protection not needed on grant
- `RevokeAdmin(tx, accountID) → error` — last-admin protection check before mutation
- `DeletePending(tx, accountID) → error` — rejects non-pending; last-admin protection if account holds admin role

**New templates:** `templates/admin/musicians/` — index, new, show, edit

**`webapp/architecture.md` update required:** mark `MembershipService`,
`ComplianceService`, `MembershipRepository`, `MusiciansHandler`, `ProfileHandler`,
`RetentionHandler` as implemented; add `AccountService` new methods.

---

## Deliverables

- Musician list (`/admin/musiciens`)
- Create musician account (with under-15 rule, invite token generated on save)
- Edit musician account (name, email, instrument, birth date, parental consent URI;
  phone/address gated on consent flag)
- Admin role grant / revoke (with last-admin protection check)
- Delete pending account (with last-admin protection check)
- Anonymization (atomic transaction; last-admin protection check before transaction)
- Consent withdrawal: clear phone, address, and consent flag together
- Toggle processing-restricted flag (display only in V1)
- Musician's own profile view (`/profil`): name, instrument, email, birth date if set,
  phone/address if consented, static consent-withdrawal notice
- Retention review list (`/admin/retention`): accounts whose last-payment season ended
  > 5 years ago
- `seed-dev` grift task (implement once schema is complete enough to generate realistic data)

---

## Acceptance Criteria

### Machine-verified

These are behavioral properties that must hold, beyond "tests pass." Each maps to at
least one integration or unit test.

**AC-M1 — Musician creation is atomic**
POST `/admin/musiciens` with valid data → 303 redirect; the DB contains a `pending`
account row, a profile row with the submitted fields, and exactly one unused
`invite_tokens` row for that account.

**AC-M2 — Under-15 rule is enforced**
POST `/admin/musiciens` with a birth date that makes the musician under 15 and no
`parental_consent_uri` → request rejected; no account created.

**AC-M3 — Last-admin protection on role revoke**
`AccountService.RevokeAdmin` called on the only active admin account → returns an error;
the `account_roles` row is unchanged.

**AC-M4 — Last-admin protection on anonymization**
`ComplianceService.Anonymize` called on the only active admin account → returns an error
before any DB mutation; account state is unchanged.

**AC-M5 — Anonymization is complete and atomic**
`ComplianceService.Anonymize` on a non-last-admin active account → in a single
transaction: `first_name`, `last_name`, `email`, `password_hash`, `birth_date`,
`phone`, `address`, `parental_consent_uri` all NULL; `status='anonymized'`;
`anonymization_token` set; zero rows in `account_roles` for that account; zero rows
in `http_sessions` for that account; all `invite_tokens` and `password_reset_tokens`
for that account marked `used=true`.

**AC-M6 — Consent withdrawal clears exactly the right fields**
`MembershipService.ConsentWithdrawal` → `phone=NULL`, `address=NULL`,
`phone_address_consent=false`; all other profile fields unchanged.

**AC-M7 — Profile page respects consent flag**
GET `/profil` for an account with `phone_address_consent=false` → response body contains
neither the phone nor the address value; for an account with `phone_address_consent=true`
and non-null phone → phone value is present.

### Human-verified

Performed by a contributor before marking the phase done.

**AC-H1 — Invite URL is usable**
Create a musician via the admin UI. The musician detail page displays an invite URL.
Follow that URL in a browser — the invite form renders with password and consent fields.

**AC-H2 — Musician detail page is complete**
Open any musician detail page. Verify all sections render: profile fields, invite/reset
token section with copyable URLs, GDPR section (consent status, processing restriction
toggle, anonymize button with confirmation step).

**AC-H3 — Anonymization is visible end-to-end**
Anonymize a musician through the admin UI. Confirm: the musician list no longer shows
their name; the detail page shows the anonymization token where the name was; attempting
to log in as that account fails.
