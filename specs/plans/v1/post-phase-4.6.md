# Post-Phase 4.6 — Adjustments Index

Improvements and infrastructure additions following phase 4.6 completion.

---

## Status

| # | Adjustment | File | Status |
|---|-----------|------|--------|
| 1 | Dashboard redesign + event description | [post-4.6/adj-1-dashboard-description.md](post-4.6/adj-1-dashboard-description.md) | Not started |
| 2 | E2e test framework + CI coverage | [post-4.6/adj-2-e2e.md](post-4.6/adj-2-e2e.md) | Not started |
| 3 | jscpd duplicate-code detection | [post-4.6/adj-3-jscpd.md](post-4.6/adj-3-jscpd.md) | Not started |

---

## Open Decisions (blocking implementation)

| ID | Adjustment | Decision | Status |
|----|-----------|---------|--------|
| D1 | adj-1 | Markdown rendering library | **Pending** |
| D2 | adj-2 | E2e browser automation framework | **Accepted: Playwright (TypeScript) — ADR 010** |

Each decision file describes the options and a recommendation; take each decision before
handing the corresponding adjustment to an agent.

---

## Suggested Sequence

```
adj-3 (jscpd) → adj-1 (dashboard + description) → adj-2 (e2e)
```

**Rationale:**
- `adj-3` first: establishes the quality gate before new code is added; any violations
  are fixed against the post-4.6 baseline so the hook doesn't block the next two PRs.
- `adj-1` second: main user-facing feature change; self-contained.
- `adj-2` last: pure infrastructure; independent of feature work but benefits from
  having all user flows in place so e2e scenarios can be written completely.
