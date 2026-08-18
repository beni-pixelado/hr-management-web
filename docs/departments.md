# Departments

The department module gives HR teams an **organizational structure** alongside the candidate pipeline. Departments let you group employees, assign a manager, and organize your company into logical units.

---

## Data Model

```go
type Department struct {
    ID             uint   `gorm:"primaryKey" json:"id"`
    UserID         uint   `gorm:"not null;index" json:"user_id"`
    OrganizationID uint   `gorm:"index;default:0" json:"organization_id"`

    Code   string `json:"code" gorm:"not null"`
    Name   string `json:"name" gorm:"not null"`
    BossID uint   `json:"boss_id"`
}
```

- `Code` — a short unique identifier (e.g. `ENG`, `SALES`).
- `Name` — the department display name.
- `BossID` — an optional manager reference.
- `OrganizationID` — tenant scoping (see [Roles & Permissions](./roles-permissions.md)).

Membership is expressed through `Employee.DepartmentID`, linking candidates to a department.

---

## Operations

### Create (`POST /department`)

Reads `name`, `code`, and optional `boss_id`; validates `name`/`code` are non-empty; parses `boss_id`; and creates the row. Requires `CanAssignDepartment()` (owner/admin/recruit).

### List (`GET /department`)

Loads all employees (to populate the manager dropdown) and all departments, then renders `departments.html`.

### Detail (`GET /department/:id`)

Paginated detail view of a single department with its members.

### Assign / Remove members

- `POST /department/:id/add_employee` — assign an employee to the department (sets `Employee.DepartmentID`).
- `POST /department/:id/remove_employee` — unassign an employee.

Both require `CanAssignDepartment()`.

### Delete (`POST /department/:id/delete`)

Removes the department.

---

## Routes

| Method | Route | Role | Description |
|---|---|---|---|
| `GET` | `/department` | any | List + creation form |
| `POST` | `/department` | assign | Create department |
| `GET` | `/department/:id` | any | Detail + members |
| `POST` | `/department/:id/add_employee` | assign | Assign employee |
| `POST` | `/department/:id/remove_employee` | assign | Remove employee |
| `POST` | `/department/:id/delete` | assign | Delete department |

---

## UI

`departments.html` splits into two sections:

1. **Add New Department** — a form with `name`, `code`, and a manager `<select>` populated from employees.
2. **Departments Created** — a responsive card grid. Each card shows the department name, a colored code badge, and the assigned manager.

Department data also feeds the **Overview** pie chart (candidates per department) via `/api/overview/departments`. See [Reports & Analytics](./reports.md).

---

## Related

- [API Reference](./api.md)
- [Backend](./backend.md)
- [Database](./database.md)