# Roadmap

## v1.1 — Current Release ✅

The primary deliverables of this release:

- [x] Multi-field search with PostgreSQL `ILIKE` across name, position, and email
- [x] Pagination for the candidates table (20 per page)
- [x] Redesigned UI with modern design system across all pages
- [x] Session-based authentication with gorilla/sessions and bcrypt
- [x] Department creation and listing module with manager assignment
- [x] Candidate delete functionality
- [x] PostgreSQL migration via Neon (from SQLite prototype)
- [x] Sequence fix utility for post-migration consistency

---

## v1.2 — Control (New Functions)

Focus: completing the department module and hardening access control.

- [ ] Department deletion (`DELETE /department/:id`)
- [ ] Add collaborators to departments (`POST /department/:id/members`)
- [ ] Remove collaborators from departments (`DELETE /department/:id/members/:employee_id`)
- [ ] Department detail view with member list
- [ ] Role-based access control (Admin, Recruiter, Viewer)
- [ ] Account management (change password, deactivate account)
- [ ] CSRF protection on all form submissions
- [ ] Rate limiting on authentication routes

---

## v1.3 — Enhanced Data Model

Focus: richer candidate and organizational data.

- [ ] Candidate notes and free-form comments
- [ ] Interview scheduling with date tracking
- [ ] iCal export for interview dates
- [ ] Full department and team assignment on candidates
- [ ] Audit trail for all status change history (who changed what, when)
- [ ] GIN trigram index on search columns for `ILIKE` at scale

---

## v2.0 — Architecture Evolution

Focus: API, observability, and deployment.

- [ ] Full REST API with OpenAPI/Swagger documentation generated from Go structs
- [ ] HTMX-powered frontend for partial page updates (no full-page reloads)
- [ ] Docker + Docker Compose for one-command local development
- [ ] CI/CD pipeline via GitHub Actions (test → lint → build → deploy)
- [ ] Structured logging with `slog` or `zap`
- [ ] Prometheus metrics endpoint
- [ ] Object storage for uploaded photos (S3 or R2)