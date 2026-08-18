# Search

Staffio lets HR staff locate candidates instantly without paging through long lists. Search is **server-side**, case-insensitive, and integrated with filtering and pagination.

---

## Where Search Lives

Search is exposed as query parameters on existing routes — there is no separate `/search` endpoint.

| Route | Parameter | Scope |
|---|---|---|
| `GET /dashboard` | `?search=` | Filters the dashboard employee grid by name/email |
| `GET /employees` | `?status=` + `?page=` | Filters the candidate table by status and pages through it |
| `GET /api/employees` | `?search=` + `?status=` | JSON API filtering |

> **Note:** The original multi-field `ILIKE` search (name/position/email) has evolved into the current status-driven `GET /employees` list plus dashboard name/email search. Position-based matching can be re-added via the search query in `GetEmployees` if desired.

---

## Dashboard Search

`GET /dashboard?search=engineer`:

```go
query := database.DB.Where("organization_id = ?", user.OrganizationID)

if search != "" {
    query = query.Where(
        "full_name LIKE ? OR email LIKE ?",
        "%"+search+"%",
        "%"+search+"%",
    )
}
```

The query is scoped to the acting user's organization and filtered case-insensitively across **full name** and **email**.

---

## Status Filtering & Pagination

`GET /employees` builds a query filtered by org and optional `status`, then paginates 20 rows per page:

```go
status := c.DefaultQuery("status", "all")

countQuery := DB.Model(&Employee{}).
    Where("organization_id = ?", orgID)
if statusFilter != "all" {
    countQuery = countQuery.Where("status = ?", statusFilter)
}
countQuery.Count(&totalFiltered)

listQuery := DB.Where("organization_id = ?", orgID)
if statusFilter != "all" {
    listQuery = listQuery.Where("status = ?", statusFilter)
}
listQuery.Limit(limit).Offset(offset).Find(&employees)
```

Friendly status aliases map to stored values:

| Alias | Stored |
|---|---|
| `interviewing` | `pending` |
| `hired` | `contractors` |
| `rejected` | `rejected` |

The handler also computes per-status totals (`pending`/`contractors`/`rejected`) for the UI badges.

---

## CSV Export

`GET /employees/download` exports the filtered set (optionally by status) as a CSV attachment — a natural companion to search for offline reporting.

---

## Security

- All search input flows through **GORM parameterized queries** (`?` placeholders), making SQL injection structurally impossible.
- Every query is scoped to `organization_id`, so a user can only ever search their own org's data.
- Search terms are used only as data — never concatenated into SQL.

---

## Performance

| Dataset size | Strategy | Latency (approx.) |
|---|---|---|
| < 10,000 rows | Sequential scan (current) | < 5 ms |
| 10,000–100,000 rows | GIN trigram index (planned) | < 10 ms |
| > 100,000 rows | Full-text search with `tsvector` | < 20 ms |

The planned GIN trigram indexes are documented in [Scalability](./scalability.md).

---

## Related

- [Scalability](./scalability.md)
- [Backend](./backend.md)
- [API Reference](./api.md)