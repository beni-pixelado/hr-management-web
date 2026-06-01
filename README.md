<div align="center">

# 🏢 HR Management Web

### A production-minded Human Resources Management System built with Go, Gin, PostgreSQL and Gorilla Sessions

[![Version](https://img.shields.io/badge/version-2.1.0-blue.svg?style=for-the-badge)](https://github.com/beni-pixelado/hr-management-web)
[![Status](https://img.shields.io/badge/status-active-brightgreen.svg?style=for-the-badge)](https://github.com/beni-pixelado/hr-management-web)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-4169E1.svg?style=for-the-badge&logo=postgresql&logoColor=white)](https://neon.tech)
[![Gin](https://img.shields.io/badge/Gin-1.12-008ECF.svg?style=for-the-badge&logo=go&logoColor=white)](https://gin-gonic.com)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=for-the-badge)](./LICENSE)

<br/>

> A full-stack HR candidate management platform that streamlines the recruitment pipeline — from candidate intake to final status resolution — featuring department management, a modern dark UI[...]

<br/>

[Overview](#-overview) · [Features](#-features) · [New UI](#-new-ui--design-system) · [Department Management](#-department-management) · [Search Engine](#-search-engine) · [Database](#-sqlite[...]

</div>

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [New UI & Design System](#-new-ui--design-system)
- [Department Management](#-department-management)
- [Search Engine](#-search-engine)
- [SQLite → PostgreSQL via Neon](#-sqlite--postgresql-via-neon)
- [Authentication & Session Management](#-authentication--session-management)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Project Structure](#-project-structure)
- [Candidate Lifecycle](#-candidate-lifecycle)
- [Installation](#-installation)
- [Environment Variables](#-environment-variables)
- [Running Locally](#-running-locally)
- [API Routes](#-api-routes)
- [CLI Utilities](#-cli-utilities)
- [Testing](#-testing)
- [Roadmap](#-roadmap)
- [License](#-license)

---

## 🔍 Overview

**HR Management Web** is a self-contained Human Resources platform built in Go, designed to manage candidate pipelines and organizational structure with clarity, efficiency, and scalability. The s[...]

The architecture follows a clean separation of concerns: the `backend/` layer handles HTTP routing and business logic through dedicated handlers, `internal/` packages encapsulate cross-cutting con[...]

The result is an application that is **portable, easy to extend, and ready for cloud deployment** — with zero client-side framework complexity and a Go binary as the single deployable artifact.

---

## ✨ Features

**Candidate Management** covers adding new candidates with name, job position, email, and optional profile photo upload. Photos are stored in `uploads/` and served statically, with an automatic fa[...]

**Department Management** allows HR teams to organize their workforce into departments. Each department has a name, a unique code, and an optional assigned manager. The module supports:
- Creating new departments via a form with name, code, and manager selection
- Listing all departments in a visual card grid
- Assigning a manager (selected from existing employees/candidates)
- Deleting departments
- Adding and removing collaborators from departments

**Status Pipeline** implements a three-state system (`Pending` → `Contractors` / `Rejected`) that is manually controlled by HR staff. Status changes are reflected immediately across both the tab[...]

**Search Engine** provides multi-field, case-insensitive, server-side search across name, position, and email — with pagination preservation and URL-bookmarkable results.

**Dashboard Metrics** give HR staff a real-time snapshot of recruitment health, displaying total counts for contractors, rejected, and pending candidates via prominently styled KPI cards.

**Authentication** provides user registration and login backed by session cookies, with middleware-enforced route protection across all sensitive routes.

**ID Card View** renders a formatted candidate profile card per employee — print-friendly and suitable for sharing.

---

## 🎨 New UI & Design System

One of the most significant improvements in v1.1 is a complete visual overhaul. The new design system introduces a modern aesthetic that prioritizes readability, structure, and user confidence.

### Design Philosophy

The interface was designed around three principles: **clarity** (every element has a clear purpose and visual weight), **density** (HR tools are data-heavy — the layout maximizes information per[...]

### Login Page

The login screen features a soft light-gray gradient background (`#e2e8f0`) with a centered white card (`border-radius: 24px`, layered `box-shadow`). The form uses uppercase field labels, rounded [...]

```
Background: #e2e8f0 gradient  →  calm, enterprise feel
Card: white + shadow           →  focus isolation
Primary button: #3f51b5        →  brand indigo, high contrast
Ghost button: border-only      →  de-emphasized without hidden
```

### Dashboard

The dashboard is the command center. It uses a sidebar navigation on the left with icon+label pairs and active-state highlighting. The main content area displays KPI cards (Contractors, Rejected,[...]

### Employees Page

The employees page combines a filterable table with status badges color-coded as chips — green, red, and amber — providing instant visual feedback. Profile photos appear as rectangular thumbn[...]

### Departments Page

The departments page presents created departments in a responsive card grid. Each card shows the department name, a colored code badge, and the assigned manager ID. The creation form includes a s[...]

### Responsiveness

All pages are built with relative units (`rem`, `%`, `vh`) and CSS media queries. The sidebar collapses to a top navigation bar on narrow viewports, and table views gracefully reduce to card-only[...]

---

## 🗂️ Department Management

The department module is one of the core v1.1 additions. It allows HR teams to create organizational structure alongside the candidate pipeline.

### How It Works

Departments are independent entities. Each has a `name`, a unique `code`, and an optional `boss_id` referencing an employee. The module is exposed at the `/department` route with both GET (list +[...]

```go
type Department struct {
    ID     uint   `gorm:"primaryKey"`
    Code   string `gorm:"not null"`
    Name   string `gorm:"not null"`
    BossID uint   `gorm:"column:boss"`
}
```

### Operations

**Creating a Department** — The creation form accepts a name, code, and an optional manager selected from a dropdown of all existing employees. On submission, the handler validates required fie[...]

```go
func CreatedepartmentHandler(c *gin.Context) {
    Code   := c.PostForm("code")
    Name   := c.PostForm("name")
    BossID := c.PostForm("boss_id")
    // validates, parses, inserts, redirects
}
```

**Listing Departments** — `DepartmentPageHandler` fetches all employees (for the manager dropdown) and all departments, then renders the full departments view in a single request.

**Deleting a Department** — Planned for v1.2. Will use `DELETE /department/:id`.

**Adding / Removing Collaborators** — Planned for v1.2. Will introduce a join table (`department_employees`) linking the `departments` and `employees` tables many-to-many, with dedicated `POST [...]

### The Departments Page

The page is split into two sections: the **Add New Department** form at the top, and the **Departments Created** card grid below it. The manager select is dynamically populated from the `{{range [...]

---

## 🔎 Search Engine

The search system allows HR staff to locate candidates instantly without paginating through long lists.

### How It Works

Search is implemented as a server-side query on the `GET /employees` route via a `q` parameter. The handler sanitizes the input and builds a parameterized SQL query using GORM's `Where` clause ✔[...]

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

### Multi-field Search

A single search term matches across three fields simultaneously: **full name**, **position**, and **email**. Typing `"eng"` surfaces candidates named `"Enrique"`, candidates with position `"Engin[...]

### Frontend Integration

The search input is a plain HTML `<form>` with `method="get"`. Submitting it produces a `GET /employees?q=...` request — no JavaScript required. The `value="{{ .search }}"` binding echoes the s[...]

### Pagination Integration

Search integrates cleanly with pagination. Every page link preserves the `q` parameter, so navigating to page 2 of `"engineer"` results does not lose the search context. This is achieved by forwa[...]

### Security

Parameterized queries make SQL injection structurally impossible. The input is also `TrimSpace`'d before use. Future hardening will add a maximum length cap and rate limiting per session.

---

## 🗄️ SQLite → PostgreSQL via Neon

This migration is one of the most important architectural decisions in the project's evolution.

### Why SQLite Was a Good Starting Point

SQLite requires zero configuration, stores everything in a single file (`data/users.db`), and the Go driver works out of the box. For validating a data model and a UI, it's the right call. GORM's[...]

### Why SQLite Becomes a Bottleneck

SQLite has a single-writer constraint. In a web application with concurrent HTTP requests, writes queue behind each other. It also can't be accessed by more than one process simultaneously — ru[...]

### Why PostgreSQL

PostgreSQL supports true concurrent reads and writes via MVCC, full-text search operators (`ILIKE`, `tsvector`), `JSONB`, row-level locking, and a rich extension ecosystem. Its `ILIKE` operator w[...]

### Why Neon

[Neon](https://neon.tech) is a serverless PostgreSQL platform that provides:

| Feature | Benefit |
|---|---|
| **Serverless scaling** | Scales to zero when idle — no cost for unused compute |
| **Branching** | Isolated DB branches for staging/testing |
| **Connection pooling** | Built-in PgBouncer-compatible pooler |
| **Cloud-native** | Works with Render, Railway, Fly.io |
| **Free tier** | Generous free tier for portfolio and small production apps |

### Connection Setup in Go

```go
// internal/db/db.go
func Connect() (*gorm.DB, error) {
    dsn := os.Getenv("DATABASE_URL")
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(10)
    sqlDB.SetMaxIdleConns(5)

    return db, nil
}
```

`MaxOpenConns(10)` prevents connection exhaustion on Neon's free tier. `MaxIdleConns(5)` avoids TCP+TLS reconnection overhead on every request.

### Local vs Production

For local development, point `DATABASE_URL` at a local PostgreSQL instance or a Neon branch. The SQLite driver remains in `go.mod` for reference and test isolation.

---

## 🔐 Authentication & Session Management

Authentication is handled via **cookie-based sessions** using `gorilla/sessions`, backed by `gorilla/securecookie` for HMAC-signed cookie values.

### How Sessions Work

When a user successfully logs in, the server creates a session entry, sets an authenticated flag and the user's ID inside the session store, and writes a `Set-Cookie` header. Every subsequent req[...]

```go
// internal/auth/session.go
func CreateSession(c *gin.Context, userID uint) error {
    session, _ := SessionStore.Get(c.Request, "hr_session")
    session.Values["user_id"]       = int(userID)
    session.Values["authenticated"] = true
    session.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   60 * 60 * 24 * 7,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
    return session.Save(c.Request, c.Writer)
}

func IsAuthenticated(c *gin.Context) (bool, uint) {
    session, err := SessionStore.Get(c.Request, "hr_session")
    if err != nil { return false, 0 }
    auth, ok := session.Values["authenticated"].(bool)
    if !ok || !auth { return false, 0 }
    userID := uint(session.Values["user_id"].(int))
    return true, userID
}
```

### Middleware

Two middleware functions are provided in `internal/middleware/auth.go`:

**`RequireAuth`** — applied to all protected routes. If the session is missing or invalid, the user is redirected to `/login` with a `302`.

**`RedirectIfAuthenticated`** — applied to `/login` and `/register`. Prevents already-logged-in users from seeing the auth pages.

```go
func RequireAuth(c *gin.Context) {
    authenticated, userID := auth.IsAuthenticated(c)
    if !authenticated {
        c.Redirect(http.StatusFound, "/login")
        c.Abort()
        return
    }
    c.Set("user_id", userID)
    c.Next()
}
```

### Security Considerations

The session cookie is HMAC-signed with the `SESSION_SECRET` key — tampering is detected and the session is invalidated. Passwords are stored hashed with `bcrypt` (`golang.org/x/crypto`). HTTPS [...]

---

## 🛠️ Tech Stack

### Backend

| Technology | Version | Role |
|---|---|---|
| **Go** | 1.21+ | Core application language |
| **gin-gonic/gin** | v1.12.0 | HTTP framework, routing, middleware, template rendering |
| **gorm.io/gorm** | v1.31.1 | ORM — schema migration, query building, model binding |
| **gorm.io/driver/postgres** | v1.6.0 | GORM PostgreSQL adapter |
| **jackc/pgx/v5** | v5.6.0 | PostgreSQL wire protocol driver |
| **jackc/puddle/v2** | v2.2.2 | Connection pool manager for pgx |
| **gorilla/sessions** | v1.4.0 | Cookie-based session management |
| **gorilla/securecookie** | v1.1.2 | HMAC cookie signing and optional encryption |
| **joho/godotenv** | v1.5.1 | `.env` file loading for local development |
| **google/uuid** | v1.6.0 | UUID generation for uploaded photo filenames |
| **gabriel-vasile/mimetype** | v1.4.12 | MIME type detection for uploaded photos |
| **go-playground/validator/v10** | v10.30.1 | Struct-level input validation |
| **golang.org/x/crypto** | v0.50.0 | `bcrypt` password hashing |

### Frontend

| Technology | Role |
|---|---|
| **HTML5 + html/template** | Server-side rendering through Gin's template engine |
| **CSS3 (vanilla)** | Custom design system — no external CSS frameworks |
| **JavaScript (vanilla)** | Client-side interactivity (delete confirmations, UI toggles) |

### Infrastructure & Tooling

| Tool | Role |
|---|---|
| **Neon** | Serverless PostgreSQL cloud database |
| **gorm.io/driver/sqlite** | SQLite driver (retained for local/test use) |
| **mattn/go-sqlite3** | CGO-based SQLite3 driver |
| **stretchr/testify** | Test assertions |
| **go.uber.org/mock** | Mock generation for unit tests |
| **Makefile** | Task automation (`run`, `test`, `seed`) |

---

## 🏗️ Architecture

The application follows a layered architecture with clear boundaries between concerns:

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Layer (Gin)                    │
│          Routes · Middleware · Template Rendering         │
└────────────────────────┬────────────────────────────────┘
                         │
          ┌───────────────┼──────────────┐
          │               │              │
          ▼               ▼              ▼
    ┌──────────┐   ┌──────────────┐  ┌──────────────┐
    │  auth.go │   │ employee.go  │  │departament.go│
    │ handler  │   │   handler    │  │   handler    │
    └────┬─────┘   └──────┬───────┘  └──────┬───────┘
         │                │                  │
         └────────────────┼──────────────────┘
                          │
                          ▼
               ┌──────────────────────┐
               │    internal/         │
               │  db · auth · middle  │
               └──────────┬───────────┘
                          │
                          ▼
               ┌──────────────────────┐
               │  Neon PostgreSQL     │
               │  (cloud, serverless) │
               └──────────────────────┘
```

The `internal/` packages (`auth`, `db`, `middleware`) are deliberately isolated from `backend/handlers/` — handlers call internal packages but not vice versa. The database connection, session l[...]

---

## 📁 Project Structure

```
hr-management-web/
├── go.mod
├── go.sum
├── makefile
│
├── backend/
│   ├── cmd/
│   │   ├── server/main.go          # Application entrypoint
│   │   ├── seed_users/main.go      # Seeds test user accounts
│   │   ├── seed_employee/main.go   # Seeds test candidate data
│   │   ├── list_users/main.go      # CLI: prints all users
│   │   └── fix_sequences/main.go  # Fixes PostgreSQL serial sequences
│   │
│   ├── database/
│   │   └── database.go             # GORM + PostgreSQL connection init
│   │
│   ├── handlers/
│   │   ├── auth.go                 # Login, register, logout handlers
│   │   ├── employee.go             # Candidate CRUD + search + status
│   │   └── departament.go          # Department CRUD handlers
│   │
│   └── templates/
│       ├── login.html
│       ├── register.html
│       ├── dashboard.html
│       ├── employees.html
│       ├── departments.html
│       └── id-card.html
│
├── internal/
│   ├── auth/
│   │   └── session.go              # Session read/write helpers
│   ├── db/
│   │   └── db.go                   # GORM + PostgreSQL connection
│   └── middleware/
│       └── auth.go                 # RequireAuth + RedirectIfAuthenticated
│
├── frontend/
│   └── css/
│       ├── style.css               # Global reset
│       ├── login.css
│       ├── register.css
│       ├── dashboard.css
│       ├── employees.css
│       ├── departments.css
│       └── id-card.css
│
├── data/
│   └── users.db                    # SQLite file (legacy / local dev)
│
└── docs/
    ├── architecture.md
    ├── backend.md
    ├── database.md
    ├── frontend.md
    ├── search.md
    ├── auth.md
    ├── departments.md
    ├── data-flow.md
    ├── scalability.md
    ├── testing.md
    └── roadmap.md
```

---

## 🔄 Candidate Lifecycle

Every candidate registered follows a defined lifecycle with three possible states:

```
           ┌──────────────────────────────────────────┐
           │           Candidate Registered            │
           └───────────────────┬──────────────────────┘
                               │
                               ▼
                       ┌───────────────┐
                       │    PENDING    │  ◄── Default on creation
                       └───────┬───────┘
                               │
              ┌────────────────┴────────────────┐
              │                                 │
              ▼                                 ▼
      ┌───────────────┐                 ┌───────────────┐
      │  CONTRACTORS  │                 │   REJECTED    │
      └───────────────┘                 └───────────────┘
```

Status transitions are **bidirectional and manually controlled** — HR staff can move a candidate from any state to any other at any time.

---

## ⚙️ Installation

### Prerequisites

Go `>= 1.21`, Git, and a PostgreSQL connection string (Neon free tier works perfectly).

```bash
git clone https://github.com/beni-pixelado/hr-management-web.git
cd hr-management-web
go mod download
```

---

## 🔑 Environment Variables

Create a `.env` file in the project root:

```env
# PostgreSQL connection string (Neon or local)
DATABASE_URL=postgres://user:password@host/dbname?sslmode=require

# Secret key for signing session cookies — use a long random string
SESSION_SECRET=your-super-secret-key-min-32-chars

# Server port (optional, defaults to 8000)
PORT=8000
```

For local development with a local PostgreSQL instance:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/hr_dev?sslmode=disable
```

---

## 🚀 Running Locally

```bash
# Copy and fill environment variables
cp .env.example .env

# (Optional) Seed the database with test users
go run ./backend/cmd/seed_users

# (Optional) Seed candidate data
go run ./backend/cmd/seed_employee

# Start the server
make run
# or: go run ./backend/cmd/server
```

The application will be available at `http://localhost:8000`.

---

## 🗺️ API Routes

| Method | Route | Auth Required | Description |
|---|---|---|---|
| `GET` | `/` | No | Redirect to `/login` |
| `GET` | `/login` | No | Login page |
| `POST` | `/login` | No | Process login |
| `GET` | `/register` | No | Registration page |
| `POST` | `/register` | No | Create account |
| `GET` | `/logout` | Yes | Destroy session |
| `GET` | `/dashboard` | Yes | Metrics overview + employee grid |
| `GET` | `/employees` | Yes | Candidate list (`?q=` and `?page=`) |
| `POST` | `/employees` | Yes | Create new candidate |
| `POST` | `/employees/:id/status` | Yes | Update candidate status |
| `DELETE` | `/employees/:id` | Yes | Delete candidate |
| `GET` | `/employees/:id/card` | Yes | Candidate ID card view |
| `GET` | `/department` | Yes | Department list + creation form |
| `POST` | `/department` | Yes | Create new department |

---

## 🧰 CLI Utilities

The `cmd/` directory contains standalone runnable programs that share the same database connection — a common Go pattern for operational tooling.

```bash
# Seed 100 test user accounts
go run ./backend/cmd/seed_users

# Seed 50 test candidate records
go run ./backend/cmd/seed_employee

# Print all users in the database
go run ./backend/cmd/list_users

# Fix PostgreSQL serial sequences after bulk imports
go run ./backend/cmd/fix_sequences
```

---

## 🧪 Testing

Integration tests validate the core application flows: candidate creation, status transitions, and authentication. The test suite uses `stretchr/testify` for assertions and `go.uber.org/mock` for[...]

```bash
make test
```

Tests run against an isolated test database instance. The `jordanlewis/gcassert` package provides compile-time assertions for performance-critical code paths.

---

## 🗺️ Roadmap

### v1.1 — Current (Quality of Life)
- [x] Multi-field search with PostgreSQL `ILIKE`
- [x] Pagination for the candidates table
- [x] Redesigned UI with modern design system
- [x] Session-based authentication with gorilla/sessions
- [x] Department creation and listing module

### v1.2 — Control (New Functions)
- [ ] Department deletion
- [ ] Add / remove collaborators from departments
- [ ] Role-based access control (Admin, Recruiter, Viewer)
- [ ] Account management (change password, deactivate)

### v1.3 — Enhanced Data Model
- [ ] Candidate notes and comments
- [ ] Interview scheduling and date tracking
- [ ] Department and team assignment
- [ ] Audit trail for status change history

### v2.0 — Architecture Evolution
- [ ] Full REST API with OpenAPI/Swagger documentation
- [ ] HTMX-powered frontend for partial page updates
- [ ] Docker + Docker Compose for containerized deployment
- [ ] CI/CD pipeline via GitHub Actions
- [ ] GIN trigram index on search columns for `ILIKE` at scale

---

## 📄 License

This project is licensed under the **MIT License**. See the [LICENSE](./LICENSE) file for full details.

---

<div align="center">

Built with ❤️ using Go + Gin + PostgreSQL · Designed for portfolio and production alike

⭐ Star the repo, please, i need buy food

</div>
