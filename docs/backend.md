# Backend — Handlers, Routing & Business Logic

This document describes the HTTP/controller layer. Infrastructure lives in `internal/` and is covered in [Architecture](./architecture.md) and [Security](./security.md).

The backend is split into `backend/handlers/` (request/response logic) and `backend/` infra (`cmd/`, `database/`). Handlers call `internal/` packages — never the reverse — keeping the layers independent and testable.

---

## Routing — `backend/cmd/server/main.go`

The composition root registers all routes. Public and protected routes are separated:

```go
r := gin.Default()
r.SetTrustedProxies(nil)

r.SetFuncMap(template.FuncMap{
    "lower":     strings.ToLower,
    "add":       func(a, b int) int { return a + b },
    "csrfField": csrf.Field,
})

r.Use(csrf.Protect())
r.LoadHTMLGlob("backend/templates/*")
r.Static("/css", "frontend/css")
r.Static("/js", "frontend/public/js")
r.Static("/static", "frontend/static")

// Public routes
r.GET("/login", ...)
r.POST("/login", middleware.RateLimit(10, time.Minute), handlers.Login)
// ... register, recuperateaccount, reset-password, landing, robots.txt

// Protected routes
protected := r.Group("/")
protected.Use(middleware.RequireAuth, middleware.LoadUser, middleware.BlockViewerWrites)
{
    protected.GET("/dashboard", ...)
    protected.GET("/employees", handlers.GetEmployees)
    protected.POST("/employees", handlers.CreateEmployee)
    // ...
}
```

The `protected` group applies authentication, user/org loading and the viewer-write block to every route inside it. Adding a protected route means registering it inside the group — no per-handler auth checks needed.

**Template functions:**

| Function | Purpose |
|---|---|
| `lower` | Maps status strings to lowercase CSS class names (`{{.Status \| lower}}`) |
| `add` | Offset arithmetic for pagination (`{{add .currentPage 1}}`) |
| `csrfField` | Renders a hidden `<input>` carrying the CSRF token |

---

## `handlers/auth.go` — Authentication & Models

Defines the core models and the auth surface.

### Data models

- `Organization{ID, Name, CreatedAt}`
- `User{ID, Username, Password, Email, Photo, OrganizationID, Role}`

Roles are string constants: `owner`, `admin`, `recruit`, `viewer`.

### Register (`POST /register`)

Reads `username`, `email`, `password`; validates presence; checks for an existing username; hashes the password with **bcrypt**; creates the user inside a fresh `Organization` (multi-tenant bootstrap). On error it re-renders `register.html` with a user-facing message.

### Login (`POST /login`)

Finds the user by username/email, verifies the credential with `bcrypt.CompareHashAndPassword`, then calls `auth.CreateSession(c, user.ID)` to write the signed session cookie and redirects to `/dashboard`. Failures return a deliberately vague error to avoid account enumeration.

### Logout (`GET /logout`)

Calls `auth.DestroySession(c)` (sets cookie `MaxAge = -1`), then redirects to `/login`.

---

## `handlers/employee.go` — Candidate Management

The largest handler file. All operations are **scoped to the actor's `OrganizationID`** and gated by `actor.CanEditEmployees()`.

### GetEmployees (`GET /employees`)

Builds a query filtered by org, with optional `?status=` and `?page=`. It computes per-status totals (`pending`/`contractors`/`rejected`) plus the filtered total, applies `OFFSET`/`LIMIT` pagination (20/page), and renders `employees.html`. Status aliases map friendly labels to stored values:

| Alias | Stored status |
|---|---|
| `interviewing` | `pending` |
| `hired` | `contractors` |
| `rejected` | `rejected` |

### DownloadEmployeesCSV (`GET /employees/download`)

Exports the org's employees (optionally filtered by status) as a CSV attachment.

### CreateEmployee (`POST /employees`)

Validates `full_name`, `email`, `position`; optionally uploads a photo; inserts the row with `Status: "pending"`.

**Photo upload** (`saveUploadedImage`):

1. Rejects files over `5 MB`.
2. Sniffs the MIME type (`http.DetectContentType`).
3. Validates against the allowlist: `image/jpeg`, `image/png`, `image/gif`, `image/webp`.
4. Uploads to **Cloudinary** under a UUID-based public ID (no path traversal, no collisions).

### UpdateEmployeeStatus (`POST /employees/:id/status`)

Sets `status` and optional `hire_date`. Uses a `map` for `Updates` so zero values update correctly. On `rejected` it records `rejected_at` and logs a `Termination`; otherwise clears `rejected_at`.

