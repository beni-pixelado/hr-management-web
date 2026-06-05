package handlers

import (
	
	"github.com/gin-gonic/gin"
)

func OverviewHandler(c *gin.Context) {
	c.HTML(200, "overview.html", nil)
}

func OverviewDataHandler(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(401, gin.H{"error": "unauthorized"})
        return
    }

    type Result struct {
        Name  string `json:"name"`
        Count int    `json:"count"`
    }

    var results []Result

    // Conta employees agrupados por departamento, filtrando pelo usuário
    DB.Table("departments").
        Select("departments.name, COUNT(employees.id) as count").
        Joins("LEFT JOIN employees ON employees.department_id = departments.id").
        Where("departments.user_id = ?", userID).
        Group("departments.id, departments.name").
        Scan(&results)

    c.JSON(200, gin.H{
        "departments": results,
    })
}