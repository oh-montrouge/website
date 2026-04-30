# Adj-3 — jscpd Duplicate-Code Detection

**Status:** Complete — all tasks done. jscpd exits 0 on both Go (<0.25%) and markup (0%) passes.

**Goal:** Add `jscpd` as a Lefthook pre-commit hook to catch duplicated code blocks, run
it against the current codebase, and fix detected violations.

---

## No Open Decisions

Implementation is fully specified. No decision gate required.

---

## Tasks

### T1 — Add jscpd to Lefthook pre-commit

Touches: `lefthook.yml`.

Add the following command to the `pre-commit` hook in `lefthook.yml`:

```yaml
jscpd:
  run: npx jscpd webapp --min-lines 10 --gitignore --threshold 4
  skip_output:
    - meta
    - empty_output
```

**Flags explained:**
- `webapp` — scope to the Go application directory only.
- `--min-lines 10` — ignore blocks shorter than 10 lines (avoids noise from short
  idiomatic patterns).
- `--gitignore` — respect `.gitignore` (excludes vendor, generated files, etc.).
- `--threshold 4` — fail if duplicated code exceeds 4% of total lines.

The hook must pass (exit 0) before it can be committed. So T2 must be completed before
this change can itself be committed.

**DoD:** `lefthook.yml` updated; `npx jscpd` is available via the existing Node/npx setup.

### T2 — Run jscpd and fix violations

Run `npx jscpd webapp --min-lines 10 --gitignore --threshold 4` against the current
codebase. For each detected duplicate:

1. Inspect the duplicate. Is it a genuine refactoring opportunity, or a false positive
   (e.g. idiomatic Go boilerplate that jscpd misidentifies as duplication)?
2. **If refactorable:** extract to a shared function/struct; ensure tests still pass.
3. **If the fix would introduce an unwanted abstraction:** wrap the block with
   `// jscpd:ignore-start` / `// jscpd:ignore-end` (or `{{/* jscpd:ignore-start */}}`
   / `{{/* jscpd:ignore-end */}}` for templates). Add an inline comment immediately
   above `jscpd:ignore-start` stating the reason:
   `// jscpd: unwanted abstraction — [brief rationale]`
   No `TECH_DEBT.md` entry required.
4. **If the duplication cannot be evaluated (lack of domain knowledge or unclear
   ownership):** raise it with the user before proceeding. If the decision is to keep
   it, apply the same ignore markers with an inline comment stating the reason:
   `// jscpd: unclear ownership — [brief rationale]`
   Record the decision in `TECH_DEBT.md` with a payback trigger.

Once all duplicates are resolved or suppressed, the tool must exit 0.

**DoD:** `npx jscpd webapp --min-lines 10 --gitignore --threshold 4` exits 0; pre-commit
passes; all existing tests pass.

### T3 — Document the suppression convention in CLAUDE.md

Touches: `CLAUDE.md`.

Add a project guardrail instructing the AI harness that every `jscpd:ignore-start` marker
must be preceded by an inline comment stating one of two reasons:
- `// jscpd: unwanted abstraction — [brief rationale]`
- `// jscpd: unclear ownership — [brief rationale]`

And that `unclear ownership` suppressions additionally require a `TECH_DEBT.md` entry.

**DoD:** `CLAUDE.md` updated; the rule is visible to the AI harness on any future task
touching `jscpd:ignore` markers.

---

## Constraints

- Do not change test behaviour to make tests pass after a refactor (Rule 14).
- If a refactored extraction would span package boundaries in a way that violates
  `webapp/architecture.md` (e.g. a model helper in the actions layer), surface the
  conflict before writing code.
- Refactoring changes introduced in T2 must be functionally neutral — behaviour is
  identical before and after.

---

## Acceptance Criteria

### Machine-verified

**AC-M1 — Pre-commit hook executes**
A commit attempt triggers `npx jscpd webapp ...` via Lefthook. The hook is listed in
`lefthook.yml` under `pre-commit`.

**AC-M2 — Tool exits 0 on current codebase**
Running `npx jscpd webapp --min-lines 10 --gitignore --threshold 4` from the repo root
exits with code 0 after T2 is complete.

**AC-M3 — All existing tests pass**
`mise run test` passes with no regressions after any refactoring done in T2.

### Human-verified

**AC-H1 — Suppressions are documented**
Every `jscpd:ignore` marker has an inline comment stating one of two reasons:
`unwanted abstraction` or `unclear ownership`. Only `unclear ownership` suppressions
require a corresponding `TECH_DEBT.md` entry with a payback trigger.
