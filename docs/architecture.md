# Architecture

**HR Management Web (Staffio)** follows a classic **layered server-side architecture**. Each layer has one well-defined responsibility, dependencies flow in a single direction, and the entire application deploys as a single Go binary.

```
┌─────────────────────────────────────────────────────────────┐
│                    Browser (Client)                           │
│          HTML rendered by Go templates + CSS + JS             │
└─────────────────────────────┬───────────────────────────────┘
                              │  HTTP (GET / POST / DELETE)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                  Gin HTTP Layer                              │
│      Router · Middleware Stack · Static Files · Templates    │
│                    backend/cmd/server/main.go                │
└──────┬──────────────────────┬──────────────────┬────────────┘
       │                      │                  │
       ▼                      ▼                  ▼
┌──────────────┐     ┌──────────────────┐  ┌──────────────────┐
│  auth.go     │     │  employee.go     │  │ departament.go   │
│  organization│     │  config.go       │  │  report.go       │
│  permissions │     │  overview.go     │  │  password_reset  │
└──────┬───────┘     └──────┬───────────┘  └──────┬───────────┘
       │                   │                      │
       └───────────────────┼──────────────────────┘
                           │  calls internal packages
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    internal/ packages                         │
│                                                               │
│  ┌────────────┐  ┌─────────────┐  ┌────────────────────────┐ │
│  │ internal/  │  │internal/auth│  │internal/middleware      │ │
│  │ storage    │  │session R/W  │  │RequireAuth · RateLimit  │ │
│  │ Cloudinary │  │             │  │LoadUser · BlockViewer   │ │
│  └──────┬─────┘  └─────┬───────┘  └───────┬───────────────┘ │
│         └──────────────┼──────────────────┘                  │
│                        ▼                                     │
│              ┌──────────────────┐                            │
│              │ internal/csrf    │  token middleware + field  │
│              └────────┬─────────┘                            │
└───────────────────────┼──────────────────────────────────────┘
                        ▼
┌─────────────────────────────────────────────────────────────┐
│            Neon PostgreSQL (cloud, serverless)                │
│         Multi-tenant · Connection Pooled · Persistent         │
└─────────────────────────────────────────────────────────────┘
```

---

## Layer Responsibilities

### 1. Composition root — `backend/cmd/server/main.go`

The Gin engine is created here. It wires everything together and contains **no business logic**:

- Loads `.env` and initializes structured logging (`log/slog`).
- Connects to PostgreSQL and runs `AutoMigrate` for all models.
- Backfills legacy users into per-user organizations.
- Registers routes, middleware, template functions and static assets.
- Starts the HTTP server on `PORT` (default `8000`).

### 2. Handlers — `backend/handlers/`

Handlers are the **controller layer**. They receive `*gin.Context`, extract input, enforce permissions, call the database, and render templates or JSON. The handlers are grouped by domain:

| File | Domain |
|---|---|
| `auth.go` | Registration, login, logout, user/org models, roles |
| `employee.go` | Candidate CRUD, search, status pipeline, absences, CSV export, badge |
| `departament.go` | Department CRUD + member assignment |
| `organization.go` | Team management, invites, role changes, ownership transfer |
| `permissions.go` | RBAC helpers, current-user resolution, authorization rules |
| `overview.go` | Overview page + chart JSON APIs |
| `report.go` | Report creation, listing, detail + analytics JSON APIs |
| `config.go` | Account/profile/photo/password/device management |
| `password_reset_token.go` | Password recovery token model + flow |

### 3. Internal packages — `internal/`

Shared, cross-cutting concerns that multiple handlers depend on. Go's `internal/` rule prevents them from being imported outside the module.

- `internal/auth` — session read/write helpers wrapping `gorilla/sessions`.
- `internal/csrf` — per-session CSRF token generation, middleware and template helper.
- `internal/middleware` — `RequireAuth`, `LoadUser`, `BlockViewerWrites`, `RedirectIfAuthenticated`, `RateLimit`.
- `internal/storage` — Cloudinary upload/destroy helpers.

### 4. Templates — `backend/templates/`

Server-rendered HTML using Go's `html/template`, which **auto-escapes output by default** — preventing XSS. Gin loads them via `r.LoadHTMLGlob("backend/templates/*")` and renders by name.

### 5. Static assets — `frontend/`

Vanilla CSS and JS served directly by Gin (`/css`, `/js`, `/static`). No build step, no bundler.

---

## Middleware Stack

The `protected` route group applies a chain in order:

1. **`RequireAuth`** — reads the session cookie; redirects to `/login` if invalid.
2. **`LoadUser`** — loads the user's `role` and `organization_id` into the context.
3. **`BlockViewerWrites`** — redirects `viewer`-role users away from any non-GET request.

Global middleware applies to **all** routes:

- **`csrf.Protect()`** — issues a CSRF token for GET requests and validates it on state-changing requests.

Auth routes (`/login`, `/register`, `/recuperateaccount`, `/reset-password`) are additionally wrapped in `middleware.RateLimit`.

---

## Request Lifecycle

A typical authenticated request (e.g. `GET /employees`):

1. Browser sends `GET /employees` with the session cookie.
2. `csrf.Protect()` issues/serves the CSRF token for the rendered page.
3. `RequireAuth` validates the cookie → calls `c.Next()` or redirects to `/login`.
4. `LoadUser` reads the user's role + org and stores them in context.
5. `BlockViewerWrites` permits the read.
6. The handler loads the current user, scopes the query to `organization_id`, applies filters/pagination, and renders `employees.html`.

---

## Multi-Tenancy

Every data model that belongs to an organization carries an `OrganizationID` field, and **every query is scoped to the acting user's org**. This guarantees that one company's candidates, departments, reports and absences are never visible to another. See [Roles & Permissions](./roles-permissions.md) and [Database](./database.md).

---

## Key Design Decisions

| Decision | Rationale |
|---|---|
| **Server-rendered HTML** (no SPA) | Single deployable binary, no frontend framework, predictable behavior |
| **Layered handlers → internal** | One-way dependency keeps the system testable and replaceable |
| **`html/template`** | Auto-escaping gives XSS protection by default |
| **Cookie sessions** (no JWT) | Simple, revocation-friendly, no token storage to manage |
| **RBAC + tenant scoping** | Enforces least privilege and hard data isolation by construction |
| **`log/slog` structured logging** | Zero-dependency, readable logs on Render |

---

## Next

- [Backend](./backend.md) — handlers and business logic in depth
- [Database](./database.md) — models, migrations and data flow
- [Security](./security.md) — CSRF, rate limiting and hardening