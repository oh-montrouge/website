# ADR 002 — Datetime Storage: UTC

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-11 |

## Context

Events have a date/time field. Account activation triggers creation of RSVP records for events
"strictly in the future." This comparison requires a consistent temporal reference.

The application is France-only (Europe/Paris timezone). France observes DST: UTC+1 (CET) in
winter, UTC+2 (CEST) in summer. DST transitions create one ambiguous local hour per year (the
"fall-back" hour, when the clock goes from 03:00 back to 02:00).

Two storage approaches:

1. **UTC everywhere** — store all datetimes as UTC; convert to/from Europe/Paris at the
   application boundary (input and display).
2. **Naive local time** — store datetimes without timezone info, implicitly assuming
   Europe/Paris.

## Decision

**Store all datetimes as UTC (timezone-aware). Accept and display in Europe/Paris.**

## Rationale

1. **DST safety.** A naive local time at 02:30 during a fall-back transition is ambiguous:
   it could be UTC+1 or UTC+2. UTC has no such ambiguity.

2. **Correct future-event comparison.** The "strictly in the future" check at account activation
   becomes `event_datetime_utc > now_utc` — always unambiguous.

3. **Hard to migrate.** Changing datetime storage semantics later requires migrating all datetime
   fields and auditing every comparison. The cost of getting this right now is low.

## Consequences

- All datetime columns are stored as timezone-aware UTC (e.g., `TIMESTAMPTZ` in PostgreSQL, or
  an explicit UTC convention in databases that lack native timezone support).
- The application converts Europe/Paris input to UTC on write, and UTC to Europe/Paris on read.
- The server clock is assumed to be UTC (standard for Linux servers).
- "Strictly in the future" is evaluated as `event_datetime_utc > utcnow()`.

## Alternatives Considered

**Naive Europe/Paris:** Rejected. Ambiguous at DST fall-back; requires discipline to never mix
timezones; harder to migrate later.
