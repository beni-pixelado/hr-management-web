# Database — Models, Migrations & Data Flow

Staffio uses **PostgreSQL** through **GORM**. The data layer is deliberately multi-tenant: every domain model carries an `OrganizationID` and all queries are scoped to the acting user's organization.

---

## Connection

`backend/database/database.go` parses `DATABASE_URL` with `pgx/v5` and opens a GORM connection over the `stdlib` adapter:

```go
connConfig, err := pgx.ParseConfig(dsn)
connConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe
sqlDB := stdlib.OpenDB(*connConfig)
DB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
```

Using the `pgx` driver directly (rather than the GORM postgres default) gives control over the wire protocol and connection behavior.

---

## Migrations

Migrations run automatically on startup via `db.AutoMigrate` for every model, in dependency order:

`Organization → User → Employee → Department → PasswordResetToken → Report → Absence → Termination`

The startup also:
- Drops a legacy `password_reset_tokens.code` column if present.
- **Backfills** legacy users (`organization_id = 0`) into their own `Organization`, promotes them to `owner`, and migrates their existing employees/departments/reports/absences into the new org.

---

## Data Models

### Organization

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `Name` | `string` | Not null |
| `CreatedAt` | `time.Time` | Auto |

### User

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `Username` | `string` | Unique, not null |
| `Password` | `string` | bcrypt hash |
| `Email` | `string` | Unique, not null |
| `Photo` | `string` | Cloudinary URL |
| `ResetToken` / `ResetExpires` | `string` / `int64` | Legacy recovery fields |
| `OrganizationID` | `uint` | Indexed tenant |
| `Role` | `string` | `owner` / `admin` / `recruit` / `viewer` |

### Employee

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `CreatedAt` / `UpdatedAt` | `time.Time` | Auto |
| `UserID` | `uint` | Creator |
| `OrganizationID` | `uint` | Indexed tenant |
| `FullName` / `Email` / `Position` | `string` | Not null |
| `Description` | `string` | Free text |
| `Status` | `string` | `pending` / `contractors` / `rejected` |
| `HireDate` | `string` | |
| `Photo` | `string` | Cloudinary URL |
| `DepartmentID` | `uint` | Optional assignment |
| `Absences` | `uint` | Running counter |
| `RejectedAt` | `*time.Time` | Set on rejection |

### Department

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `UserID` | `uint` | Creator |
| `OrganizationID` | `uint` | Indexed tenant |
| `Code` | `string` | Not null |
| `Name` | `string` | Not null |
| `BossID` | `uint` | Optional manager |

### Absence

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `CreatedAt` | `time.Time` | Auto |
| `UserID` | `uint` | Who recorded |
| `EmployeeID` | `uint` | Indexed |
| `OrganizationID` | `uint` | Indexed tenant |

### Termination

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `CreatedAt` | `time.Time` | Auto |
| `OrganizationID` | `uint` | Indexed tenant |
| `EmployeeID` | `uint` | |
| `EmployeeName` | `string` | Snapshot |
| `Reason` | `string` | `rejected` / `deleted` |

### Report

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `CreatedAt` | `time.Time` | Auto |
| `UserID` | `uint` | Creator |
| `OrganizationID` | `uint` | Indexed tenant |
| `Type` | `string` | `absences` / `hired` / `fired` |
| `Period` | `string` | `week` / `month` / `year` |
| `Days` | `int` | Derived period length (7/30/365) |
| `Total` | `int` | Computed count |

### PasswordResetToken

| Field | Type | Notes |
|---|---|---|
| `ID` | `uint` | PK |
| `Email` | `string` | Indexed |
| `Token` | `string` | Unique, one-time |
| `ExpiresAt` | `int64` | Expiry epoch |
| `CreatedAt` | `time.Time` | Auto |

---

## Tenant Scoping Pattern

Every organization-bound query is filtered by `organization_id` **and** the row's owner scoping where applicable. Example from `GetEmployees`:

```go
countQuery := DB.Model(&Employee{}).
    Where("organization_id = ?", orgID)
```

This pattern appears throughout: employees, departments, reports, absences and terminations. It makes cross-tenant leakage structurally impossible as long as handlers resolve the actor's org before querying.

---

## Data Flows

### Flow 1 — Login

```
Browser → POST /login → auth.go handler
  → DB.Where(username/email).First(user)
  → bcrypt.CompareHashAndPassword
  → auth.CreateSession → Set-Cookie: hr_session (HMAC-signed)
  → 302 → /dashboard
```

### Flow 2 — Candidate creation (with photo)

```
Browser → POST /employees (multipart)
  → validate full_name/email/position
  → saveUploadedImage → Cloudinary (size + MIME checked)
  → DB.Create(&Employee{Status:"pending"})
  → 302 → /employees
```

### Flow 3 — Status update → rejection

```
Browser → POST /employees/42/status {status:"rejected"}
  → RBAC guard (recruit vs superior)
  → DB.Model(...).Updates(map)  // status, hire_date, rejected_at
  → DB.Create(&Termination{Reason:"rejected"})
  → 302 → /employees
```

### Flow 4 — Absence marking (atomic)

```
Browser → POST /employees/42/absence
  → DB.Transaction:
       create Absence{EmployeeID:42}
       UpdateColumn("absences", absences + 1)
  → 200 JSON
```

### Flow 5 — Report metrics

```
Report Absences / Hired / Fired
  → countAbsencesSince(org, days)     // Absence rows within window
  → countHiredSince(org, days)        // contractors with hire_date within window
  → countFiredSince(org, days)        // rejected_at + terminations within window
  → store Report{Type, Period, Days, Total}
```

---

## Related

- [Architecture](./architecture.md)
- [Reports & Analytics](./reports.md)
- [Roles & Permissions](./roles-permissions.md)