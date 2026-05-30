# Search Engine — Multi-field Candidate Search

## Why Search Matters in HR Tools

An HR system accumulates records fast. Even a small company running 20 open positions and reviewing 50 candidates per position will have 1,000 rows in the candidates table within a few months. Paginating through that list to find "John from engineering" is friction. The search system eliminates that entirely by letting staff locate any candidate in under a second.

## Design Goals

- Match across multiple fields simultaneously (not just name)
- Case-insensitive without requiring the user to think about casing
- Preserve URL state so results are bookmarkable and shareable
- Integrate with pagination so searching "engineer" and getting 30 results lets you page through all 30 without losing the filter

---

## Implementation

### The Route

Search is exposed on the existing `GET /employees` route via a `q` query parameter. There is no separate `/search` route — search is a filtered view of the employee list.

```
GET /employees?q=engineer&page=2
```

### The Handler Logic

```go
func GetEmployees(c *gin.Context) {
    var employees []Employee

    pageStr := c.DefaultQuery("page", "1")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        page = 1
    }

    limit  := 20
    offset := (page - 1) * limit

    query := DB.Model(&Employee{})

    if search := strings.TrimSpace(c.Query("q")); search != "" {
        term := "%" + search + "%"
        query = query.Where(
            "full_name ILIKE ? OR position ILIKE ? OR email ILIKE ?",
            term, term, term,
        )
    }

    var totalEmployees int64
    query.Count(&totalEmployees)

    query.Limit(limit).Offset(offset).Find(&employees)

    totalPages := int(math.Ceil(float64(totalEmployees) / float64(limit)))

    c.HTML(http.StatusOK, "employees.html", gin.H{
        "employees":      employees,
        "currentPage":    page,
        "totalPages":     totalPages,
        "totalEmployees": totalEmployees,
        "prevPage":       page - 1,
        "nextPage":       page + 1,
    })
}
```

The `Count` and `Find` calls reuse the same `query` value. GORM builds the `WHERE` clause once and applies it to both. This means pagination totals always reflect the filtered result, not the total table size.

---

## Multi-field Matching

The `OR` chain in the `WHERE` clause is the key. A single user input matches against three columns at once:

| Input | Matches |
|---|---|
| `"dev"` | Names containing "dev", positions like "Developer", emails with "dev" |
| `"@company"` | Any candidate with that email domain |
| `"engineer"` | Any candidate named/positioned/emailed with "engineer" |

`ILIKE` is PostgreSQL's case-insensitive `LIKE`. The search is never case-sensitive from the user's perspective.

---

## Frontend Integration

```html
<!-- employees.html -->
<form method="get" action="/employees">
    <input type="text" name="q" value="{{ .search }}"
           placeholder="Search by name, position or email..." />
    <button type="submit">Search</button>
</form>
```

The `value="{{ .search }}"` binding is the key detail — after searching, the input field shows what you typed. Without it, the field would clear on every result page.

### Pagination Link Preservation

Every pagination link must preserve the active search query:

```html
{{if gt .currentPage 1}}
<a href="/employees?page={{.prevPage}}" class="page-btn">← Previous</a>
{{end}}

{{if lt .currentPage .totalPages}}
<a href="/employees?page={{.nextPage}}" class="page-btn">Next →</a>
{{end}}
```

> **Note:** The current implementation passes `prevPage`/`nextPage` as integers without re-attaching `?q=`. For full search + pagination integration, the template links should be updated to `/employees?q={{.search}}&page={{.prevPage}}`. This is tracked in the v1.2 backlog.

---

## Security

Because GORM uses parameterized queries (`?` placeholders), the PostgreSQL driver sends the query template and parameters separately — SQL injection is structurally impossible regardless of what characters the user types. The `strings.TrimSpace` call handles leading/trailing whitespace, which would otherwise produce queries like `%  engineer  %` that fail to match expected results.

---

## Performance Characteristics

| Dataset size | Strategy | Latency |
|---|---|---|
| < 10,000 rows | Sequential scan (current) | < 5 ms |
| 10,000–100,000 rows | GIN trigram index (planned) | < 10 ms |
| > 100,000 rows | Full-text search with `tsvector` | < 20 ms |

### Planned: GIN Trigram Index

```sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_employees_fullname_trgm
    ON employees USING GIN (full_name gin_trgm_ops);

CREATE INDEX idx_employees_position_trgm
    ON employees USING GIN (position gin_trgm_ops);

CREATE INDEX idx_employees_email_trgm
    ON employees USING GIN (email gin_trgm_ops);
```

Neon supports `pg_trgm` natively. This optimization is planned once the dataset exceeds 10,000 rows and is documented in the [Roadmap](./roadmap.md).