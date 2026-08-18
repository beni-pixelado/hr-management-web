# API Reference

> **Machine-readable spec:** a full OpenAPI 3.0 description is available at [`openapi.yaml`](./openapi.yaml). Use it with Swagger UI, Redoc or any API client generator.

This is the complete route table as registered in `backend/cmd/server/main.go`. The application is primarily **server-rendered HTML**; the `/api/*` endpoints return JSON for charts and live analytics.

Legend — **Auth**: whether the route requires a valid session. **Role**: the minimum access needed (see [Roles & Permissions](./roles-permissions.md)).

---

## Public Routes

| Method | Route | Description |
|---|---|---|
| `GET` | `/` | Product landing page |
| `GET` | `/login` | Login form |
| `POST` | `/login` | Authenticate (rate-limited: 10/min) |
| `GET` | `/register` | Registration form |
| `POST` | `/register` | Create account + organization (rate-limited: 10/min) |
| `GET` | `/recuperateaccount` | Password recovery form |
| `POST` | `/recuperateaccount` | Send recovery e-mail (rate-limited: 5/min) |
| `GET` | `/reset-password` | Reset form (`?token=`) |
| `POST` | `/reset-password` | Set new password (rate-limited: 5/min) |
| `GET` | `/robots.txt` | SEO robots file |
| — | *(anything else)* | Renders `404.html` |

---

## Protected Routes

All routes below are wrapped in `RequireAuth → LoadUser → BlockViewerWrites`.

### Dashboard

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/dashboard` | any (recruit → `/employees`) | Metrics + employee grid (`?search=` / `?all=`) |
| `GET` | `/logout` | any | Destroy session |

### Candidates

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/employees` | view+ | List, `?status=`/`?page=` filters |
| `GET` | `/employees/download` | view+ | Export CSV (optionally filtered) |
| `POST` | `/employees` | edit | Create candidate (photo upload) |
| `GET` | `/employees/:id/edit` | edit | Edit form |
| `POST` | `/employees/:id/edit` | edit | Update candidate |
| `POST` | `/employees/:id/status` | edit | Set status + hire date |
| `POST` | `/employees/:id/absence` | edit | Mark an absence |
| `DELETE` | `/employees/:id` | edit | Delete (JSON) |
| `POST` | `/employees/:id/delete` | edit | Delete (form) |
| `GET` | `/badge/:id` | any | Print-friendly ID card |

### Departments

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/department` | any | List + creation form |
| `POST` | `/department` | assign | Create department |
| `GET` | `/department/:id` | any | Detail + members |
| `POST` | `/department/:id/add_employee` | assign | Assign employee |
| `POST` | `/department/:id/remove_employee` | assign | Remove employee |
| `POST` | `/department/:id/delete` | assign | Delete department |

### Analytics & Reports

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/overview` | admin/owner | Overview page with charts |
| `GET` | `/report` | admin/owner | Report list |
| `GET` | `/report/new` | admin/owner | Report creation form |
| `POST` | `/report/new` | admin/owner | Create report |
| `GET` | `/report/:id` | admin/owner | Report detail |

### JSON API Endpoints

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/api/overview/departments` | any | Candidates per department (pie) |
| `GET` | `/api/overview/employees` | any | New candidates + fired per day (line) |
| `GET` | `/api/report/absences` | any | Absence count (`?days=`) |
| `GET` | `/api/report/hired` | any | Hired count (`?days=`) |
| `GET` | `/api/report/fired` | any | Fired count (`?days=`) |

### Account & Configuration

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/config` | any | Configuration menu |
| `GET` | `/config/account` | any | Account settings |
| `POST` | `/config/account/profile` | any | Update username/email |
| `POST` | `/config/account/photo` | any | Upload profile photo |
| `POST` | `/config/account/password` | any | Change password |
| `POST` | `/config/account/delete` | any | Delete account |
| `GET` | `/config/device` | any | Device settings |

### Team (multi-tenant)

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/team` | manage | List members |
| `POST` | `/team/invite` | manage | Invite / update a member |
| `POST` | `/team/:id/role` | manage | Change a member's role |
| `POST` | `/team/:id/remove` | manage | Remove a member |
| `POST` | `/team/:id/transfer` | owner | Transfer ownership |

---

## Error Conventions

- **HTML pages**: render a form with an actionable user-facing message (re-render on error).
- **API endpoints**: return JSON envelopes like `{"error": "..."}` or `{"total": N}`.
- **Authorization failures**: `403` via `abortUnauthorized`, or a redirect to `/dashboard` for viewer-write blocks.
- **Rate limiting**: `429 Too Many Requests`.

---

## Related

- [Backend](./backend.md)
- [Reports & Analytics](./reports.md)
- [Roles & Permissions](./roles-permissions.md)