package middleware

import (
	"net/http"

	"hr-management-web/internal/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var DB *gorm.DB

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

// LoadUser loads the authenticated user's role and organization into context.
func LoadUser(c *gin.Context) {
	userID := c.GetUint("user_id")
	var role string
	var orgID uint
	DB.Raw("SELECT role, organization_id FROM users WHERE id = ?", userID).Row().Scan(&role, &orgID)
	c.Set("user_role", role)
	c.Set("user_org", orgID)
	c.Next()
}

// BlockViewerWrites prevents viewers from performing any mutating request.
func BlockViewerWrites(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		role, _ := c.Get("user_role")
		if role == "viewer" {
			c.Redirect(http.StatusFound, "/dashboard")
			c.Abort()
			return
		}
	}
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
