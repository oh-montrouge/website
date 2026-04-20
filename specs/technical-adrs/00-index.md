# Technical ADRs — Index

This folder records decisions about **how the system is built**: stack, infrastructure, data storage, deployment, and technical patterns. Entries here are derived from the functional specs and the non-functional requirements in the vision.

A technical ADR answers: *"We chose this implementation approach, because..."*

A technical ADR may supersede a functional one when a technical constraint requires a behavioral trade-off. In that case, the superseded functional ADR is marked `superseded` and the affected functional spec section is updated before technical specs are written.

## Conventions

- File naming: `NNN-short-title.md` (e.g. `001-deployment-target.md`)
- Status values: `proposed` | `accepted` | `superseded` | `deprecated`
- One decision per file

## Decision Index

| ID | Title | Status |
|----|-------|--------|
| [001](001-authentication.md) | Authentication: Server-Side Sessions | Accepted |
| [002](002-datetime-storage.md) | Datetime Storage: UTC | Accepted |
| [003](003-ovh-deployment-target.md) | OVH Deployment Target | Accepted |
| [004](004-framework-and-language.md) | Framework and Language | Accepted |
| [005](005-role-model.md) | Role Model: Association Table over Boolean Flag | Accepted |
| [006](006-backup-strategy.md) | Database Backup Strategy | Accepted |
| [007](007-account-musician-dtos.md) | Two-DTO Pattern: AccountDTO and MusicianProfile | Accepted |
| [008](008-frontend-interactivity.md) | Frontend Interactivity: Alpine.js | Accepted |

## Open Decisions

*(none — all blocking technical decisions resolved)*
