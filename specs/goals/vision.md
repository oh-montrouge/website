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

**Account model:** Admins create and manage musician accounts via an admin-mediated invite and password-reset flow; no system email infrastructure is required. Accounts progress through three states: pending → active → anonymized (terminal). Anonymization is the sole mechanism for revoking access. See [`functional-specs/01-account-lifecycle.md`](../functional-specs/01-account-lifecycle.md).

In scope:
- Individual musician accounts; no shared credentials; the admin role is a permission, not a separate account type
- Admin role grant/revoke with last-admin protection
- Season management with a single designated current season; no PhPMyAdmin required
- Musician record management including GDPR-sensitive fields (phone, address, parental consent)
- Fee payment recording per season; first inscription date derived automatically — no more Google Sheet duplication
- Account anonymization (GDPR right to erasure) with fee payment pseudonymisation
- Consent withdrawal and deletion of pending accounts
- Event management (concert / rehearsal / other) with musician RSVPs; event deletion is the GDPR retention path — no more OHM Agenda
- Musician profile view
- Sheet music access via a configured Google Drive link
- GDPR processing restriction flag and data retention review list
- Public homepage with static content

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
- [x] Define a high-level roadmap — see `specs/plans/v1/v1.md`

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
- In French
- Data from old system should be migrated (see [specs/technical-specs/07-data-migration.md](07-data-migration.md))

### Technical
- DB schema must be versioned
- Must be testable locally
- Must be straightforward to deploy to OVH
- Developer tooling must allow bootstrapping the first system admin account
- No system email infrastructure — invite and password reset links are generated by the system and sent manually by an admin
- Privacy notice is a static page bundled with the application
