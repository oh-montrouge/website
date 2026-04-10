# ADR 002 — RSVP Eligibility: Active Account as Membership Proxy

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-15 |

## Context

The RSVP spec creates attendance records for every `active` account when an event is created.
`active` is an authentication state — it means "completed the invite flow and not yet
anonymized." It does not mean "paid fees this season" or "currently participating."

A musician who has not attended or paid in several seasons, whose account has not been
anonymized, is treated identically to a current member: they receive RSVP records for every
new event, appear in every event's RSVP list, and will be counted by any future aggregate
that filters on `status = active`.

The domain concept of **current membership** is absent from the model. This was identified as
a blind spot in `architecture/ddd-issues.md`.

The vision is explicit on a related constraint: *"There is no separate deactivation —
anonymization is the only mechanism for revoking login access from an active account."*
A suspended/deactivated state would require amending this clause — which is in scope if the
decision warrants it.

Two interpretations are possible:

1. **Active account = member for RSVP purposes.** The assumption is that admins will
   anonymize departed members in a timely manner. RSVP list quality is an administrative
   discipline, not a system-enforced invariant.

2. **RSVP eligibility requires a membership signal** beyond account status — a fee payment
   gate, an explicit flag, or a membership record.

## Decision

**Active account is the RSVP eligibility criterion. No additional membership gate is applied.**

The current RSVP creation logic (`active` accounts receive records on event creation;
newly activated accounts receive records for future events) is correct as specified.

This decision explicitly accepts that former members whose accounts have not yet been
anonymized will appear in RSVP lists. That is an administrative hygiene problem, not a
system design problem.

## Rationale

**A fee payment gate is fragile at season boundaries.** If RSVP eligibility required a fee
payment in the current season, no account would be eligible at the start of a new season
before fees are recorded. Widening to "current or previous season" introduces a sliding
window that must be maintained and reasoned about throughout the codebase.

**An explicit membership flag adds admin work with unclear benefit.** A separate
`is_active_member` flag would require admins to manage a new state transition in addition
to everything else. It duplicates the signal already carried by account status and fee
payment history without adding precision — it would still drift if admins forget to unset it.

**A suspended/deactivated state adds meaningful scope in V1.** A `suspended` state between
`active` and `anonymized` would require a vision amendment plus new lifecycle transitions,
auth middleware changes, admin UI, and answers to new questions (can suspended accounts be
unsuspended? do they count toward last-admin protection?). It is a valid future path, not a
V1 scope item.

**The Retention Review list is the intended hygiene mechanism.** Accounts whose retention
period has elapsed (5 years after last fee payment season end) are surfaced to admins for
anonymization. This is the system's existing path for clearing former members. RSVP list
quality is a downstream benefit of performing that review regularly.

**This is consistent with KISS.** The simplest model that works: if you have an active
account, you are in the orchestra until an admin anonymizes you.

## Consequences

- Active accounts that have not paid fees for one or more seasons will appear in RSVP lists.
  This is an accepted consequence. Admins control it by anonymizing departed members.
- The `05-events-and-rsvp.md` spec must be updated to state this assumption explicitly:
  RSVP eligibility is `status = active`; fee payment history has no bearing on eligibility.
- The `context-map.md` note "current member — implicitly: an active account with a recent
  fee payment" should be revised: the explicit definition is `status = active`, period.
- If RSVP list clutter from former members becomes a practical pain point post-launch,
  the right lever is a reminder mechanism or tighter cadence on the Retention Review list —
  not a new eligibility concept.

## Reversibility

### Before implementation

Free. Any option can be chosen by updating this ADR and the affected functional specs.
No code exists to change.

### After implementation

Reversing this decision requires schema and/or application changes of moderate scope.

| Target | Schema change | Application changes | Admin work |
|--------|---------------|---------------------|------------|
| Explicit membership flag | Add column to `accounts` | RSVP creation logic; admin UI for toggling flag | One-time: identify and flag former members among active accounts |
| Suspended state | Add status value to account lifecycle | Auth middleware, RSVP creation, account lifecycle transitions, admin UI | One-time: identify and suspend former members; vision amendment required |
| Fee payment gate | None | RSVP creation logic | Ongoing: handle season-boundary edge cases operationally |

RSVP records accumulated for former members under this decision are inert — they do not
corrupt account data or financial records. They persist until deleted but have no downstream
effect. If a future decision changes eligibility, those records can be cleaned up, but doing
so requires identifying which accounts were "former members" — a human judgment call per
account, not a scriptable migration.

### What accumulates over time

Nothing irreversible. The practical cost of reversal grows proportionally with the number
of accounts that need reclassification, not with time itself. For an association of ~30
active members with a handful of departed-but-not-anonymized accounts, this is a small
admin task.

### Cheapest reversal path

**Explicit membership flag** — single column addition, one RSVP creation logic change, no
new auth state. If the pain point is specifically RSVP list clutter and nothing else, this
is the minimum viable fix.

**Suspended state** — richer model (handles "on leave" members, not just former ones) but
higher implementation cost: new lifecycle state touching auth, admin UI, and the account
state machine. Worth the cost if the vision is being amended anyway for other reasons.

## Alternatives Considered

**Fee payment in current season (rejected):** Breaks at season start. Would prevent all
accounts from receiving RSVPs until the first payment of the new season is recorded.

**Fee payment in current or previous season (rejected):** More robust than above but
introduces a sliding window concept with edge cases (what if the previous season's end date
is ambiguous? what about the transition period?). Complexity not justified by the benefit.

**Explicit `is_active_member` flag (rejected):** New admin burden. Drifts if not maintained.
Duplicates signal already in the model without adding precision.

**Suspended/deactivated state (out of scope for V1):** Viable with a vision amendment.
Adds a genuine domain concept ("member on leave, data preserved") that the current model
cannot express. The implementation cost is higher than the other alternatives: new lifecycle
state, auth middleware update, admin UI, and answers to new operational questions
(unsuspend path, last-admin interaction). Deferred to the "Improve account model" Later item;
worth revisiting before that feature is scoped.
