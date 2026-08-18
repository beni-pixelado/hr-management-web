package handlers

import (
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Report struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UserID         uint      `gorm:"not null;index"`
	OrganizationID uint      `gorm:"index;default:0"`
	Type           string    `json:"type" gorm:"not null"`
	Period         string    `json:"period" gorm:"not null"`
	Days           int       `json:"days" gorm:"not null"`
	Total          int       `json:"total" gorm:"not null;default:0"`
}

func periodDays(period string) int {
	switch period {
	case "month":
		return 30
	case "year":
		return 365
	default:
		return 7
	}
}

func countAbsencesSince(orgID uint, days int) int64 {
	since := time.Now().AddDate(0, 0, -days)

	var total int64
	DB.Model(&Absence{}).
		Where("organization_id = ? AND created_at >= ?", orgID, since).
		Count(&total)

	return total
}

func countHiredSince(orgID uint, days int) int64 {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	var total int64
	DB.Model(&Employee{}).
		Where("organization_id = ? AND status = ? AND hire_date <> '' AND hire_date >= ?", orgID, "contractors", since).
		Count(&total)

	return total
}

func countFiredSince(orgID uint, days int) int64 {
	since := time.Now().AddDate(0, 0, -days)

	var rejected int64
	DB.Model(&Employee{}).
		Where("organization_id = ? AND rejected_at IS NOT NULL AND rejected_at >= ?", orgID, since).
		Count(&rejected)

	var terminated int64
	DB.Model(&Termination{}).
		Where("organization_id = ? AND created_at >= ?", orgID, since).
		Count(&terminated)

	return rejected + terminated
}

func ReportAbsencesContent(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil || days < 1 {
		days = 7
	}

	c.JSON(http.StatusOK, gin.H{
		"days":  days,
		"total": countAbsencesSince(actor.OrganizationID, days),
	})
}

func ReportHiredContent(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil || days < 1 {
		days = 7
	}

	c.JSON(http.StatusOK, gin.H{
		"days":  days,
		"total": countHiredSince(actor.OrganizationID, days),
	})
}

func ReportFiredContent(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil || days < 1 {
		days = 7
	}

	c.JSON(http.StatusOK, gin.H{
		"days":  days,
		"total": countFiredSince(actor.OrganizationID, days),
	})
}

func ReportHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if actor.IsRecruit() {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	var reports []Report
	DB.Where("organization_id = ?", actor.OrganizationID).Order("created_at DESC").Find(&reports)

	totalAbsences := 0
	totalHired := 0
	totalFired := 0
	for _, r := range reports {
		if r.Type == "hired" {
			totalHired += r.Total
		} else if r.Type == "fired" {
			totalFired += r.Total
		} else {
			totalAbsences += r.Total
		}
	}

	c.HTML(http.StatusOK, "report.html", gin.H{
		"user":          actor,
		"reports":       reports,
		"totalAbsences": totalAbsences,
		"totalHired":    totalHired,
		"totalFired":    totalFired,
	})
}

func ReportNewHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if actor.IsRecruit() {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	c.HTML(http.StatusOK, "report-new.html", gin.H{
		"user": actor,
	})
}

