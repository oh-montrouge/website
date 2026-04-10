# OHM Website — Privacy and Consent

> **Depends on:** [Information Model](00-information-model.md),
> [Account Lifecycle](01-account-lifecycle.md)
>
> **Authority:** GDPR requirements are documented in [`specs/goals/gdpr.md`](../goals/gdpr.md).
> This spec defines where and how those requirements surface in the application.

---

## Privacy Notice

A privacy notice (politique de confidentialité) is a static page bundled with the application.
Its required content is defined in `specs/goals/gdpr.md` §5. The content is fixed at deploy
time; there is no admin UI to edit it.

**Placement:**
- The invite-link account-setup form includes a checkbox requiring the musician to acknowledge
  the privacy notice before completing account setup. The checkbox links to the privacy notice
  page.
- A persistent link to the privacy notice is available in the logged-in interface (e.g., in the
  footer) for all authenticated users.
- The privacy notice page is also accessible to unauthenticated users (i.e., reachable via the
  invite-link form without being logged in).

**Acknowledgement:** Checking the checkbox during the invite flow constitutes a one-time
acknowledgement. No re-acknowledgement is required on subsequent logins or when the privacy
notice is updated.

---

## Phone/Address Consent

Consent for phone and address is collected during the invite flow via the combined
phone/address consent checkbox on the account-setup form.

- Consent covers both fields together. They cannot be consented to independently.
- Until consent is given, both fields are locked: the admin cannot fill them.
- Once consent is given (checkbox checked on the invite form), the admin can fill phone and
  address fields.
- Consent, once given, persists until explicitly withdrawn.

**Withdrawal:** A musician who wishes to withdraw consent contacts an admin. The admin performs
the "clear phone and address" operation (see [Musician Management](02-musician-management.md#consent-withdrawal-clearing-phone-and-address)).
Withdrawal clears both fields and the consent flag simultaneously.

**Notice on musician profile:** The musician's own profile page always displays the following
static notice: "Pour retirer votre consentement concernant téléphone et adresse,
contactez un administrateur." This notice is shown regardless of whether consent has been
given.

---

## Under-15 Parental Consent URI

The parental consent URI documents that parental consent has been obtained for a member who is
under 15 (Art. 8 GDPR). It is stored on the account record and is visible in the admin account
detail view. It is not displayed on the musician's own profile page.

How the parental consent is obtained (email, paper form, etc.) is a process decision for the
association; the system stores only the document URI.
