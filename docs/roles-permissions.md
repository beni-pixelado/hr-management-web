# Roles & Permissions (RBAC)

Staffio implements **role-based access control** (RBAC) combined with **multi-tenant organization isolation**. Every user belongs to exactly one organization and holds one of four roles.

---

## Roles

| Role | Constant | Rank | Summary |
|---|---|---|---|
| **Owner** | `owner` | 3 | Full control, including team management and ownership transfer |
| **Admin** | `admin` | 2 | Team management, editing, reports — cannot transfer ownership |
| **Recruit** | `recruit` | 1 | Hands-on recruiting: create/edit candidates and departments |
| **Viewer** | `viewer` | 0 | Read-only access |

---

## Permission Matrix

Implemented in `handlers/permissions.go` and enforced via the middleware chain in `main.go`.

| Permission | Owner | Admin | Recruit | Viewer |
|---|---|---|---|---|
| View candidates / dashboard | ✅ | ✅ | ✅ | ✅ |
| Create / edit / delete candidates | ✅ | ✅ | ✅ | ❌ |
| Assign candidates to departments | ✅ | ✅ | ✅ | ❌ |
| View reports & analytics | ✅ | ✅ | ❌ | ❌ |
| Invite / remove / change team roles | ✅ | ✅ | ❌ | ❌ |
| Transfer ownership | ✅ | ❌ | ❌ | ❌ |
| Any mutating request | ✅ | ✅ | ✅ | ❌ |

Predicates:

```go
func (u *User) CanManageTeam() bool       { return u.Role == owner || u.Role == admin }
func (u *User) CanEditEmployees() bool    { return u.Role == owner || u.Role == admin || u.Role == recruit }
func (u *User) CanAssignDepartment() bool { return u.Role == owner || u.Role == admin || u.Role == recruit }
func (u *User) CanViewEmployees() bool    { return true }
```

### Middleware enforcement

The `protected` group runs `RequireAuth → LoadUser → BlockViewerWrites`:

- `LoadUser` puts `role` and `organization_id` into context.
- `BlockViewerWrites` redirects any non-GET request from a `viewer` back to `/dashboard`.

### Handler-level guards

Handlers also call predicates explicitly (e.g. `actor.CanEditEmployees()`) and return `403` via `abortUnauthorized` when denied. This double enforcement (middleware + handler) is defense-in-depth.

---

## Object-Level Rules

Beyond role rank, the code enforces nuanced rules:

- **`employeeLinksToSuperior`** — a `recruit` cannot reject or delete a candidate whose email belongs to a member with a higher role than them.
- **`canChangeRole`** — you cannot change your own role; only the owner may promote to `admin`; admins may only reassign `recruit`/`viewer`; ownership changes only via transfer.
- **`canRemoveMember`** — you cannot remove yourself, and `owner`/`admin` members are untouchable.
- **Ownership transfer** is transactional: the target becomes `owner` and the previous owner becomes `admin`.

---

## Multi-Tenant Isolation

- Registration creates a fresh `Organization` and makes the registrant its `owner`.
- Every org-owned model stores `OrganizationID`.
- **All** queries are scoped to `actor.OrganizationID` — a member can only ever see their own company's data.
- Team management (invites, role changes, removal, transfer) operates strictly within the actor's org.

---

## Team Lifecycle

- **Invite** (`POST /team/invite`) — creates a member with a temporary password (or updates an existing member's role) and emails credentials.
- **Change role** (`POST /team/:id/role`) — reassigns a role within the matrix.
- **Remove** (`POST /team/:id/remove`) — removes a member and **reassigns their records to the actor** so data stays valid.
- **Transfer** (`POST /team/:id/transfer`) — hands the `owner` role to another member.

---

## Related

- [Security](./security.md)
- [Backend](./backend.md)
- [Database](./database.md)