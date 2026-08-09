package handlers

import (
	"log"
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
	for _, r := range reports {
		totalAbsences += r.Total
	}

	c.HTML(http.StatusOK, "report.html", gin.H{
		"user":          user,
		"reports":       reports,
		"totalAbsences": totalAbsences,
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

	report := Report{
		UserID: userID,
		Type:   reportType,
		Period: period,
		Days:   days,
		Total:  int(countAbsencesSince(userID, days)),
	}

	if err := DB.Create(&report).Error; err != nil {
		log.Println("Error creating report:", err)
		c.Redirect(http.StatusFound, "/report")
		return
	}

	log.Printf("Report created: type=%s period=%s days=%d", report.Type, report.Period, report.Days)

	c.Redirect(http.StatusFound, "/report")
}
