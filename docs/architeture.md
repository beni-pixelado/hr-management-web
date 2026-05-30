# Architecture — HR Management Web

## Overview

HR Management Web is structured around a classic **layered server-side architecture** where each layer has a single, well-defined responsibility. Understanding this layering is the key to navigating the codebase and knowing where to add new functionality.

```
┌─────────────────────────────────────────────────────────────┐
│                   Browser (Client)                           │
│         HTML rendered by Go templates + CSS + JS             │
└─────────────────────────────┬───────────────────────────────┘
                              │  HTTP (GET/POST/DELETE)
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Gin HTTP Layer                             │
│         Router · Middleware Stack · Static Files             │
│                  backend/cmd/server                          │
└──────┬──────────────────────┬──────────────────┬────────────┘
       │                      │                  │
       ▼                      ▼                  ▼
┌─────────────┐     ┌──────────────────┐  ┌──────────────────┐
│  auth.go    │     │  employee.go     │  │ departament.go   │
│  handler    │     │   handler        │  │   handler        │
│ (login,     │     │ (CRUD, search,   │  │ (create, list,   │
│  register,  │     │  status update,  │  │  manage members) │
│  logout)    │     │  pagination)     │  │                  │
└──────┬──────┘     └──────┬───────────┘  └──────┬───────────┘
       │                   │                     │
       └───────────────────┼─────────────────────┘
                           │  calls internal packages
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                  internal/ packages                          │
│                                                              │
│  ┌──────────────┐  ┌─────────────┐  ┌────────────────────┐ │
│  │ internal/db  │  │internal/auth│  │internal/middleware │ │
│  │  GORM setup  │  │session R/W  │  │  RequireAuth gate  │ │
│  │  pool config │  │             │  │  RedirectIfAuth    │ │
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

The Gin engine is created here. All routes are registered, the database connection is initialized, templates and static files are loaded, and the HTTP server is started. It is deliberately thin — it wires dependencies together but contains no business logic. Think of it as the composition root of the application.

### 2. Handlers — `backend/handlers/`

Handlers are the controller layer. They receive `*gin.Context`, extract input (path params, query params, form values), call internal packages for data access, and render templates or return HTTP responses. There are three handler files:

- `auth.go` — owns the authentication flow (login, register, logout)
- `employee.go` — owns all candidate operations (create, list, search, filter, status update, delete, card view)
- `departament.go` — owns all department operations (create, list, member management)

### 3. Internal packages — `internal/`

The `internal/` directory holds cross-cutting concerns that multiple handlers share. Go's `internal` visibility rule ensures these packages cannot be imported by code outside the module.

- `internal/db` — GORM initialization and connection pool configuration
- `internal/auth` — session read/write helpers (wrapping `gorilla/sessions`)
- `internal/middleware` — `RequireAuth` and `RedirectIfAuthenticated` Gin middleware

### 4. Templates — `backend/templates/`

Templates use Go's `html/template` package, which provides auto-escaping of HTML output by design — preventing XSS by default. Gin loads them at startup via `r.LoadHTMLGlob("backend/templates/*")` and renders them by name.

### 5. Static Assets — `frontend/`

CSS files are served as static assets by Gin's `r.Static("/css", "frontend/css")` directive. Uploaded photos are served from `r.Static("/uploads", "./uploads")`. No build steps, no bundlers.

## Request Lifecycle

A typical authenticated request follows this path:

1. Browser sends `GET /employees?q=engineer&page=2`.
2. Gin's router matches the path and checks the middleware stack — `RequireAuth` runs first.
3. `RequireAuth` calls `internal/auth.IsAuthenticated(c)`, which reads the session cookie. If valid, it calls `c.Next()`; if not, it redirects to `/login` and aborts.
4. The `employee.go` handler extracts `c.Query("q")` and `c.Query("page")`, builds a GORM query with `ILIKE` conditions, applies `Offset`/`Limit`, and executes.
5. GORM sends a parameterized SQL statement to Neon via the `pgx` driver.
6. The result set is passed to `c.HTML(200, "employees.html", gin.H{...})`.
7. Gin renders the template and writes the response body.

## Why No REST API (Yet)

The current architecture uses server-rendered HTML. This eliminates the need for a separate frontend framework, keeps the deployment artifact to a single Go binary, and reduces cognitive overhead. The v2.0 Roadmap item adds a full REST API layer — but the internal handler logic will remain largely unchanged, as business logic is already separated from the rendering concern.