package handlers

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Report struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UserID    uint      `gorm:"not null;index"`
	Type      string    `json:"type" gorm:"not null"`
	Period    string    `json:"period" gorm:"not null"`
	Days      int       `json:"days" gorm:"not null"`
	Total     int       `json:"total" gorm:"not null;default:0"`
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

func countAbsencesSince(userID uint, days int) int64 {
	since := time.Now().AddDate(0, 0, -days)

	var total int64
	DB.Model(&Absence{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&total)

	return total
}

func countHiredSince(userID uint, days int) int64 {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	var total int64
	DB.Model(&Employee{}).
		Where("user_id = ? AND status = ? AND hire_date <> '' AND hire_date >= ?", userID, "contractors", since).
		Count(&total)

	return total
}

func ReportAbsencesContent(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil || days < 1 {
		days = 7
	}

	c.JSON(http.StatusOK, gin.H{
		"days":  days,
		"total": countAbsencesSince(userID, days),
	})
}

func ReportHiredContent(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))
	if err != nil || days < 1 {
		days = 7
	}

	c.JSON(http.StatusOK, gin.H{
		"days":  days,
		"total": countHiredSince(userID, days),
	})
}

func ReportHandler(c *gin.Context) {
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

	var reports []Report
	DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&reports)

	totalAbsences := 0
	totalHired := 0
	for _, r := range reports {
		if r.Type == "hired" {
			totalHired += r.Total
		} else {
			totalAbsences += r.Total
		}
	}

	c.HTML(http.StatusOK, "report.html", gin.H{
		"user":          user,
		"reports":       reports,
		"totalAbsences": totalAbsences,
		"totalHired":    totalHired,
	})
}

func ReportNewHandler(c *gin.Context) {
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

	c.HTML(http.StatusOK, "report-new.html", gin.H{
		"user": user,
	})
}

func CreateReportHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
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
		total = countHiredSince(userID, days)
	default:
		total = countAbsencesSince(userID, days)
	}

	report := Report{
		UserID: userID,
		Type:   reportType,
		Period: period,
		Days:   days,
		Total:  int(total),
	}

	if err := DB.Create(&report).Error; err != nil {
		log.Println("Error creating report:", err)
		c.Redirect(http.StatusFound, "/report")
		return
	}

	log.Printf("Report created: type=%s period=%s days=%d", report.Type, report.Period, report.Days)

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

func countByStatus(userID uint, status string) int {
	var n int64
	DB.Model(&Employee{}).Where("user_id = ? AND status = ?", userID, status).Count(&n)
	return int(n)
}

func countAllAbsences(userID uint) int {
	var n int64
	DB.Model(&Absence{}).Where("user_id = ?", userID).Count(&n)
	return int(n)
}

func countEmployeesWithAbsences(userID uint) int {
	var n int64
	DB.Model(&Absence{}).Where("user_id = ?", userID).Distinct("employee_id").Count(&n)
	return int(n)
}

func ReportDetailHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/report")
		return
	}

	var report Report
	if err := DB.Where("id = ? AND user_id = ?", id, userID).First(&report).Error; err != nil {
		c.Redirect(http.StatusFound, "/report")
		return
	}

	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
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

		totalHired := countByStatus(userID, "contractors")
		pending := countByStatus(userID, "pending")
		rejected := countByStatus(userID, "rejected")
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

	default:
		typeLabel = "Absences"

		allTime := countAllAbsences(userID)
		employeesAbsent := countEmployeesWithAbsences(userID)
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
		"user":        user,
		"report":      report,
		"typeLabel":   typeLabel,
		"periodLabel": "Last " + strconv.Itoa(report.Days) + " days",
		"startDate":   startLabel,
		"endDate":     endLabel,
		"summary":     summary,
		"stats":       stats,
	})
}
