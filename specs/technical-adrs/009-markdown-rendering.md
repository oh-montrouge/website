# ADR 009 — Markdown Rendering: goldmark

| Field | Value |
|-------|-------|
| Status | Accepted |
| Date | 2026-04-30 |

## Context

Event descriptions (column `description TEXT`, already in DB) will be entered by admins and
rendered for members on the dashboard card and the event detail page. The descriptions are
user-supplied rich text; the chosen library must convert markdown to safe HTML without
exposing the application to XSS through injected `<script>` tags or on-attribute event handlers.

---

## Alternatives Considered

### goldmark

CommonMark-compliant Go parser. Used by Hugo, the Go Playground, and pkg.go.dev.
Strips raw HTML by default — `WithUnsafe()` must be explicitly enabled to pass HTML through.
Actively maintained as of 2026.

### blackfriday v2

Established in the Go ecosystem. Not CommonMark-compliant; its last substantive release was
2019, with maintenance uncertain since 2021. Produces similar output but with subtle spec
divergence and no clear path forward.

**Rejected because:** maintenance trajectory is poor and CommonMark compliance is worth having
for predictable rendering. goldmark has no meaningful disadvantages relative to blackfriday v2.

---

## Decision

**goldmark** (`github.com/yuin/goldmark`).

XSS safety is non-negotiable: all descriptions are user-supplied. goldmark is safe by default —
raw HTML in input is stripped unless the renderer is configured with `WithUnsafe()`. This
option MUST NOT be enabled in this application.

---

## Template Helper Contract

A `markdownToHTML` helper is registered in `actions/render.go` (or `actions/helpers.go` if
that file is introduced). It converts a markdown string to HTML and returns `template.HTML`,
signalling to Buffalo/Plush that the value is already safe and must not be double-escaped.

```go
// markdownToHTML converts user-supplied markdown to safe HTML.
// XSS safety: goldmark strips raw HTML by default; WithUnsafe() is never enabled.
func markdownToHTML(s string) template.HTML {
    var buf bytes.Buffer
    if err := goldmark.Convert([]byte(s), &buf); err != nil {
        return template.HTML(html.EscapeString(s))
    }
    return template.HTML(buf.String())
}
```

**Invariants:**
- `goldmark.New()` is called without `renderer.WithNodeRenderers()` that enable `WithUnsafe()`.
- The return type is `template.HTML` — callers must never cast the output back to `string`
  and feed it into another template as raw data.
- Empty input returns empty output without error.

---

## Consequences

- `go.mod` gains one new direct dependency: `github.com/yuin/goldmark`.
- `markdownToHTML` is registered as a Plush helper alongside existing helpers in
  `actions/render.go` (or a new `actions/helpers.go`).
- `webapp/architecture.md` is updated to record the helper and its XSS invariant.
- Templates call the helper as `<%= markdownToHTML(event.Description) %>` and must guard
  against empty descriptions before rendering the surrounding HTML block.
