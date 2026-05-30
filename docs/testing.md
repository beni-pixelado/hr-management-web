# Testing

## Overview

The test suite covers the two layers most likely to break silently: handler logic and the authentication flow. Unit tests for individual utility functions (password hashing, session helpers) are also included.

## Tools

| Tool | Role |
|---|---|
| `testing` (stdlib) | Test runner |
| `stretchr/testify` | Assertion helpers (`assert.Equal`, `require.NotNil`) |
| `go.uber.org/mock` | Generated mock types for database interfaces |
| `jordanlewis/gcassert` | Compile-time assertions for performance-critical paths |

## Running Tests

```bash
make test
# or directly:
go test ./...
```

## Integration Tests

Integration tests validate complete HTTP request cycles against a test database. They spin up a Gin router with the same middleware stack as production, send requests via `httptest.NewRecorder`, and assert on status codes, headers, and body content.

```go
func TestCreateEmployee(t *testing.T) {
    db := testDB()
    router := setupRouter(db)

    body := strings.NewReader(`full_name=Alice&position=Engineer&email=alice@example.com`)
    req  := httptest.NewRequest("POST", "/employees", body)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    addValidSession(req)

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, http.StatusFound, w.Code)

    var count int64
    db.Model(&Employee{}).Where("email = ?", "alice@example.com").Count(&count)
    assert.Equal(t, int64(1), count)
}
```

## What Is Not Tested Yet

- **Load tests** — planned using `k6` or `vegeta`. Target: search query under 100 concurrent users on a 10,000-row table should remain under 50 ms.
- **Department CRUD integration tests** — tracked for v1.2 alongside the collaborator membership feature.
- **Session expiry edge cases** — cookie tamper detection is covered by `gorilla/securecookie` internally but should have an explicit test.