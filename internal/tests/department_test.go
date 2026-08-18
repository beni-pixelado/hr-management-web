package tests

import (
	"net/url"
	"testing"

	"hr-management-web/backend/handlers"
)

func TestDepartmentCreateAndAssignMember(t *testing.T) {
	client, user := registerAndLogin(t, "dept_owner")

	// create a department
	resp := post(t, client, "/department", url.Values{
		"code": []string{"ENG"},
		"name": []string{"Engineering"},
	})
	resp.Body.Close()

	var dept handlers.Department
	if err := testDB.Where("name = ?", "Engineering").First(&dept).Error; err != nil {
		t.Fatalf("department not created: %v", err)
	}
	if dept.OrganizationID != user.OrganizationID {
		t.Errorf("expected org %d, got %d", user.OrganizationID, dept.OrganizationID)
	}

	// create an employee
	resp = post(t, client, "/employees", url.Values{
		"full_name": []string{"Nikola Tesla"},
		"email":     {"nikola@example.com"},
		"position":  {"Engineer"},
	})
	resp.Body.Close()

	var emp handlers.Employee
	testDB.Where("email = ?", "nikola@example.com").First(&emp)

	// assign employee to department
	resp = post(t, client, "/department/"+itoa(dept.ID)+"/add_employee", url.Values{
		"employee_id": []string{itoa(emp.ID)},
	})
	resp.Body.Close()

	testDB.First(&emp, emp.ID)
	if emp.DepartmentID != dept.ID {
		t.Errorf("expected employee in dept %d, got %d", dept.ID, emp.DepartmentID)
	}

	// remove employee from department
	resp = post(t, client, "/department/"+itoa(dept.ID)+"/remove_employee", url.Values{
		"employee_id": []string{itoa(emp.ID)},
	})
	resp.Body.Close()

	testDB.First(&emp, emp.ID)
	if emp.DepartmentID != 0 {
		t.Errorf("expected employee removed from dept, got %d", emp.DepartmentID)
	}
}

func TestDeleteDepartmentClearsMembers(t *testing.T) {
	client, _ := registerAndLogin(t, "dept_del")

	resp := post(t, client, "/department", url.Values{"code": []string{"QA"}, "name": []string{"Quality"}})
	resp.Body.Close()

	var dept handlers.Department
	testDB.Where("code = ?", "QA").First(&dept)

	resp = post(t, client, "/employees", url.Values{
		"full_name": []string{"Marie Curie"},
		"email":     {"marie@example.com"},
		"position":  {"Analyst"},
	})
	resp.Body.Close()

	var emp handlers.Employee
	testDB.Where("email = ?", "marie@example.com").First(&emp)

	testDB.Model(&handlers.Employee{}).Where("id = ?", emp.ID).Update("department_id", dept.ID)

	resp = post(t, client, "/department/"+itoa(dept.ID)+"/delete", url.Values{})
	resp.Body.Close()

	var deptCount int64
	testDB.Model(&handlers.Department{}).Where("id = ?", dept.ID).Count(&deptCount)
	if deptCount != 0 {
		t.Errorf("expected department deleted, still %d", deptCount)
	}

	testDB.First(&emp, emp.ID)
	if emp.DepartmentID != 0 {
		t.Errorf("expected member's department_id cleared, got %d", emp.DepartmentID)
	}
}