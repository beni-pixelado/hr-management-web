# Deployment

Staffio is designed to run entirely on **free tiers** — a single stateless Go binary, a serverless PostgreSQL database, and cloud image storage. This document covers local Docker, production on Render, and the external services it depends on.

---

## Architecture at a Glance

```
                    ┌─────────────────────────┐
                    │      Render (web)        │
                    │  ./server  (Go binary)   │
                    │  stateless — no disk      │
                    └────┬──────────────┬───────┘
                         │              │
               DATABASE_URL            Cloudinary (photos)
                         │              │
                 ┌───────▼───────┐   ┌──▼──────────────┐
                 │ Neon Postgres │   │ Cloudinary CDN  │
                 │ (serverless)  │   │ image storage    │
                 └───────────────┘   └─────────────────┘
```

Nothing stateful lives on the app host — sessions are cookie-based, images live on Cloudinary, and data lives in Neon. This is what makes the app fully deployable to a free, ephemeral instance.

---

## Local Development with Docker

A `Dockerfile` + `docker-compose.yml` bring up the app **and** PostgreSQL together:

```bash
docker compose up --build
```

- App: http://localhost:8000
- PostgreSQL: localhost:5432 (`hr` / `hr` / db `hr`)

The `app` service waits for the `db` health check before starting. Set the required variables (`SESSION_SECRET`, Cloudinary, SMTP) as environment variables — the Compose file provides safe defaults but they are **not production secrets**.

The Dockerfile is multi-stage: it builds a **static, `CGO_ENABLED=0` binary** (the server does not import the cgo SQLite driver) and copies only the runtime assets (templates, `frontend/`, `robots.txt`) into a slim Alpine image.

---

## Production: Render (free tier)

The repository ships a `render.yaml` Blueprint, so deployment is mostly dashboard-driven.

### 1. Push the blueprint

```bash
git add render.yaml
git commit -m "chore: add render.yaml blueprint"
git push
```

### 2. Create the service

1. Go to <https://render.com> → **New** → **Blueprint** → select the `hr-management-web` repository.
2. Render detects the `hr-management-web` web service from `render.yaml`.

### 3. Set environment variables

All variables are `sync: false` — you enter them manually in the Render dashboard:

| Variable | Where to get it |
|---|---|
| `DATABASE_URL` | Neon → project → Connection details (pooled connection string) |
| `SESSION_SECRET` | `openssl rand -hex 32` (any long random string) |
| `CLOUDINARY_CLOUD_NAME` | Cloudinary dashboard |
| `CLOUDINARY_API_KEY` | Cloudinary dashboard |
| `CLOUDINARY_API_SECRET` | Cloudinary dashboard |
| `SMTP_HOST` | e.g. `smtp.gmail.com` |
| `SMTP_PORT` | e.g. `587` |
| `SMTP_EMAIL` | SMTP account e-mail |
| `SMTP_PASSWORD` | SMTP account / app password |
| `SMTP_FROM` | Sender address shown in e-mails |
| `SMTP_FROM_NAME` | Sender display name (default `Staffio`) |
| `SITE_URL` | e.g. `https://your-service.onrender.com` |

`PORT` is injected by Render automatically; the server reads it via `os.Getenv("PORT")`. Do **not** put `PORT` in `.env`.

### 4. Deploy

Render builds with `go build -o server ./backend/cmd/server` and starts `./server`. Health check: `/login`. You receive a `https://your-service.onrender.com` URL.

---

## First-Time Setup on Production

Tables are created automatically on startup (`AutoMigrate`). To seed sample data against the production database:

```bash
DATABASE_URL="postgres://..." go run ./backend/cmd/seed_users
DATABASE_URL="postgres://..." go run ./backend/cmd/seed_employee
```

---

## External Services

### Neon (PostgreSQL)

Serverless PostgreSQL. Scales to zero when idle (no cost for unused compute), supports branching for staging, and includes a built-in connection pooler. Design for its connection limits — the app caps its pool (`SetMaxOpenConns(10)` / `SetMaxIdleConns(5)`).

### Cloudinary (image storage)

Employee and profile photos are uploaded to Cloudinary and destroyed when replaced or deleted. The free quota comfortably covers this scale. Files are addressed by UUID public IDs — no local disk, no path traversal.

### SMTP (e-mail)

Recovery and team-invite e-mails use `net/smtp` with `PlainAuth`. Gmail works with an app password (~500/day cap, fine at this scale).

---

## Free-Tier Reality

| Service | Behavior to design for |
|---|---|
| Render | Spins down after ~15 min idle; cold start on the first hit after idle |
| Neon | Connection limit; scales to zero when idle |
| Cloudinary | Quota capped; destroy images on delete/replace |
| Gmail SMTP | ~500/day sending cap |

The whole production stack runs at **$0/month**.

---

## Related

- [Getting Started](./getting-started.md)
- [Scalability](./scalability.md)
- [Roadmap](./roadmap.md)