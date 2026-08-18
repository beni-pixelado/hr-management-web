package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Department struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID         uint `gorm:"not null;index" json:"user_id"`
	OrganizationID uint `gorm:"index;default:0" json:"organization_id"`

	Code   string `json:"code" gorm:"not null"`
	Name   string `json:"name" gorm:"not null"`
	BossID uint   `json:"boss_id"`
}

func DepartmentsHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "departments.html", nil)
}

func DepartmentHandler(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.Redirect(http.StatusFound, "/department")
		return
	}

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if actor.IsRecruit() {
		c.Redirect(http.StatusFound, "/employees")
		return
	}
	orgID := actor.OrganizationID

	var dept Department

	if err := DB.
		Where("id = ? AND organization_id = ?", idParam, orgID).
		First(&dept).Error; err != nil {

		c.String(http.StatusNotFound, "Department not found")
		return
	}

	var allEmployees []Employee
	if err := DB.
		Where("organization_id = ?", orgID).
		Find(&allEmployees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching employees"})
		return
	}

	var members []Employee
	_ = DB.Where("department_id = ? AND organization_id = ?", dept.ID, orgID).Find(&members).Error

	c.HTML(http.StatusOK, "department.html", gin.H{
		"Department": dept,
		"Employees":  allEmployees,
		"Members":    members,
		"user":       actor,
		"canEdit":    actor.CanManageTeam() || actor.IsOwner(),
	})
}

func AssignEmployeeToDepartment(c *gin.Context) {
	deptID := c.Param("id")
	empID := c.PostForm("employee_id")

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !actor.CanAssignDepartment() {
		abortUnauthorized(c)
		return
	}

	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required"})
		return
	}

	if err := DB.
		Model(&Employee{}).
		Where(
			"id = ? AND organization_id = ?",
			empID,
			actor.OrganizationID,
		).
		Update("department_id", deptID).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error assigning employee to department"})
		return
	}

	c.Redirect(http.StatusFound, "/department/"+deptID)
}

func DeleteDepartment(c *gin.Context) {
	deptID := c.Param("id")

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !actor.CanManageTeam() {
		abortUnauthorized(c)
		return
	}

	_ = DB.
		Model(&Employee{}).
		Where("department_id = ? AND organization_id = ?", deptID, actor.OrganizationID).
		Update("department_id", 0).Error

	if err := DB.
		Where(
			"id = ? AND organization_id = ?",
			deptID,
			actor.OrganizationID,
		).
		Delete(&Department{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting department"})
		return
	}

	c.Redirect(http.StatusFound, "/department")
}

func CreatedepartmentHandler(c *gin.Context) {
	code := c.PostForm("code")
	name := c.PostForm("name")
	bossIDStr := c.PostForm("boss_id")

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !actor.CanManageTeam() {
		abortUnauthorized(c)
		return
	}

	if code == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Required fields: code and name",
		})
		return
	}

	var bossID uint
	if bossIDStr != "" {
		parsed, err := strconv.ParseUint(bossIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid manager ID",
			})
			return
		}
		bossID = uint(parsed)
	}

	department := Department{
		UserID:         actor.ID,
		OrganizationID: actor.OrganizationID,
		Name:           name,
		Code:           code,
		BossID:         bossID,
	}

	slog.Info("Creating department", "department", department)

	if err := DB.Create(&department).Error; err != nil {
		slog.Error("Error creating department", "error", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Error saving department",
			"detail": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/department")
}

func DepartmentPageHandler(c *gin.Context) {
	var employees []Employee
	var departments []Department

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if actor.IsRecruit() {
		c.Redirect(http.StatusFound, "/employees")
		return
	}
	orgID := actor.OrganizationID

	if err := DB.
		Where("organization_id = ?", orgID).
		Find(&employees).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching employees",
		})
		return
	}

	if err := DB.
		Where("organization_id = ?", orgID).
		Find(&departments).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching departments",
		})
		return
	}

	c.HTML(http.StatusOK, "departments.html", gin.H{
		"Employees":   employees,
		"Departments": departments,
		"user":        actor,
		"canEdit":     actor.CanManageTeam(),
	})
}

func DeleteEmployeeFromDepartment(c *gin.Context) {
	deptID := c.Param("id")
	empID := c.PostForm("employee_id")

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !actor.CanAssignDepartment() {
		abortUnauthorized(c)
		return
	}

	if empID == "" {
		empID = c.Param("employee_id")
	}

	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required"})
		return
	}

	if err := DB.
		Model(&Employee{}).
		Where(
			"id = ? AND organization_id = ?",
			empID,
			actor.OrganizationID,
		).
		Update("department_id", 0).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error removing employee from department",
		})
		return
	}

	if deptID == "" {
		c.Redirect(http.StatusFound, "/department")
		return
	}

	c.Redirect(http.StatusFound, "/department/"+deptID)
}
