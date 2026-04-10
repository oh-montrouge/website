# OHM Website — GDPR Requirements

> **Status:** Reviewed (2026-04-10).
> **Scope:** New OHM website handling personal data of musicians and administrators.
> **Authority:** CNIL (Commission Nationale de l'Informatique et des Libertés).

## 1. Controller identity

OHM is the **data controller** (responsable du traitement) under GDPR Art. 4(7). The controller's
identity and contact details must appear in the privacy notice.

No DPO is required: the association has no large-scale or systematic processing of special-category
data (Art. 37). A single internal contact for data subject requests is sufficient.

## 2. Data inventory

Every field must have a documented legal basis and retention period before it may be stored.

| Field | Source | Purpose | Proposed legal basis | Retention |
|-------|--------|---------|---------------------|-----------|
| First and last name | Registration | Identification, membership | Contract (Art. 6(1)(b)) | Active + 5 yr |
| Email address | Registration | Authentication, communication | Contract (Art. 6(1)(b)) | Active + 5 yr |
| Password (hashed) | Registration | Authentication | Contract (Art. 6(1)(b)) | Active + 5 yr |
| Instrument | Registration | Membership management | Contract (Art. 6(1)(b)) | Active + 5 yr |
| Fee payment per season | Admin | Association finances | Legitimate interest (Art. 6(1)(f)) | 5 yr then anonymise |
| Event RSVPs | Musician | Event management | Legitimate interest (Art. 6(1)(f)) | 2 yr after event |
| Phone number | Google Sheet | Communication | Consent (Art. 6(1)(a)) | Active + 5 yr |
| Address | Google Sheet | Administrative correspondence | Consent (Art. 6(1)(a)) | Active + 5 yr |
| Birth date (optional) | Google Sheet | Age-pyramid and diversity statistics; supports subsidy requests; under-15 parental consent gate | Legitimate interest (Art. 6(1)(f)) | Active + 5 yr |
| **Job** | Google Sheet | **Deferred — member-demographic statistics** | Legitimate interest (Art. 6(1)(f)) — collection decision deferred | — |
| **Birth place** | Google Sheet | **Deferred — member-demographic statistics** | Legitimate interest (Art. 6(1)(f)) — collection decision deferred | — |
| **Nationality** | Google Sheet | **Deferred — member-demographic statistics** | Legitimate interest (Art. 6(1)(f)) — collection decision deferred | — |

> **First inscription date is a derived value.** It is computed at query time from the musician's
> first recorded fee payment and is not stored as a separate field. No legal basis entry or
> retention period is required for it.

> **Data minimisation decision pending (Art. 5(1)(c)):** Job, birth place, and nationality are
> under consideration for member-demographic statistics (similar to the age pyramid). The legal
> basis would be legitimate interest — same framework as birth date. Whether to collect these
> fields at all is deferred to story breakdown. No migration until that decision is made.

### Retention notes

- "Active + 5 yr" in the table above means: retained while the musician has a fee payment on record, plus 5 years from the end of the season of their last recorded fee payment. This is a derived status, not an admin-set flag.
- 5 years after end of membership is standard practice for French associations; confirm with legal
  reviewer.
- Fee payment records: the 10-year accounting obligation (Code de commerce Art. L123-22) applies
  to the association's authoritative financial documents, held in its separate accounting system.
  Website records are secondary: anonymise after 5 years (replace name with an opaque token such
  as "Membre #4821"; retain season and payment status for historical continuity).
- **Anonymisation satisfies the retention obligation.** Once all identifying fields are removed,
  the remaining record is no longer personal data (GDPR Recital 26) and may be retained
  indefinitely. This is the appropriate approach for fee payment records after membership ends:
  keep the financial history, erase the identity.
- **Automatic deletion is not required.** A manual annual review by an admin — listing accounts
  whose retention period has elapsed — is compliant, provided it actually occurs. The system must
  make eligible accounts identifiable; it need not delete or anonymise them automatically.

## 3. Legal basis summary

| Basis | Applied to |
|-------|-----------|
| Contract (Art. 6(1)(b)) | Name, email, instrument, first inscription date |
| Legitimate interest (Art. 6(1)(f)) | Fee payment records (website records are secondary — the association's separate financial system holds authoritative accounting records), event RSVPs, statistics |
| Consent (Art. 6(1)(a)) | Phone and address, under one combined "extended profile" consent; must be freely given, specific, and withdrawable |

When consent is the legal basis, withdrawal must be as easy as giving it, and must not affect
membership access.

## 4. Data subject rights

Each right must be exercisable by the data subject directly, or on request to an admin.

### 4.1 Right of access (Art. 15)

A musician can see all personal data held about them. The controller must respond within one
month (Art. 12(3)); a self-service UI is not required for compliance.

**Requirements:**
- A "My Profile" page is a V1 product feature (independent of GDPR) and covers basic profile
  fields (name, instrument, status, etc.).
- Fee payment history and event RSVPs need not appear on that page in V1. An admin can compile
  and send the full dataset on request, in the same way as 4.4 (portability).

### 4.2 Right to rectification (Art. 16)

A musician can correct inaccurate data; an admin can correct any field.

**Requirement:** In V1, musicians cannot self-edit; profile changes are handled by an admin on request. Admins can edit all fields.

### 4.3 Right to erasure (Art. 17)

A musician can request erasure of their account. Erasure does not override legal-obligation
retention (fee records).

**Requirements:**
- An admin can anonymize a musician account. Anonymisation is the V1 mechanism for satisfying
  erasure requests — it is performed immediately on request, without waiting for the 5-year
  retention period to elapse.
- Anonymisation erases all personal fields (name, email, birth date, phone, address, parental
  consent URI); main instrument is retained (it is not personal data once de-linked from an
  identity, and is needed for future aggregate statistics). Fee payment records are anonymised:
  the musician's name is replaced with a single randomly generated opaque token (e.g. "Membre
  #4821"), the same token applied to all of that musician's payment records, so they can be
  counted as a unit in aggregates without re-identifying the person. The token must never be
  derived from a stable identifier (such as user ID). Season, amount, and payment type are
  retained. The exact format will be defined in the user story.
- The 10-year accounting obligation lives with the association's separate financial system.
- Anonymisation is a deliberate action with explicit confirmation.

### 4.4 Right to data portability (Art. 20)

A musician can receive their personal data in a machine-readable format (JSON or CSV) and
transmit it elsewhere. Applies to data processed on the basis of contract or consent.

**Requirement:** In V1, portability requests are fulfilled by an admin via direct database
access — no export UI is required. An email request fulfilled by an admin within one month
(Art. 12(3)) is compliant.

### 4.5 Right to restriction (Art. 18)

A musician can request that their data be retained but not actively processed (e.g., while a
rectification is under review).

**Requirement:** Admins can flag an account as "processing restricted." In V1, the flag has no
operational effect beyond being stored — exclusion from bulk operations will be enforced when
those features are built.

### 4.6 Right to object (Art. 21)

A musician can object to processing based on legitimate interest without losing membership access.

**Requirement:** Statistics-based opt-out will be addressed when the statistics feature is built (V2). In V1, birth date is collected but not yet used for statistics; no opt-out UI is required. The "processing restricted" flag (Art. 18) covers the V1 obligation for data subjects who raise concerns.

### 4.7 Right to withdraw consent (Art. 7(3))

A musician can withdraw consent for optional fields (phone, address) at any time.

**Requirement:** Consent for phone and address is collected during the invite-link flow (when the musician sets their password for the first time). To withdraw consent, a musician contacts an admin; the admin deletes both fields from the musician's record. This process must be documented in the privacy notice. Withdrawal must not affect membership access. The musician profile page displays a notice: "Pour retirer votre consentement concernant téléphone et adresse, contactez un administrateur."

## 5. Privacy notice

A privacy notice (politique de confidentialité) must be published on the website and presented
at registration. Required content under Art. 13:

- Identity and contact details of the data controller
- Purposes and legal basis per data category
- Recipients (internal only, unless otherwise decided — see open questions)
- Retention periods
- How to exercise data subject rights
- Right to lodge a complaint with CNIL (cnil.fr)
- Whether providing data is contractually required and the consequence of refusal

**Requirements:**
- The invite-link form is the first-login acknowledgement step: completing the form (including the privacy notice checkbox) constitutes first login and grants access. No re-acknowledgement is required on subsequent logins or when the privacy notice is updated.
- The logged-in interface must include a persistent link to the privacy notice (e.g., footer).

**Under-15 members:** OHM accepts members under 15. France sets the digital consent age at 15
(Art. 8 GDPR, loi informatique et libertés). Accounts are admin-created.

Birth date is optional. If provided and indicates the member is under 15, the parental consent
URI field becomes required before the account can be saved. If birth date is left blank, no
parental consent check is triggered — the admin implicitly asserts the member is ≥ 15.

How consent is obtained (email, paper form, etc.) is a process decision for the association; the system stores the document URI as a prerequisite to account creation.

## 6. Security measures (Art. 32)

GDPR requires appropriate technical and organisational measures. The following are GDPR-driven;
they complement, not replace, general security requirements.

| Measure | Notes |
|---------|-------|
| Individual accounts | Replaces shared credentials — already planned |
| Password hashing | Bcrypt or Argon2 |
| HTTPS everywhere | TLS in transit |
| Role-based access control | Musicians see their own personal data; musicians also see the full event RSVP list (all musicians with their state: yes / no / maybe / unanswered; for concerts, the instrument of each musician who answered yes is also shown); admins see all |

## 7. Data breach response (Art. 33–34)

- CNIL must be notified **within 72 hours** of becoming aware of a breach (Art. 33).
- If the breach is likely to result in high risk to individuals, affected data subjects must also
  be notified without undue delay (Art. 34).

**Process requirement (not a system feature):** The association must designate an internal
contact responsible for breach detection and CNIL notification. A response checklist should
be documented separately and rehearsed.

## 8. Record of processing activities (ROPA)

Not mandatory below 250 employees (Art. 30(5)). This document serves as OHM's ROPA baseline;
update it whenever data fields are added or removed.
