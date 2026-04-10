# OHM Website — Functional Information Model

> **Purpose:** Defines the entities, fields, and constraints that all other functional specs
> refer to. Does not prescribe a data-storage technology or schema.

## Entities

### Account

An Account represents a registered user. All accounts are musician accounts. The admin role is a
role assignment on an account, not a separate account type.

| Field | Required | Notes |
|-------|----------|-------|
| First name | Yes | Cleared on anonymization |
| Last name | Yes | Cleared on anonymization |
| Email | Yes | Unique among non-anonymized accounts. Used for login. Cleared on anonymization. |
| Password | Post-invite | Set by the musician during the invite flow. Never retrievable in plain text. Cleared on anonymization. |
| Main instrument | Yes | From the controlled instrument list. Retained on anonymization. |
| Birth date | No | If set and indicates under-15: parental consent URI is required before saving. Cleared on anonymization. |
| Parental consent URI | Conditional | Required if birth date indicates under-15. Cleared on anonymization. |
| Phone | No | Admin-editable only after the musician has given phone/address consent. Cleared on anonymization and on consent withdrawal. |
| Address | No | Admin-editable only after the musician has given phone/address consent. Cleared on anonymization and on consent withdrawal. |
| Phone/address consent | No | Set by the musician during the invite flow. Cleared on anonymization and on consent withdrawal. |
| Status | Yes | `pending` \| `active` \| `anonymized` |
| Roles | — | The set of roles granted to this account. May be empty. All roles are removed on anonymization. |
| Processing restricted | Yes | Default: false. GDPR Art. 18 flag. No operational effect in V1 beyond storage and display. Cleared on anonymization. |
| Anonymization token | Conditional | Opaque value generated at anonymization time. Not derived from any stable identifier. Applied to all FeePayment records for this account. |

**Account status values:**

| Status | Can log in | Terminal |
|--------|------------|---------|
| `pending` | No | No |
| `active` | Yes | No |
| `anonymized` | No | Yes |

**Role invariant:** At least one `active` account must hold the `admin` role at all times.
Enforced by last-admin protection checks before any operation that would remove the last admin.

**Under-15 rule:** If birth date is set and the calculated age at save time is under 15, the
parental consent URI field is required before the account can be saved.

---

### Instrument

A musical instrument from the controlled list. The list is sourced from OHM Agenda at migration,
plus "Chef d'orchestre". The list is fixed at application configuration time; there is no admin
UI to add or remove instruments in V1.

| Field | Notes |
|-------|-------|
| Name | Unique |

---

### Season

A named membership period with a configurable date range.

| Field | Required | Notes |
|-------|----------|-------|
| Label | Yes | e.g., "2025-2026" |
| Start date | Yes | — |
| End date | Yes | — |
| Is current | Yes | Exactly one season carries this flag at all times |

Seasons cannot be edited or deleted after creation.

---

### FeePayment

A single fee payment record associating an account with a season.

| Field | Required | Notes |
|-------|----------|-------|
| Account | Yes | On account anonymization: replaced by the account's anonymization token. |
| Season | Yes | — |
| Amount | Yes | — |
| Date | Yes | The date the payment was made |
| Type | Yes | `chèque` \| `espèces` \| `virement bancaire` |
| Comment | No | — |

**Constraint:** At most one FeePayment per account per season.

---

### Event

A scheduled event.

| Field | Required | Notes |
|-------|----------|-------|
| Name | Yes | — |
| Date/time | Yes | — |
| Type | Yes | `concert` \| `rehearsal` \| `other` |

---

### RSVP ("Répondez s'il vous plaît")

An account's attendance response for an event.

| Field | Required | Notes |
|-------|----------|-------|
| Account | Yes | — |
| Event | Yes | — |
| State | Yes | `unanswered` \| `yes` \| `no` \| `maybe` |
| Instrument | Conditional | Present only when state is `yes` and event type is `concert` |

**Constraint:** Exactly one RSVP record per (account, event) pair, for eligible accounts (see
[Events and RSVP spec](05-events-and-rsvp.md) for eligibility rules).

---

### EventField

A custom field defined by an admin for a specific `other`-type event. Determines what
additional information is collected from musicians who RSVP `yes`.

| Field | Required | Notes |
|-------|----------|-------|
| Event | Yes | Must be of type `other` |
| Label | Yes | Question text shown to the musician |
| Field type | Yes | `choice` \| `integer` \| `text` |
| Required | Yes | Whether a `yes` RSVP must provide a value |
| Position | Yes | Display order among the event's fields |

Fields can be added at any time before or after the event. A field can be edited or deleted
only if no responses have been recorded for it.

---

### EventFieldChoice

A selectable option for a `choice`-type EventField.

| Field | Required | Notes |
|-------|----------|-------|
| EventField | Yes | — |
| Label | Yes | Text shown to the musician |
| Position | Yes | Display order among the field's choices |

---

### RsvpFieldResponse

A musician's response to a single EventField, captured as part of a `yes` RSVP.

| Field | Required | Notes |
|-------|----------|-------|
| RSVP | Yes | — |
| EventField | Yes | — |
| Value | Yes | Stored as text regardless of field type |

**Constraint:** Exactly one response per (RSVP, EventField) pair. Responses are deleted when
the RSVP state changes from `yes` to `no` or `maybe`.

---

### InviteToken

A one-time token enabling a pending account holder to complete account setup.

| Field | Notes |
|-------|-------|
| Account | The pending account this token is for |
| Token | Opaque, random |
| Expires at | 7 days from generation |
| Used | Marked true when the invite flow is completed |

At most one active (unused, non-expired) invite token per account. Generating a new token
invalidates any existing one for that account.

---

### PasswordResetToken

A one-time token enabling an active account holder to reset their password.

| Field | Notes |
|-------|-------|
| Account | The account requesting the reset |
| Token | Opaque, random |
| Expires at | 7 days from generation |
| Used | Marked true when the password is successfully changed |

At most one active (unused, non-expired) password reset token per account. Generating a new
token invalidates any existing one for that account.

---

## Derived Values

### First inscription date

The payment date of the account's earliest FeePayment record, ordered by payment date. Computed
on demand; not stored. Undefined if the account has no FeePayment records.

---

## Application-wide Requirements

- **Language:** The application is in French. All UI text, error messages, and labels are in
  French. Internationalisation is not planned.
- **Accessibility:** WCAG 2.1 Level A; Lighthouse accessibility score ≥ 80.
- **Authentication:** Only `active` accounts can log in. `pending` and `anonymized` accounts
  cannot authenticate.
- **Bootstrap:** The system provides a developer-tooling mechanism for creating the first admin
  account outside the normal invite flow. This account is created in `active` state with the
  admin flag set. The mechanism is defined in the technical specs.