> **RBAC guard:** a `recruit` may not reject a candidate whose email belongs to a member of higher rank (see `employeeLinksToSuperior`).

### MarkAbsence (`POST /employees/:id/absence`)

Within a transaction: creates an `Absence` row and atomically increments `Employee.Absences` via `UpdateColumn("absences", gorm.Expr("absences + 1"))`.

### DeleteEmployee (`DELETE /employees/:id`) / DeleteEmployeeForm (`POST /employees/:id/delete`)

Soft-deletes the employee, logs a `Termination` (reason `deleted`), and destroys the photo on Cloudinary. Both a JSON and a form-path variant exist.

### EditEmployeePage / UpdateEmployee

Render the edit form and apply updates (including photo replacement, destroying the previous image).

### GetEmployeesAPI (`GET /api/employees`)

JSON-only endpoint supporting `?search=` and `?status=`, scoped to the org. Intended for a future REST API layer.

### BadgeHandler (`GET /badge/:id`)

Renders a print-friendly employee ID card (`id-card.html`).

---

## `handlers/departament.go` — Departments

- **DepartmentPageHandler (`GET /department`)** — loads all employees (for the manager dropdown) and departments, renders `departments.html`.
- **CreatedepartmentHandler (`POST /department`)** — validates `name`/`code`, parses optional `boss_id`, creates the `Department` row.
- **DepartmentHandler (`GET /department/:id`)** — paginated detail view with members.
- **AssignEmployeeToDepartment / DeleteEmployeeFromDepartment** — manage membership.
- **DeleteDepartment (`POST /department/:id/delete`)** — remove a department.

See [Departments](./departments.md).

---

## `handlers/organization.go` — Team & Multi-Tenant Management

Requires `CanManageTeam()` (owner/admin).

- **TeamPageHandler (`GET /team`)** — lists the org's members.
- **InviteMemberHandler (`POST /team/invite`)** — creates a member with a temporary password (or updates an existing member's role) and emails credentials. Role assignability depends on the actor's rank.
- **ChangeRoleHandler (`POST /team/:id/role`)** — reassigns a role respecting the permission matrix.
- **RemoveMemberHandler (`POST /team/:id/remove`)** — removes a member, reassigning their records to the actor to preserve data integrity.
- **TransferOwnershipHandler (`POST /team/:id/transfer`)** — the owner hands over the `owner` role (transactional swap to `admin`).

See [Roles & Permissions](./roles-permissions.md).

---

## `handlers/permissions.go` — Authorization

Central RBAC helpers: `roleRank`, `GetCurrentUser`, and the permission predicates `CanManageTeam`, `CanEditEmployees`, `CanViewEmployees`, `CanAssignDepartment`, plus the nuanced `canChangeRole` / `canRemoveMember` rules. `abortUnauthorized` renders a 403.

---

## `handlers/overview.go` — Analytics

- **OverviewHandler (`GET /overview`)** — renders `overview.html` (blocked for `recruit`).
- **OverviewDataHandlerDepartments (`GET /api/overview/departments`)** — employees grouped by department (JSON).
- **OverviewDataHandlerEmployees (`GET /api/overview/employees`)** — new candidates per day plus a merged fired series (rejected + terminations) (JSON).

---

## `handlers/report.go` — Reports

- **ReportHandler (`GET /report`)** — lists reports with aggregated totals (absences/hired/fired).
- **ReportNewHandler / CreateReportHandler** — build a report over a period (`week`/`month`/`year`).
- **ReportDetailHandler (`GET /report/:id`)** — narrative summary + stats for a report.
- **ReportAbsencesContent / ReportHiredContent / ReportFiredContent** — JSON API endpoints for live counts.

See [Reports & Analytics](./reports.md).

---

## `handlers/config.go` — Account Management

Profile update, password change (requires current password + bcrypt), account deletion (destroys session), profile photo (Cloudinary), and a device page that infers browser/platform from the `User-Agent`.

---

## `handlers/password_reset_token.go` — Recovery

`PasswordResetToken{ID, Email, Token, ExpiresAt}`. A one-time token with an expiry, created on recovery request and consumed on reset. See [Authentication](./authentication.md).

---

## Error Handling Pattern

Handlers follow a consistent pattern: **validate input → attempt operation → on error, re-render the form with a user-facing message**. Raw 4xx/5xx JSON responses are used for API endpoints; template-rendered responses always give the user an actionable message. Success and error paths are logged with `log/slog`.