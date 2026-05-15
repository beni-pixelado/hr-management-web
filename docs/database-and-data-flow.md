# Database — SQLite → PostgreSQL via Neon

## Evolution of the Data Layer

The database story of HR Management Web is a classic and important progression: start simple, validate your data model, then migrate to a production-grade system when the limitations become real. Understanding *why* each decision was made is as important as understanding *how* it was implemented.

## Phase 1 — SQLite

SQLite was the right choice for the initial version. It requires no separate server process, stores everything in a single file (`data/users.db`), and integrates with Go through the `mattn/go-sqlite3` CGO driver. For a developer validating a schema and a UI, the cycle of `go run` → test → adjust is faster when there's no database server to manage.

GORM's SQLite driver (`gorm.io/driver/sqlite`) wraps this transparently, so all queries were written against GORM's ORM API — and those queries remained valid when the driver was swapped.

## Phase 2 — PostgreSQL via Neon

### The Hard Limits of SQLite in Web Applications

SQLite enforces a **single-writer, file-level lock**. On any web application with concurrent HTTP requests, write operations queue behind each other, creating latency spikes under real load. Additionally, the database is a local file — which means:

- A cloud deployment (Render, Railway, Fly.io) using an ephemeral filesystem will **lose data on every redeploy**.
- You cannot run database migrations from a separate process or CI job without file access.
- Horizontal scaling (multiple application instances) is structurally impossible.

None of these were problems during the prototyping phase, but all of them become blockers the moment the application needs to be hosted or shared.

### Why PostgreSQL

PostgreSQL handles true concurrent reads and writes via MVCC (Multi-Version Concurrency Control), meaning readers never block writers and writers never block readers. It supports the `ILIKE` operator used by the search system, native `UUID` types, `JSONB` columns for future unstructured data, and a mature extension ecosystem. It is also the database that Neon is built on — making this a natural combination.

### Why Neon Specifically

[Neon](https://neon.tech) is a serverless PostgreSQL platform that solves deployment friction:

The **serverless model** means compute scales to zero when the database has no connections, eliminating idle-time cost. For a portfolio project or low-traffic production app, this makes the cost effectively zero until real traffic arrives.

**Database branching** is a standout feature: Neon allows you to create an isolated copy of your production database in seconds, which can be used as a staging environment or a test fixture that is reset after each test run. This is invaluable for testing schema migrations safely.

The **built-in connection pooler** (compatible with PgBouncer) handles connection multiplexing at the platform level, which means the application's `MaxOpenConns` setting is an additional layer of safety rather than the only protection against connection exhaustion.

## Connection Configuration

The connection is managed through `internal/db/db.go`. GORM uses the `pgx/v5` driver (`gorm.io/driver/postgres`), which in turn uses `jackc/pgx/v5` as the wire protocol implementation and `jackc/puddle/v2` for connection pooling at the driver level.

```go
// internal/db/db.go
package db

import (
    "os"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/joho/godotenv"
)

func Connect() (*gorm.DB, error) {
    // Load .env in development — production environments inject vars directly
    godotenv.Load()

    dsn := os.Getenv("DATABASE_URL")

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }

    // Configure the underlying sql.DB connection pool
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // MaxOpenConns: limits total connections to the database.
    // Neon's free tier supports up to 100 concurrent connections;
    // keeping this at 10 ensures we never exhaust the pool under burst traffic.
    sqlDB.SetMaxOpenConns(10)

    // MaxIdleConns: connections kept open even when idle.
    // 5 avoids the overhead of re-establishing TCP+TLS connections on every request.
    sqlDB.SetMaxIdleConns(5)

    return db, nil
}
```

### The `DATABASE_URL` Format

For Neon, the connection string looks like this:

```
postgres://username:password@ep-xyz.us-east-2.aws.neon.tech/dbname?sslmode=require
```

The `?sslmode=require` parameter is essential for Neon — all connections must use TLS. For a local PostgreSQL instance during development, replace with `?sslmode=disable`.

## Schema Management

GORM's `AutoMigrate` function creates and updates tables based on Go struct definitions. This is run at application startup before the HTTP server begins accepting requests:

```go
// Inside main.go (conceptual)
db.AutoMigrate(&models.User{}, &models.Employee{})
```

`AutoMigrate` is **additive only** — it will create missing tables and add missing columns, but it will never drop columns or tables. This is safe to run on every startup and avoids the need for a migration runner in the early stages of the project. For v2.0, a proper migration tool (like `goose` or `atlas`) is planned to handle destructive changes safely.

## Data Models

The two primary entities are `User` (authentication records) and `Employee` (candidate records). Both use GORM's embedded `gorm.Model` struct, which provides `ID` (uint primary key), `CreatedAt`, `UpdatedAt`, and `DeletedAt` (soft delete support).

```go
type Employee struct {
    gorm.Model
    Name     string `gorm:"not null"`
    Position string `gorm:"not null"`
    Email    string `gorm:"uniqueIndex;not null"`
    Status   string `gorm:"default:'Pending'"`
    PhotoURL string
}
```

The `uniqueIndex` on `Email` enforces uniqueness at the database level (not just application level), and the `default:'Pending'` tag means new rows always start in the correct initial state even if the application layer fails to set it explicitly.

## Future Optimizations

When the dataset grows beyond a few thousand rows, `ILIKE` queries with leading wildcards (`%term%`) require a full sequential scan. The planned fix is a GIN index on the search columns using PostgreSQL's `pg_trgm` extension:

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_employees_name_trgm ON employees USING GIN (name gin_trgm_ops);
CREATE INDEX idx_employees_position_trgm ON employees USING GIN (position gin_trgm_ops);
```

This turns trigram-based `ILIKE` search from `O(n)` to effectively `O(log n)` on large datasets, and Neon supports `pg_trgm` natively.