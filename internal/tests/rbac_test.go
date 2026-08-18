package tests

import (
	"net/http"
	"net/url"
	"testing"

	"hr-management-web/backend/handlers"
)

// createUserInOrg inserts a user directly into the given org with a role and
// returns a logged-in client for them.
func loginAs(t *testing.T, orgID uint, username, email, password, role string) *http.Client {
	t.Helper()
	hashed := hashPassword(t, password)
	u := handlers.User{
		Username:       username,
		Email:          email,
		Password:       hashed,
		OrganizationID: orgID,
		Role:           role,
	}
	if err := testDB.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	client := newClient()
	resp := post(t, client, "/login", url.Values{
		"username": []string{username},
		"email":    {email},
		"password": []string{password},
	})
	resp.Body.Close()
	return client
}

func TestViewerCannotCreateEmployee(t *testing.T) {
	_, owner := registerAndLogin(t, "rbac_viewer")

	viewer := loginAs(t, owner.OrganizationID, "viewer1", "viewer1@example.com", "Secret123!", handlers.RoleViewer)

	resp := post(t, viewer, "/employees", url.Values{
		"full_name": []string{"Should Block"},
		"email":     {"block@example.com"},
		"position":  {"Analyst"},
	})
	defer resp.Body.Close()

	var count int64
	testDB.Model(&handlers.Employee{}).Where("email = ?", "block@example.com").Count(&count)
	if count != 0 {
		t.Error("viewer should not be able to create employees")
	}
}

func TestRecruitRedirectedFromDashboard(t *testing.T) {
	_, owner := registerAndLogin(t, "rbac_recruit")

	recruit := loginAs(t, owner.OrganizationID, "recruit1", "recruit1@example.com", "Secret123!", handlers.RoleRecruit)

	resp := clientGet(t, recruit, "/dashboard")
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound || resp.Header.Get("Location") != "/employees" {
		t.Errorf("expected recruit redirected to /employees, got %d -> %s", resp.StatusCode, resp.Header.Get("Location"))
	}
}

func TestAdminCannotPromoteAdmin(t *testing.T) {
	_, owner := registerAndLogin(t, "rbac_admin")

	admin := loginAs(t, owner.OrganizationID, "admin1", "admin1@example.com", "Secret123!", handlers.RoleAdmin)
	loginAs(t, owner.OrganizationID, "recruit2", "recruit2@example.com", "Secret123!", handlers.RoleRecruit)

	// admin attempts to promote recruit2 to admin
	resp := post(t, admin, "/team/"+itoa(recruitID(t, "recruit2"))+"/role", url.Values{"role": []string{"admin"}})
	resp.Body.Close()

	var updated handlers.User
	testDB.Where("username = ?", "recruit2").First(&updated)
	if updated.Role == "admin" {
		t.Error("admin should not be able to promote another user to admin")
	}
}

func TestOwnerCanPromoteAdmin(t *testing.T) {
	ownerClient, owner := registerAndLogin(t, "rbac_owner_promote")

	loginAs(t, owner.OrganizationID, "recruit3", "recruit3@example.com", "Secret123!", handlers.RoleRecruit)

	resp := post(t, ownerClient, "/team/"+itoa(recruitID(t, "recruit3"))+"/role", url.Values{"role": []string{"admin"}})
	resp.Body.Close()

	var updated handlers.User
	testDB.Where("username = ?", "recruit3").First(&updated)
	if updated.Role != "admin" {
		t.Errorf("owner should be able to promote to admin, got %q", updated.Role)
	}
}

func recruitID(t *testing.T, username string) uint {
	var u handlers.User
	if err := testDB.Where("username = ?", username).First(&u).Error; err != nil {
		t.Fatalf("load recruit: %v", err)
	}
	return u.ID
}