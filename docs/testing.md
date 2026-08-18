# Testing

Staffio ships a real test suite that exercises the core flows most likely to break silently: authentication, candidate lifecycle, departments, RBAC, CSRF, rate limiting and multi-tenant isolation. Tests run against an **isolated in-memory database** so they are fast, deterministic and safe.

---

## Running

```bash
make test          # go test ./...
make test-suite    # go test ./internal/tests/ -v
```

GitHub Actions CI runs `go vet`, `go test ./...` and a production build on every push/PR (see `.github/workflows/ci.yml`).

---

## Coverage Areas

The suite lives in `internal/tests/`:

| File | Covers |
|---|---|
| `auth_test.go` | Registration (creates org + owner), login, logout, unauthenticated redirect |
| `employee_test.go` | Candidate creation, field validation, status pipeline, deletion, org scoping |
| `department_test.go` | Department create + member assignment, member cleanup on delete, cross-org isolation |
| `csrf_test.go` | Accepts valid token, rejects missing and wrong tokens |
| `rbac_test.go` | Owner promotes admin, admin cannot promote admin, viewer cannot create, recruit redirected from dashboard |
| `tenant_test.go` | Multi-tenant data isolation between organizations |

Additional middleware behaviors (rate limiting) are validated in the same package.

---

## Testing Approach

The harness (`harness_test.go`) stands up a Gin router with the same middleware stack as production and drives requests through `httptest.NewRecorder` against an in-memory database. Example:

```go
func TestCreateEmployeeScopedToOrg(t *testing.T) {
    db := testDB()
    router := setupRouter(db)

    // register + login a user, then create an employee as that user
    // assert the row belongs to the user's organization
}
```

Assertions use the standard library plus `stretchr/testify` (`assert`/`require`).

---

## Tooling

| Tool | Role |
|---|---|
| `testing` (stdlib) | Test runner |
| `stretchr/testify` | Assertion helpers |
| `go.uber.org/mock` | Generated mocks where interfaces are isolated |
| `jordanlewis/gcassert` | Compile-time assertions for hot paths |

---

## Load Testing (k6)

Optional `k6` scripts validate behavior under concurrent load (requires [k6](https://k6.io)):

```bash
make k6-dashboard        # dashboard under concurrent users
make k6-department       # department operations
make k6-create           # employee creation throughput
```

Targets/expectations are set via env vars (`BASE_URL`, `ACCOUNTS`, `EMPLOYEES_PER_ACCOUNT`, `SLEEP_MS`, `ITERATIONS`).

---

## What Is Not Tested Yet

- **Session expiry edge cases** — tamper detection is delegated to `gorilla/securecookie`; an explicit test is planned.
- **E-mail delivery** — recovery/invite flows are mocked at the SMTP boundary.

---

## Related

- [Getting Started](./getting-started.md)
- [Architecture](./architecture.md)
- [Roadmap](./roadmap.md)