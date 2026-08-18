package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"hr-management-web/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	fieldName = "_csrf"
	sessionKey = "_csrf"
)

// token is the CSRF token for the request currently being rendered. Template
// functions execute synchronously inside the handler goroutine, so this holder
// is only read during that same request and is safe under the lock.
var (
	mu    sync.RWMutex
	token string
)

func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "0"
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func setToken(t string) {
	mu.Lock()
	token = t
	mu.Unlock()
}

func currentToken() string {
	mu.RLock()
	defer mu.RUnlock()
	return token
}

// Field renders a hidden input carrying the CSRF token. Intended to be used
// inside every <form> via the template func map.
func Field() template.HTML {
	return template.HTML(fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`, fieldName, currentToken()))
}

// Protect validates the CSRF token for state-changing requests and guarantees
// a token exists (and is exposed to templates) for read requests.
func Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := auth.SessionStore.Get(c.Request, "hr_session")
		if err != nil {
			session, _ = auth.SessionStore.New(c.Request, "hr_session")
		}

		tok, _ := session.Values[sessionKey].(string)

		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead {
			if tok == "" {
				tok = newToken()
				session.Values[sessionKey] = tok
				if err := session.Save(c.Request, c.Writer); err != nil {
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}
			}
			setToken(tok)
			c.Next()
			return
		}

		if tok == "" || c.PostForm(fieldName) != tok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		setToken(tok)
		c.Next()
	}
}