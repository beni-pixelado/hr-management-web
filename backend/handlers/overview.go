package handlers

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

func OverviewHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if actor.IsRecruit() {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	c.HTML(http.StatusOK, "overview.html", gin.H{
		"user": actor,
	})
}

func OverviewDataHandlerDepartments(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	orgID := actor.OrganizationID

	type Result struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	var results []Result

	DB.Table("departments").
		Select("departments.name, COUNT(employees.id) as count").
		Joins("LEFT JOIN employees ON employees.department_id = departments.id AND employees.organization_id = ?", orgID).
		Where("departments.organization_id = ?", orgID).
		Group("departments.id, departments.name").
		Scan(&results)

	c.JSON(200, gin.H{
		"departments": results,
	})
}

func OverviewDataHandlerEmployees(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}
	orgID := actor.OrganizationID

	type Result struct {
		Day   string `json:"day"`
		Count int    `json:"count"`
	}

	// New candidates, bucketed by the day they were added.
	var created []Result
	createdResult := DB.Table("employees").
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as count").
		Where("organization_id = ?", orgID).
		Group("day").
		Order("day ASC").
		Scan(&created)
	if createdResult.Error != nil {
		c.JSON(500, gin.H{"error": createdResult.Error.Error()})
		return
	}

	// Fired (rejected) candidates, bucketed by the day they were rejected.
	var rejected []Result
	rejectedResult := DB.Table("employees").
		Select("TO_CHAR(rejected_at, 'YYYY-MM-DD') as day, COUNT(*) as count").
		Where("organization_id = ? AND rejected_at IS NOT NULL", orgID).
		Group("day").
		Order("day ASC").
		Scan(&rejected)
	if rejectedResult.Error != nil {
		c.JSON(500, gin.H{"error": rejectedResult.Error.Error()})
		return
	}

	// Deleted/rejected records logged as terminations (keeps deletions visible).
	var terminated []Result
	termResult := DB.Table("terminations").
		Select("TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as count").
		Where("organization_id = ?", orgID).
		Group("day").
		Order("day ASC").
		Scan(&terminated)
	if termResult.Error != nil {
		c.JSON(500, gin.H{"error": termResult.Error.Error()})
		return
	}

	// Merge both sources into a single fired series by day.
	byDay := map[string]int{}
	for _, r := range rejected {
		byDay[r.Day] += r.Count
	}
	for _, r := range terminated {
		byDay[r.Day] += r.Count
	}

	type DayResult struct {
		Day   string `json:"day"`
		Count int    `json:"count"`
	}
	fired := make([]DayResult, 0, len(byDay))
	for day, count := range byDay {
		fired = append(fired, DayResult{Day: day, Count: count})
	}
	sort.Slice(fired, func(i, j int) bool { return fired[i].Day < fired[j].Day })

	c.JSON(200, gin.H{
		"employees": created,
		"fired":     fired,
	})
}
