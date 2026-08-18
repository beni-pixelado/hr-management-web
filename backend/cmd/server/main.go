package main

import (
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"hr-management-web/backend/database"
	"hr-management-web/backend/handlers"
	"hr-management-web/internal/auth"
	"hr-management-web/internal/csrf"
	"hr-management-web/internal/middleware"
	"hr-management-web/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := godotenv.Load(); err != nil {
		slog.Info("Error loading .env")
	}

	database.Connect()
	storage.Init()
	auth.InitSessionStore()

	handlers.DB = database.DB
	db := database.DB

	if err := db.AutoMigrate(&handlers.Organization{}); err != nil {
		log.Fatal("Organization migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.User{}); err != nil {
		log.Fatal("User migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Employee{}); err != nil {
		log.Fatal("Employee migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Department{}); err != nil {
		log.Fatal("Department migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.PasswordResetToken{}); err != nil {
		log.Fatal("PasswordResetToken migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Report{}); err != nil {
		log.Fatal("Report migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Absence{}); err != nil {
		log.Fatal("Absence migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Termination{}); err != nil {
		log.Fatal("Termination migration failed:", err)
	}
	if db.Migrator().HasColumn(&handlers.PasswordResetToken{}, "code") {
		if err := db.Migrator().DropColumn(&handlers.PasswordResetToken{}, "code"); err != nil {
			log.Fatal("Failed to drop legacy password_reset_tokens.code column:", err)
		}
		log.Println("Dropped legacy password_reset_tokens.code column")
	}

	// Backfill: assign every legacy user without an organization to their own org.
	var orphanUsers []handlers.User
	if err := db.Where("organization_id = 0").Find(&orphanUsers).Error; err != nil {
		log.Fatal("Failed to query orphan users:", err)
	}
	for i := range orphanUsers {
		org := handlers.Organization{Name: orphanUsers[i].Email}
		if err := db.Create(&org).Error; err != nil {
			log.Fatal("Failed to create backfill org:", err)
		}
		if err := db.Model(&handlers.User{}).Where("id = ?", orphanUsers[i].ID).Updates(map[string]interface{}{
			"organization_id": org.ID,
			"role":            handlers.RoleOwner,
		}).Error; err != nil {
			log.Fatal("Failed to assign backfill org:", err)
		}
		// Migrate that user's data to the new org.
		db.Model(&handlers.Employee{}).Where("user_id = ?", orphanUsers[i].ID).Update("organization_id", org.ID)
		db.Model(&handlers.Department{}).Where("user_id = ?", orphanUsers[i].ID).Update("organization_id", org.ID)
		db.Model(&handlers.Report{}).Where("user_id = ?", orphanUsers[i].ID).Update("organization_id", org.ID)
		db.Model(&handlers.Absence{}).Where("user_id = ?", orphanUsers[i].ID).Update("organization_id", org.ID)
	}
	if len(orphanUsers) > 0 {
		slog.Info("Backfilled legacy users into organizations", "count", len(orphanUsers))
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.SetFuncMap(template.FuncMap{
		"lower":     strings.ToLower,
		"add":       func(a, b int) int { return a + b },
		"csrfField": csrf.Field,
	})

	r.Use(csrf.Protect())

	r.LoadHTMLGlob("backend/templates/*")
	r.Static("/css", "frontend/css")
	r.Static("/js", "frontend/public/js")
	r.Static("/static", "frontend/static")

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.POST("/register", middleware.RateLimit(10, time.Minute), handlers.Register)
	r.POST("/login", middleware.RateLimit(10, time.Minute), handlers.Login)

	r.GET("/recuperateaccount", func(c *gin.Context) {
		c.HTML(http.StatusOK, "recover.html", nil)
	})
	r.POST("/recuperateaccount", middleware.RateLimit(5, time.Minute), handlers.RecoverAccount)
	r.GET("/reset-password", handlers.ResetPasswordPage)
	r.POST("/reset-password", middleware.RateLimit(5, time.Minute), handlers.ResetPassword)

	r.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "landing.html", nil) })

	r.GET("/robots.txt", func(c *gin.Context) {
		c.File("robots.txt")
	})

	r.NoRoute(func(c *gin.Context) { c.HTML(http.StatusOK, "404.html", nil) })

	middleware.DB = database.DB

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth, middleware.LoadUser, middleware.BlockViewerWrites)
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			search := strings.TrimSpace(c.DefaultQuery("search", ""))

			var user handlers.User
			if err := database.DB.Where("id = ?", handlers.GetCurrentUserID(c)).First(&user).Error; err != nil {
				c.Redirect(http.StatusFound, "/login")
				return
			}
			if user.Role == handlers.RoleRecruit {
				c.Redirect(http.StatusFound, "/employees")
				return
			}

			query := database.DB.Where("organization_id = ?", user.OrganizationID)

			var employees []handlers.Employee

			if search != "" {
				query = query.Where(
					"full_name LIKE ? OR email LIKE ?",
					"%"+search+"%",
					"%"+search+"%",
				)
			}

			showAll := c.Query("all") == "true"
			showOff := c.Query("all") == "false"

			if showOff {
				query = query.Limit(100000)
			}
			if !showAll {
				query = query.Limit(20)
			}

			var totalEmployees int64
			database.DB.
				Model(&handlers.Employee{}).
				Where("organization_id = ?", user.OrganizationID).
				Count(&totalEmployees)

			query.Find(&employees)

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"employees":      employees,
				"search":         search,
				"showAll":        showAll,
				"showOff":        showOff,
				"totalEmployees": totalEmployees,
				"user":           user,
				"canEdit":        user.CanEditEmployees(),
			})

		})

		protected.GET("/badge/:id", handlers.BadgeHandler)
		protected.GET("/employees", handlers.GetEmployees)
		protected.GET("/employees/download", handlers.DownloadEmployeesCSV)
		protected.POST("/employees", handlers.CreateEmployee)
		protected.GET("/employees/:id/edit", handlers.EditEmployeePage)
		protected.POST("/employees/:id/edit", handlers.UpdateEmployee)
		protected.POST("/employees/:id/status", handlers.UpdateEmployeeStatus)
		protected.POST("/employees/:id/absence", handlers.MarkAbsence)
		protected.DELETE("/employees/:id", handlers.DeleteEmployee)
		protected.POST("/employees/:id/delete", handlers.DeleteEmployeeForm)

		protected.GET("/department", handlers.DepartmentPageHandler)
		protected.POST("/department", handlers.CreatedepartmentHandler)
		protected.GET("/department/:id", handlers.DepartmentHandler)
		protected.POST("/department/:id/add_employee", handlers.AssignEmployeeToDepartment)
		protected.POST("/department/:id/remove_employee", handlers.DeleteEmployeeFromDepartment)
		protected.POST("/department/:id/delete", handlers.DeleteDepartment)

		protected.GET("/overview", handlers.OverviewHandler)
		protected.GET("/report", handlers.ReportHandler)
		protected.GET("/report/new", handlers.ReportNewHandler)
		protected.POST("/report/new", handlers.CreateReportHandler)
		protected.GET("/report/:id", handlers.ReportDetailHandler)
		protected.GET("/api/overview/departments", handlers.OverviewDataHandlerDepartments)
		protected.GET("/api/overview/employees", handlers.OverviewDataHandlerEmployees)
		protected.GET("/api/report/absences", handlers.ReportAbsencesContent)
		protected.GET("/api/report/hired", handlers.ReportHiredContent)
		protected.GET("/api/report/fired", handlers.ReportFiredContent)

		protected.GET("/config", handlers.ConfigPageHandler)
		protected.GET("/config/account", handlers.AccountPageHandler)
		protected.GET("/config/device", handlers.DevicePageHandler)
		protected.POST("/config/account/profile", handlers.UpdateProfileHandler)
		protected.POST("/config/account/photo", handlers.UpdateProfilePhotoHandler)
		protected.POST("/config/account/password", handlers.ChangePasswordHandler)
		protected.POST("/config/account/delete", handlers.DeleteAccountHandler)

		protected.GET("/team", handlers.TeamPageHandler)
		protected.POST("/team/invite", handlers.InviteMemberHandler)
		protected.POST("/team/:id/role", handlers.ChangeRoleHandler)
		protected.POST("/team/:id/remove", handlers.RemoveMemberHandler)
		protected.POST("/team/:id/transfer", handlers.TransferOwnershipHandler)

		protected.GET("/recover", handlers.RecoverAccountPageHandler)

		protected.GET("/logout", handlers.Logout)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	slog.Info("Server running", "port", port)
	r.Run(":" + port)
}
