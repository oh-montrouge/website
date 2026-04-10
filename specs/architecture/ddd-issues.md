# DDD Issues

Registry of domain modeling gaps identified through a DDD strategic and tactical review.
Covers both findings that are consciously deferred and those that warrant a decision before
implementation.

---

### ~~Active Account Does Not Equal Current Member~~

**Skill:** ddd-review
**Category:** BLIND SPOT
**Status:** RESOLVED — see `functional-adrs/002-active-account-rsvp-eligibility.md`.

`status = active` is the explicit RSVP eligibility criterion. Fee payment history has no
bearing. RSVP list hygiene for departed members is an administrative responsibility exercised
through anonymization. The suspended state is deferred to the "Improve account model" Later
item.

**Issue:** The RSVP spec creates records for every `active` account on event creation. But
`active` means "completed invite flow and not yet anonymized" — it says nothing about whether
the person is a current-season member. A musician who has not paid fees for three seasons and
has not been anonymized is treated identically to a current member: they receive RSVP records
for every new event, appear in every event's RSVP list, and are counted by any future
aggregate that filters on `status = active`.

The domain concept of **membership** (being part of the orchestra for a given season) is absent
from the model. `FeePayment` is the closest proxy, but it is a payment record, not a membership
record. The implicit assumption is that admins will promptly anonymize departed members. If they
do not, former members accumulate silently in every event's RSVP list with no system signal.

**Implication:** Event RSVP lists may include former members as long as their accounts remain
active. This is not explicitly decided anywhere in the specs — it is an undefended assumption in
the RSVP creation logic. If the answer is "yes, admins manage this via anonymization," that
decision should be stated. If the answer is "the RSVP list should reflect current membership,"
the RSVP creation logic needs a membership check.

**Current mitigation:** The Retention Review list surfaces accounts whose retention period has
elapsed, providing a path to anonymization. However, this is a 5-year GDPR trigger, not a
membership trigger — a musician could be a former member for 4 years and still appear in every
event's RSVP list.

**Future options:**
— Explicit domain decision: "active account = member for RSVP purposes."
— Add a seasonal membership check to RSVP creation (requires a `Membership` or equivalent concept).

---

### Account and Musician Are Conflated

**Skill:** ddd-review
**Category:** TENSION

**Issue:** The domain has two distinct things:
- An **auth credential** (email, password, status, roles, tokens): something that lets a person
  log in.
- A **musician record** (name, instrument, birth date, fee history): something that says a
  person was in the orchestra.

These live in a single `Account` entity. The anonymized state reveals the tension: after
anonymization, the row retains `main_instrument` and `status = anonymized` — it is no longer
an Account (cannot log in, holds no identity) but a **residual musician stub** kept for
aggregate statistics. `FeePayment` records post-anonymization are linked to an
`anonymization_token`, not a live account — the aggregate reference is broken by design.

**Implication:** Any feature that queries "who has played in this orchestra" must reason about
both `active` and `anonymized` accounts. The concept of musician identity leaks across the
anonymization boundary. This is contained in V1 but will create friction when later features
(statistics, blog authorship, Commission Artistique membership) need to distinguish "person who
can act" from "historical record of a person."

**Current mitigation:** The specs handle this consistently — anonymization fields and behaviors
are explicitly documented. The issue is conceptual clarity, not a functional bug.

**Future options:**
— If a statistics feature is built, consider naming the two concepts explicitly even if the
   schema does not separate them.

---

### ~~No Bounded Contexts Named~~

**Skill:** ddd-review
**Category:** BLIND SPOT
**Status:** RESOLVED — see `architecture/context-map.md`.

**Issue:** The information model is a single flat entity set that spans four conceptually
distinct sub-domains: Identity & Access (Account, tokens), Membership (musician profile, fee
history), Event Coordination (Event, RSVP, custom fields), and GDPR Compliance (consent,
anonymization, retention). These are not named or bounded anywhere in the spec set.

For a KISS V1 on a VPS, a single schema is appropriate. The gap is the absence of a map that
says which entities belong to which concern. The future-compatibility document covers schema
impact but not conceptual impact — it is not clear which later feature (blog, Commission
Artistique, assemblées générales) touches which sub-domain.

**Implication:** When later features are added, contributors will attach them to whichever
existing entity looks closest, without a principled boundary to guide the decision. This is how
accidental coupling accumulates.

**Current mitigation:** The spec INDEX provides a loose grouping by feature area (account
lifecycle, season management, events). This partially substitutes for a context map.

**Future options:**
— Add a context map section to the spec INDEX naming the four sub-domains and the entities
   that belong to each. No implementation change required.

---

### `other` Event Type Is an Embedded Subdomain

**Skill:** ddd-review
**Category:** STRESS POINT

**Issue:** The `EventField` / `EventFieldChoice` / `RsvpFieldResponse` pattern is an
Entity-Attribute-Value (EAV) form builder embedded within the RSVP domain. It is a
mini-subdomain with its own edit guards ("only if no responses recorded"), type integrity
concerns (already documented in `architectural-issues.md`), and cascading deletion rules. The
type-change matrix in the events spec has six asymmetric rules. This complexity signals that
`concert`, `rehearsal`, and `other` are not the same entity with a type flag but meaningfully
different domain objects that happen to share a table.

