package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Department struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Code   string `json:"code" gorm:"not null"`
	Name   string `json:"Name" gorm:"not null"`
	BossID uint   `gorm:"column:boss" json:"boss_id"`
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
	if err := DB.First(&dept, idParam).Error; err != nil {
		c.String(http.StatusNotFound, "Departamento não encontrado")
		return
	}

	var allEmployees []Employee
	if err := DB.Find(&allEmployees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar funcionários"})
		return
	}

	var members []Employee
	if err := DB.Where("department_id = ?", dept.ID).Find(&members).Error; err != nil {

		members = []Employee{}
	}

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

	if err := DB.Model(&Employee{}).Where("id = ?", empID).Update("department_id", deptID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atribuir funcionário ao departamento"})
		return
	}

	c.Redirect(http.StatusFound, "/department/"+deptID)
}

func DeleteDepartment(c *gin.Context) {
	deptID := c.Param("id")

	if err := DB.Model(&Employee{}).Where("department_id = ?", deptID).Update("department_id", 0).Error; err != nil {

	}

	if err := DB.Delete(&Department{}, deptID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar departamento"})
		return
	}

	c.Redirect(http.StatusFound, "/department")
}

func CreatedepartmentHandler(c *gin.Context) {
	Code := c.PostForm("code")
	Name := c.PostForm("name")
	BossIDStr := c.PostForm("boss_id")

	if Code == "" || Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Campos obrigatórios: code e name",
		})
		return
	}

	var bossID uint
	if BossIDStr != "" {
		bossIDUintParsed, err := strconv.ParseUint(BossIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "ID do chefe inválido",
			})
			return
		}
		bossID = uint(bossIDUintParsed)
	}

	department := Department{
		Name:   Name,
		Code:   Code,
		BossID: bossID,
	}

	log.Printf("Tentando criar departamento: %+v", department)

	if err := DB.Create(&department).Error; err != nil {
		log.Println("Erro ao criar departamento:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Erro ao salvar departamento no banco de dados",
			"detail": err.Error(),
		})

		return
	}

	log.Printf("A new department has been added: %s", Name)

	c.Redirect(http.StatusFound, "/department")
}

func DepartmentPageHandler(c *gin.Context) {
	var employees []Employee
	var departments []Department

	if err := DB.Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar funcionários",
		})
		return
	}

	if err := DB.Find(&departments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar departamentos",
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

	if err := DB.Model(&Employee{}).Where("id = ?", empID).Update("department_id", 0).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover funcionário do departamento"})
		return
	}

	if deptID == "" {
		c.Redirect(http.StatusFound, "/department")
		return
	}

	c.Redirect(http.StatusFound, "/department/"+deptID)
}
