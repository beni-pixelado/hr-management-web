package handlers

import (
	"log"
	"math"
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

func DepartmentHandler(c *gin.Context) {
	var department []Department

	pageStr := c.DefaultQuery("page", "1")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 20
	offset := (page - 1) * limit

	var totalDepartments int64
	DB.Model(&Department{}).Count(&totalDepartments)

	DB.
		Limit(limit).
		Offset(offset).
		Find(&department)

	totalPages := int(math.Ceil(float64(totalDepartments) / float64(limit)))

	c.HTML(http.StatusOK, "departments.html", gin.H{
		"departments":      department,
		"currentPage":      page,
		"totalPages":       totalPages,
		"totalDepartments": totalDepartments,
		"prevPage":         page - 1,
		"nextPage":         page + 1,
	})
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
