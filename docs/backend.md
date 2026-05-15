# Backend — Handlers, Routing & Business Logic

## Structure

The backend is split between `backend/handlers/` — where HTTP request/response logic lives — and `internal/` — where infrastructure and cross-cutting concerns live. This split is intentional and important: handlers should be thin translators between HTTP and business logic, not fat blobs that mix SQL queries, cookie manipulation, and template rendering in one function.

## `backend/handlers/auth.go`

This file owns the entire authentication surface: registering new accounts, processing login credentials, and destroying sessions on logout.

The **register handler** (`POST /register`) accepts a form submission with username, email, and password. It validates the inputs using `go-playground/validator/v10` struct tags, hashes the password with bcrypt, and inserts a new `User` row via GORM. If the email already exists (caught by the unique index), it re-renders the registration form with an error message rather than returning a raw 500.

The **login handler** (`POST /login`) queries the users table for a matching email, then calls `bcrypt.CompareHashAndPassword` to verify the credential. On success, it calls `auth.SetUser(...)` to write the session cookie, then redirects to `/dashboard`. On failure, it re-renders the login form with a generic "invalid credentials" message — deliberately vague to avoid confirming whether the email exists.

The **logout handler** (`POST /logout`) calls `auth.Logout(...)`, which sets the cookie's `MaxAge` to `-1`, instructing the browser to delete it immediately, then redirects to `/login`.

## `backend/handlers/employee.go`

This is the most complex handler file. It handles five distinct operations:

**List + Search** (`GET /employees`) builds a conditional GORM query based on the `q` query parameter, applies `OFFSET`/`LIMIT` pagination, and passes both the result set and pagination metadata to the `employees.html` template. The count query runs on the same filtered condition so pagination totals are always accurate.

**Create** (`POST /employees`) reads a multipart form (because the form includes a file upload for the profile photo), validates required fields, runs MIME type validation on the uploaded file using `gabriel-vasile/mimetype` to ensure only image types are accepted, saves the file to `backend/uploads/` with a UUID-based filename (preventing collisions and path traversal attacks), and inserts the `Employee` record. If no photo is uploaded, `PhotoURL` is left empty and the template falls back to a default avatar.

**Status Update** (`PUT /employees/:id/status`) accepts the candidate ID from the route parameter and a `status` field from the JSON body. It validates that the status is one of the three allowed values (`Pending`, `Accepted`, `Rejected`) before writing to the database — rejecting arbitrary string values at the application layer rather than relying solely on the database.

**Card View** (`GET /employees/:id/card`) fetches a single employee record and renders the `id-card.html` template, which displays a formatted candidate profile card suitable for printing or sharing.

## `backend/cmd/` — CLI Utilities

The `cmd/` directory contains standalone runnable programs that share the same `internal/db` connection. This is a common Go pattern: multiple binaries from one module, each with its own `main.go`.

`cmd/seed_users/main.go` creates a set of test user accounts for local development. `cmd/seed_employee/main.go` creates test candidate records. `cmd/list_users/main.go` prints all user accounts — useful for verifying the database state without needing a database GUI. Having these as separate binaries means they can be run independently without starting the HTTP server.

## Routing in `cmd/server/main.go`

The router setup in `main.go` follows the Gin best-practice pattern of grouping routes by their middleware requirements. Public routes (login, register) are registered directly on the root engine. Protected routes are registered on a `Group` that has `middleware.RequireAuth()` applied. Static file serving for CSS and uploads is also registered here.

```go
r := gin.Default()

// Load HTML templates from the templates directory
r.LoadHTMLGlob("backend/templates/*")

// Serve CSS files
r.Static("/css", "./frontend/css")

// Serve uploaded profile photos
r.Static("/uploads", "./backend/uploads")

// Public routes
r.GET("/login", handlers.LoginPage)
r.POST("/login", handlers.Login(db))
r.GET("/register", handlers.RegisterPage)
r.POST("/register", handlers.Register(db))

// Protected routes — RequireAuth middleware applied to the whole group
protected := r.Group("/")
protected.Use(middleware.RequireAuth())
{
    protected.GET("/dashboard", handlers.Dashboard(db))
    protected.GET("/employees", handlers.ListEmployees(db))
    protected.POST("/employees", handlers.CreateEmployee(db))
    protected.PUT("/employees/:id/status", handlers.UpdateStatus(db))
    protected.GET("/employees/:id/card", handlers.EmployeeCard(db))
    protected.POST("/logout", handlers.Logout)
}

r.Run(":" + port)
```

## Input Validation

The `go-playground/validator/v10` library provides struct-tag-based validation. Fields marked `binding:"required"` are automatically checked by Gin's `ShouldBind` / `ShouldBindJSON` calls. Custom validations (like checking that `Status` is one of an allowed set) are implemented as validator functions registered at startup.

## Error Handling Pattern

Handlers follow a consistent pattern: validate input, attempt the operation, handle the error by re-rendering the form with a user-facing message rather than returning a raw HTTP error code. Raw 4xx/5xx responses are reserved for API endpoints (future REST layer). Template-rendered responses always give the user an actionable message.