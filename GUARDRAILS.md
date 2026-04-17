# Project Guardrails

Project-specific constraints. Uses and extends the Tier 0–3 system defined in CORE.md.

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
