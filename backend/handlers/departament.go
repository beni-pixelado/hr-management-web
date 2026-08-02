package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Department struct {
	ID uint `gorm:"primaryKey" json:"id"`

	UserID uint `gorm:"not null;index" json:"user_id"`
	User   User `gorm:"constraint:OnDelete:CASCADE;"`

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

	var dept Department

	if err := DB.
		Where("id = ? AND user_id = ?", idParam, GetCurrentUserID(c)).
		First(&dept).Error; err != nil {

		c.String(http.StatusNotFound, "Department not found")
		return
	}

	var allEmployees []Employee
	if err := DB.
		Where("user_id = ?", GetCurrentUserID(c)).
		Find(&allEmployees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching employees"})
		return
	}

	var members []Employee
	_ = DB.Where("department_id = ? AND user_id = ?", dept.ID, GetCurrentUserID(c)).Find(&members).Error

	c.HTML(http.StatusOK, "department.html", gin.H{
		"Department": dept,
		"Employees":  allEmployees,
		"Members":    members,
	})
}

func AssignEmployeeToDepartment(c *gin.Context) {
	deptID := c.Param("id")
	empID := c.PostForm("employee_id")

	if empID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "employee_id is required"})
		return
	}

	if err := DB.
		Model(&Employee{}).
		Where(
			"id = ? AND user_id = ?",
			empID,
			GetCurrentUserID(c),
		).
		Update("department_id", deptID).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error assigning employee to department"})
		return
	}

	c.Redirect(http.StatusFound, "/department/"+deptID)
}

func DeleteDepartment(c *gin.Context) {
	deptID := c.Param("id")

	_ = DB.
		Model(&Employee{}).
		Where("department_id = ? AND user_id = ?", deptID, GetCurrentUserID(c)).
		Update("department_id", 0).Error

	if err := DB.
		Where(
			"id = ? AND user_id = ?",
			deptID,
			GetCurrentUserID(c),
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
		UserID: GetCurrentUserID(c),
		Name:   name,
		Code:   code,
		BossID: bossID,
	}

	log.Printf("Creating department: %+v", department)

	if err := DB.Create(&department).Error; err != nil {
		log.Println("Error creating department:", err)

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

	userID := GetCurrentUserID(c)

	if err := DB.
		Where("user_id = ?", userID).
		Find(&employees).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching employees",
		})
		return
	}

	if err := DB.
		Where("user_id = ?", userID).
		Find(&departments).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching departments",
		})
		return
	}

	c.HTML(http.StatusOK, "departments.html", gin.H{
		"Employees":   employees,
		"Departments": departments,
	})
}

func DeleteEmployeeFromDepartment(c *gin.Context) {
	deptID := c.Param("id")
	empID := c.PostForm("employee_id")

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
			"id = ? AND user_id = ?",
			empID,
			GetCurrentUserID(c),
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
