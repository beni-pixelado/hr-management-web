# Roadmap

This roadmap describes where **Staffio** is going and how it gets there. It is organized into time-boxed phases rather than version numbers, so work ships continuously in small, deployable increments.

> **Guiding constraint:** the entire platform runs on **$0/month** infrastructure (Neon, Render, Cloudinary, Gmail SMTP, GitHub Actions). Every phase is designed to stay within free tiers while remaining genuinely useful.

---

## ✅ Shipped

The following capabilities are complete and available today:

- **Candidate management** — CRUD, status pipeline, pagination, ID card, photo uploads
- **Multi-field search** — dashboard name/email search + status filtering + CSV export
- **Departments** — create, list, detail, delete, member assign/remove
- **Overview analytics** — live pie + line charts via JSON APIs
- **Reports** — absences, hired, fired snapshots with narrative detail
- **Account & device management** — profile, photo, password change, account deletion
- **Team management** — invites, role changes, removal, ownership transfer
- **Authentication** — session auth, password recovery/reset, CSRF protection, rate limiting
- **RBAC** — Owner / Admin / Recruit / Viewer with tenant isolation
- **Quality** — Go test suite, GitHub Actions CI, Docker + Compose, structured logging (`log/slog`)
- **Deployment** — Render blueprint, Neon PostgreSQL, Cloudinary storage

---

## Phase 1 — Polish, safety & onboarding

Focus: make first-time users successful and the product audit-ready.

- [ ] Hardened, tested e-mail delivery for invites and recovery
- [ ] Empty states and guided first-run onboarding
- [ ] Demo account / one-click sample-data seeder
- [ ] Live demo link in the README
- [ ] Lightweight analytics (e.g. GoatCounter) to measure usage
- [ ] Remove remaining dev leftovers

**Exit criteria:** a new user can register, understand the product, and be productive in under five minutes.

---

## Phase 2 — Core recruiting value

Features HR practitioners miss immediately.

- [ ] **Candidate notes** — free-form text per candidate with author and timestamp
- [ ] **Interview scheduling** — a date + status per candidate and an "upcoming interviews" list on the dashboard

**Explicitly deferred:** iCal export, calendar sync, e-mail reminders (Gmail SMTP covers reminders on request).

---

## Phase 3 — Scale & reliability

Prepare for growth without adding cost.

- [ ] GIN trigram indexes for search (once >10k rows)
- [ ] Connection-pool tuning for higher Neon plans
- [ ] Explicit session-expiry and e-mail-delivery tests
- [ ] Structured logging enrichment for request tracing

---

## Phase 4 — Launch & distribution

The only free growth lever is distribution.

- [ ] Product Hunt, Hacker News, dev.to, LinkedIn write-ups
- [ ] A compelling "how I built a $0/mo SaaS" case study
- [ ] Collect feedback from the first real users and feed it back into the roadmap

---

## Later — only if real users ask

- [ ] Audit trail (who changed what/when)
- [ ] Full REST API + OpenAPI documentation
- [ ] HTMX partial reloads (only if full-page reloads are friction)
- [ ] iCal export / calendar sync

---

## Explicitly out of scope (solo + $0)

- Prometheus / Grafana / APM dashboards (slog + Render logs suffice)
- Kubernetes / microservices / multi-node orchestration
- Redis, Kafka or message queues (unless multi-instance sharing demands it)
- Paid e-mail / monitoring / AI features (revisit only when revenue exists)

---

## Cost Sheet

| Service | Job | Free-tier reality |
|---|---|---|
| Neon | PostgreSQL | Scales to zero when idle; limits connections — design for it |
| Render | Hosting | Spins down after ~15 min idle; cold start on first hit |
| Cloudinary | Photo storage | Free quota sufficient at this scale |
| Gmail SMTP | E-mail | ~500/day cap; fine alone |
| GitHub + Actions | Repo + CI | Free private repo + build minutes |
| **Total** | | **$0.00/month** |

---

## Related

- [Scalability](./scalability.md)
- [Deployment](./deployment.md)
- [Testing](./testing.md)