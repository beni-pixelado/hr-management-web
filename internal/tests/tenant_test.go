package tests

import (
	"net/url"
	"testing"

	"hr-management-web/backend/handlers"
)

func TestMultiTenantIsolation(t *testing.T) {
	clientA, orgA := registerAndLogin(t, "tenant_a")
	clientB, orgB := registerAndLogin(t, "tenant_b")

	if orgA.OrganizationID == orgB.OrganizationID {
		t.Fatal("test requires two distinct organizations")
	}

	// org A creates an employee
	resp := post(t, clientA, "/employees", url.Values{
		"full_name": []string{"Tenant A Person"},
		"email":     {"a-only@example.com"},
		"position":  {"Engineer"},
	})
	resp.Body.Close()

	// org B creates an employee
	resp = post(t, clientB, "/employees", url.Values{
		"full_name": []string{"Tenant B Person"},
		"email":     {"b-only@example.com"},
		"position":  {"Engineer"},
	})
	resp.Body.Close()

	var aEmps int64
	var bEmps int64
	testDB.Model(&handlers.Employee{}).Where("organization_id = ?", orgA.OrganizationID).Count(&aEmps)
	testDB.Model(&handlers.Employee{}).Where("organization_id = ?", orgB.OrganizationID).Count(&bEmps)

	if aEmps != 1 {
		t.Errorf("org A should see exactly 1 employee, got %d", aEmps)
	}
	if bEmps != 1 {
		t.Errorf("org B should see exactly 1 employee, got %d", bEmps)
	}

	// A direct cross-tenant read must be empty for B.
	var bSeesA int64
	testDB.Model(&handlers.Employee{}).Where("organization_id = ?", orgB.OrganizationID).Where("email = ?", "a-only@example.com").Count(&bSeesA)
	if bSeesA != 0 {
		t.Error("org B must not see org A's employee")
	}
}

func TestDepartmentIsolationBetweenOrgs(t *testing.T) {
	clientA, orgA := registerAndLogin(t, "dept_iso_a")
	_, orgB := registerAndLogin(t, "dept_iso_b")

	resp := post(t, clientA, "/department", url.Values{"code": []string{"ISO"}, "name": []string{"Isolated"}})
	resp.Body.Close()

	var dept handlers.Department
	testDB.Where("code = ?", "ISO").First(&dept)

	if dept.OrganizationID != orgA.OrganizationID {
		t.Errorf("department should belong to org A")
	}

	// org B cannot see it: direct fetch with B's org must fail.
	var found int64
	testDB.Model(&handlers.Department{}).
		Where("id = ? AND organization_id = ?", dept.ID, orgB.OrganizationID).
		Count(&found)
	if found != 0 {
		t.Error("org B must not see org A's department")
	}
}