# Frontend Reference

Reference for HTML and CSS work in `webapp/templates/` and `webapp/public/assets/`.
Read this before touching any template or stylesheet.

---

## Design Tokens

All tokens are in `webapp/public/assets/application.css`. Never hardcode colours, sizes, or spacing.

| Family | Examples |
|---|---|
| Colour — brand | `--ohm-bordeaux`, `--ohm-bordeaux-ink`, `--ohm-bordeaux-50`, `--ohm-gold`, `--ohm-gold-ink` |
| Colour — neutral | `--ink` `--ink-2` … `--ink-5`, `--paper` `--paper-2` `--paper-3`, `--line`, `--line-strong` |
| Colour — semantic | `--ok` / `--ok-bg`, `--warn` / `--warn-bg`, `--danger` / `--danger-bg`, `--info` / `--info-bg` |
| Spacing | `--sp-1` (4px) → `--sp-20` (80px) |
| Font size | `--fs-12` → `--fs-56` |
| Radius | `--r-1` (2px), `--r-2` (4px), `--r-3` (6px), `--r-4` (10px), `--r-pill` |
| Shadow | `--shadow-1` (subtle) → `--shadow-3` (elevated) |

---

## Component Patterns

### Page shell

```html
<div class="page-shell">
    <div class="page-shell__header">
        <div>
            <nav class="breadcrumb" aria-label="fil d'Ariane">
                <a href="/">Accueil</a>
                <span class="breadcrumb__sep">›</span>
                <span>Current page</span>
            </nav>
            <h1>Page title</h1>
        </div>
        <!-- optional: action button -->
        <a href="..." class="btn btn--primary">Action</a>
    </div>
    <!-- page content -->
</div>
```

Use `page-shell__header` whenever a page has a title — even without an action button. It provides the visual separator and correct spacing.

### Card

```html
<section class="card">
    <div class="card__header">
        <h2>Section title</h2>
        <!-- optional: badge, action button -->
    </div>
    <div class="card__body">
        <!-- content -->
    </div>
    <!-- optional -->
    <div class="card__footer">...</div>
</section>
```

`card__body` is required — it provides inner padding. Never place content directly inside `.card`.

### Key-value pairs

```html
<dl class="kv">
    <dt>Label</dt>
    <dd>Value — or <span class="muted">—</span> if empty</dd>
</dl>
```

Use `<dl class="kv">` for profile data, detail views, and any label/value layout. The grid collapses to single-column on mobile automatically.

### Form field

```html
<div class="field">
    <label class="field__label" for="field-id">
        Label <span class="req" aria-hidden="true">*</span>
    </label>
    <input class="input" id="field-id" type="text" name="field"
           required aria-required="true" />
    <span class="field__error" role="alert">Error message</span>
    <span class="field__hint">Hint text</span>
</div>
```

### Table

```html
<table class="table" aria-label="Descriptive label">
    <thead>
        <tr>
            <th scope="col">Name</th>
            <th scope="col"><span class="sr-only">Actions</span></th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td data-label="Name">Value</td>
            <td data-label=""><!-- action cell --></td>
        </tr>
    </tbody>
</table>
```

`data-label` on every `<td>` is required — the mobile CSS uses it for stacked labels. Action cells with no visible label: `data-label=""`.

### Confirmation modal

Use for any destructive action that needs user confirmation before submitting a form.
The modal is powered by `Alpine.store('confirmModal')`, declared in `_confirm_modal.plush.html`.

**Requirements per page:**
1. Include `<%= partial("confirm_modal.html") %>` once (just before the closing tag of the page root element)
2. The page root element (or any ancestor of the trigger buttons) must have `x-data` so Alpine processes the `@click` directives

**Trigger button:**
```html
<button type="button" class="btn btn--danger btn--sm"
        @click="$store.confirmModal.open({
            title: 'Supprimer l\'élément',
            body: 'Cette action est irréversible.',
            confirmLabel: 'Supprimer',
            action: '/admin/ressource/123',
            httpMethod: 'DELETE'
        })">Supprimer</button>
```

**`open(opts)` options:**

| Field | Default | Description |
|---|---|---|
| `title` | `''` | Dialog heading |
| `body` | `''` | Explanatory text |
| `confirmLabel` | `'Confirmer'` | Submit button label |
| `cancelLabel` | `'Annuler'` | Cancel button label |
| `confirmWord` | `''` | If set, user must type this exact word before confirm enables (for irreversible actions) |
| `action` | `''` | Form `action` URL |
| `httpMethod` | `''` | `_method` override value (e.g. `'DELETE'`) |

---

### Buttons

```html
<!-- primary action -->
<a href="..." class="btn btn--primary">Label</a>

<!-- secondary / neutral -->
<button type="button" class="btn btn--ghost">Label</button>

<!-- destructive -->
<button type="submit" class="btn btn--danger btn--sm">Supprimer</button>
```

Always use `<button type="button">` for JS-triggered actions and `<button type="submit">` inside forms. Never use `<div>` or `<a>` for actions with no URL.

---

## Pre-ship Checklist

Run against every modified template before marking done:

**Structure**
- [ ] All CSS classes referenced in the template exist in `application.css`
- [ ] Cards use `card__header` + `card__body` — no bare content directly inside `.card`
- [ ] Key-value data uses `<dl class="kv">`, not an ad-hoc class
- [ ] Page has exactly one `<h1>`; heading levels are sequential (no h3 after h1)
- [ ] Inline styles are justified — used only for values not expressible as a token or utility class

**Accessibility**
- [ ] Headings communicate the organization of the content on the page
- [ ] Every `<nav>` has an `aria-label`
- [ ] Every `<table>` has `aria-label` or `<caption>`; `<th>` elements have `scope`
- [ ] Interactive elements (`<button>`, `<a>`) have a visible label or `aria-label`
- [ ] Form inputs are associated to their `<label>` via `for`/`id`
- [ ] Images have meaningful `alt` (or `alt=""` if decorative)
- [ ] Error messages use `role="alert"` so they are announced by screen readers

**Responsiveness**
- [ ] No fixed pixel widths that would overflow a 375px viewport
- [ ] Tables include `data-label` on every `<td>` for the mobile stacked layout
- [ ] Touch targets (buttons, links) are at least 44×44px or use a `.btn` variant (min-height 30px for `btn--sm`, 38px default)

---

## Anti-patterns

| Don't | Do instead |
|---|---|
| `<dl class="definition-list">` or any invented class | `<dl class="kv">` |
| `<h2 class="card__title">` without `card__header` | `<div class="card__header"><h2>…</h2></div>` |
| Bare content inside `.card` | Wrap in `<div class="card__body">` |
| `style="color:#8b1e3f"` | `color: var(--ohm-bordeaux)` |
| `style="margin-top:20px"` | `style="margin-top:var(--sp-5)"` |
| Repeated `max-width` on sibling cards | One wrapper `<div>` with `max-width` |
| `<div onclick="…">` or `<a>` without `href` | `<button type="button">` |
| `<br>` for vertical spacing | Margins via tokens or wrapper elements |
