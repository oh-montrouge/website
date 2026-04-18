# Project Guardrails

Project-specific constraints. Uses the Tier 0–3 system def:
- Tier 0 — Hard Invariants (NEVER Violated)
- Tier 1 — Epistemic Integrity (Suspended Only with Explicit Waiver)
- Tier 2 — Process Quality (Best-Effort Under Pressure)
- Tier 3 — Collaboration Quality (Degraded Gracefully)

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
