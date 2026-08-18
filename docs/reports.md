# Reports & Analytics

Staffio turns raw recruitment and HR activity into **live metrics and shareable reports**. The analytics are scoped per organization and available to admins/owners.

---

## Overview Dashboard

The overview page (`/overview`) renders live charts backed by JSON APIs:

| Endpoint | Chart | Data |
|---|---|---|
| `/api/overview/departments` | Pie | Candidates grouped by department |
| `/api/overview/employees` | Line | New candidates per day + a merged "fired" series |

The fired series merges `rejected_at` timestamps with logged **terminations**, so deleted/rejected records stay visible in history.

**Access:** `admin`/`owner` only; `recruit` users are redirected to `/employees`.

---

## Absences

- An absence is recorded via `POST /employees/:id/absence`, which creates an `Absence` row and **atomically increments** the employee's `Absences` counter in a single transaction.
- Absence events are used both for reports and the overview.

---

## Terminations

A `Termination` is logged automatically whenever:

- A candidate is **rejected** (`UpdateEmployeeStatus` with `rejected`), or
- An employee is **deleted** (both JSON and form delete paths).

`Reason` distinguishes `rejected` vs `deleted`. This keeps an audit trail even after soft-deletes.

---

## Reports

Reports let admins capture a snapshot metric over a period. Periods map to day windows:

| Period | Days |
|---|---|
| `week` | 7 |
| `month` | 30 |
| `year` | 365 |

### Report types

| Type | Counts |
|---|---|
| `absences` | `Absence` rows created within the window |
| `hired` | `contractors` with a `hire_date` within the window |
| `fired` | `rejected_at` + `Termination` rows within the window |

### Routes

| Method | Route | Description |
|---|---|---|
| `GET` | `/report` | Report list + aggregated totals |
| `GET` | `/report/new` | Report creation form |
| `POST` | `/report/new` | Create a report snapshot |
| `GET` | `/report/:id` | Report detail with narrative summary + stats |
| `GET` | `/api/report/absences` | Live absence count (`?days=`) |
| `GET` | `/api/report/hired` | Live hired count (`?days=`) |
| `GET` | `/api/report/fired` | Live fired count (`?days=`) |

### Report detail

`ReportDetailHandler` renders a human-readable narrative (e.g. *"Between Jan 01 and Jan 07, a span of 7 days, your team recorded 12 absences — an average of 2 per day…"*) plus a stat list, computed per report type:

- **Hired** — hires in period, average per day, total hired, in-interview count, rejected count.
- **Fired** — fired in period, average per day, current rejected, total hired.
- **Absences** — absences in period, average per day, all-time absences, affected employees.

---

## Related

- [API Reference](./api.md)
- [Database](./database.md)
- [Backend](./backend.md)