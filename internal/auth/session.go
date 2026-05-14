package auth

import (
	"net/http"
	"os"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var SessionStore *sessions.CookieStore

func InitSessionStore() {
	secret := os.Getenv("SESSION_SECRET")

	if secret == "" {
		panic("SESSION_SECRET não definido")
	}

	SessionStore = sessions.NewCookieStore([]byte(secret))

	SessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7, // 7 dias
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func CreateSession(c *gin.Context, userID uint) error {
	session, err := SessionStore.Get(c.Request, "hr_session")
	if err != nil {
		return err
	}

	session.Values["user_id"] = int(userID)
	session.Values["authenticated"] = true

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	return session.Save(c.Request, c.Writer)
}

func IsAuthenticated(c *gin.Context) (bool, uint) {
	session, err := SessionStore.Get(c.Request, "hr_session")
	if err != nil {
		return false, 0
	}

	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		return false, 0
	}

	userIDValue, ok := session.Values["user_id"].(int)
if !ok {
	return false, 0
}

userID := uint(userIDValue)
	if !ok {
		return false, 0
	}

	fmt.Printf("AUTH VALUE: %#v\n", session.Values["authenticated"])
fmt.Printf("AUTH TYPE: %T\n", session.Values["authenticated"])

	return true, userID

}

func DestroySession(c *gin.Context) error {
	session, err := SessionStore.Get(c.Request, "hr_session")
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1

	return session.Save(c.Request, c.Writer)
}