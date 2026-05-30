# Data Flow — Request Lifecycle & Key Flows

## Overview

This document traces the path of data through the system for the most common operations: login, candidate creation, department creation, search, and status update.

---

## Flow 1 — Login

```
Browser                  Gin Router              auth.go handler        internal/auth
   │                          │                        │                      │
   │  POST /login             │                        │                      │
   │  {username,email,pass}   │                        │                      │
   │─────────────────────────▶│                        │                      │
   │                          │  handler(c)            │                      │
   │                          │───────────────────────▶│                      │
   │                          │                        │  DB.Where(user)      │
   │                          │                        │──────────────────────▶ PostgreSQL
   │                          │                        │◀──────────────────────│
   │                          │                        │  bcrypt.Compare()     │
   │                          │                        │  (fail → re-render)   │
   │                          │                        │                      │
   │                          │                        │  auth.CreateSession() │
   │                          │                        │─────────────────────▶│
   │                          │                        │  Set-Cookie: hr_session
   │◀─────────────────────────│  302 → /dashboard      │                      │
```

---

## Flow 2 — Authenticated Request (Middleware Gate)

```
Browser              Gin Router          RequireAuth          handler
   │                     │                   │                   │
   │  GET /employees     │                   │                   │
   │  Cookie: hr_session │                   │                   │
   │────────────────────▶│                   │                   │
   │                     │  middleware runs   │                   │
   │                     │──────────────────▶│                   │
   │                     │                   │  IsAuthenticated() │
   │                     │                   │──▶ read cookie     │
   │                     │                   │◀── (bool, userID)  │
   │                     │                   │                   │
   │                     │  [invalid] 302 /login                 │
   │◀────────────────────│───────────────────│                   │
   │                     │                   │                   │
   │                     │  [valid] c.Next() │                   │
   │                     │                   │──────────────────▶│
   │                     │                   │                   │  DB query
   │                     │                   │                   │──▶ PostgreSQL
   │                     │                   │                   │◀──
   │◀────────────────────│  200 HTML         │◀──────────────────│
```

---

## Flow 3 — Candidate Creation (with Photo Upload)

```
Browser                         CreateEmployee handler
   │                                     │
   │  POST /employees                    │
   │  multipart/form-data                │
   │  {full_name, email, position, photo}│
   │────────────────────────────────────▶│
   │                                     │
   │                                     │  validate required fields
   │                                     │  (400 if missing)
   │                                     │
   │                                     │  file != nil?
   │                                     │  ├─ check size ≤ 5MB
   │                                     │  ├─ detect MIME type
   │                                     │  ├─ validate allowed types
   │                                     │  ├─ generate UUID filename
   │                                     │  └─ SaveUploadedFile → ./uploads/
   │                                     │
   │                                     │  DB.Create(&Employee{...})
   │                                     │──────────────────▶ PostgreSQL
   │                                     │◀──────────────────
   │◀────────────────────────────────────│  302 → /employees
```

---

## Flow 4 — Search + Pagination

```
Browser                         GetEmployees handler              PostgreSQL
   │                                     │                             │
   │  GET /employees?q=eng&page=2        │                             │
   │────────────────────────────────────▶│                             │
   │                                     │                             │
   │                                     │  parse q="eng", page=2      │
   │                                     │  offset = (2-1)*20 = 20     │
   │                                     │                             │
   │                                     │  query = DB.Model(&Employee{})
   │                                     │  query.Where(ILIKE x3)      │
   │                                     │                             │
   │                                     │  query.Count(&total) ───────▶
   │                                     │◀──────────────────── total  │
   │                                     │                             │
   │                                     │  query.Limit(20).Offset(20) │
   │                                     │  .Find(&employees) ─────────▶
   │                                     │◀──────────────────── rows   │
   │                                     │                             │
   │                                     │  render employees.html      │
   │◀────────────────────────────────────│  {employees, page, total}   │
```

The `Count` and `Find` calls share the same `query` object — the `WHERE` clause is applied once and used for both, ensuring pagination totals reflect the filtered result.

---

## Flow 5 — Department Creation

```
Browser                         CreatedepartmentHandler
   │                                     │
   │  POST /department                   │
   │  {name, code, boss_id}              │
   │────────────────────────────────────▶│
   │                                     │
   │                                     │  validate name != "" && code != ""
   │                                     │  (400 JSON if invalid)
   │                                     │
   │                                     │  parse boss_id as uint64
   │                                     │  (400 JSON if non-empty and invalid)
   │                                     │
   │                                     │  DB.Create(&Department{...})
   │                                     │──────────────────▶ PostgreSQL
   │                                     │◀──────────────────
   │◀────────────────────────────────────│  302 → /department
```

---

## Flow 6 — Status Update

```
Browser                         UpdateEmployeeStatus handler
   │                                     │
   │  POST /employees/42/status          │
   │  {status: "contractors",            │
   │   hire_date: "2026-01-15"}          │
   │────────────────────────────────────▶│
   │                                     │
   │                                     │  id = c.Param("id")  → "42"
   │                                     │  status = c.PostForm("status")
   │                                     │  hireDate = c.PostForm("hire_date")
   │                                     │
   │                                     │  DB.Model(&employee).
   │                                     │    Where("id = ?", id).
   │                                     │    Updates(map[string]interface{}{
   │                                     │      "status":    status,
   │                                     │      "hire_date": hireDate,
   │                                     │    })
   │                                     │──────────────────▶ PostgreSQL
   │                                     │◀──────────────────
   │◀────────────────────────────────────│  302 → /employees
```

Using a `map` for `Updates` (rather than a struct) allows zero-value fields to be updated explicitly. GORM skips zero-value fields when updating from a struct.

---

## Session Cookie Lifecycle

```
Registration / Login
  └─ bcrypt hash password
  └─ DB.Create(user) / DB.Where(user).First()
  └─ auth.CreateSession() → gorilla/sessions → Set-Cookie: hr_session (HMAC-signed)

Every protected request
  └─ RequireAuth middleware
  └─ auth.IsAuthenticated() → gorilla/sessions.Get() → verify HMAC → decode values
  └─ c.Set("user_id", userID) → available to downstream handlers

Logout
  └─ auth.DestroySession() → session.Options.MaxAge = -1 → Set-Cookie (expiry in past)
  └─ Browser deletes cookie
```