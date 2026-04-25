# OHM Website — Events and RSVP

> **Depends on:** [Information Model](00-information-model.md)

Event management (create, edit, delete) is admin-only. Viewing and responding to events is
available to all authenticated users.

---

## Creating an Event

Admin provides:
- Name
- Date and time
- Type: `concert` | `rehearsal` | `other`

On save, the system creates an RSVP record in the `unanswered` state for every `active` account.

---

## Editing an Event

Admin can edit: name, date/time, type.

**Effect of type change on existing RSVP records and fields:**

| From | To | Effect |
|------|----|--------|
| `rehearsal` | `concert` | All `yes` RSVP records reset to `unanswered` (instrument selection not collected). |
| `rehearsal` | `other` | No effect on RSVP records. No fields exist to delete. |
| `concert` | `rehearsal` | RSVP states retained. Instrument selections on `yes` records cleared. |
| `concert` | `other` | RSVP states retained. Instrument selections on `yes` records cleared. No fields exist yet. |
| `other` | `concert` | All custom fields (and their choices and any collected responses) are deleted. All `yes` RSVP records reset to `unanswered`. |
| `other` | `rehearsal` | All custom fields (and their choices and any collected responses) are deleted. RSVP states retained. |
| Any type | Same type | No effect on RSVP records or fields. |

Name and date/time changes have no effect on RSVP records or fields.

---

## Field Management for `other` Events

Admin can define custom fields on any `other`-type event to collect additional information from
musicians who RSVP `yes`.

**Adding a field:** Admin provides a label, field type (`choice`, `integer`, or `text`),
whether the field is required, and the display order. For `choice` fields, admin also provides
the list of selectable options (each with a label and display order).

**Editing a field:** Admin can change any field property. Only allowed if no responses have
been recorded for that field yet.

**Deleting a field:** Admin can delete a field and its choices. Only allowed if no responses
have been recorded for that field yet.

Fields can be added at any time. The edit/delete restriction protects data integrity once
responses exist.

---

## Deleting an Event

Admin can delete an event. All RSVP records for the event are deleted. Deletion is immediate and
irreversible.

Deleting events is the GDPR compliance path for RSVP records: once an event is deleted, its
attendance data is gone. There is no separate RSVP cleanup tool in V1.

---

## RSVP States

An RSVP record captures an account's intention for an event:

| State | Meaning |
|-------|---------|
| `unanswered` | No response yet |
| `yes` | Will attend |
| `no` | Will not attend |
| `maybe` | Uncertain |

Any authenticated user can set their own RSVP state to `yes`, `no`, or `maybe` at any time. They
can also change it at any time.

For `concert` events, a `yes` RSVP additionally requires an instrument selection (see below).

For `other` events with required custom fields, a `yes` RSVP must include responses to all
required fields before it can be saved.

---

## Instrument Selection for Concerts

When an authenticated user sets their RSVP to `yes` on a `concert` event:
- They are prompted to select which instrument they will play.
- Their main instrument is pre-selected.
- They may change the selection to any other instrument from the controlled list.

When a `yes` RSVP on a `concert` event is changed to `no` or `maybe`:
- The instrument selection is discarded.

---

## RSVP List Visibility

All authenticated users can view the full RSVP list for any event.

**For all event types:**
- Each account's name and RSVP state (`yes` / `no` / `maybe` / `unanswered`) is shown.
- Admin can modify RSVP state for other accounts.
- Filters: Filter by instrument played for this concert (eventType=concert only), Filter by RSVP state (oui / peut-être / non / sans rép.), Search by name, Quick chip: "Sans réponse" (highlight who hasn't answered)

**For `concert` events additionally:**
- Pupitre headcount table on the event page, above the musicians list. Rows per instrument × Présents / Peut-être / Absents / Sans rép., with a Total row.
- For each account with state `yes`: the selected instrument is shown alongside their name.

**For `other` events with custom fields:**
- For each account with state `yes`: their responses to all custom fields are shown alongside
  their name.

---

## Event List Views

### Dashboard (`/tableau-de-bord`)

Shows upcoming events only: events whose date/time is today or in the future
(`datetime >= CURRENT_DATE`). No past events. Events are ordered chronologically ascending.

For each event: name, date/time, type, and the viewer's own RSVP state (`unanswered` if no
response given). For events where the viewer has no RSVP record (events that predated their
account activation), no RSVP state is shown.

### Full Event List (`/evenements`)

Shows all events with no date filter. Events are ordered chronologically ascending.

Displays the same per-event information as the dashboard. For admin viewers, create/edit/delete
controls are shown inline (links to the admin forms).

---

## RSVP Record Eligibility

RSVP records are created proactively; there is no implicit "no record = unanswered" logic.

**Eligibility criterion:** `status = active`. Fee payment history has no bearing on RSVP
eligibility. An account that has not paid fees in one or more seasons but has not been
anonymized is treated as a member for RSVP purposes. RSVP list hygiene for departed members
is an administrative responsibility, exercised through the anonymization flow. See
[ADR 002](../functional-adrs/002-active-account-rsvp-eligibility.md).

**On event creation:** An RSVP record (state `unanswered`) is created for every `active` account.

**On account activation:** When a musician completes the invite flow and their account
transitions to `active`, RSVP records (state `unanswered`) are created for every event whose
date/time is strictly in the future at that moment.

Events whose date/time has already passed at the time of account activation receive no RSVP
record. The musician will not appear in the RSVP list for those events.

**On account anonymization:** All RSVP records belonging to the anonymized account are deleted
(see [ADR 001](../functional-adrs/001-rsvp-records-at-anonymization.md)).
