# OHM Website — Vision

> **Purpose of this document:** This document expresses needs and constraints. It does not
> prescribe a technical stack. Stack proposals are to be derived from this vision by the
> architect, once the vision is stable.

The "Orchestre d'Harmonie de Montrouge" (OHM) is a French concert band, legally constituted as an association under the French law of 1901. This document describes the vision for its new website.

## Current state

### The current website

http://oh.montrouge.free.fr/

- Built as static HTML, decades ago
- Hosted on Free.fr — storage quota exceeded, migration is required
- Sheet music is uploaded manually by one person
- All musicians share a single login and password, giving access to personal data
- Likely not GDPR compliant
- Visually outdated

> **Open:** Source code location unknown.

### OHM Agenda

An internal tool ([ohm-agenda](https://www.ohm-agenda.ovh/)) used to manage the association's activity:

- Musicians subscribe to events
- Admins create events (concerts, rehearsals, other), manage musician records (name, instrument), and track annual fee payments per season

Season changes are manual: an admin must connect to [PhPMyAdmin](https://phpmyadmin.cluster021.hosting.ovh.net) and insert a row in the `season` table. The current season runs September–August; switching to January–December is under discussion and should be easy to support.

The database must not be lost. Migration is acceptable if the data model improves.

- DB schema: `current/ohm-agenda/db_schema.sql`
- DB export (git-ignored): `current/ohm-agenda/db_export.sql`

### Google Sheets

[Drive folder](https://drive.google.com/drive/u/1/folders/1S5GoohHv_6DakVC0fzNoqGaLWUjp1nFB) — export available in the git-ignored folder `current/gdrive-liste-musiciens`

A spreadsheet is maintained each season with musician information. It holds fields absent from OHM Agenda: phone number, address, job, birth date and place, nationality, and first inscription date. The schema is denormalized, making statistics unreliable and maintenance tedious. Fee payment is tracked here in parallel with OHM Agenda, duplicating work.

## Problem

OHM's operations depend on a handful of people, fragile tools, and a security model that cannot be defended.

The current website is maintained by a single person — if they leave, no one can update it. All musicians share one login and password to access a page that exposes personal data, creating both a security risk and a GDPR liability. Each season, admins must duplicate work across two tools, and creating a new season requires PhPMyAdmin access — a technical step most members cannot perform. Building statistics from a denormalized spreadsheet is painful, even though the underlying data could make it trivial.

The storage limit on the current host makes inaction impossible: staying would mean paying 10€/month to Free.fr for hosting already covered by OVH — money the association would rather spend on sheet music.

This project is primarily an **internal operations project**. The goal is to replace fragile, person-dependent processes with something any admin can run, that protects member data, and that consolidates the tools already in use into one place. A better public face is a secondary benefit.

## Principles

- **Keep it simple**: prefer the simplest solution that solves the problem.
- **Be lean**: defer decisions that have no present impact; do not design for hypothetical future needs.

## Personas

- **External users** — general public or potential new members
- **Musicians** — association members, regular performers
- **Admins** — association members with management privileges

## Milestones

### V1 — Replace the old tools

**Success condition:** OHM Agenda and the seasonal Google Sheet can be decommissioned, and Free.fr hosting can be abandoned.

**Account model:** Admins create musician accounts. The system generates a one-time invite link, displayed in the UI for the admin to copy and send manually (e.g. from the association's email address); the musician follows it to a single-page form presenting: a password field, a privacy notice acknowledgement checkbox, and a combined phone/address consent checkbox. Submitting the form completes account setup and grants access. Password reset is also admin-mediated: the admin generates a one-time reset link in the UI, displayed for copying and sending manually. Invite links and password-reset links expire after 7 days. An expired invite link leaves the account in a pending state (cannot log in) — the admin can generate a new invite link at any time. An expired password-reset link leaves the account unaffected — the admin generates a new reset link. No system email infrastructure required.

Phone and address share a single combined consent, given on the invite-link form. These fields are locked until consent is recorded; the admin can fill them only after consent is given. Birth date is optional. If provided and indicates the member is under 15, the parental consent URI field becomes required before the account can be saved. If birth date is left blank, no parental consent check is triggered. The system prevents removing the admin role from the last remaining admin account.

An account that has completed the invite flow is active. Account states: pending → active → anonymized (terminal). There is no separate deactivation — anonymization is the only mechanism for revoking login access.

In scope:
- Individual accounts — no more shared credentials; all accounts are musician accounts; the admin role is a permission that can be granted to any account
- Admin can grant or revoke the admin role on any account; the system prevents removing the admin role from the last remaining admin account
- Admin can create a season with a configurable date range and designate one as current — no more PhPMyAdmin; exactly one season is designated current at all times; the current designation cannot be removed without designating another season as current
- Admin can manage musicians; musician record fields: name, email, main instrument (from a controlled list sourced from OHM Agenda, plus "Chef d'orchestre"), birth date (optional; if provided and indicates under 15, a parental consent URI is required before the account can be saved), and phone/address (admin-editable only after the musician has given consent); parental consent URI, if set, is visible in the admin account detail view
- Admin can record a single fee payment per season for a musician, for any season (amount, date, type: chèque / espèces / virement bancaire; optional comment); payments can be edited or deleted; first inscription date is derived automatically from the musician's first recorded fee payment — no more Google Sheet duplication
- Admin can anonymize a musician account: all personal fields (name, email, birth date, phone, address, parental consent URI) are erased; main instrument is retained; fee payment records are anonymised (name replaced with a single opaque token generated at anonymization time and applied to all of that musician's payment records; season, amount, and payment type retained)
- Admin can delete a pending account (invite never accepted; no consent recorded)
- Admin can clear a musician's phone and address fields on the musician's behalf (consent withdrawal)
- Admin can create, edit, and delete events (name, date/time, type: concert / rehearsal / other — mirroring OHM Agenda); deleting an event removes all associated RSVP records (event deletion is the GDPR compliance path for RSVP records — RSVPs have a 2-year post-event retention; no separate cleanup tool in V1); new events start with all active accounts in an "unanswered" RSVP state; musicians can RSVP (yes / no / maybe) and change their RSVP at any time; all authenticated users can see the full RSVP list for an event (yes / no / maybe / unanswered; for concerts, the instrument of each musician who answered yes is also shown); when RSVPing yes to a concert, a musician selects which instrument they will play (main instrument pre-selected by default; changing a yes RSVP to no or maybe discards the selection); musicians who join after event creation start in the "unanswered" state for events that have not yet occurred — no more OHM Agenda
- Musician can view their own profile, showing: name, main instrument, email, birth date (if set), phone and address (if consented), and a static notice explaining how to withdraw consent for phone and address
- Musician can access sheet music via a single Google Drive link, configured at deploy time (no admin UI), displayed as a dedicated menu item to all authenticated users; the menu item is hidden if the link is not configured; access control is handled at the Google Drive level
- Admin can flag an account as "processing restricted" (GDPR Art. 18); has no operational effect in V1 beyond storage and display, and does not affect login
- Admin can view a list of accounts whose data retention period has elapsed
- Homepage: a public page presenting the association briefly, accessible to all visitors (authenticated and unauthenticated); content is static (bundled with the application)

#### Steps

- [x] Iterate on the vision until a fresh-context prompt passes the gate:
  > "Review `specs/goals/vision.md` as a functional specification writer. In the next steps, you'll have to propose functional solutions — not technical ones. Apply KISS — default to the simplest reasonable interpretation — before flagging. Only flag items where the simplest interpretation is itself ambiguous or where two equally simple interpretations contradict each other."
  - If you do any modification, ask the LLM to review the entire document for potentially introduced inconsistancies.
- [x] Write functional specs in `specs/functional-specs/`
  > "You are a functional specificator. Your role is to propose functional solutions — not technical ones. Write functional
    specs in `specs/functional-specs/`, applying KISS. When a genuinely blocking decision is encountered mid-writing, record
    it as a functional ADR in `specs/functional-adrs/` before proceeding. Ensure the work is not lost if you reach usage limit."
  - Gate: all blocking functional decisions recorded.
- [x] Iterate on the functional specs until a fresh-context prompt passes the gate:
  > "Review `specs/goals/vision.md` and `specs/functional-specs/` as a technical specification writer. In the next steps, you'll have to propose technical solutions. Apply KISS — default to the simplest reasonable interpretation — before flagging. Only flag items where the simplest interpretation is itself ambiguous or where two equally simple interpretations contradict each other."
- [x] Write technical specs in `specs/technical-specs/`
  > "You are a technical specificator. Your role is to propose technical solutions. Write technical
    specs in `specs/technical-specs/`, applying KISS. When a genuinely blocking decision is encountered mid-writing, record
    it as a technical ADR in `specs/technical-adrs/` before proceeding. If the decision is impactful, stop right away to initialize the ADR process.
    If a technical ADR supersedes a functional one, update the affected functional spec sections before proceeding.
    Ensure the work is not lost if you reach usage limit."
  - Gate: all blocking technical decisions recorded.
- [x] Do a final pass with the systemic-thinking skill.
- [x] Ask analysis of a DDD expert agent (should have been done before)
> "You are a Domain Driven Design expert. Can you analyze this project specs, especially the vision and the functional specifications, and challenge it regarding your expertise. Are there things we missed, in either DDD strategic or tactical thinking? What would be the impacts?"
- [x] Iterate until a fresh-context prompt passes the gate:
  > "Review `specs/functional-specs/` and `specs/technical-specs/` as if you were about to decompose them into implementation tasks. Apply KISS — default to the simplest reasonable interpretation — before flagging. Only flag items where the simplest interpretation is itself ambiguous or where two equally simple interpretations contradict each other."
- [x] Get stakeholder alignment.
- [x] Define a high-level roadmap?
- [ ] Break down into deliverable Epic Stories.
- [ ] Cycle:
  - [ ] Break down the next Epic Story into User Stories.
  - [ ] Implement the User Stories.
  - [ ] Review the vision and remaining Epic Stories.

### Later

- **Public face** — Blog and articles will live in the same application, so the architecture must avoid blog-hostile choices from day one (routing namespaces, schema extensibility — no blog features in V1). The timeline and the community manager role (Admin or a dedicated role) remain open decisions.
- **Statistics page** — Age pyramid per season, reflecting each member's age and membership status at the time
- **Improve account model** — In V1, account model is admin-heavy. We'll consider to improve it.
- **Sheets database and search engine** — In V1, only a Google Drive link. Later, we can consider an integrated experience. 
- **Trombinoscope** — The old website contain a trombinoscope. It can serve for newcomers but could cause GDPR concerns. We'll re-evaluate.
- **Commission artistique** — Some association members meet regularly to discuss around the next pieces to be played. We can consider tooling around it.
- **Assemblées générales** (and other legal stuff) — The association has bunch of legal documents that should legally be accessible by members. We can consider using the website for this.

## Requirements

### Functional
- Secure
- GDPR compliant (see [`specs/goals/gdpr.md`](gdpr.md))
- Accessible (WCAG A; Lighthouse accessibility score ≥ 80)
- In French — i18n not planned
- Data from old system should be migrated (see [specs/technical-specs/07-data-migration.md](07-data-migration.md))

### Technical
- DB schema must be versioned
- Must be testable locally
- Must be straightforward to deploy to OVH
- Developer tooling must allow bootstrapping the first system admin account
- No system email infrastructure — invite and password reset links are generated by the system and sent manually by an admin
- Privacy notice is a static page bundled with the application
