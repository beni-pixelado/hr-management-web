# Scalability

Staffio is built to be **correct at the current scale** and to have a clear, low-cost path forward as it grows. This document is honest about current limits and how each is intended to be addressed.

---

## Design for Scale

The current architecture is deliberately simple (a single stateless Go binary + serverless Postgres) so that scaling decisions are postponed until real usage demands them — not made preemptively with paid infrastructure.

| Aspect | Current state | Growth path |
|---|---|---|
| **Deployment** | Single instance, stateless | Stateless by design — horizontal scaling is a config change on Render, no code changes |
| **Database** | Neon PostgreSQL, pool capped (10 conns) | Neon auto-scales compute; add indexes and connection tuning as needed |
| **Sessions** | Cookie-based (gorilla/sessions) | Scales trivially (stateless); switch to encrypted + Redis-backed store only if needed |
| **File storage** | Cloudinary (external CDN) | Already off-host and CDN-served; scale is Cloudinary's problem |
| **Search** | Sequential `ILIKE` scan | Add GIN trigram indexes, then full-text (`tsvector`) — see [Search](./search.md) |
| **Concurrency** | In-memory rate limiting | Distribute with Redis if multiple instances share auth routes |
| **Observability** | `log/slog` to stdout | Structured logs already; add lightweight analytics as desired |

---

## Current Limits

1. **PostgreSQL connection pool** — capped at 10 open / 5 idle connections to respect Neon's free-tier limits. High concurrent traffic could queue; raise `MaxOpenConns` as the Neon plan grows.
2. **Search performance** — `ILIKE '%term%'` cannot use a B-tree index and scans rows. Fine under ~10k rows; needs trigram indexes beyond that.
3. **Rate limiting is in-memory** — it resets on restart and is per-instance. Correct for a single instance; needs shared state (Redis) for multi-instance.
4. **Photo/upload processing** — synchronous Cloudinary upload; acceptable at this scale.

---

## Planned Optimizations

### Search indexes

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_employees_fullname_trgm
    ON employees USING GIN (full_name gin_trgm_ops);

CREATE INDEX idx_employees_position_trgm
    ON employees USING GIN (position gin_trgm_ops);

CREATE INDEX idx_employees_email_trgm
    ON employees USING GIN (email gin_trgm_ops);
```

Neon supports `pg_trgm` natively. Recommended once the employees table exceeds ~10,000 rows.

### Database tuning

- Raise pool sizes (`SetMaxOpenConns` / `SetMaxIdleConns`) as the Neon plan allows.
- Add composite indexes on the most common filtered/ordered columns (e.g. `(organization_id, status)`).

### Load testing

The repo includes `k6` scripts (`make k6-dashboard`, `make k6-department`, `make k6-create`) to validate behavior under concurrent users before tuning.

---

## What Is Intentionally Not Planned

Per the [Roadmap](./roadmap.md), the following are **explicitly out of scope** for a solo, $0-budget project:

- Kubernetes / microservices / multi-node orchestration
- Redis, Kafka or message queues (only if/when multi-instance sharing demands it)
- Prometheus / Grafana / APM dashboards (slog + Render logs suffice at this scale)

These are deferred until real users and revenue justify them.

---

## Related

- [Deployment](./deployment.md)
- [Search](./search.md)
- [Roadmap](./roadmap.md)