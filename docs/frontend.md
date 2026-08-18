# Frontend — Design System & UI Architecture

Staffio's frontend is built **entirely without CSS frameworks or JavaScript bundlers**. Every style is handcrafted vanilla CSS, every interaction is vanilla JavaScript, and the HTML is rendered server-side by Go's `html/template`.

This keeps the deployment artifact to a single Go binary plus a few static directories — no `node_modules`, no build pipeline, no transpilation.

---

## File Organization

Each page owns a dedicated CSS file in `frontend/css/`, mirroring the component-per-file pattern of CSS modules without any tooling.

```
frontend/css/
├── style.css           ← Global reset & shared base styles
├── login.css           ← Login card layout & form styling
├── register.css        ← Registration (mirrors login)
├── recover.css         ← Password recovery form
├── reset-password.css  ← Password reset form
├── landing.css         ← Product landing page
├── dashboard.css       ← Sidebar, KPI cards, employee grid
├── employees.css       ← Table, form, search bar, status badges
├── departments.css     ← Department cards, manager select, grid
├── overview.css        ← Analytics page & charts
├── reports.css         ← Reports list & detail
├── config.css          ← Account / device settings
└── id-card.css         ← Candidate profile card, print-optimized
```

JavaScript lives in `frontend/public/js/`:

```
├── dark-mode.js        ← Theme toggle
├── departments-pie.js  ← Overview pie chart (Chart.js)
├── employees-line.js   ← Overview line chart
├── employees-line-v2.js← Updated line chart (created/fired)
├── absence-menu.js     ← Absence recording interactions
├── description-clamp.js← Truncates long descriptions
└── dynamic-text.js     ← Lightweight text animation
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
| Success | `#059669` | Hired (`contractors`) status |
| Danger | `#dc2626` | Rejected status, delete buttons |
| Warning | `#d97706` | Pending status |
| Radius (sm/md/lg) | `10px` / `12px` / `20px` | Buttons / inputs·badges / window cards |

A **dark-mode** toggle is implemented via `dark-mode.js`, switching the design tokens through CSS custom properties.

---

## Pages

### Login & Register

Light, calm surface: `#e2e8f0` gradient background, centered white card (`border-radius: 24px`, layered shadow), rounded inputs (`12px`) with a focus transition, and an indigo CTA with a `translateY(-2px)` hover. An `fadeInUp` entrance animation plays on load.

### Dashboard

Two-column layout: fixed sidebar (`260px`) + fluid main. KPI cards use CSS Grid `auto-fit`/`minmax(180px, 1fr)` and reflow on resize. Each card's number uses a distinct gradient text color (indigo, pink, amber, red) for fast scanning. The employee grid uses `auto-fill`/`minmax(260px, 1fr)` cards linking to ID-card views. Recruit-role users are redirected to `/employees`.

### Employees

The densest screen — a table (`border-collapse: collapse`) with alternating hover backgrounds. Status badges are pills (`border-radius: 999px`):

| Status | Background | Text |
|---|---|---|
| `pending` | `#fef3c7` | `#d97706` |
| `contractors` | `#d1fae5` | `#059669` |
| `rejected` | `#fee2e2` | `#dc2626` |

Photos render as `48×48px` thumbnails (`object-fit: cover`); the fallback is a purple gradient with a `👤` glyph.

### Departments

Two-column creation form with a styled manager `<select>` (SVG data-URI arrow). Cards render in `repeat(auto-fill, minmax(300px, 1fr))` with a `translateY(-2px)` hover + indigo border.

### Overview & Reports

Chart pages load Chart.js for the department pie and the employees created/fired line series.

### ID Card

Print-optimized (`@media print` hides nav/actions, expands the card, forces white background). Flexbox row: `260×260px` photo left, info right; collapses to a centered column below `850px`.

---

## Template Engine

Go's `html/template` provides rendering with **auto-escaping** — any user-supplied string rendered via `{{ .field }}` is HTML-escaped by default (XSS-safe).

Helper functions registered in `main.go`:

| Function | Purpose |
|---|---|
| `lower` | Maps status to lowercase CSS classes (`{{.Status \| lower}}`) |
| `add` | Pagination arithmetic (`{{add .currentPage 1}}`) |
| `csrfField` | Renders the hidden CSRF `<input>` (see [Security](./security.md)) |

---

## JavaScript

JavaScript is minimal and purposeful: chart rendering, dark-mode toggle, absence interactions, description clamping, and confirm/delete actions — all vanilla, all fetch-based where needed. There is no client-side routing; navigation is full-page requests, keeping behavior predictable.

---

## Responsiveness

Media queries at `900px` (employees/departments) and `800px` (dashboard) switch the layout to a single column: the sidebar becomes a full-width horizontal nav, the form grid collapses to one column, and padding reduces. A secondary `768px` breakpoint tightens the table and stacks action buttons.

---

## Related

- [Architecture](./architecture.md)
- [Security](./security.md)
- [Backend](./backend.md)