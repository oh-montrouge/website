# Context Map

Four sub-domains, one deployment. The contexts below are conceptual boundaries within the
monolith — they do not map to separate services or schemas. Their value is linguistic: each
context owns its vocabulary, and concepts that cross boundaries are named explicitly here.

---

## Contexts

### Identity & Access

Concerned with who can authenticate, what they are permitted to do, and how credentials
are established and revoked.

**Entities:** Account (auth fields: email, password, status, roles), InviteToken,
PasswordResetToken

**Language:**
- **Account** — a set of credentials and a permission set; the authentication identity
- **Active / Pending / Anonymized** — authentication states, not membership states
- **Admin role** — a permission grant on an account, not a separate account type
- **Invite flow** — the process by which a pending account acquires credentials
- **Last-admin protection** — the invariant that at least one active admin account must exist at all times

**Specs:** `functional-specs/01-account-lifecycle.md`

---

### Membership

Concerned with who is (or was) a member of the orchestra, what they play, and their
financial history with the association.

**Entities:** Musician profile fields on Account (name, instrument, birth date, phone,
address, consent), Season, FeePayment

**Language:**
- **Musician** — the person behind an account; characterized by their instrument and
  membership history
- **Season** — a named membership period; immutable after creation
- **Fee payment** — a single recorded contribution; the system's proxy for membership in a season
- **First inscription date** — derived from the earliest fee payment; represents tenure
- **Current member** — explicitly: an account with `status = active`. Fee payment history
  has no bearing on RSVP eligibility or membership status in the system. See
  [ADR 002](../functional-adrs/002-active-account-rsvp-eligibility.md).

**Specs:** `functional-specs/02-musician-management.md`,
`functional-specs/03-season-management.md`, `functional-specs/04-fee-payments.md`

---

### Event Coordination

Concerned with scheduling events, collecting attendance intentions, and knowing who will be
present with which instrument.

**Entities:** Event, RSVP, EventField, EventFieldChoice, RsvpFieldResponse

**Language:**
- **Event** — a scheduled activity: concert, rehearsal, or open-form (other)
- **RSVP** — an account's attendance intention: unanswered / yes / no / maybe
- **Instrument selection** — the instrument a musician will play at a specific concert;
  distinct from their main instrument
- **Custom field** — an ad-hoc data point collected from `yes` respondents on `other` events

**Specs:** `functional-specs/05-events-and-rsvp.md`

---

### Compliance

Concerned with the association's GDPR obligations: consent, erasure, restriction, and
retention. This context does not own domain entities; it applies policies to the other three
and is the authority on what data may be stored, for how long, and under which legal basis.

**Language:**
- **Consent** — freely given, specific, and withdrawable agreement to process optional data
- **Anonymization** — irreversible erasure of identifying fields; satisfies Art. 17 erasure requests
- **Anonymization token** — the opaque identifier that replaces a musician's name on fee
  payment records post-anonymization; a pseudonymization artifact, not true anonymization
  (see `architectural-issues.md`)
- **Retention period** — active membership + 5 years from the end of the last recorded fee
  payment season
- **Processing restriction** — Art. 18 flag; no operational effect in V1 beyond storage

**Specs:** `functional-specs/06-privacy-and-consent.md`, `goals/gdpr.md`

---

## Map

```
┌──────────────────────┐                    ┌──────────────────────┐
│  Identity & Access   │◄── same row ───────►│     Membership       │
│                      │   (tension: see     │                      │
│  Account (auth)      │    ddd-issues.md)   │  Musician profile    │
│  InviteToken         │                     │  Season              │
│  PasswordResetToken  │                     │  FeePayment          │
└──────────┬───────────┘                     └──────────┬───────────┘
           │                                            │
           │ account activation                         │ active accounts
           │ → create RSVPs for future events           │ → RSVP recipients
           │                                            │
           └──────────────────┬─────────────────────────┘
                              ▼
                 ┌────────────────────────┐
                 │   Event Coordination   │
                 │                        │
                 │  Event                 │
                 │  RSVP                  │
                 │  EventField            │
                 │  RsvpFieldResponse     │
                 └────────────────────────┘

  Compliance ──── policy authority over all three contexts ────►
```

---

## Relationship Notes

### Identity & Access ↔ Membership — same row, two concerns

These two contexts share a single `accounts` table row. They are not separated in the
implementation. The tension: an `active` account in the Identity & Access sense (can log in)
does not necessarily mean a current member in the Membership sense (paid fees this season).
The two concepts are conflated in the current model. See `ddd-issues.md`.

### Membership → Event Coordination — upstream dependency

Event Coordination depends on Membership to know who exists. On event creation, RSVP records
are seeded for every active account. The absence of an explicit membership concept means Event
Coordination uses `status = active` as a proxy for "should receive an RSVP," which may include
former members whose accounts have not yet been anonymized.

### Compliance — cross-cutting policy

Compliance does not drive features; it constrains them. The most structurally significant
Compliance operation is **anonymization**, which reaches across all three contexts in a single
atomic transaction:

| Context | Effect |
|---------|--------|
| Identity & Access | Email, password, roles, tokens cleared; status set to `anonymized` |
| Membership | Name, birth date, phone, address, consent cleared; fee payment account references replaced with anonymization token |
| Event Coordination | All RSVP records deleted |

---

## Where Later Features Land

| Feature | Primary context | Note |
|---------|----------------|-------|
| Blog / articles | Identity & Access (authorship) + new Content context | Needs a `content_editor` role; authorship references accounts |
| Statistics page | Membership (fee history) + Event Coordination (RSVPs) | Read-only aggregation; no new context |
| Improved account model (self-service) | Identity & Access | Requires email infrastructure; no other context affected |
| Trombinoscope | Membership + Compliance | Photo requires a new Compliance consent field |
| Commission Artistique | Identity & Access (new role) + new context TBD | Scope undefined |
| Assemblées Générales | Identity & Access (access control) + new Documents context | Purely additive |
| Sheet music search | Membership? + new Repertoire context | No existing context is a natural home |
