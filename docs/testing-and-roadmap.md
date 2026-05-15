# Testing Strategy

## Overview

The test suite for HR Management Web covers the two layers most likely to break silently: the handler logic (does a request to `POST /employees` actually create a row?) and the authentication flow (does a request without a valid session get rejected?). Unit tests for individual utility functions (password hashing, session helpers) are also included.

## Tools

The testing stack uses the standard `testing` package extended by `stretchr/testify` for assertion helpers (`assert.Equal`, `assert.NoError`, `require.NotNil`) that produce readable failure messages. `go.uber.org/mock` provides generated mock types for database interfaces, allowing handler tests to run without a live database connection. `jordanlewis/gcassert` provides compile-time assertion macros for validating performance-critical code properties at build time rather than runtime.

## Running Tests

```bash
make test
# or directly:
go test ./...
```

## Integration Tests

Integration tests validate complete HTTP request cycles against a test database. They spin up a Gin router with the same middleware stack as production, send HTTP requests using `httptest.NewRecorder`, and assert on the response status code, headers, and body. The test database is either a Neon branch (in CI) or a local PostgreSQL instance with a separate `DATABASE_URL_TEST` environment variable.

```go
func TestCreateEmployee(t *testing.T) {
    db := testDB() // connects to the test database
    router := setupRouter(db)

    body := strings.NewReader(`name=Alice&position=Engineer&email=alice@example.com`)
    req := httptest.NewRequest("POST", "/employees", body)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    addValidSession(req) // helper that attaches a valid session cookie

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusFound, w.Code) // expects redirect after creation

    var count int64
    db.Model(&Employee{}).Where("email = ?", "alice@example.com").Count(&count)
    assert.Equal(t, int64(1), count)
}
```

## What Is Not Tested (Yet)

Load and stress tests (`docs/` references a `tests/loads/` path) are planned but not yet implemented. These would use a tool like `k6` or `vegeta` to validate response times under concurrent user load, particularly for the search query path. Search under 100 concurrent users hitting `ILIKE` on a 10,000-row table should remain under 50ms — a benchmark that becomes important before enabling GIN trigram indexes.

---

# Roadmap

## v1.1 — Current Release ✅

The primary deliverables of this release were a modernized UI, a functional search system, the database migration from SQLite to PostgreSQL via Neon, session-based authentication, and pagination. All of these are live in the current codebase.

## v1.2 — Authentication Hardening

The next milestone focuses on making the authentication layer more robust. JWT-based authentication will be added as an option for API consumers (the future REST layer). Role-based access control will restrict sensitive operations — account deletion, bulk status changes, user management — to users with the Admin role. Password change and account deactivation flows are also planned, making the system usable as a long-term multi-user tool rather than a single-team deployment.

## v1.3 — Enhanced Data Model

This release focuses on the richness of candidate information. A notes and comments system will allow HR staff to attach freeform context to each candidate (interview impressions, contact history, concerns). Interview scheduling will add date tracking and calendar-friendly iCal export. Department and team assignment adds organizational structure to the data model. An audit trail will log every status change with a timestamp and the user who made it — essential for compliance in real HR workflows.

## v2.0 — Architecture Evolution

The most significant architectural change: a full REST API backend with OpenAPI/Swagger documentation generated from Go struct annotations. The server-rendered HTML frontend will be progressively enhanced with HTMX for partial page updates (avoiding a full React rewrite while gaining the responsiveness of a SPA). Docker and Docker Compose will provide a one-command local development environment. GitHub Actions CI will run the test suite, lint, and build on every push. GIN trigram indexes on the search columns will be added once the dataset exceeds 10,000 rows.