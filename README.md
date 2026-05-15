<div align="center">

# 🏢 HR Management Web

### A production-minded Human Resources Management System built with Go, Gin, PostgreSQL and Gorilla Sessions

[![Version](https://img.shields.io/badge/version-1.1.0-blue.svg?style=for-the-badge)](https://github.com/beni-pixelado/hr-management-web)
[![Status](https://img.shields.io/badge/status-active-brightgreen.svg?style=for-the-badge)](https://github.com/beni-pixelado/hr-management-web)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-4169E1.svg?style=for-the-badge&logo=postgresql&logoColor=white)](https://neon.tech)
[![Gin](https://img.shields.io/badge/Gin-1.12-008ECF.svg?style=for-the-badge&logo=go&logoColor=white)](https://gin-gonic.com)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=for-the-badge)](./LICENSE)

<br/>

> A full-stack HR candidate management platform that streamlines the recruitment pipeline — from candidate intake to final status resolution — featuring a modern dark UI, a powerful search engine, PostgreSQL via Neon, and session-based authentication.

<br/>

[Overview](#-overview) · [New UI](#-new-ui--design-system) · [Search Engine](#-search-engine) · [Database Migration](#-sqlite--postgresql-via-neon) · [Auth & Sessions](#-authentication--session-management) · [Tech Stack](#-tech-stack) · [Installation](#-installation) · [Roadmap](#-roadmap)

</div>

---

## 📋 Table of Contents

- [Overview](#-overview)
- [New UI & Design System](#-new-ui--design-system)
- [Search Engine](#-search-engine)
- [SQLite → PostgreSQL via Neon](#-sqlite--postgresql-via-neon)
- [Authentication & Session Management](#-authentication--session-management)
- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Project Structure](#-project-structure)
- [Candidate Lifecycle](#-candidate-lifecycle)
- [Installation](#-installation)
- [Environment Variables](#-environment-variables)
- [Running Locally](#-running-locally)
- [API Routes](#-api-routes)
- [Testing](#-testing)
- [Roadmap](#-roadmap)
- [License](#-license)

---

## 🔍 Overview

**HR Management Web** is a self-contained Human Resources platform built in Go, designed to manage candidate pipelines with clarity, efficiency, and scalability. The system evolved from a lightweight local prototype into a production-oriented application backed by a cloud PostgreSQL database, a session-based authentication system, a redesigned UI, and a fast, multi-field search engine.

The architecture follows a clean separation of concerns: the `backend/` layer handles HTTP routing and business logic through dedicated handlers, `internal/` packages encapsulate cross-cutting concerns like database connections, authentication, and middleware, and the `frontend/` layer is a server-rendered HTML/CSS interface rendered through Go's `html/template` engine via Gin.

The result is an application that is **portable, easy to extend, and ready for cloud deployment** — with zero client-side framework complexity and a Go binary as the single deployable artifact.

---

## 🎨 New UI & Design System

One of the most significant improvements in v1.1 is a complete visual overhaul of the interface. The previous UI was functional but lacked polish and visual hierarchy. The new design system introduces a dark, modern aesthetic that prioritizes readability, structure, and user confidence.

#### Inspirad in apple OS

### Design Philosophy

The new interface was designed around three principles: **clarity** (every element has a clear purpose and visual weight), **density** (HR tools are data-heavy — the layout maximizes information per screen without feeling cluttered), and **consistency** (reusable CSS classes and component patterns across every page).

### Login Page

The login screen establishes the visual identity of the application immediately. It features a soft light-gray gradient background (`#e8eaf0`) with a centered white card (`border-radius: 16px`, `box-shadow` for depth). The form uses clear field labels in uppercase, rounded pill-style input fields, and a bold indigo CTA button. A secondary "Sign up" link is styled as a ghost button to communicate its secondary priority without removing it from visibility.

```
Background: #e8eaf0 gradient  →  communicates calm, enterprise feel
Card: white + shadow           →  focus isolation, modal-style intent
Primary button: #3d52a0        →  brand indigo, high contrast
Ghost button: border-only      →  de-emphasized without being hidden
```

### Dashboard

The dashboard is the command center of the application. It uses a dark sidebar navigation on the left with icon+label pairs and clear active-state highlighting. The main content area displays KPI cards (Accepted, Rejected, Pending counts) using colored accent borders and icon badges, making it scannable at a glance. The layout is built on a CSS Grid structure, ensuring consistent alignment across all viewport sizes.

### Employees Page

The employees page combines a filterable table view with a card view toggle. Table rows use alternating subtle backgrounds for readability, and status badges (`Accepted` / `Rejected` / `Pending`) are color-coded chips — green, red, and amber respectively — providing instant visual feedback without needing to read the text. Profile photos appear as circular avatars with a gray fallback SVG when no image was uploaded.

### Responsiveness

All pages are built mobile-first using relative units (`rem`, `%`, `vh`) and CSS media queries. The sidebar collapses to a top navigation bar on narrow viewports, and table views gracefully reduce to card-only layout below `768px`. There are no external CSS frameworks — every style is handcrafted in the `frontend/css/` directory, giving full control over the design language.

### Screenshots

**Login Page**

<img src="docs/assets/rh-login.png" width="800" style="border-radius: 12px;" />

**Dashboard**

<img src="docs/assets/rh-dashboard.png" width="800" style="border-radius: 12px;" />

**Candidate Management**

<img src="docs/assets/hr-manage.png" width="800" style="border-radius: 12px;" />

---

## 🔎 Search Engine

The search system is one of the core productivity features of v1.1. It allows HR staff to locate candidates instantly without paginating through long lists.

### How It Works

Search is implemented as a server-side query that accepts a `q` parameter on the `GET /employees` route. The handler reads the parameter, sanitizes it, and builds a parameterized SQL query using GORM's `Where` clause. This means the database engine performs the filtering — there is no in-memory iteration over a loaded slice.

```go
// Simplified logic inside employee handler
query := db.Model(&Employee{})

if search := c.Query("q"); search != "" {
    term := "%" + strings.TrimSpace(search) + "%"
    query = query.Where(
        "name ILIKE ? OR position ILIKE ? OR email ILIKE ?",
        term, term, term,
    )
}

query.Find(&employees)
```

### Multi-field Search

A single search term matches across three fields simultaneously: **name**, **position**, and **email**. This means typing "eng" will surface candidates named "Enrique", candidates with position "Engineer", and candidates whose email contains "eng". The `ILIKE` operator is used (PostgreSQL case-insensitive `LIKE`), so the search is never case-sensitive from the user's perspective.

### Frontend Integration

The search input is part of the employee listing page's header bar. When the user types and submits, the form issues a `GET` request with the `q` parameter appended to the URL — making search results **bookmarkable and shareable** via URL. The current search term is echoed back into the input field value on the server-rendered response, so the user always sees what they searched for.

### Pagination Integration

Search integrates cleanly with pagination. The paginator respects the `q` parameter — each page link in the paginated result set preserves the active search query, so navigating to page 2 of "engineer" results does not lose the search context. This is achieved by forwarding the `q` value into every pagination link's `href` on the template side.

### Security & Sanitization

Because the search uses parameterized queries (`?` placeholders in GORM / pgx), SQL injection is structurally impossible — the user input is never interpolated into the query string. The input is also `TrimSpace`'d before use. Future hardening could add a maximum length cap (e.g., 100 characters) and rate limiting per session.

### Performance

For the current SQLite-era data size this was fast enough as a full-scan. After migrating to PostgreSQL (Neon), a GIN index on the relevant columns would make `ILIKE` search logarithmic rather than linear — this is documented in the [Roadmap](#-roadmap).

---

## 🗄️ SQLite → PostgreSQL via Neon

This migration is one of the most important architectural decisions in the project's evolution. It transforms the system from a local-only tool into a cloud-ready, production-capable application.

### Why SQLite Was a Good Starting Point

SQLite is excellent for prototyping: zero configuration, file-based, no server process, and the Go driver (`mattn/go-sqlite3`) works with CGO out of the box. For a solo developer validating a data model and a UI, it's the right call. The initial `data/users.db` file is testament to this — the entire database is a single file on disk.

### Why SQLite Becomes a Bottleneck

SQLite has a single-writer constraint. In a web application with concurrent HTTP requests, writes queue behind each other, which creates latency under any meaningful load. It also can't be accessed by more than one process (or machine) simultaneously — ruling out horizontal scaling, separate migration runners, or cloud deployment where the filesystem is ephemeral.

### Why PostgreSQL

PostgreSQL is the most feature-complete open-source relational database. It supports true concurrent reads and writes, full-text search operators (`ILIKE`, `tsvector`), `JSONB`, row-level locking, and a rich ecosystem of extensions. Its `ILIKE` operator was directly leveraged for the multi-field search described above. GORM supports both SQLite and PostgreSQL with the same query API, making the migration a matter of swapping the driver — not rewriting queries.

### Why Neon

[Neon](https://neon.tech) is a serverless PostgreSQL platform that provides:

| Feature | Benefit |
|---|---|
| **Serverless scaling** | Database scales to zero when idle — no cost for unused compute |
| **Branching** | Create isolated DB branches for staging/testing without duplication |
| **Connection pooling** | Built-in PgBouncer-compatible pooler at the platform level |
| **Cloud-native** | Accessible from any cloud region; works with Render, Railway, Fly.io |
| **Free tier** | Generous free tier suitable for portfolio and small production apps |

### Connection Setup in Go

The connection is managed in `internal/db/db.go` using GORM with the `pgx/v5` driver via `gorm.io/driver/postgres`. The DSN is loaded from environment variables via `joho/godotenv`.

```go
// internal/db/db.go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "os"
)

func Connect() (*gorm.DB, error) {
    dsn := os.Getenv("DATABASE_URL") // Neon connection string
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

The `jackc/pgx/v5` library handles the low-level wire protocol, while `jackc/puddle/v2` manages the connection pool. The pool settings above are sensible defaults for a low-to-medium traffic application; they prevent connection exhaustion on Neon's free tier while maintaining responsiveness.

### Local vs Production

For local development, you can still point `DATABASE_URL` at a local PostgreSQL instance or a Neon branch. The SQLite driver (`gorm.io/driver/sqlite`) remains in `go.mod` for reference and potential test isolation, but the primary path is PostgreSQL.

---

## 🔐 Authentication & Session Management

Authentication is handled via **cookie-based sessions** using the `gorilla/sessions` library, backed by `gorilla/securecookie` for HMAC-signed, optionally encrypted cookie values.

### How Sessions Work

When a user successfully logs in (username + email + password matched against the users table), the server creates a session entry using `gorilla/sessions`, sets an authenticated flag and the user's ID inside the session store, and writes a `Set-Cookie` header. Every subsequent request carries that cookie, and the middleware at `internal/middleware/auth.go` validates it before allowing access to protected routes.

```go
// internal/auth/session.go — simplified
var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))

func SetUser(c *gin.Context, userID uint) {
    session, _ := store.Get(c.Request, "hr-session")
    session.Values["authenticated"] = true
    session.Values["user_id"] = userID
    session.Save(c.Request, c.Writer)
}

func IsAuthenticated(c *gin.Context) bool {
    session, _ := store.Get(c.Request, "hr-session")
    auth, ok := session.Values["authenticated"].(bool)
    return ok && auth
}
```

### Middleware

The auth middleware (`internal/middleware/auth.go`) is applied as a Gin middleware group around all protected routes (`/dashboard`, `/employees`). If the session is missing or invalid, the user is redirected to `/login` with a `302`. This is a clean gate that requires no changes to individual handlers.

```go
// internal/middleware/auth.go
func RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        if !auth.IsAuthenticated(c) {
            c.Redirect(http.StatusFound, "/login")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### Security Considerations

The session cookie is signed with an HMAC key stored in `SESSION_SECRET`. The `gorilla/securecookie` library ensures that any tampering with the cookie value is detected and the session is invalidated. Passwords are stored hashed — `golang.org/x/crypto` is in the dependency tree, which provides `bcrypt` support. HTTPS in production (via a reverse proxy like Caddy or nginx) ensures the cookie is never transmitted in plaintext.

---

## ✨ Features

**Candidate Management** covers adding new candidates with name, job position, email, and optional profile photo upload. Photos are stored in `backend/uploads/` and served statically, with an automatic fallback to a default gray avatar SVG when no image is provided. The full candidate list is displayed in a sortable, readable table with pagination.

**Status Pipeline** implements a three-state system (`Pending` → `Accepted` / `Rejected`) that is manually controlled by HR staff. Status changes are reflected immediately across both the table view and the dashboard metrics.

**Search Engine** provides multi-field, case-insensitive, server-side search across name, position, and email — with pagination preservation and URL-bookmarkable results.

**Dashboard Metrics** give HR staff a real-time snapshot of recruitment health, displaying total counts for accepted, rejected, and pending candidates via prominently styled KPI cards.

**Authentication** provides user registration and login backed by session cookies, with middleware-enforced route protection.

---

## 🛠️ Tech Stack

### Backend

| Technology | Version | Role |
|---|---|---|
| **Go** | 1.21+ | Core application language |
| **gin-gonic/gin** | v1.12.0 | HTTP framework, routing, middleware, template rendering |
| **gorm.io/gorm** | v1.31.1 | ORM — schema migration, query building, model binding |
| **gorm.io/driver/postgres** | v1.6.0 | GORM PostgreSQL adapter |
| **jackc/pgx/v5** | v5.6.0 | PostgreSQL wire protocol driver (used by GORM under the hood) |
| **jackc/puddle/v2** | v2.2.2 | Connection pool manager for pgx |
| **gorilla/sessions** | v1.4.0 | Server-side session management via signed cookies |
| **gorilla/securecookie** | v1.1.2 | HMAC cookie signing and optional encryption |
| **joho/godotenv** | v1.5.1 | `.env` file loading for local development |
| **google/uuid** | v1.6.0 | UUID generation for entity identifiers |
| **gabriel-vasile/mimetype** | v1.4.12 | MIME type detection for uploaded photos |
| **go-playground/validator/v10** | v10.30.1 | Struct-level input validation |
| **golang.org/x/crypto** | v0.50.0 | `bcrypt` password hashing |

### Frontend

| Technology | Role |
|---|---|
| **HTML5 + html/template** | Server-side rendering through Gin's template engine |
| **CSS3 (vanilla)** | Custom design system — no external CSS frameworks |
| **JavaScript (vanilla)** | Client-side interactivity (status updates, UI toggling) |

### Infrastructure & Tooling

| Tool | Role |
|---|---|
| **Neon** | Serverless PostgreSQL cloud database |
| **gorm.io/driver/sqlite** | SQLite driver (retained for local/test use) |
| **mattn/go-sqlite3** | CGO-based SQLite3 driver |
| **stretchr/testify** | Test assertions |
| **go.uber.org/mock** | Mock generation for unit tests |
| **Makefile** | Task automation (`run`, `test`, `seed`) |

### Supporting Libraries (from `go list`)

The `bytedance/sonic` + `bytedance/sonic/loader` + `cloudwego/base64x` trio is pulled in by Gin for high-performance JSON encoding on supported architectures (amd64/arm64). `json-iterator/go` and `goccy/go-json` provide additional fast-path JSON paths. `pelletier/go-toml/v2` and `goccy/go-yaml` are Gin configuration utilities. `klauspost/compress` handles gzip response compression.

---

## 🏗️ Architecture

The application follows a layered architecture with clear boundaries between concerns:

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Layer (Gin)                    │
│          Routes · Middleware · Template Rendering         │
└────────────────────────┬────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
   ┌──────────┐   ┌──────────────┐  ┌──────────────┐
   │  auth.go │   │ employee.go  │  │  Middleware   │
   │ handler  │   │   handler    │  │  (auth gate)  │
   └────┬─────┘   └──────┬───────┘  └──────┬───────┘
        │                │                  │
        └────────────────┼──────────────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │    internal/db       │
              │  GORM + pgx/v5 Pool  │
              └──────────┬───────────┘
                         │
                         ▼
              ┌──────────────────────┐
              │  Neon PostgreSQL     │
              │  (cloud, serverless) │
              └──────────────────────┘
```

The `internal/` packages (`auth`, `db`, `middleware`) are deliberately isolated from `backend/handlers/` — handlers call internal packages but not vice versa. This means the database connection, session logic, and auth gate can be tested and swapped independently.

---

## 📁 Project Structure

```
hr-management-web/
├── go.mod                          # Module definition
├── go.sum                          # Dependency checksums
├── makefile                        # Build and task automation
│
├── backend/
│   ├── cmd/
│   │   ├── server/main.go          # Application entrypoint
│   │   ├── seed_users/main.go      # Seeds test user accounts
│   │   ├── seed_employee/main.go   # Seeds test candidate data
│   │   └── list_users/main.go      # CLI: prints all users
│   │
│   ├── database/
│   │   └── database.go             # Legacy DB init (SQLite era)
│   │
│   ├── handlers/
│   │   ├── auth.go                 # Login, register, logout handlers
│   │   └── employee.go             # Candidate CRUD + search + status
│   │
│   └── templates/
│       ├── login.html
│       ├── register.html
│       ├── dashboard.html
│       ├── employees.html
│       └── id-card.html
│
├── internal/
│   ├── auth/
│   │   └── session.go              # Session read/write helpers
│   ├── db/
│   │   └── db.go                   # GORM + PostgreSQL connection
│   └── middleware/
│       └── auth.go                 # RequireAuth Gin middleware
│
├── frontend/
│   └── css/
│       ├── style.css               # Global design tokens
│       ├── login.css
│       ├── register.css
│       ├── dashboard.css
│       ├── employees.css
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
      │   ACCEPTED    │                 │   REJECTED    │
      └───────────────┘                 └───────────────┘
```

Status transitions are **bidirectional and manually controlled** — HR staff can move a candidate from any state to any other at any time, reflecting real-world hiring process flexibility.

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

For local development with a Neon database, the `DATABASE_URL` is available in your Neon project dashboard under **Connection Details**. For local PostgreSQL, use `postgres://postgres:password@localhost:5432/hr_dev?sslmode=disable`.

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
| `GET` | `/login` | No | Login page |
| `POST` | `/login` | No | Process login |
| `GET` | `/register` | No | Registration page |
| `POST` | `/register` | No | Create account |
| `POST` | `/logout` | Yes | Destroy session |
| `GET` | `/dashboard` | Yes | Metrics overview |
| `GET` | `/employees` | Yes | Candidate list (supports `?q=` and `?page=`) |
| `POST` | `/employees` | Yes | Create new candidate |
| `PUT` | `/employees/:id/status` | Yes | Update candidate status |
| `GET` | `/employees/:id/card` | Yes | Candidate ID card view |

---

## 🧪 Testing

Integration tests validate the core application flows: candidate creation, status transitions, and authentication. The test suite uses `stretchr/testify` for assertions and `go.uber.org/mock` for mocking the database layer.

```bash
make test
```

Tests run against an isolated test database instance. The `jordanlewis/gcassert` package provides compile-time assertions for performance-critical code paths.

---

## 🗺️ Roadmap

### v1.1 — Current (Quality of Life)
- [x] Multi-field search with PostgreSQL `ILIKE`
- [x] Pagination for the candidates table
- [x] Redesigned dark UI with modern design system
- [x] Session-based authentication with gorilla/sessions

### v1.2 — Authentication Hardening
- [ ] JWT-based authentication as an alternative to session cookies
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
- [ ] GIN index on search columns for `ILIKE` at scale

---

## 📄 License

This project is licensed under the **MIT License**. See the [LICENSE](./LICENSE) file for full details.

---

<div align="center">

Built with ❤️ using Go + Gin + PostgreSQL · Designed for portfolio and production alike

⭐ Star the repo if it helped you — it helps more than you think

</div>