# Security

Staffio applies defense-in-depth across the application: CSRF protection on all state-changing requests, rate limiting on auth routes, RBAC-based authorization, tenant scoping, and safe-by-default rendering.

---

## CSRF Protection

Every mutating request (POST/DELETE) must carry a per-session CSRF token.

**Mechanism** (`internal/csrf/csrf.go`):

- A 32-byte token is generated and stored in the session.
- On **GET/HEAD**, the middleware guarantees a token exists and exposes it to templates via the `csrfField` template function.
- On **state-changing** requests, the submitted `_csrf` field must equal the session token, otherwise the request is rejected with `403 Forbidden`.

```go
// Every <form> includes the token automatically:
{{csrfField}}

// Global middleware on the whole router:
r.Use(csrf.Protect())
```

Because the middleware runs globally (not just on the protected group), even public forms (login, register, recovery) are protected.

---

## Rate Limiting

Auth endpoints are rate-limited per client IP using an in-memory fixed-window limiter (`internal/middleware/ratelimit.go`):

| Route | Limit |
|---|---|
| `POST /login` | 10 / minute |
| `POST /register` | 10 / minute |
| `POST /recuperateaccount` | 5 / minute |
| `POST /reset-password` | 5 / minute |

Exceeding the limit returns `429 Too Many Requests`. The limiter resets on restart, which is acceptable at this scale and avoids external infrastructure.

---

## Authentication Hardening

- **bcrypt** password hashing (never plaintext).
- **HMAC-signed** session cookies (`SESSION_SECRET`).
- Cookie flags: `HttpOnly`, `Secure`, `SameSite=Lax`.
- Vague login errors and recovery responses to prevent **account enumeration**.
- Single-use, time-limited password-reset tokens.
- `SESSION_SECRET` is required at startup — the server panics if it is unset.

---

## Authorization (RBAC)

Authorization is enforced at three layers:

1. **Middleware** — `RequireAuth` gates every protected route; `BlockViewerWrites` rejects viewer write attempts globally.
2. **Handler guards** — domain handlers call predicates like `CanEditEmployees()`, `CanManageTeam()`, `CanAssignDepartment()` before acting.
3. **Object-level rules** — e.g. a `recruit` cannot reject a candidate whose email belongs to a superior, and only the owner can transfer ownership.

See [Roles & Permissions](./roles-permissions.md) for the full matrix.

---

## Multi-Tenant Isolation

Every organization-owned model carries an `OrganizationID`, and all queries are scoped to the acting user's org. Combined with object-level guards, cross-tenant data access is structurally prevented.

---

## Injection & XSS Prevention

| Vector | Mitigation |
|---|---|
| **SQL injection** | GORM parameterized queries (`?` placeholders) — user input is never interpolated into SQL |
| **XSS** | `html/template` auto-escapes all rendered user data; `csrfField` returns `template.HTML` only for the server-generated token |
| **File upload** | Size cap (5 MB), MIME-type sniffing against an allowlist, UUID filenames, Cloudinary-hosted (no local path traversal) |
| **Path traversal** | Uploads are addressed by UUID public IDs |

---

## Trusted Proxies & Headers

`r.SetTrustedProxies(nil)` disables trust of `X-Forwarded-For`, so client IPs (used by rate limiting) are taken directly from the connection. Configure trust explicitly if you deploy behind a reverse proxy.

---

## Missing / Legacy Surface

- A legacy `/debug/cookie` route and leftover SQLite driver remain in the codebase; these are tracked for removal in the [Roadmap](./roadmap.md).
- Load/DoS testing is planned via the included `k6` scripts.

---

## Related

- [Authentication](./authentication.md)
- [Roles & Permissions](./roles-permissions.md)
- [Architecture](./architecture.md)