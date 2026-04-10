# Architectural Issues

Curated registry of known, acknowledged architectural issues. Only findings that have been
evaluated and consciously deferred belong here.

---

### ~~Admin Lockout Recovery Gap~~

**Skill:** systemic-thinking
**Category:** FRAGILITY
**Status:** RESOLVED — see `specs/technical-specs/02-configuration.md` § Emergency Admin Recovery.

`mise run recover-admin` force-resets an active account's password directly in the DB via a
grift task. Requires SSH access to the VPS; does not require an active session.

---

### ~~No Database Backup Strategy~~

**Skill:** systemic-thinking
**Category:** FRAGILITY
**Status:** RESOLVED — see `specs/technical-adrs/006-backup-strategy.md`.

Daily `pg_dump` to OVH Object Storage (30-day rolling retention). Restore procedure
documented in the ADR.

---

### RSVP 2-Year Retention Has No Enforcement Mechanism

**Skill:** systemic-thinking
**Category:** TENSION

**Issue:** The spec names a 2-year post-event retention target for RSVP records and identifies
event deletion as the GDPR compliance path. There is no retention review list for events, no
reminder mechanism, and no automated cleanup. Compliance depends entirely on admins manually
deleting events more than two years old — an undocumented administrative discipline with no
system support. A separate cleanup tool is explicitly deferred to post-V1.

**Implication:** The system states a compliance requirement it has no mechanism to enforce,
making GDPR compliance for RSVP data dependent on human discipline rather than system behavior.

**Current mitigation:** Acknowledged in functional spec as a known V1 gap ("no separate RSVP
cleanup tool in V1").

**Future options:**
—

---

### ~~Data Migration Prerequisite Unspecified~~

**Skill:** systemic-thinking
**Category:** BLIND SPOT
**Status:** RESOLVED — see `specs/technical-specs/07-data-migration.md`.

The migration spec covers source mapping, conflict resolution, encoding repair, all
transformation decisions, verification queries, and the post-migration admin checklist.

---

### Anonymization Token Is a Permanent Quasi-Identifier

**Skill:** systemic-thinking
**Category:** ASSUMPTION

**Issue:** The anonymization token is described as "not derived from any stable identifier."
But once set, it is permanent, unique, and links all of a musician's historical fee payment
records to a single opaque value stored on the account row. If this token is exposed — through
a backup, a log, or an export — all payment records belonging to that person can be grouped
without knowing their name. The spec treats it as privacy-preserving because it replaces the
name; it is actually a pseudonymization artifact, not anonymization in the GDPR sense.

**Implication:** The system may not satisfy GDPR Art. 4(5)'s definition of anonymisation,
which requires that re-identification be irreversible; the token design preserves re-grouping
capability.

**Current mitigation:** Token is described as CSPRNG-generated and "not derived from stable
identifier," which reduces — but does not eliminate — re-identification risk.

**Future options:**
—

---

### Event Deletion Is Simultaneously Routine Management and GDPR Compliance Action

**Skill:** systemic-thinking
**Category:** LOAD-BEARING

**Issue:** Event deletion is both a routine administrative operation (correcting mistakes,
removing cancelled events) and the sole mechanism for GDPR-compliant disposal of RSVP
attendance data. These two motivations have incompatible risk profiles. No audit trail of
deletions exists, and there is no distinction in the data model between administrative and
compliance-motivated deletions. The confirmation step is the only safeguard.

**Implication:** An admin who deletes an event for operational reasons is simultaneously
performing a GDPR disposal action with no record that it occurred, making compliance audits
impossible.

**Current mitigation:** A confirmation step is required before deletion. Functional spec
explicitly identifies event deletion as the GDPR compliance path.

**Future options:**
—

---

### ADR 005 Role Extensibility Claim Is Schema-Only

**Skill:** systemic-thinking
**Category:** TENSION

**Issue:** ADR 005 justifies the roles table by claiming "extension cost for new roles is zero
schema work." The schema claim is accurate. The application claim is not: a third role requires
new middleware, a new route group, new UI for assignment, and changes to every handler that
checks permissions. The routing structure has exactly two access levels. The decision is
documented as providing extensibility, which may reduce future scrutiny of the architectural
work actually required when a second role is introduced.

**Implication:** Future roles will be treated as low-effort additions until implementation
begins, at which point the full application-layer scope will surface as unplanned work.

**Current mitigation:** ADR 005 acknowledges the KISS tension and notes that schema migration
is the hardest artifact to change post-launch.

**Future options:**
—

---

### Human Invite/Reset Workflow Depends on Undocumented Shared Credential

**Skill:** systemic-thinking
**Category:** BLIND SPOT

**Issue:** The invite and password reset flows require an admin to copy a generated link and
send it manually from "the association's email address." This is a human workflow with no
documentation anywhere in the specs. Who has access to that email address? What happens when
the person who knows the password is unavailable? The technical system is correctly designed
for no email infrastructure, but the operational dependency — a shared email credential held
by a person — is invisible in the spec set.

**Implication:** The invite flow's operational reliability depends on an undocumented human
process and shared credential that will degrade as the association's membership changes over
time.

**Current mitigation:** None documented.

**Future options:**
—

---

### ~~First Inscription Date Correctness Depends Entirely on Migration Completeness~~

**Skill:** systemic-thinking
**Category:** ASSUMPTION
**Status:** RESOLVED — see `specs/technical-specs/07-data-migration.md` Step 5b.

The migration spec uses the `Inscription` column from the Google Sheet seasonal exports to
create a guaranteed fee_payment record for each active musician's first recorded season. The
verification query in the migration spec asserts that every active account has at least one
fee_payment record before go-live.

---

### EAV Field Type Integrity Is Application-Enforced Only

**Skill:** systemic-thinking
**Category:** STRESS POINT

**Issue:** The `event_fields` table defines a `field_type` column (`choice | integer | text`)
and `rsvp_field_responses` stores all values as `TEXT`. Nothing at the schema level prevents
a `yes` RSVP with missing required fields, an integer field containing a non-numeric string,
or a choice field containing a value not in `event_field_choices`. Integer validation, choice
membership validation, and required-field enforcement all live in application code. Over time,
as the codebase evolves and edge cases accumulate, the gap between declared type and stored
value will widen.

**Implication:** The EAV model's flexibility creates a data integrity surface that grows with
every code path that writes responses, with no schema-level backstop.

**Current mitigation:** Application-layer validation at write time (specified but not
detailed).

**Future options:**
—
