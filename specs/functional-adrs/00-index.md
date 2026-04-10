# Functional ADRs — Index

This folder records decisions about **what the system does**: user-visible behavior, business rules, information model, and user flows. Decisions here are independent of implementation — they do not prescribe a technical stack.

A functional ADR answers: *"We decided the system behaves this way, because..."*

Technical ADRs (in `specs/technical-adrs/`) may reference or supersede entries here when a technical constraint requires a functional trade-off. When that happens, the affected section of the functional spec must be updated before technical specs are written.

## Conventions

- File naming: `NNN-short-title.md` (e.g. `001-invite-link-flow.md`)
- Status values: `proposed` | `accepted` | `superseded` | `deprecated`
- One decision per file

## Decision Index

| ID | Title | Status |
|----|-------|--------|
| [001](001-rsvp-records-at-anonymization.md) | RSVP records at anonymization | Accepted |
| [002](002-active-account-rsvp-eligibility.md) | RSVP eligibility: active account as membership proxy | Accepted |

## Open Decisions

Decisions that must be resolved before functional specs can be written without guessing.

- *(none)*
