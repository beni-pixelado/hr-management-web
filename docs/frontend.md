# Frontend — Design System & UI Architecture

## Philosophy

The frontend of HR Management Web is built entirely without CSS frameworks or JavaScript bundlers. Every style rule is handcrafted, every interaction is vanilla JavaScript, and the HTML is generated server-side by Go's `html/template` engine. This is a deliberate choice: no build pipeline, no `node_modules`, no transpilation step — just static files that Gin serves directly.

This approach keeps the deployment artifact to a single Go binary plus a few directories, makes the CSS fully transparent and debuggable, and avoids the version churn and bundle bloat that come with framework dependencies.

## File Organization

Each page has its own dedicated CSS file in `frontend/css/`, plus a shared `style.css` that defines global design tokens (colors, typography, spacing scale). This mirrors the component-per-file pattern common in CSS module systems, but without requiring any tooling.

```
frontend/css/
├── style.css         ← Global tokens: colors, fonts, reset
├── login.css         ← Login card layout and form styling
├── register.css      ← Register page (mirrors login structure)
├── dashboard.css     ← Sidebar navigation, KPI cards, layout grid
├── employees.css     ← Table, card grid, search bar, status badges
└── id-card.css       ← Candidate profile card, print-optimized
```

Each template loads only its own CSS file plus the shared `style.css`, keeping the CSS payload minimal per page.

## Design Tokens

The global `style.css` establishes CSS custom properties (variables) that are used throughout all page-specific stylesheets:

```css
:root {
    --color-bg: #0f1117;          /* Dark page background */
    --color-surface: #1a1d27;     /* Card and panel surface */
    --color-surface-alt: #22253a; /* Slightly elevated surfaces */
    --color-border: #2e3150;      /* Subtle border color */
    --color-primary: #3d52a0;     /* Brand indigo — buttons, accents */
    --color-primary-hover: #4a63bf;
    --color-text: #e2e8f0;        /* Primary body text */
    --color-text-muted: #8892b0;  /* Secondary text, labels */
    --color-success: #4caf82;     /* Accepted status */
    --color-danger: #e05555;      /* Rejected status */
    --color-warning: #d4a843;     /* Pending status */

    --radius-sm: 6px;
    --radius-md: 12px;
    --radius-lg: 20px;

    --shadow-card: 0 4px 24px rgba(0, 0, 0, 0.35);
}
```

The benefit of this centralization is that changing the brand color from indigo to another value requires editing a single line — all buttons, links, and highlights update simultaneously.

## Page-by-Page Breakdown

### Login Page (`login.css`)

The login page uses a light mode exception to the dark theme: a soft `#e8eaf0` gradient background is chosen to create a calm, trustworthy first impression without the darkness of the authenticated area. This is intentional — the login screen is the "outside" of the system, and the visual shift to dark on entry creates a clear threshold moment.

The white card is centered using absolute positioning with a `transform: translate(-50%, -50%)` technique, ensuring it stays centered regardless of viewport height. `border-radius: 16px` and a layered `box-shadow` give it a lifted, floating feel. Input fields use `border-radius: 999px` (pill shape) with a transition on the border color on focus, providing gentle visual feedback. The CTA button uses the brand indigo (`#3d52a0`) with a hover darkening transition.

### Dashboard (`dashboard.css`)

The dashboard layout is built on a two-column CSS Grid: a fixed-width sidebar (`240px`) and a fluid content area (`1fr`). This creates a classic application shell that is immediately recognizable to enterprise software users.

The sidebar uses a dark surface (`var(--color-surface)`) with navigation items styled as full-width rows. The active item is indicated by a left border accent in brand indigo and a slightly lighter background, making the current location unmistakeable without heavy visual weight. Icon + label pairs use flexbox alignment to stay consistent regardless of label length.

KPI cards (Accepted / Rejected / Pending counts) use the same surface background as the sidebar but with a 2px left border in the corresponding status color. The number is displayed in a large, heavy weight font with the label in muted text below it. This creates a scannable information hierarchy — the number is read first, the label second.

### Employees Page (`employees.css`)

This page has the highest information density in the application. The table uses `border-collapse: collapse` with alternating row backgrounds for readability. Column widths are set with a mix of fixed values (for photo and status columns, which have predictable content) and `auto` (for name, position, email, which vary in length).

Status badges are `display: inline-flex` chips with `border-radius: 999px`, a background color at 15% opacity of the status color, and the full status color for the text. This keeps them readable without overwhelming the table row. The color mapping is: `Accepted → var(--color-success)`, `Rejected → var(--color-danger)`, `Pending → var(--color-warning)`.

Profile photos are rendered as `40px × 40px` circles using `border-radius: 50%` and `object-fit: cover`, ensuring square photos don't appear stretched. When no photo exists, the template renders an SVG avatar placeholder at the same dimensions.

The search bar sits in the page header row alongside the "Add Candidate" button. It uses a dark input field that blends with the surface while providing clear focus highlighting, keeping the header area clean and uncluttered.

### ID Card (`id-card.css`)

The ID card template is optimized for both screen display and printing. `@media print` rules hide navigation and action buttons, expand the card to full page width, and use a white background to ensure the card prints cleanly on paper. This is a small but professional touch that makes the feature genuinely usable in real HR workflows.

## Template Engine

Go's `html/template` package provides the rendering layer. Key properties of this engine that shape how the frontend works:

**Auto-escaping** means any user-supplied string rendered in a template is automatically HTML-escaped. Outputting `{{ .employee.Name }}` where `Name` is `<script>alert(1)</script>` will render as the literal text, not execute the script. This is XSS protection by default, without any extra effort.

**Template composition** is achieved with `{{ template "name" . }}` and `{{ define "name" }}` blocks. Shared layout elements (navigation, header, footer) are defined once and included in each page template, reducing duplication.

**Helper functions** can be registered with the Gin template engine for things like formatting dates, constructing pagination URLs with query parameters, or mapping status strings to CSS class names. These are registered in `main.go` via `r.SetFuncMap(template.FuncMap{...})` before `LoadHTMLGlob` is called.

## JavaScript

JavaScript is kept minimal and purposeful. The primary use cases are status update requests (which use `fetch` to send `PUT /employees/:id/status` without a full page reload) and any UI toggling behavior (like opening an "Add Candidate" modal). There are no JavaScript frameworks, no build steps, and no significant client-side state management. This is an intentional constraint that keeps the application predictable and the behavior transparent.