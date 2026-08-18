# Getting Started

This guide walks you through running **HR Management Web (Staffio)** locally, configuring it, and getting the test suite green.

---

## Prerequisites

| Tool | Minimum version | Purpose |
|---|---|---|
| [Go](https://golang.org/dl/) | `1.25` | Compile and run the server |
| [Git](https://git-scm.com/) | any | Clone the repository |
| [Docker](https://www.docker.com/) + Compose | any | *(optional)* one-command local stack |
| [Make](https://www.gnu.org/software/make/) | any | Convenience targets in the `makefile` |

A PostgreSQL connection string is required. The **Neon** free tier works perfectly, or you can run PostgreSQL locally via Docker (see below).

---

## 1. Clone and install

```bash
git clone https://github.com/beni-pixelado/hr-management-web.git
cd hr-management-web
go mod download
```

---

## 2. Configure environment variables

Copy the example file and fill it in:

```bash
cp .env.example .env
```

```env
# PostgreSQL connection string (Neon or local)
DATABASE_URL=postgres://user:password@host/dbname?sslmode=require

# Secret key for HMAC-signing session cookies — use a long random string
SESSION_SECRET=replace-me-with-32-plus-random-chars

# Cloudinary credentials (employee & profile photos)
CLOUDINARY_CLOUD_NAME=your-cloud-name
CLOUDINARY_API_KEY=your-api-key
CLOUDINARY_API_SECRET=your-api-secret

# SMTP for password-recovery and team-invite e-mail
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=you@example.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=you@example.com
SMTP_FROM_NAME=Staffio

# Base URL used in e-mail links (falls back to the request host)
SITE_URL=http://localhost:8000

# Server port (optional, defaults to 8000)
PORT=8000
```

> **Security note:** `SESSION_SECRET` must be a unique, high-entropy value. Generate one with `openssl rand -hex 32`. Never ship the default value.

For a local PostgreSQL instance:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/hr_dev?sslmode=disable
```

---

## 3. Run the server

```bash
make run
# or: go run ./backend/cmd/server
```

The application is served at **http://localhost:8000**.

On startup the server:

1. Loads `.env` (if present).
2. Connects to PostgreSQL and runs `AutoMigrate` for every model.
3. Backfills any legacy users without an organization into their own org.
4. Starts the Gin HTTP server.

---

## 4. Seed sample data *(optional)*

```bash
# Seed test user accounts
go run ./backend/cmd/seed_users

# Seed candidate records
go run ./backend/cmd/seed_employee
```

---

## 5. Running with Docker (one command)

A `Dockerfile` and `docker-compose.yml` bring up the app **and** a PostgreSQL database together:

```bash
make docker-up
# or: docker compose up --build
```

- App: http://localhost:8000
- PostgreSQL: localhost:5432

Override any variable (e.g. `SESSION_SECRET`, `DATABASE_URL`) through environment variables before running — **don't ship the Compose defaults**.

---

## 6. Running the test suite

```bash
make test
# or: go test ./...
```

The suite spins up an isolated in-memory database and covers authentication, candidate/status lifecycle, departments, RBAC, CSRF and multi-tenant isolation. See [Testing](./testing.md).

For load testing with `k6`:

```bash
make k6-dashboard
make k6-department
make k6-create
```

---

## 7. CLI utilities

The `backend/cmd/` directory contains standalone operational tools sharing the same database connection:

```bash
go run ./backend/cmd/seed_users        # Seed test user accounts
go run ./backend/cmd/seed_employee     # Seed candidate records
go run ./backend/cmd/list_users        # Print all users
go run ./backend/cmd/fix_sequences     # Fix PostgreSQL serial sequences after bulk imports
```

---

## Next steps

- Understand the system's shape → [Architecture](./architecture.md)
- Configure cloud deployment → [Deployment](./deployment.md)
- Explore the data model → [Database](./database.md)