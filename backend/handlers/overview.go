package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OverviewHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "overview.html", gin.H{
		"user": user,
	})
}

func OverviewDataHandlerDepartments(c *gin.Context) {
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

func OverviewDataHandlerEmployees(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	type Result struct {
		Day   string `json:"day"`
		Count int    `json:"count"`
	}

	var results []Result

	result := DB.Table("employees").
		Select(`CASE
			WHEN created_at IS NOT NULL THEN TO_CHAR(created_at, 'YYYY-MM-DD')
			WHEN hire_date IS NOT NULL AND hire_date <> '' THEN SUBSTRING(hire_date, 1, 10)
			ELSE TO_CHAR(CURRENT_DATE, 'YYYY-MM-DD')
		END as day, COUNT(*) as count`).
		Where("user_id = ?", userID).
		Group("day").
		Order("day ASC").
		Scan(&results)

	if result.Error != nil {
		c.JSON(500, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"employees": results,
	})
}
