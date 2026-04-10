# OHM Website — Fee Payments

> **Depends on:** [Information Model](00-information-model.md),
> [Season Management](03-season-management.md)

All operations in this spec are admin-only.

---

## Recording a Fee Payment

Admin can record a fee payment for any account (any status except `anonymized`) in any season.

Required fields:
- Season
- Amount
- Date (the date the payment was made)
- Type: `chèque` | `espèces` | `virement bancaire`

Optional:
- Comment

**Constraint:** At most one fee payment per account per season. If a payment already exists for
the selected account/season combination, the system rejects the save and prompts the admin to
edit the existing payment instead.

---

## Editing a Fee Payment

Admin can edit any fee payment. All fields (amount, date, type, comment) can be changed. The
account and season cannot be changed; to move a payment to a different season, delete it and
record a new one.

---

## Deleting a Fee Payment

Admin can delete any fee payment. Deletion is immediate.

If the deleted payment was the account's only fee payment, the first inscription date becomes
undefined for that account.

---

## First Inscription Date

The first inscription date for an account is the payment date of the account's earliest
FeePayment record, ordered by payment date. It is computed on demand and not stored.

If an account has no FeePayment records, the first inscription date is not displayed.

The first inscription date is shown in the admin account detail view. It is not shown on the
musician's own profile page in V1.

---

## Fee Payments After Anonymization

When an account is anonymized, its FeePayment records are not deleted. Instead, the account
name reference is replaced with the account's anonymization token (see
[Musician Management](02-musician-management.md#anonymizing-an-account)). Season, amount,
payment type, date, and comment are retained.
