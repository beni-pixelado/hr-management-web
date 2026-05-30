# Frontend — Design System & UI Architecture

## Philosophy

The frontend of HR Management Web is built entirely without CSS frameworks or JavaScript bundlers. Every style rule is handcrafted, every interaction is vanilla JavaScript, and the HTML is generated server-side by Go's `html/template` engine. No build pipeline, no `node_modules`, no transpilation — just static files that Gin serves directly.

This approach keeps the deployment artifact to a single Go binary plus a few directories, makes the CSS fully transparent and debuggable, and avoids the version churn and bundle bloat that come with framework dependencies.

---

## File Organization

Each page has its own dedicated CSS file in `frontend/css/`. This mirrors the component-per-file pattern common in CSS module systems, without requiring any tooling.

```
frontend/css/
├── style.css           ← Global reset
├── login.css           ← Login card layout and form styling
├── register.css        ← Register page (mirrors login structure)
├── dashboard.css       ← Sidebar, KPI cards, employee grid
├── employees.css       ← Table, form, search bar, status badges
├── departments.css     ← Department cards, manager select, grid
└── id-card.css         ← Candidate profile card, print-optimized
```

---

## Design Tokens

The shared visual language across all pages:

| Token | Value | Usage |
|---|---|---|
| Page background | `#e2e8f0` | All page `body` backgrounds |
| Surface (card) | `#ffffff` | Main content areas, form sections |
| Surface alt | `#f8fafc` | Table headers, sidebar background |
| Border | `#edf2f7` | Separators, card outlines |
| Brand indigo | `#5c6bc0` / `#3f51b5` | Buttons, active nav, code badges |
| Text primary | `#1e293b` | Headings, labels |
| Text secondary | `#475569` | Nav items, form labels |
| Text muted | `#64748b` | Subtitles, helper text |
| Success | `#059669` | Contractors status |
| Danger | `#dc2626` | Rejected status, delete buttons |
| Warning | `#d97706` | Pending status |
| Border radius (sm) | `10px` | Buttons |
| Border radius (md) | `12px` | Inputs, nav items, badges |
| Border radius (lg) | `20px` | Window cards |

---

## Page-by-Page Breakdown

### Login & Register Pages

Light mode pages (unlike the authenticated area): `#e2e8f0` gradient background, centered white card (`border-radius: 24px`, layered `box-shadow`). Input fields use `border-radius: 12px` with a transition on `border-color` and `box-shadow` on focus. The CTA button uses the brand indigo gradient with a `translateY(-2px)` hover effect. An entrance animation (`fadeInUp`) runs once on page load.

### Dashboard (`dashboard.css`)

Two-column layout: fixed-width sidebar (`260px`) and fluid main area (`flex: 1`). Sidebar uses `#f9fafb` with icon+label nav items — active item gets the indigo gradient. KPI cards use CSS Grid `auto-fit` with `minmax(180px, 1fr)` so the grid reflows gracefully on resize. Each card's number uses a gradient text color unique to that card (indigo, pink, amber, red) for fast scanning. The employee grid below uses `auto-fill` with `minmax(260px, 1fr)` cards that link to individual ID card views.

### Employees Page (`employees.css`)

Highest information density in the application. Table uses `border-collapse: collapse` with alternating hover backgrounds (`#f8fafc`). Status badges use `border-radius: 999px` (pill shape) with background and text color derived from the status value:

| Status | Background | Text |
|---|---|---|
| `pending` | `#fef3c7` | `#d97706` |
| `contractors` | `#d1fae5` | `#059669` |
| `rejected` | `#fee2e2` | `#dc2626` |

Profile photos render as `48×48px` thumbnails with `border-radius: 12px` and `object-fit: cover`. The fallback is a purple gradient div with a `👤` emoji. The form section uses a gradient background (`#f8f9fc → #a8a8b0`) to visually separate it from the table below.

### Departments Page (`departments.css`)

Mirrors the window/sidebar/main-content structure of the other pages. The creation form is a two-column CSS Grid; the manager select is styled with a custom dropdown arrow via `background-image` (SVG data URI). Department cards are rendered in a `repeat(auto-fill, minmax(300px, 1fr))` grid. Cards hover with `translateY(-2px)` and a stronger `box-shadow` and `border-color: #5c6bc0`.

### ID Card (`id-card.css`)

Print-optimized. `@media print` rules hide navigation and action buttons, expand the card to full page width, and use a white background. The card body is a flexbox row: `260×260px` photo on the left (with `border-radius: 24px`) and employee info on the right. On viewports below `850px`, the layout switches to column with centered content.

---

## Template Engine

Go's `html/template` package provides the rendering layer.

**Auto-escaping** — any user-supplied string rendered with `{{ .field }}` is automatically HTML-escaped. XSS protection by default, no extra effort.

**Template composition** — shared layout elements can be defined with `{{ define "name" }}` and included with `{{ template "name" . }}`. Currently each page is a standalone template, but partials can be extracted as the project grows.

**Helper functions** — two custom functions are registered at startup in `main.go`:

```go
r.SetFuncMap(template.FuncMap{
    "lower": strings.ToLower,
    "add":   func(a, b int) int { return a + b },
})
```

`lower` is used to map `Status` strings to lowercase CSS class names: `{{.Status | lower}}` → `"pending"`, `"contractors"`, `"rejected"`.

`add` is used for pagination arithmetic: `{{add .currentPage 1}}` avoids needing a dedicated `nextPage` field in some contexts.

---

## JavaScript

JavaScript is kept minimal and purposeful. The primary use case is the delete confirmation for candidates:

```javascript
function deleteEmployee(id) {
    if (confirm('Are you sure you want to delete this candidate?')) {
        fetch('/employees/' + id, { method: 'DELETE' })
            .then(() => location.reload());
    }
}
```

There are no JavaScript frameworks, no build steps, and no client-side routing. All navigation is full-page requests. This is an intentional constraint that keeps the application predictable and behavior transparent.

---

## Responsiveness

All pages include a media query at `@media (max-width: 900px)` (employees/departments) or `@media (max-width: 800px)` (dashboard) that:

- Switches `window-body` from `flex-direction: row` to `column`
- Expands the sidebar to full width and removes the right border
- Switches the `nav-menu` from vertical column to horizontal `flex-wrap` row
- Collapses the form grid to a single column
- Reduces padding on `main-content`

A secondary breakpoint at `768px` reduces table padding further and stacks action buttons vertically.