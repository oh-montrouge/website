# OHM Website — Spec Index

Entry point for the specification tree. Read this first, then load only the files relevant
to your task.

---

## Goals

| File | Purpose |
|------|---------|
| [goals/vision.md](goals/vision.md) | Project vision, personas, V1 scope, milestones, requirements. **Start here** for full context. |
| [goals/gdpr.md](goals/gdpr.md) | GDPR obligations, applicable rights, retention rules, and compliance approach. |

---

## Functional Specs

Define **what** the system does. No implementation decisions.

| File | Purpose |
|------|---------|
| [functional-specs/00-information-model.md](functional-specs/00-information-model.md) | All entities, fields, and constraints. The canonical reference for domain concepts. Read before any other functional spec. |
| [functional-specs/01-account-lifecycle.md](functional-specs/01-account-lifecycle.md) | Account states, invite flow, password reset, login/logout, sheet music access, last-admin protection. |
| [functional-specs/02-musician-management.md](functional-specs/02-musician-management.md) | Creating, editing, viewing, anonymizing, and deleting accounts. Admin role grant/revoke. Consent withdrawal. Retention review. |
| [functional-specs/03-season-management.md](functional-specs/03-season-management.md) | Season creation, current-season designation, immutability rules. |
| [functional-specs/04-fee-payments.md](functional-specs/04-fee-payments.md) | Recording, editing, and deleting fee payments. First inscription date derivation. |
| [functional-specs/05-events-and-rsvp.md](functional-specs/05-events-and-rsvp.md) | Event CRUD, type-change effects, custom fields for `other` events, RSVP states, instrument selection for concerts, RSVP list visibility. |
| [functional-specs/06-privacy-and-consent.md](functional-specs/06-privacy-and-consent.md) | Privacy notice, consent collection, withdrawal, GDPR rights handling. |
| [functional-specs/07-homepage.md](functional-specs/07-homepage.md) | Public homepage: static content, navigation by auth state. |

### Functional ADRs

| File | Decision |
|------|---------|
| [functional-adrs/00-index.md](functional-adrs/00-index.md) | Index of all functional ADRs. |
| [functional-adrs/001-rsvp-records-at-anonymization.md](functional-adrs/001-rsvp-records-at-anonymization.md) | RSVP records are deleted (not anonymized) when an account is anonymized. |

---

## Technical Specs

Define **how** the system is built. Depend on functional specs.

| File | Purpose |
|------|---------|
| [technical-specs/00-data-model.md](technical-specs/00-data-model.md) | PostgreSQL schema: all tables, columns, types, indexes, and constraints. |
| [technical-specs/01-auth-and-security.md](technical-specs/01-auth-and-security.md) | Session-based auth, password hashing, CSRF, token generation, role checks. |
| [technical-specs/02-configuration.md](technical-specs/02-configuration.md) | Environment variables, bootstrap CLI (`seed-admin`), dev seed data, schema migration commands. |
| [technical-specs/03-stack.md](technical-specs/03-stack.md) | Runtime components, Docker, Mise tasks, Lefthook, local dev setup, build/deploy, backup, testing. |
| [technical-specs/04-routing.md](technical-specs/04-routing.md) | All HTTP routes grouped by access level (public / authenticated / admin), middleware stack, conventions. |
| [technical-specs/05-implementation-notes.md](technical-specs/05-implementation-notes.md) | Non-trivial implementation logic: SQL for invite flow, anonymization transaction, last-admin protection, event type changes, retention review, and more. |
| [technical-specs/06-ci-cd.md](technical-specs/06-ci-cd.md) | Pre-commit hooks (Lefthook), GitHub Actions CI jobs, manual deploy task, future CD option, LICENSE. |
| [technical-specs/07-data-migration.md](technical-specs/07-data-migration.md) | One-time migration from OHM Agenda (MySQL) and Google Sheets to the new PostgreSQL schema. Sources, field mappings, encoding repair, verification queries, post-migration checklist. |

### Technical ADRs

| File | Decision |
|------|---------|
| [technical-adrs/00-index.md](technical-adrs/00-index.md) | Index of all technical ADRs. |
| [technical-adrs/001-authentication.md](technical-adrs/001-authentication.md) | Server-side sessions (pgstore) over JWT. |
| [technical-adrs/002-datetime-storage.md](technical-adrs/002-datetime-storage.md) | Store all datetimes as UTC; display in Europe/Paris. |
| [technical-adrs/003-ovh-deployment-target.md](technical-adrs/003-ovh-deployment-target.md) | VPS + Docker Compose (not shared hosting or managed PaaS). |
| [technical-adrs/004-framework-and-language.md](technical-adrs/004-framework-and-language.md) | Go + Buffalo framework. |
| [technical-adrs/005-role-model.md](technical-adrs/005-role-model.md) | Roles table + association table instead of `is_admin` boolean. |
| [technical-adrs/006-backup-strategy.md](technical-adrs/006-backup-strategy.md) | Daily `pg_dump` to OVH Object Storage; restore procedure. |

---

## Architecture

| File | Purpose |
|------|---------|
| [architecture/architectural-issues.md](architecture/architectural-issues.md) | Known, consciously deferred architectural issues (systemic risks, fragilities, assumptions). Read before making structural changes. |
| [architecture/future-compatibility.md](architecture/future-compatibility.md) | Compatibility assessment for each Later feature: current fit, main architectural consideration, where the biggest change lands. |
