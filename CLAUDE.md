# Project Guardrails

Project-specific constraints. Uses the Tier 0–3 system def:
- Tier 0 — Hard Invariants (NEVER Violated)
- Tier 1 — Epistemic Integrity (Suspended Only with Explicit Waiver)
- Tier 2 — Process Quality (Best-Effort Under Pressure)
- Tier 3 — Collaboration Quality (Degraded Gracefully)

---

## Component Architecture — Tier 2

Every implementation task touching `webapp/` must:

1. **Before implementing:** Read `webapp/architecture.md`. Verify the planned change
   fits the component boundaries, service ownership, and interface contracts defined there.
   If it doesn't fit, surface the conflict before writing code.

2. **As part of DoD:** If the change introduces or modifies a service, repository interface,
   DTO, handler file, or template directory — update `architecture.md` to reflect the new
   state. A structural change without a corresponding doc update is incomplete.

---

## Test Coverage — Tier 1

Every file in `webapp/actions/`, `webapp/services/`, and `webapp/models/` that contains
at least one function or method body must have a corresponding `*_test.go` file.

**Exempt:**
- Files containing only type/interface/struct declarations with no method bodies
- Files whose sole content is an `init()` for framework wiring (e.g. `render.go`, `app.go`)
- TODO-stub files: all declared methods are unimplemented (no body or single `panic`)

**Suspension** requires an explicit waiver:
```
WAIVER: test coverage — [file] — [reason] — [payback trigger]
```
Record in `TECH_DEBT.md`.

See `webapp/TESTING.md` for stub patterns, test app setup, and the integration test harness.

---

## HTML/CSS Quality — Tier 2

Every implementation task touching `webapp/templates/` or `webapp/public/assets/` must:

1. **Before implementing:** Read `webapp/FRONTEND.md`. Verify the planned change uses
   existing CSS classes and component patterns. If a needed pattern is missing from the
   design system, surface that before writing ad-hoc CSS or inline styles.

2. **As part of DoD:** Run the pre-ship checklist in `webapp/FRONTEND.md` against all
   modified templates.

---

# Design references

Some plans may refer to a design (wireframes). Those are guidelines, not hard constraints. If the implementation sounds too complexe, stop and ask.

Do not blindly copy wireframes code, adapt to ensure accessibility, SEO and other web development best practices.
