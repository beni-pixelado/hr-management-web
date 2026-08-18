package handlers

import (
	"crypto/rand"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TeamPageHandler lists the members of the current organization.
func TeamPageHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanManageTeam() {
		abortUnauthorized(c)
		return
	}

	var members []User
	DB.Where("organization_id = ?", actor.OrganizationID).Order("created_at ASC").Find(&members)

	c.HTML(http.StatusOK, "team.html", gin.H{
		"user":    actor,
		"members": members,
	})
}

func randomPassword() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "Staffio2026!"
	}
	for i := range b {
		b[i] = chars[int(b[i])%len(chars)]
	}
	return string(b)
}

// InviteMemberHandler creates (or updates) a member account for an email with a role,
// and emails them their credentials and role.
func InviteMemberHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanManageTeam() {
		abortUnauthorized(c)
		return
	}

	email := strings.TrimSpace(c.PostForm("email"))
	role := strings.TrimSpace(c.PostForm("role"))
	if email == "" {
		c.HTML(http.StatusBadRequest, "team.html", gin.H{
			"user": actor, "error": "Email is required",
		})
		return
	}

	allowedRoles := map[string]bool{RoleRecruit: true, RoleViewer: true}
	if actor.IsOwner() {
		allowedRoles[RoleAdmin] = true
	}
	if !allowedRoles[role] {
		c.HTML(http.StatusBadRequest, "team.html", gin.H{
			"user": actor, "error": "You cannot assign this role",
		})
		return
	}

	var existing User
	err := DB.Where("email = ?", email).First(&existing).Error

	if err == nil {
		if existing.OrganizationID != actor.OrganizationID {
			c.HTML(http.StatusBadRequest, "team.html", gin.H{
				"user": actor, "error": "This email already belongs to another company",
			})
			return
		}
		if existing.Role == RoleOwner || existing.Role == RoleAdmin {
			c.HTML(http.StatusBadRequest, "team.html", gin.H{
				"user": actor, "error": "This member cannot be reassigned",
			})
			return
		}
		if err := DB.Model(&User{}).Where("id = ?", existing.ID).Update("role", role).Error; err != nil {
			slog.Error("Error updating member role", "error", err)
			c.HTML(http.StatusInternalServerError, "team.html", gin.H{"user": actor, "error": "Error updating member"})
			return
		}
		sendRoleEmail(email, role, existing.Username, "")
		c.Redirect(http.StatusFound, "/team")
		return
	}

	tempPassword := randomPassword()
	hashed, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "team.html", gin.H{"user": actor, "error": "Internal error"})
		return
	}

	newUser := User{
		Username:       email,
		Password:       string(hashed),
		Email:          email,
		OrganizationID: actor.OrganizationID,
		Role:           role,
	}
	if err := DB.Create(&newUser).Error; err != nil {
		slog.Error("Error inviting member", "error", err)
		c.HTML(http.StatusInternalServerError, "team.html", gin.H{"user": actor, "error": "Error creating member"})
		return
	}

	sendRoleEmail(email, role, newUser.Username, tempPassword)
	c.Redirect(http.StatusFound, "/team")
}

func ChangeRoleHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanManageTeam() {
		abortUnauthorized(c)
		return
	}

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/team")
		return
	}
	newRole := strings.TrimSpace(c.PostForm("role"))

	var target User
	if err := DB.Where("id = ? AND organization_id = ?", targetID, actor.OrganizationID).First(&target).Error; err != nil {
		c.Redirect(http.StatusFound, "/team")
		return
	}

	if !actor.canChangeRole(&target, newRole) {
		c.Redirect(http.StatusFound, "/team")
		return
	}

	if err := DB.Model(&User{}).Where("id = ?", target.ID).Update("role", newRole).Error; err != nil {
		slog.Error("Error changing role", "error", err)
	}
	c.Redirect(http.StatusFound, "/team")
}

func RemoveMemberHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanManageTeam() {
		abortUnauthorized(c)
		return
	}

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/team")
		return
	}

	var target User
	if err := DB.Where("id = ? AND organization_id = ?", targetID, actor.OrganizationID).First(&target).Error; err != nil {
		c.Redirect(http.StatusFound, "/team")
		return
	}

	if !actor.canRemoveMember(&target) {
		c.Redirect(http.StatusFound, "/team")
		return
	}

	// Reassign records created by the member to the acting member so the org
	// data is preserved and the FK reference stays valid.
	DB.Model(&Employee{}).Where("user_id = ?", target.ID).Update("user_id", actor.ID)
	DB.Model(&Department{}).Where("user_id = ?", target.ID).Update("user_id", actor.ID)
	DB.Model(&Report{}).Where("user_id = ?", target.ID).Update("user_id", actor.ID)

	if err := DB.Delete(&target).Error; err != nil {
		slog.Error("Error removing member", "error", err)
	}
	c.Redirect(http.StatusFound, "/team")
}

// TransferOwnershipHandler lets the owner hand over their role to another member.
func TransferOwnershipHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.IsOwner() {
		abortUnauthorized(c)
		return
	}

	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/team")
		return
	}

	var target User
	if err := DB.Where("id = ? AND organization_id = ?", targetID, actor.OrganizationID).First(&target).Error; err != nil {
		c.Redirect(http.StatusFound, "/team")
		return
	}
	if target.ID == actor.ID || target.Role == RoleOwner || target.Role == RoleAdmin {
		c.Redirect(http.StatusFound, "/team")
		return
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&User{}).Where("id = ?", target.ID).Update("role", RoleOwner).Error; err != nil {
			return err
		}
		return tx.Model(&User{}).Where("id = ?", actor.ID).Update("role", RoleAdmin).Error
	})
	if err != nil {
		slog.Error("Error transferring ownership", "error", err)
	}
	c.Redirect(http.StatusFound, "/team")
}

func sendRoleEmail(to, role, username, tempPassword string) {
	baseURL := strings.TrimRight(getEnv("SITE_URL", ""), "/")
	roleLabel := strings.Title(role)

	var body string
	if tempPassword != "" {
		body = `<p>Your account for <strong>Staffio</strong> has been created.</p>
			<p>Your role is: <strong>` + roleLabel + `</strong></p>
			<p>Username: <strong>` + username + `</strong></p>
			<p>Temporary password: <strong>` + tempPassword + `</strong></p>
			<p>Use these to <a href="` + baseURL + `/login">log in</a>, then change your password in Configuration.</p>`
	} else {
		body = `<p>Your role in <strong>Staffio</strong> has been updated to: <strong>` + roleLabel + `</strong>.</p>`
	}

	html := `<div style="font-family:Arial,Helvetica,sans-serif;background:#f4f6f9;padding:30px;">
		<div style="max-width:500px;margin:auto;background:#fff;border-radius:12px;padding:30px;">
			<h2 style="color:#2563eb;margin:0;">Staffio</h2>
			<h3 style="color:#222;">Team membership</h3>
			` + body + `
		</div>
	</div>`

	if err := sendEmail(to, "Your Staffio team role", html); err != nil {
		slog.Error("Error sending role email", "error", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
