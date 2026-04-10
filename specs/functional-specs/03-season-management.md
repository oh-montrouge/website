# OHM Website — Season Management

> **Depends on:** [Information Model](00-information-model.md)

All operations in this spec are admin-only.

---

## Creating a Season

Admin provides:
- Label (e.g., "2025-2026")
- Start date
- End date

The newly created season is not automatically designated current.

Seasons cannot be edited or deleted after creation. If a season is created with incorrect data,
a new season with correct data should be created and designated current.

---

## Designating the Current Season

Admin can designate any season as current.

When a season is designated current:
- The current designation is transferred from the previously current season to the newly
  designated one.
- The previously current season remains in the system and continues to accept fee payments.

**Invariant:** Exactly one season is designated current at all times. This designation cannot
be removed without simultaneously transferring it to another season.
