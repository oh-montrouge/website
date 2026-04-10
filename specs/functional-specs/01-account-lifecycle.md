# OHM Website — Account Lifecycle

> **Depends on:** [Information Model](00-information-model.md)

## Account States and Transitions

```
[Admin creates account]
        |
        v
    PENDING ──── Admin deletes ──→ (removed)
        |
    Musician completes invite flow
        |
        v
    ACTIVE ──── Admin anonymizes ──→ ANONYMIZED (terminal)
```

- `pending` → `active`: musician completes the invite flow
- `active` → `anonymized`: admin anonymizes the account (see [Musician Management](02-musician-management.md))
- `pending` → deleted: admin deletes a pending account (see [Musician Management](02-musician-management.md))

There is no separate "deactivated" state. Anonymization is the only mechanism for revoking
login access from an active account.

---

## Invite Flow

1. Admin creates a musician account (see [Musician Management](02-musician-management.md)).
   The account is created in `pending` state.
2. On creation, the system generates an InviteToken. The corresponding invite URL is displayed
   in the admin UI for the admin to copy and send manually (e.g., from the association's email
   address).
3. The InviteToken expires 7 days after generation. If it expires unused, the account remains
   `pending`. The admin can generate a new InviteToken at any time from the account detail view;
   doing so invalidates any existing token for that account.
4. The musician follows the invite URL and is presented with a single account-setup form
   containing:
   - A password field
   - A privacy notice acknowledgement checkbox (with a link to the privacy notice page)
   - A combined phone/address consent checkbox (with a brief explanation of what consent covers)
5. The musician submits the form. The following occur atomically:
   - The account transitions to `active`
   - The password is set
   - The phone/address consent flag is set according to the checkbox state
   - The InviteToken is marked used
   - The musician is logged in
6. After the invite flow, the musician lands on their home page.

**Expired or already-used invite URL:** The system displays an informative message. The account
remains `pending`. The admin can generate a fresh invite token.

---

## Password Reset

1. Admin generates a PasswordResetToken for an `active` account from the account detail view.
   The corresponding reset URL is displayed in the admin UI for copying and sending manually.
2. The PasswordResetToken expires 7 days after generation. If it expires unused, the account is
   unaffected. The admin can generate a new reset token at any time; doing so invalidates any
   existing token for that account.
3. The musician follows the reset URL and is presented with a new password field.
4. The musician submits the new password. The account remains `active` with the updated
   password. The PasswordResetToken is marked used.

**Expired or already-used reset URL:** The system displays an informative message. The account
is unaffected. The admin generates a new reset token.

Password reset is not available for `pending` or `anonymized` accounts.

---

## Last-Admin Protection

The system must always have at least one `active` account with the admin flag set.

The following operations are blocked if they would result in zero active admin accounts:
- Removing the admin flag from an account
- Anonymizing an admin account
- Deleting a pending admin account

In each case, the system rejects the operation with an informative message.

---

## Sheet Music Access

When a Google Drive link is configured at deploy time, a "Partitions" menu item is shown to all
authenticated (logged-in) users. The item opens the configured Drive URL. If no link is
configured, the menu item is hidden.

There is no admin UI to manage this link; it is set at application configuration time.
