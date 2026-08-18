package tests

import (
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"hr-management-web/backend/handlers"
)

func itoa(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

func TestCreateEmployeeScopedToOrg(t *testing.T) {
	client, user := registerAndLogin(t, "emp_create")

	form := url.Values{
		"full_name": []string{"Ada Lovelace"},
		"email":     {"ada@example.com"},
		"position":  {"Engineer"},
	}
	resp := post(t, client, "/employees", form)
	resp.Body.Close()

	var emp handlers.Employee
	if err := testDB.Where("full_name = ?", "Ada Lovelace").First(&emp).Error; err != nil {
		t.Fatalf("employee not created: %v", err)
	}
	if emp.OrganizationID != user.OrganizationID {
		t.Errorf("expected org %d, got %d", user.OrganizationID, emp.OrganizationID)
	}
	if emp.Status != "pending" {
		t.Errorf("expected pending status, got %q", emp.Status)
	}
}

func TestCreateEmployeeRequiresFields(t *testing.T) {
	client, _ := registerAndLogin(t, "emp_validate")

	form := url.Values{
		"email":    {"nofields@example.com"},
		"position": []string{"Engineer"},
	}
	resp := post(t, client, "/employees", form)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing fields, got %d", resp.StatusCode)
	}
}

func TestUpdateEmployeeStatusPipeline(t *testing.T) {
	client, _ := registerAndLogin(t, "emp_status")

	// create
	resp := post(t, client, "/employees", url.Values{
		"full_name": []string{"Grace Hopper"},
		"email":     {"grace@example.com"},
		"position":  {"Developer"},
	})
	resp.Body.Close()

	var emp handlers.Employee
	testDB.Where("email = ?", "grace@example.com").First(&emp)

	// pending -> contractors (hired)
	resp = post(t, client, "/employees/"+itoa(emp.ID)+"/status", url.Values{
		"status":    []string{"contractors"},
		"hire_date": []string{"2026-01-10"},
	})
	resp.Body.Close()

	testDB.First(&emp, emp.ID)
	if emp.Status != "contractors" {
		t.Errorf("expected contractors, got %q", emp.Status)
	}

	// contractors -> rejected
	resp = post(t, client, "/employees/"+itoa(emp.ID)+"/status", url.Values{
		"status": []string{"rejected"},
	})
	resp.Body.Close()

	testDB.First(&emp, emp.ID)
	if emp.Status != "rejected" {
		t.Errorf("expected rejected, got %q", emp.Status)
	}

	var terminations int64
	testDB.Model(&handlers.Termination{}).Where("employee_id = ?", emp.ID).Count(&terminations)
	if terminations == 0 {
		t.Error("expected a termination record on rejection")
	}
}

func TestDeleteEmployee(t *testing.T) {
	client, _ := registerAndLogin(t, "emp_del")

	resp := post(t, client, "/employees", url.Values{
		"full_name": []string{"Alan Turing"},
		"email":     {"alan@example.com"},
		"position":  {"Scientist"},
	})
	resp.Body.Close()

	var emp handlers.Employee
	testDB.Where("email = ?", "alan@example.com").First(&emp)

	resp = post(t, client, "/employees/"+itoa(emp.ID)+"/delete", url.Values{})
	resp.Body.Close()

	var count int64
	testDB.Model(&handlers.Employee{}).Where("id = ?", emp.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected employee to be deleted, still %d rows", count)
	}
}