func CreateReportHandler(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if actor.IsRecruit() {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	reportType := c.PostForm("report-type")
	if reportType == "" {
		reportType = "absences"
	}

	period := c.PostForm("time-report")
	if period == "" {
		period = "week"
	}

	days := periodDays(period)

	var total int64
	switch reportType {
	case "hired":
		total = countHiredSince(actor.OrganizationID, days)
	case "fired":
		total = countFiredSince(actor.OrganizationID, days)
	default:
		total = countAbsencesSince(actor.OrganizationID, days)
	}

	report := Report{
		UserID:         actor.ID,
		OrganizationID: actor.OrganizationID,
		Type:           reportType,
		Period:         period,
		Days:           days,
		Total:          int(total),
	}

	if err := DB.Create(&report).Error; err != nil {
		slog.Error("Error creating report", "error", err)
		c.Redirect(http.StatusFound, "/report")
		return
	}

	slog.Info("Report created", "type", report.Type, "period", report.Period, "days", report.Days)

	c.Redirect(http.StatusFound, "/report")
}

type reportStat struct {
	Label string
	Value string
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func verbWas(n int) string {
	if n == 1 {
		return "was"
	}
	return "were"
}

func verbIs(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}

func verbHas(n int) string {
	if n == 1 {
		return "has"
	}
	return "have"
}

func countByStatus(orgID uint, status string) int {
	var n int64
	DB.Model(&Employee{}).Where("organization_id = ? AND status = ?", orgID, status).Count(&n)
	return int(n)
}

func countAllAbsences(orgID uint) int {
	var n int64
	DB.Model(&Absence{}).Where("organization_id = ?", orgID).Count(&n)
	return int(n)
}

func countEmployeesWithAbsences(orgID uint) int {
	var n int64
	DB.Model(&Absence{}).Where("organization_id = ?", orgID).Distinct("employee_id").Count(&n)
	return int(n)
}

func ReportDetailHandler(c *gin.Context) {
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

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/report")
		return
	}

	var report Report
	if err := DB.Where("id = ? AND organization_id = ?", id, orgID).First(&report).Error; err != nil {
		c.Redirect(http.StatusFound, "/report")
		return
	}

	startDate := time.Now().AddDate(0, 0, -report.Days)
	endDate := time.Now()
	startLabel := startDate.Format("Jan 02, 2006")
	endLabel := endDate.Format("Jan 02, 2006")

	typeLabel := report.Type
	summary := ""
	var stats []reportStat

	switch report.Type {
	case "hired":
		typeLabel = "Hired"

		totalHired := countByStatus(orgID, "contractors")
		pending := countByStatus(orgID, "pending")
		rejected := countByStatus(orgID, "rejected")
		avg := 0
		if report.Days > 0 {
			avg = int(math.Round(float64(report.Total) / float64(report.Days)))
		}

		summary = fmt.Sprintf(
			"Between %s and %s, a span of %d days, %d candidate%s %s hired, bringing your team to %d hired employee%s in total. In the rest of the pipeline, %d candidate%s %s still being interviewed and %d %s been rejected.",
			startLabel, endLabel, report.Days,
			report.Total, plural(report.Total), verbWas(report.Total),
			totalHired, plural(totalHired),
			pending, plural(pending), verbIs(pending),
			rejected, verbHas(rejected),
		)

		stats = []reportStat{
			{"Hired in this period", strconv.Itoa(report.Total)},
			{"Average hires per day", strconv.Itoa(avg)},
			{"Total hired employees", strconv.Itoa(totalHired)},
			{"Candidates in interview", strconv.Itoa(pending)},
			{"Rejected candidates", strconv.Itoa(rejected)},
		}

	case "fired":
		typeLabel = "Fired"

		totalRejected := countByStatus(orgID, "rejected")
		totalHired := countByStatus(orgID, "contractors")
		avg := 0
		if report.Days > 0 {
			avg = int(math.Round(float64(report.Total) / float64(report.Days)))
		}

		summary = fmt.Sprintf(
			"Between %s and %s, a span of %d days, %d employee%s %s fired or rejected, taking the team from %d active hire%s to %d current hire%s in the pipeline. A total of %d candidate%s currently remain in the rejected bucket.",
			startLabel, endLabel, report.Days,
			report.Total, plural(report.Total), verbWas(report.Total),
			totalHired, plural(totalHired),
			totalHired, plural(totalHired),
			totalRejected, plural(totalRejected),
		)

		stats = []reportStat{
			{"Fired in this period", strconv.Itoa(report.Total)},
			{"Average per day", strconv.Itoa(avg)},
			{"Current rejected candidates", strconv.Itoa(totalRejected)},
			{"Total hired employees", strconv.Itoa(totalHired)},
		}

	default:
		typeLabel = "Absences"

		allTime := countAllAbsences(orgID)
		employeesAbsent := countEmployeesWithAbsences(orgID)
		avg := 0
		if report.Days > 0 {
			avg = int(math.Round(float64(report.Total) / float64(report.Days)))
		}

		summary = fmt.Sprintf(
			"Between %s and %s, a span of %d days, your team recorded %d absence%s — an average of %d per day. Since tracking began, %d absence%s in total %s been recorded, affecting %d employee%s.",
			startLabel, endLabel, report.Days,
			report.Total, plural(report.Total), avg,
			allTime, plural(allTime), verbHas(allTime),
			employeesAbsent, plural(employeesAbsent),
		)

		stats = []reportStat{
			{"Total absences in period", strconv.Itoa(report.Total)},
			{"Average per day", strconv.Itoa(avg)},
			{"All-time absences", strconv.Itoa(allTime)},
			{"Employees with absences", strconv.Itoa(employeesAbsent)},
		}
	}

	c.HTML(http.StatusOK, "report-detail.html", gin.H{
		"user":        actor,
		"report":      report,
		"typeLabel":   typeLabel,
		"periodLabel": "Last " + strconv.Itoa(report.Days) + " days",
		"startDate":   startLabel,
		"endDate":     endLabel,
		"summary":     summary,
		"stats":       stats,
	})
}
