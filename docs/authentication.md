# Authentication

Staffio uses **cookie-based sessions** via `gorilla/sessions`, backed by `gorilla/securecookie` for HMAC-signing. Passwords are hashed with **bcrypt**. The auth surface covers registration, login, logout, and password recovery/reset.

---

## Session Flow

When a user logs in, the server creates a session, stores `user_id` and an `authenticated` flag, and writes a signed `hr_session` cookie. Every subsequent request validates that cookie.

```go
// internal/auth/session.go
func InitSessionStore() {
    secret := os.Getenv("SESSION_SECRET")
    SessionStore = sessions.NewCookieStore([]byte(secret))
    SessionStore.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   60 * 60 * 24 * 7,   // 7 days
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    }
}

func CreateSession(c *gin.Context, userID uint) error {
    session, _ := SessionStore.Get(c.Request, "hr_session")
    session.Values["user_id"]       = int(userID)
    session.Values["authenticated"] = true
    // ...saves the signed cookie
}
```

Cookie properties:

| Property | Value | Rationale |
|---|---|---|
| `MaxAge` | 7 days | Persistent session |
| `HttpOnly` | `true` | JS cannot read the cookie (mitigates XSS) |
| `Secure` | `true` | Sent only over HTTPS |
| `SameSite` | `Lax` | CSRF defense-in-depth |

The cookie payload is HMAC-signed with `SESSION_SECRET`. Tampering is detected and the session is rejected.

---

## Middleware

`internal/middleware/auth.go` provides:

- **`RequireAuth`** — protects every route in the `protected` group. If `auth.IsAuthenticated` fails, it redirects (302) to `/login` and aborts; otherwise it stores `user_id` in the context.
- **`RedirectIfAuthenticated`** — intended for `/login` and `/register`; sends already-logged-in users to `/dashboard`.
- **`LoadUser`** — after auth, loads the user's `role` and `organization_id` into context (used by RBAC and tenant scoping).
- **`BlockViewerWrites`** — blocks `viewer`-role users from any non-GET request.

---

## Registration

`POST /register` (`handlers/auth.go`):

1. Validates `username`, `email`, `password` are present.
2. Rejects an existing username.
3. Hashes the password with `bcrypt.GenerateFromPassword` (`DefaultCost`).
4. In a transaction, creates an **Organization** and the user as its **owner**.
5. Redirects to login.

The transaction guarantees an org is never created without its owner, and vice versa.

---

## Login

`POST /login`:

1. Finds the user by `username AND email`.
2. Verifies with `bcrypt.CompareHashAndPassword`.
3. On success calls `auth.CreateSession` and redirects to `/dashboard`.

Both wrong-credential and missing-user paths render the same message — *"Incorrect username, email or password"* — to prevent account enumeration.

---

## Logout

`GET /logout` calls `auth.DestroySession` (cookie `MaxAge = -1`) and redirects to `/login`.

---

## Password Recovery & Reset

### Recovery (`POST /recuperateaccount`)

1. Looks up the user by email.
2. Generates a **one-time token** (`uuid.NewString`) with a 1-hour expiry and stores it in `PasswordResetToken`.
3. Emails a reset link `SITE_URL/reset-password?token=...` via SMTP.
4. Whether or not the email exists, it responds with *"If this email exists, a recovery link was sent"* — no user enumeration.

### Reset (`GET/POST /reset-password`)

- The page validates the token exists and has not expired.
- `POST` requires `new_password == confirm_password` and a minimum length of **6 characters**, re-hashes, saves, and **deletes the reset record** so the token is single-use.

### SMTP

`sendEmail` uses `net/smtp` with `PlainAuth` and reads `SMTP_HOST`, `SMTP_PORT`, `SMTP_EMAIL`, `SMTP_PASSWORD`, `SMTP_FROM`, `SMTP_FROM_NAME` from the environment. The same helper is reused for team-invite e-mails.

---

## Security Notes

- Passwords are stored only as bcrypt hashes.
- Session cookies are HMAC-signed; secrets come from `SESSION_SECRET`.
- Recovery tokens are single-use and time-limited.
- Auth endpoints are additionally rate-limited (see [Security](./security.md)).

---

## Related

- [Security](./security.md)
- [Roles & Permissions](./roles-permissions.md)
- [Getting Started](./getting-started.md)