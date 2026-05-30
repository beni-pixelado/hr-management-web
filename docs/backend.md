# Backend — Handlers, Routing & Business Logic

## Structure

The backend is split between `backend/handlers/` — where HTTP request/response logic lives — and `internal/` — where infrastructure and cross-cutting concerns live. Handlers call internal packages but not vice versa; this one-way dependency keeps the system testable and the layers independent.

## `backend/handlers/auth.go`

Owns the entire authentication surface: registering new accounts, processing login credentials, and destroying sessions on logout.

### Register (`POST /register`)

Accepts a form submission with `username`, `email`, and `password`. Validates all fields are present, checks for an existing username via a GORM `Where` query, hashes the password with `bcrypt.GenerateFromPassword`, and inserts a new `User` row. On error, re-renders `register.html` with a user-facing message rather than returning a raw HTTP error code.

### Login (`POST /login`)

Queries the users table with `username AND email`, then calls `bcrypt.CompareHashAndPassword` to verify the credential. On success, calls `auth.CreateSession(c, user.ID)` to write the signed session cookie, then redirects to `/dashboard`. On failure, renders a generic error message — deliberately vague to avoid confirming whether the username/email pair exists.

### Logout (`GET /logout`)

Calls `auth.DestroySession(c)`, which sets the session cookie's `MaxAge` to `-1`, instructing the browser to delete it immediately, then redirects to `/login`.

---

## `backend/handlers/employee.go`

The most complex handler file. Handles six distinct operations.

### GetEmployees (`GET /employees`)

Builds a conditional GORM query based on the optional `q` query parameter. When `q` is present, adds a `WHERE` clause with `ILIKE` matching across `full_name`, `position`, and `email`. Applies `OFFSET`/`LIMIT` pagination (20 per page by default). The `Count` query runs on the same filtered condition, ensuring pagination totals always reflect the filtered result set.

```go
query := DB.Model(&Employee{})

if search := strings.TrimSpace(c.Query("q")); search != "" {
    term := "%" + search + "%"
    query = query.Where(
        "full_name ILIKE ? OR position ILIKE ? OR email ILIKE ?",
        term, term, term,
    )
}

var total int64
query.Count(&total)

var employees []Employee
query.Offset(offset).Limit(limit).Find(&employees)
```

### CreateEmployee (`POST /employees`)

Reads a multipart form (required for the photo upload), validates that `full_name`, `email`, and `position` are present. If a file is provided:

1. Checks file size (`MaxFileSize = 5 MB`)
2. Detects MIME type via file extension and `http.DetectContentType` as a fallback
3. Validates that the MIME type is in the allowed list (`image/jpeg`, `image/png`, `image/gif`, `image/webp`)
4. Generates a UUID-based filename to prevent collisions and path traversal
5. Saves to `./uploads/`

Inserts the `Employee` row with `Status` defaulting to `"pending"`.

### UpdateEmployeeStatus (`POST /employees/:id/status`)

Accepts `status` and optional `hire_date` from the form body. Updates both fields in a single GORM `Updates` call using a map (not a struct, to allow zero-value updates). Redirects back to `/employees` on success.

### DeleteEmployee (`DELETE /employees/:id`)

Performs a GORM soft-delete (sets `DeletedAt`) on the employee matching the route `:id` parameter. Returns JSON on both success and error.

### BadgeHandler (`GET /employees/:id/card`)

Fetches a single employee record with `DB.First` and renders `id-card.html`. Returns a plain `404` string if the ID is not found.

### GetEmployeesAPI (`GET /api/employees`)

JSON-only endpoint. Supports `?search=` and `?status=` query parameters for filtering. Returns a JSON envelope with `employees` array and `total` count. Intended for the future REST API layer.

---

## `backend/handlers/departament.go`

Handles all department operations.

### DepartmentPageHandler (`GET /department`)

Fetches all employees (to populate the manager `<select>` dropdown) and all departments in two separate queries, then renders `departments.html` with both datasets. This is a single-request page load — no AJAX.

### CreatedepartmentHandler (`POST /department`)

Reads `name`, `code`, and optional `boss_id` from the form. Validates that `name` and `code` are not empty. Parses `boss_id` as `uint64` (returns a 400 if the string is non-empty but invalid). Creates the `Department` row and redirects to `/department`.

```go
department := Department{
    Name:   Name,
    Code:   Code,
    BossID: bossID,
}
DB.Create(&department)
```

### DepartmentHandler (paginated listing)

Handles `GET /department` with pagination support (20 per page). Computes `totalPages` and passes pagination metadata to the template for next/prev link generation.

---

## Routing in `backend/cmd/server/main.go`

```go
r := gin.Default()
r.SetFuncMap(template.FuncMap{
    "lower": strings.ToLower,
    "add":   func(a, b int) int { return a + b },
})
r.LoadHTMLGlob("backend/templates/*")
r.Static("/css", "frontend/css")
r.Static("/uploads", "./uploads")

// Public routes
r.GET("/login",    ...)
r.POST("/login",   handlers.Login)
r.GET("/register", ...)
r.POST("/register", handlers.Register)
r.GET("/", func(c *gin.Context) { c.Redirect(302, "/login") })

// Protected routes
protected := r.Group("/")
protected.Use(middleware.RequireAuth)
{
    protected.GET("/dashboard",               dashboardHandler)
    protected.GET("/employees",               handlers.GetEmployees)
    protected.POST("/employees",              handlers.CreateEmployee)
    protected.POST("/employees/:id/status",   handlers.UpdateEmployeeStatus)
    protected.DELETE("/employees/:id",        handlers.DeleteEmployee)
    protected.GET("/department",              handlers.DepartmentPageHandler)
    protected.POST("/department",             handlers.CreatedepartmentHandler)
    protected.GET("/logout",                  handlers.Logout)
}
```

The `protected` group applies `middleware.RequireAuth` to every route inside it. Adding a new protected route is a matter of registering it inside the group — no per-handler auth checks needed.

## Template Functions

Two custom template functions are registered at startup:

| Function | Usage | Purpose |
|---|---|---|
| `lower` | `{{.Status \| lower}}` | Maps status strings to lowercase CSS class names |
| `add` | `{{add .currentPage 1}}` | Offset arithmetic for pagination link hrefs |

## Error Handling Pattern

Handlers follow a consistent pattern: validate input → attempt the operation → on error, re-render the form with a user-facing message. Raw 4xx/5xx responses are used only for API endpoints. Template-rendered responses always give the user an actionable message.