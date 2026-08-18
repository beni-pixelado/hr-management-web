<div align="center">

# 🏢 Staffio — HR Management Web

### A production-minded, multi-tenant Human Resources Management System

Go · Gin · PostgreSQL (Neon) · Gorilla Sessions

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![Gin](https://img.shields.io/badge/Gin-1.12-008ECF.svg?style=for-the-badge&logo=go&logoColor=white)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-4169E1.svg?style=for-the-badge&logo=postgresql&logoColor=white)](https://neon.tech)
[![CI](https://github.com/beni-pixelado/hr-management-web/actions/workflows/ci.yml/badge.svg)](https://github.com/beni-pixelado/hr-management-web/actions)
[![License](https://img.shields.io/badge/license-MIT-green.svg?style=for-the-badge)](./LICENSE)

</div>

<br/>

> **Staffio** is a full-stack Human Resources platform that streamlines the recruitment pipeline — from candidate intake to final status resolution — with department management, live analytics, reports, team collaboration with role-based access, account configuration, password recovery, and a modern dark UI. Designed to run entirely on **$0/month** infrastructure.

<br/>

## ✨ Highlights

- **Multi-tenant by design** — every organization is isolated; teams collaborate with RBAC
- **Roles & permissions** — Owner / Admin / Recruit / Viewer
- **Candidate pipeline** — pending → hired / rejected, with status analytics
- **Departments** — organizational structure and member management
- **Live analytics** — pie & line charts backed by JSON APIs
- **Reports** — absences, hires and firings over configurable periods
- **Secure by default** — CSRF protection, rate limiting, bcrypt, HMAC-signed sessions
- **$0/month stack** — Neon + Render + Cloudinary + Gmail SMTP + GitHub Actions

---

## 📋 Table of Contents

- [Screenshots](#-screenshots)
- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Architecture](#-architecture)
- [Getting Started](#-getting-started)
- [Environment Variables](#-environment-variables)
- [Docker](#-docker)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Documentation](#-documentation)
- [Roadmap](#-roadmap)
- [License](#-license)

---

## 📸 Screenshots

| Dashboard | Candidates |
|---|---|
| ![Dashboard](./frontend/static/rh-dashboard.png) | ![Candidates](./frontend/static/rh-login.png) |

> More visuals live in [`frontend/static/`](./frontend/static/).

---

## ✨ Features

**Candidate Management** — add candidates with name, position, email, description and an optional profile photo (uploaded to Cloudinary). Full CRUD, a three-state status pipeline (`pending` → `contractors`/`rejected`), pagination, and a print-friendly ID card.

**Search & Export** — server-side, case-insensitive search across name and email, status filtering, pagination, and one-click CSV export of the filtered set.

**Department Management** — organize your workforce into departments with a name, unique code and optional manager; assign and remove collaborators.

**Live Analytics** — an overview with two live charts: **Candidates by Department** (pie) and **New candidates over time** (line), fed by dedicated JSON APIs scoped to your organization.

**Reports** — capture snapshot metrics for absences, hires and firings over week/month/year periods, each rendered with a human-readable narrative and stats.

**Team & Roles (RBAC)** — invite teammates with temporary credentials, assign roles (Owner/Admin/Recruit/Viewer), change or remove members, and transfer ownership. Permissions are enforced at both middleware and handler level.

**Account & Device Management** — manage your profile, profile photo, change your password, review device details, and delete your account.

**Authentication & Recovery** — session-based auth with bcrypt hashing, CSRF protection, per-IP rate limiting, and e-mail password recovery with single-use tokens.

**Landing Page** — a marketing-grade entry point with a hero, an animated pipeline mock, and product showcase.

---

## 🛠️ Tech Stack

### Backend

| Technology | Version | Role |
|---|---|---|
| **Go** | 1.25 | Core application language |
| **gin-gonic/gin** | v1.12.0 | HTTP framework, routing, middleware, rendering |
| **gorm.io/gorm** | v1.31.1 | ORM — migrations, queries, model binding |
| **jackc/pgx/v5** | v5.6.0 | PostgreSQL wire protocol driver |
| **gorilla/sessions** | v1.4.0 | Cookie-based session management |
| **gorilla/securecookie** | v1.1.2 | HMAC cookie signing |
| **golang.org/x/crypto** | v0.50.0 | `bcrypt` password hashing |
| **cloudinary-go/v2** | v2.16.0 | Photo upload/storage |
| **joho/godotenv** | v1.5.1 | `.env` loading |
| **google/uuid** | v1.6.0 | UUID tokens & filenames |
| **gabriel-vasile/mimetype** | v1.4.12 | Upload MIME detection |

### Frontend

| Technology | Role |
|---|---|
| **HTML5 + html/template** | Server-side rendering (auto-escaping = XSS-safe) |
| **CSS3 (vanilla)** | Custom design system, dark mode, no frameworks |
| **JavaScript (vanilla)** | Charts (Chart.js), dark mode, UI interactions |

### Infrastructure & Tooling

| Tool | Role |
|---|---|
| **Neon** | Serverless PostgreSQL (free tier) |
| **Cloudinary** | Image CDN/storage |
| **Render** | Cloud hosting (free tier) |
| **Gmail SMTP** | Recovery / invite e-mail |
| **GitHub Actions** | CI (vet, test, build) |
| **Docker + Compose** | One-command local dev |
| **k6** | Load testing |

---

## 🏗️ Architecture

Layered, server-rendered, single-binary:

```
Browser (Go templates + CSS + JS)
        │  HTTP
        ▼
Gin HTTP Layer (router · middleware · static · templates)
        │
        ▼
Handlers (auth · employee · department · org · report · config …)
        │
        ▼
internal/ (auth · csrf · middleware · storage)
        │
        ▼
Neon PostgreSQL (multi-tenant)
```

- **Server-rendered HTML** — one deployable binary, no frontend framework.
- **RBAC + tenant scoping** — least privilege and data isolation by construction.
- **Cookie sessions** — simple, revocation-friendly, no token storage.

See the full breakdown in [docs/architecture.md](./docs/architecture.md).

---

## 🚀 Getting Started

### Prerequisites

Go `>= 1.25`, Git, and a PostgreSQL connection string (Neon free tier works perfectly).

```bash
git clone https://github.com/beni-pixelado/hr-management-web.git
cd hr-management-web
go mod download
```

### Configure environment

```bash
cp .env.example .env
# then fill in DATABASE_URL, SESSION_SECRET, and service keys (see below)
```

### Run

```bash
make run
# or: go run ./backend/cmd/server
```

Visit **http://localhost:8000**. Tables are created automatically via `AutoMigrate`.

### Seed sample data *(optional)*

```bash
go run ./backend/cmd/seed_users
go run ./backend/cmd/seed_employee
```

The full walkthrough is in [docs/getting-started.md](./docs/getting-started.md).

---

## 🔑 Environment Variables

Create a `.env` in the project root:

```env
# PostgreSQL connection string (Neon or local)
DATABASE_URL=postgres://user:password@host/dbname?sslmode=require

# Secret for HMAC-signing session cookies — use a long random string
SESSION_SECRET=openssl-rand-hex-32

# Cloudinary (profile & employee photos)
CLOUDINARY_CLOUD_NAME=your-cloud-name
CLOUDINARY_API_KEY=your-api-key
CLOUDINARY_API_SECRET=your-api-secret

# SMTP (password recovery + team invites)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=you@example.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=you@example.com
SMTP_FROM_NAME=Staffio

# Base URL for e-mail links (falls back to request host)
SITE_URL=http://localhost:8000

# Server port (optional, defaults to 8000)
PORT=8000
```

For a local PostgreSQL instance: `DATABASE_URL=postgres://postgres:password@localhost:5432/hr_dev?sslmode=disable`.

---

## 🐳 Docker

Bring up the app **and** PostgreSQL with one command:

```bash
make docker-up
# or: docker compose up --build
```

- App: http://localhost:8000
- PostgreSQL: localhost:5432

Set `SESSION_SECRET`, Cloudinary and SMTP values via environment variables before running — the Compose defaults are for local dev only.

---

## 🧪 Testing

```bash
make test
# or: go test ./...
```

The suite covers authentication, candidate lifecycle, departments, RBAC, CSRF, rate limiting and multi-tenant isolation against an isolated in-memory database. CI (`.github/workflows/ci.yml`) runs `go vet`, tests and a build on every push.

Load tests are available via `k6` (`make k6-dashboard`, `make k6-department`, `make k6-create`).

See [docs/testing.md](./docs/testing.md).

---

## ☁️ Deployment

Deployable to **Render** free tier in minutes via the included Blueprint:

1. Push the repo (including `render.yaml`).
2. In Render → **New** → **Blueprint**, select the repository.
3. Fill in the environment variables (`DATABASE_URL`, `SESSION_SECRET`, Cloudinary, SMTP, `SITE_URL`).
4. **Apply** — Render builds `go build -o server ./backend/cmd/server` and serves it.

Nothing stateful lives on the app host: sessions are cookies, photos are on Cloudinary, and data lives in Neon. The full stack runs at **$0/month**.

See [docs/deployment.md](./docs/deployment.md).

---

## 📚 Documentation

A complete technical reference lives in [`docs/`](./docs/README.md):

| Area | Document |
|---|---|
| Start here | [Getting Started](./docs/getting-started.md) |
| Architecture | [Architecture](./docs/architecture.md) · [Backend](./docs/backend.md) |
| Data | [Database](./docs/database.md) |
| Frontend | [Frontend](./docs/frontend.md) |
| Auth & Security | [Authentication](./docs/authentication.md) · [Security](./docs/security.md) · [Roles & Permissions](./docs/roles-permissions.md) |
| Modules | [Search](./docs/search.md) · [Departments](./docs/departments.md) · [Reports](./docs/reports.md) |
| Reference | [API](./docs/api.md) · [Testing](./docs/testing.md) · [Scalability](./docs/scalability.md) |
| Direction | [Roadmap](./docs/roadmap.md) · [Deployment](./docs/deployment.md) |

---

## 🗺️ Roadmap

The current focus is onboarding polish and a few high-value recruiting features (candidate notes, interview scheduling), then distribution. See [docs/roadmap.md](./docs/roadmap.md).

---

## 📄 License

This project is licensed under the **MIT License**. See the [LICENSE](./LICENSE) file for full details.

---

<div align="center">

Built with ❤️ using Go + Gin + PostgreSQL · Designed for portfolio and production alike

⭐ Star the repo if you find it useful

</div>