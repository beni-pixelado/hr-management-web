<div align="center">

# 📚 HR Management Web — Documentation

**Staffio** is a production-minded, multi-tenant Human Resources management platform. This directory is the authoritative technical reference for the codebase: architecture, data model, security, deployment and everything in between.

Go · Gin · PostgreSQL (Neon) · Gorilla Sessions

</div>

---

## 🧭 How to use this documentation

- **New to the project?** Start with [Getting Started](./getting-started.md) to run it locally in minutes.
- **Evaluating the architecture?** Read [Architecture](./architecture.md) and [Backend](./backend.md).
- **Extending a feature?** Jump straight to the relevant module doc below.
- **Deploying?** Follow [Deployment](./deployment.md).

Every document stays in sync with the codebase. If you find a discrepancy, open an issue or a pull request.

---

## 📑 Document Index

### Get started

| Document | Purpose |
|---|---|
| [Getting Started](./getting-started.md) | Prerequisites, environment variables, running locally, Docker, test suite |
| [Deployment](./deployment.md) | Shipping to Render, Neon PostgreSQL, Cloudinary and one-command Docker |

### Architecture & Core

| Document | Purpose |
|---|---|
| [Architecture](./architecture.md) | Layered architecture, request lifecycle, design decisions |
| [Backend](./backend.md) | Handlers, routing and business logic |
| [Database](./database.md) | Data models, migrations, data flow and lifecycle |
| [Frontend](./frontend.md) | Design system, templates, CSS/JS architecture and responsiveness |

### Security & Access

| Document | Purpose |
|---|---|
| [Authentication](./authentication.md) | Sessions, bcrypt, password recovery and reset flow |
| [Roles & Permissions](./roles-permissions.md) | RBAC model, permission matrix and multi-tenant isolation |
| [Security](./security.md) | CSRF, rate limiting and hardening practices |

### Modules

| Document | Purpose |
|---|---|
| [Candidate Search](./search.md) | Multi-field server-side search and pagination |
| [Departments](./departments.md) | Organizational structure and member management |
| [Reports & Analytics](./reports.md) | Absences, hirings, terminations and overview charts |

### Reference

| Document | Purpose |
|---|---|
| [API Reference](./api.md) | Complete route table with auth requirements (plus [`openapi.yaml`](./openapi.yaml)) |
| [Testing](./testing.md) | Test strategy, tooling and coverage |
| [Scalability](./scalability.md) | Current limits and growth strategy |
| [Roadmap](./roadmap.md) | Product direction and phase planning |

---

## 🗺️ Project Map

```
hr-management-web/
├── backend/
│   ├── cmd/                     # Runnable entrypoints (server, seeders, CLI)
│   │   └── server/main.go       # Composition root — routing, middleware, startup
│   ├── database/                # GORM + PostgreSQL connection init
│   ├── handlers/                # HTTP controllers (auth, employees, departments, …)
│   └── templates/               # Server-rendered HTML (html/template)
├── internal/
│   ├── auth/                    # Cookie session read/write helpers
│   ├── csrf/                    # CSRF token middleware + template helper
│   ├── middleware/              # RequireAuth, LoadUser, BlockViewerWrites, RateLimit
│   ├── storage/                 # Cloudinary upload/destroy helpers
│   └── tests/                   # Integration tests + k6 load scripts
├── frontend/
│   ├── css/                     # Per-page vanilla CSS design system
│   ├── public/js/               # Vanilla JS (charts, dark mode, UI helpers)
│   └── static/                  # Favicon and screenshot assets
├── docs/                        # ← You are here
├── docker-compose.yml           # App + PostgreSQL for local development
├── render.yaml                  # Render Blueprint for cloud deployment
└── makefile                     # Task automation (run, test, build, k6, …)
```

---

## 🔧 Quick Reference

Run everything you need from the `makefile`:

```bash
make run             # Start the server (go run ./backend/cmd/server)
make build           # Compile a static binary to ./server
make test            # Run the full Go test suite
make docker-up       # App + PostgreSQL via Docker Compose
make k6-dashboard    # Run the dashboard load test
```

---

<div align="center">

_Technical documentation for **HR Management Web (Staffio)** · MIT License_

</div>