**Implication:** The EAV pattern grows with usage. Future `other` events with complex field
requirements accumulate application-layer validation to compensate for what the schema cannot
enforce. Each new code path that writes `RsvpFieldResponse` values is a new surface for the
integer/choice/text integrity gap noted in `architectural-issues.md`. The type-change complexity
will compound with new event types if they are added later.

**Current mitigation:** The `architectural-issues.md` entry on EAV field type integrity
captures the schema-level concern. The type-change matrix in the events spec is explicit and
testable.

**Future options:**
— If a third event type is ever added, treat this as a trigger to re-evaluate event type
   modeling.

---

### ~~FeePayment.Amount Has No Domain Type~~

**Skill:** ddd-review
**Category:** ASSUMPTION
**Status:** RESOLVED — see `functional-specs/00-information-model.md` § FeePayment.

Amount is EUR, two decimal places, ≥ 0, zero permitted (honorary membership or subsidised
fee). Both the functional information model and the technical data model (`NUMERIC(10,2)`)
are now consistent.

---

### ~~Cross-Aggregate Invariants Are Undeclared~~

**Skill:** ddd-review
**Category:** FRAGILITY
**Status:** RESOLVED — see `functional-specs/00-information-model.md` § Domain Invariants.

**Issue:** Four invariants span multiple records and are enforced by application code, but they
are not collected in one place:

| Invariant | Aggregate boundary crossed |
|-----------|---------------------------|
| At most one FeePayment per account per season | Account ↔ FeePayment |
| Exactly one RSVP per (account, event) pair | Account ↔ Event ↔ RSVP |
| Last-admin protection (≥ 1 active admin at all times) | All Account records |
| Exactly one current season at all times | All Season records |

Each is specified correctly in its own functional spec, but no document names them together as
cross-aggregate invariants requiring transactional enforcement.

**Implication:** New code paths (e.g., a bulk import, a batch operation, a future API endpoint)
are likely to miss enforcement of these invariants if contributors are not aware of them. The
risk is proportional to the number of entry points that write to these tables.

**Current mitigation:** Database constraints cover the RSVP and FeePayment uniqueness rules
at the schema level. Last-admin and current-season invariants are application-enforced with
explicit checks documented in the implementation notes.

**Future options:**
— Add a "Domain Invariants" subsection to the information model listing these four rules and
   their enforcement mechanism (DB constraint vs. application check). No implementation
   change required.

---

### `Chef d'orchestre` Is a Role, Not an Instrument

**Skill:** ddd-review
**Category:** TENSION

**Issue:** The instrument controlled list includes "Chef d'orchestre," inherited from OHM
Agenda. When a musician RSVPs `yes` to a concert event, the system asks "which instrument will
you play?" — and the conductor's answer is "Chef d'orchestre." This is a domain language
inconsistency: the list conflates **instruments** (Clarinette, Trompette) with **ensemble
roles** (Chef d'orchestre).

**Implication:** If a statistics page is built that shows instrument counts per concert (e.g.,
"8 clarinets, 4 trumpets"), `Chef d'orchestre` will appear in that list and require special
handling. The concept the controlled list actually models is closer to "part in the ensemble"
or "role at the event" than "musical instrument."

**Current mitigation:** Accepted for V1 as inherited from OHM Agenda. The instrument list is
fixed and has no admin UI, so the inconsistency cannot proliferate.

**Future options:**
— Rename the domain concept from "Instrument" to "Ensemble role" or "Part" when the
   statistics feature is defined. Scope impact: information model, functional specs, data
   model, and any stats queries that filter by instrument type.

---

### Privacy Notice Re-Acknowledgement Policy Is an Undocumented Decision

**Skill:** ddd-review
**Category:** BLIND SPOT

**Issue:** The privacy spec states: "No re-acknowledgement is required on subsequent logins or
when the privacy notice is updated." This is legally correct — the privacy notice is an Art. 13
information obligation, not a consent — but it is a deliberate domain decision with GDPR
accountability implications. If the privacy notice is materially updated (new data fields, new
purposes, new retention periods), musicians will not be re-informed through any system
mechanism. The association's compliance accountability for "members were informed of changes"
rests entirely on the initial invite-flow acknowledgement and whatever process the association
follows externally.

**Implication:** This decision is currently a sentence in the privacy spec, not a recorded
functional ADR with explicit rationale. If a future spec change introduces new personal data
categories or changes a processing purpose, there is no documented trigger to re-evaluate this
policy.

**Current mitigation:** The GDPR spec (`goals/gdpr.md`) accurately separates Art. 13 notice
from Art. 6(1)(a) consent. The legal framing is correct.

**Future options:**
— Promote this to a functional ADR recording: (a) that this is a deliberate decision,
   (b) the legal basis (Art. 13 is information, not consent), and (c) the trigger for
   revisiting it (material change to processing purposes or data categories).
