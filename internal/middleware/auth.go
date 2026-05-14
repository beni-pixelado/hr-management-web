
package middleware

import (
	"net/http"

	"hr-management-web/internal/auth"

	"github.com/gin-gonic/gin"
)



func RequireAuth(c *gin.Context) {
	authenticated, userID := auth.IsAuthenticated(c)

	if !authenticated {
		c.Redirect(http.StatusFound, "/login")
		c.Abort()
		return
	}

	
	c.Set("user_id", userID)

	c.Next()
}



func RedirectIfAuthenticated(c *gin.Context) {
	authenticated, _ := auth.IsAuthenticated(c)

	if authenticated {
		c.Redirect(http.StatusFound, "/dashboard")
		c.Abort()
		return
	}

	c.Next()
}