# Architecture — HR Management Web

## Overview

HR Management Web is structured around a classic **layered server-side architecture** where each layer has a single, well-defined responsibility. Understanding this layering is the key to navigating the codebase and knowing where to add new functionality.

```
┌─────────────────────────────────────────────────────────────┐
│                   Browser (Client)                           │
│         HTML rendered by Go templates + CSS + JS             │
└─────────────────────────────┬───────────────────────────────┘
                              │  HTTP (GET/POST/PUT)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Gin HTTP Layer                             │
│         Router · Middleware Stack · Static Files             │
│                  backend/cmd/server                          │
└──────┬──────────────────────┬───────────────────────────────┘
       │                      │
       ▼                      ▼
┌─────────────┐     ┌──────────────────┐
│  auth.go    │     │  employee.go      │
│  handler    │     │  handler          │
│ (login,     │     │ (CRUD, search,    │
│  register,  │     │  status update,   │
│  logout)    │     │  pagination)      │
└──────┬──────┘     └────────┬─────────┘
       │                     │
       └──────────┬──────────┘
                  │  calls internal packages
                  ▼
┌─────────────────────────────────────────────────────────────┐
│                  internal/ packages                          │
│                                                              │
│  ┌──────────────┐  ┌─────────────┐  ┌────────────────────┐ │
│  │ internal/db  │  │internal/auth│  │internal/middleware │ │
│  │  GORM setup  │  │session R/W  │  │  RequireAuth gate  │ │
│  │  pool config │  │             │  │                    │ │
│  └──────┬───────┘  └─────────────┘  └────────────────────┘ │
└─────────┼───────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│              Neon PostgreSQL (cloud)                         │
│         Serverless · Connection Pooled · Persistent          │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Entrypoint — `backend/cmd/server/main.go`

This is where the Gin engine is created, all routes are registered, the database connection is initialized, and the HTTP server is started. It is deliberately thin — it wires dependencies together but contains no business logic. Think of it as the composition root of the application.

### 2. Handlers — `backend/handlers/`

Handlers are the controller layer. They receive `*gin.Context`, extract input (path params, query params, form values), call internal packages for data access, and render templates or return HTTP responses. There are two handler files: `auth.go` owns the authentication flow (login, register, logout) and `employee.go` owns all candidate operations (create, list, search, filter, status update, card view).

The key design decision here is that handlers are stateless — they receive a `*gorm.DB` reference (or similar) via closure or dependency injection rather than using a global. This makes them testable in isolation.

### 3. Internal packages — `internal/`

The `internal/` directory holds cross-cutting concerns that multiple handlers share. Go's `internal` visibility rule ensures these packages cannot be imported by code outside the module, preventing accidental coupling from external consumers if the module is ever published.

`internal/db` holds the GORM initialization function and connection pool configuration. `internal/auth` holds session read/write helpers (wrapping `gorilla/sessions`). `internal/middleware` holds the `RequireAuth` Gin middleware function that gates protected routes.

### 4. Templates — `backend/templates/`

Templates use Go's standard `html/template` package, which provides auto-escaping of HTML output by design — preventing XSS by default. Gin loads them at startup and renders them by name. Templates receive a `gin.H` map of data from handlers.

### 5. Static Assets — `frontend/`

CSS files are served as static assets by Gin's `Static` router directive. JavaScript is kept minimal and inline or in small files. There are no build steps, no bundlers, no transpilation — what you write is what gets served.

## Request Lifecycle

A typical authenticated request follows this path through the system:

1. The browser sends `GET /employees?q=engineer&page=2`.
2. Gin's router matches the path and checks the middleware stack — `RequireAuth` runs first.
3. `RequireAuth` calls `internal/auth.IsAuthenticated(c)`, which reads the session cookie. If valid, it calls `c.Next()`; if not, it redirects to `/login` and aborts.
4. The `employee.go` handler function executes. It reads `c.Query("q")` and `c.Query("page")`, builds a GORM query with `ILIKE` conditions, applies `Offset`/`Limit` for pagination, and executes against the PostgreSQL connection.
5. GORM translates the query to a parameterized SQL statement and sends it to Neon via the `pgx` driver.
6. The result set is passed to `c.HTML(200, "employees.html", gin.H{...})`.
7. Gin renders the template and writes the response body.

## Dependency Injection Pattern

The application uses a simple closure-based DI pattern. The database connection (`*gorm.DB`) is created once in `main.go` and passed into handler constructors that return `gin.HandlerFunc` closures:

```go
// main.go (conceptual)
db := internal_db.Connect()

r.GET("/employees", handlers.ListEmployees(db))
r.POST("/employees", handlers.CreateEmployee(db))

// handlers/employee.go
func ListEmployees(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // db is available via closure
    }
}
```

This pattern avoids global state while remaining straightforward to follow. It also makes unit testing direct — you can pass a test database (or a mock) without any framework magic.

## Why No REST API (Yet)

The current architecture uses server-rendered HTML rather than a JSON API. This is a deliberate choice for this phase: it eliminates the need for a separate frontend framework, keeps the deployment artifact to a single Go binary, and reduces cognitive overhead. The Roadmap item for v2.0 includes a full REST API layer — but the internal handler logic will remain largely unchanged, as the business logic is already separated from the rendering concern.