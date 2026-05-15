# Search Engine — Multi-field Candidate Search

## Why Search Matters in HR Tools

An HR system accumulates records fast. Even a small company running 20 open positions and reviewing 50 candidates per position will have 1,000 rows in the candidates table within a few months. Paginating through that list to find "John from engineering" is not a workflow — it's friction. The search system eliminates that friction entirely by letting staff locate any candidate in under a second.

## Design Goals

The search was designed with four goals in mind: it should match across multiple fields simultaneously (not just name), it should be case-insensitive without requiring the user to think about casing, it should preserve URL state so results are bookmarkable and shareable, and it should integrate cleanly with pagination so searching "engineer" and getting 30 results lets you page through all 30 without losing the filter.

## Implementation

### The Route

Search is exposed on the existing `GET /employees` route via a `q` query parameter. There is no separate `/search` route — search is simply a filtered view of the employee list. This is both simpler to implement and more correct from a REST perspective.

```
GET /employees?q=engineer&page=2
```

### The Handler Logic

Inside the `ListEmployees` handler, the search logic is a conditional GORM query builder. When `q` is empty, all candidates are returned (paginated). When `q` has a value, a `WHERE` clause is added before the query executes.

```go
func ListEmployees(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        search := strings.TrimSpace(c.Query("q"))
        page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
        limit := 10
        offset := (page - 1) * limit

        query := db.Model(&Employee{})

        if search != "" {
            // Wrap the term in % wildcards for partial matching.
            // ILIKE is PostgreSQL's case-insensitive LIKE.
            term := "%" + search + "%"
            query = query.Where(
                "name ILIKE ? OR position ILIKE ? OR email ILIKE ?",
                term, term, term,
            )
        }

        var total int64
        query.Count(&total)

        var employees []Employee
        query.Offset(offset).Limit(limit).Find(&employees)

        c.HTML(http.StatusOK, "employees.html", gin.H{
            "employees":   employees,
            "search":      search,   // echoed back to the template
            "page":        page,
            "total":       total,
            "totalPages":  int(math.Ceil(float64(total) / float64(limit))),
        })
    }
}
```

Notice that the `Count` and `Find` calls reuse the same `query` value — GORM builds the `WHERE` clause once and applies it to both the count query (for pagination math) and the data query (for the result set). This means the pagination total always reflects the filtered result, not the total table size.

### Multi-field Matching

The `OR` chain in the `WHERE` clause is the key to multi-field search. A single user input matches against three columns at once. The practical effect is that a recruiter who types "dev" will surface anyone named "Devlin", any candidate applying for a "Developer" role, and any candidate with "dev" in their email — all in one query, one round-trip to the database.

The `ILIKE` operator is PostgreSQL-specific (standard SQL uses `LIKE`, which is case-sensitive). Since the application migrated to PostgreSQL, `ILIKE` is available without any workarounds. On the SQLite version, this would have required `LOWER(name) LIKE LOWER(?)` instead.

### Security

Because GORM uses parameterized queries, the `?` placeholders are never string-interpolated with user input. The PostgreSQL driver sends the query template and the parameters separately — the database engine binds them, making SQL injection structurally impossible regardless of what characters the user types. The `strings.TrimSpace` call handles leading/trailing whitespace, which otherwise produces queries like `%  engineer  %` that would fail to match expected results.

## Frontend Integration

The search input sits in the header bar of the employees page. It is a plain HTML form with `method="get"` and `action="/employees"`, so submitting it produces a `GET /employees?q=...` request — no JavaScript required for the basic flow.

```html
<form method="get" action="/employees">
    <input type="text" name="q" value="{{ .search }}" placeholder="Search by name, position or email..." />
    <button type="submit">Search</button>
</form>
```

The `value="{{ .search }}"` binding is the key detail that makes the UX feel correct: after searching, the input field shows what you typed. Without this, the field would clear on every result page — a classic UX failure mode in server-rendered search.

### Pagination Link Preservation

Every pagination link must preserve the active search query. In the template, page links are constructed to always include both `q` and `page` parameters:

```html
{{ if gt .page 1 }}
<a href="/employees?q={{ .search }}&page={{ sub .page 1 }}">Previous</a>
{{ end }}
{{ if lt .page .totalPages }}
<a href="/employees?q={{ .search }}&page={{ add .page 1 }}">Next</a>
{{ end }}
```

If this were omitted, clicking "Next" would drop the `q` parameter and return to the full unfiltered list — breaking the search experience completely.

## Performance Characteristics

At the current scale (hundreds to low thousands of records), `ILIKE` with a wrapping `%term%` pattern performs a sequential scan. For PostgreSQL, this is fast on small tables — a few milliseconds. Once the dataset grows into the tens of thousands, the `pg_trgm` GIN index described in the database documentation should be added. That index allows PostgreSQL to use trigram matching to skip most rows without a full scan, keeping search latency sub-10ms even at large scale.