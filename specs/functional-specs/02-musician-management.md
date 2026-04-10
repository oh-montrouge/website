# OHM Website — Musician Management

> **Depends on:** [Information Model](00-information-model.md),
> [Account Lifecycle](01-account-lifecycle.md)

All operations in this spec are admin-only unless stated otherwise.

---

## Creating a Musician Account

Admin fills in:

| Field | Required at creation |
|-------|---------------------|
| First name | Yes |
| Last name | Yes |
| Email | Yes |
| Main instrument | Yes |
| Birth date | No |
| Parental consent URI | Conditional (required if birth date indicates under-15) |
| Phone | No — only fillable after consent is given |
| Address | No — only fillable after consent is given |

On save:
- The account is created in `pending` state.
- An InviteToken is generated automatically. The invite URL is displayed for the admin to copy
  and send.

Email must be unique among non-anonymized accounts. If the email is already in use, the system
rejects the save with an informative message.

---

## Editing a Musician Account

Admin can edit: first name, last name, email, main instrument, birth date, parental consent URI.

Admin can edit phone and address only after the musician has given phone/address consent (i.e.,
the consent flag is true).

The under-15 rule applies on every save: if birth date is set and indicates under-15, parental
consent URI is required.

Email uniqueness is enforced on edit (same rule as creation).

Editing is available for `pending` and `active` accounts. `anonymized` accounts cannot be edited.

---

## Viewing a Musician Account (Admin)

Admin sees all fields including:
- Parental consent URI (if set)
- Phone/address consent status
- Processing restricted flag status
- Account status (`pending` / `active` / `anonymized`)
- First inscription date (derived; shown if at least one FeePayment exists)

---

## Viewing Own Profile (Musician)

Active musicians can view their own profile. Displayed fields:
- First name, last name
- Email
- Main instrument
- Birth date (if set)
- Phone (if consent given and value is set)
- Address (if consent given and value is set)
- Static notice: "Pour retirer votre consentement concernant téléphone et adresse,
  contactez un administrateur."

Musicians cannot edit their own profile in V1. Profile changes are handled by an admin on
request.

---

## Granting and Revoking the Admin Role

Admin can grant the `admin` role to any `pending` or `active` account.
Admin can revoke the `admin` role from any `pending` or `active` account.

Both operations are subject to last-admin protection (see
[Account Lifecycle](01-account-lifecycle.md#last-admin-protection)): the operation is blocked if
it would leave zero `active` admin accounts.

The `admin` role cannot be granted to or revoked from an `anonymized` account (no effect and no
meaningful state).

---

## Deleting a Pending Account

Admin can delete a `pending` account. This removes the account record entirely.

Conditions: the account must be in `pending` state (invite flow not completed; no consent
recorded).

Deletion is subject to last-admin protection: blocked if the account is the last active-admin
candidate (see [Account Lifecycle](01-account-lifecycle.md#last-admin-protection)).

Deletion is immediate and irreversible.

---

## Anonymizing an Account

Anonymization is the mechanism for satisfying erasure requests (GDPR Art. 17). It is available
for `active` accounts.

A confirmation step is required before anonymization proceeds.

On confirmation, the following occur atomically:

**Fields cleared on the account:**
- First name, last name, email, password, birth date, phone, address, parental consent URI,
  phone/address consent flag, processing-restricted flag

**Roles removed:**
- All role assignments for this account are removed.

**Fields retained on the account:**
- Main instrument
- Status (transitions to `anonymized`)

**Anonymization token:**
- A new opaque token is generated (not derived from account ID or any stable identifier)
- The token is stored on the account record

**FeePayment records:**
- For each FeePayment belonging to this account: the account name reference is replaced with
  the anonymization token. Season, amount, payment type, date, and comment are retained.
  All of this account's fee payment records share the same token, so they can be counted as a
  unit in aggregates without re-identifying the person.

**RSVP records:**
- All RSVP records belonging to this account are deleted.
  (See [ADR 001](../functional-adrs/001-rsvp-records-at-anonymization.md) for rationale.)

**Tokens:**
- Any pending InviteToken or PasswordResetToken for this account is invalidated.

Anonymization is immediate and irreversible. The account transitions to the `anonymized`
terminal state and can no longer log in.

Anonymization is subject to last-admin protection: blocked if the account holds the admin flag
and is the last active admin (see [Account Lifecycle](01-account-lifecycle.md#last-admin-protection)).

---

## Consent Withdrawal (Clearing Phone and Address)

Admin can clear the phone and address fields of an account on the musician's behalf.

On clearing:
- Phone field is cleared
- Address field is cleared
- Phone/address consent flag is set to false

Both fields are always cleared together; they cannot be cleared independently.

This operation is available on `active` accounts where phone/address consent is currently true.

---

## Processing Restriction Flag (GDPR Art. 18)

Admin can toggle the processing-restricted flag on any `active` account.

In V1, toggling this flag has no operational effect beyond recording and displaying the flag
value in the admin account view.

---

## Retention Review List

Admin can view a list of accounts whose data retention period has elapsed.

An account appears in this list when **all** of the following are true:
1. The account has at least one FeePayment record.
2. The end date of the season of the account's most recent FeePayment is more than 5 years ago.
3. The account has not been anonymized (status is not `anonymized`).

The list displays, for each eligible account: name, main instrument, and the label of the
season of their most recent fee payment.

The system does not automatically anonymize accounts. The admin reviews the list and performs
anonymization manually for each account.

Accounts with no FeePayment records are not included in the retention review list.
