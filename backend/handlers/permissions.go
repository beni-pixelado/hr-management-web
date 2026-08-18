package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func roleRank(r string) int {
	switch r {
	case RoleOwner:
		return 3
	case RoleAdmin:
		return 2
	case RoleRecruit:
		return 1
	default:
		return 0
	}
}

// GetCurrentUser loads the authenticated user (with role + organization).
func GetCurrentUser(c *gin.Context) (*User, bool) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		return nil, false
	}
	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, false
	}
	return &user, true
}

func (u *User) IsOwner() bool  { return u.Role == RoleOwner }
func (u *User) IsAdmin() bool  { return u.Role == RoleAdmin }
func (u *User) IsRecruit() bool { return u.Role == RoleRecruit }

// CanManageTeam reports whether the user may invite/remove/assign team members.
func (u *User) CanManageTeam() bool {
	return u.Role == RoleOwner || u.Role == RoleAdmin
}

// CanEditEmployees reports whether the user may create/edit/delete candidates.
func (u *User) CanEditEmployees() bool {
	return u.Role == RoleOwner || u.Role == RoleAdmin || u.Role == RoleRecruit
}

// CanViewEmployees is true for every member (viewers may only read).
func (u *User) CanViewEmployees() bool { return true }

// CanAssignDepartment reports whether the user may assign candidates to departments.
func (u *User) CanAssignDepartment() bool {
	return u.Role == RoleOwner || u.Role == RoleAdmin || u.Role == RoleRecruit
}

// canRemoveMember reports whether actor may remove (fire) target from the org.
func (u *User) canRemoveMember(target *User) bool {
	if target.ID == u.ID {
		return false // cannot fire yourself
	}
	if target.Role == RoleOwner || target.Role == RoleAdmin {
		return false // owner and admin are untouchable
	}
	// owner and admin may remove recruit/viewer members
	return u.CanManageTeam()
}

// canChangeRole reports whether actor may assign the given role to target.
func (u *User) canChangeRole(target *User, newRole string) bool {
	if target.ID == u.ID {
		return false // cannot change your own role here (use transfer)
	}
	if newRole == RoleOwner {
		return false // ownership only changes via transfer
	}
	if target.Role == RoleAdmin {
		return false // only owner may act on admins, and never demote them here
	}
	if target.Role == RoleOwner {
		return false
	}

	switch u.Role {
	case RoleOwner:
		// owner may promote to admin, or set recruit/viewer
		return newRole == RoleAdmin || newRole == RoleRecruit || newRole == RoleViewer
	case RoleAdmin:
		// admin may only reassign recruit/viewer, never create admins
		return newRole == RoleRecruit || newRole == RoleViewer
	default:
		return false
	}
}

func abortUnauthorized(c *gin.Context) {
	c.HTML(http.StatusForbidden, "login.html", gin.H{"error": "You don't have permission to do that"})
	c.Abort()
}

// employeeLinksToSuperior reports whether an employee's email belongs to a
// member of the same org with a role higher than the acting user.
func employeeLinksToSuperior(employeeID string, orgID uint, actor *User) (bool, error) {
	var emp Employee
	if err := DB.Where("id = ? AND organization_id = ?", employeeID, orgID).First(&emp).Error; err != nil {
		return false, err
	}
	if emp.Email == "" {
		return false, nil
	}
	var member User
	if err := DB.Where("organization_id = ? AND email = ?", orgID, emp.Email).First(&member).Error; err != nil {
		return false, nil
	}
	return roleRank(member.Role) > roleRank(actor.Role), nil
}
