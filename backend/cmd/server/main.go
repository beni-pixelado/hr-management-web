package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"hr-management-web/backend/database"
	"hr-management-web/backend/handlers"
	"hr-management-web/internal/auth"
	"hr-management-web/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Erro ao carregar .env")
	}

	database.Connect()
	auth.InitSessionStore()

	handlers.DB = database.DB
	db := database.DB

	// migrations
	if err := db.AutoMigrate(&handlers.User{}); err != nil {
		log.Fatal("User migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Employee{}); err != nil {
		log.Fatal("Employee migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Department{}); err != nil {
		log.Fatal("Department migration failed:", err)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.SetFuncMap(template.FuncMap{
		"lower": strings.ToLower,
		"add":   func(a, b int) int { return a + b },
	})

	r.LoadHTMLGlob("backend/templates/*")
	r.Static("/css", "frontend/css")
	r.Static("/uploads", "./uploads")

	// auth routes
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	// debug route
	r.GET("/debug/cookie", func(c *gin.Context) {
		cookie, err := c.Cookie("hr_session")
		if err != nil {
			c.String(http.StatusOK, "COOKIE NÃO ENCONTRADO: %v", err)
			return
		}

		authenticated, userID := auth.IsAuthenticated(c)
		c.String(
			http.StatusOK,
			fmt.Sprintf("Cookie: %s, Autenticado: %v, UserID: %d", cookie, authenticated, userID),
		)
	})

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth)
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			search := strings.TrimSpace(c.DefaultQuery("search", ""))

			userID := handlers.GetCurrentUserID(c)
			if userID == 0 {
				c.Redirect(http.StatusFound, "/login")
				return
			}

			query := database.DB.Where("user_id = ?", userID)

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
				Where("user_id = ?", userID).
				Count(&totalEmployees)

			query.Find(&employees)

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"employees":      employees,
				"search":         search,
				"showAll":        showAll,
				"showOff":        showOff,
				"totalEmployees": totalEmployees,
			})

		})

		protected.GET("/badge/:id", handlers.BadgeHandler)
		protected.GET("/employees", handlers.GetEmployees)
		protected.POST("/employees", handlers.CreateEmployee)
		protected.POST("/employees/:id/status", handlers.UpdateEmployeeStatus)
		protected.DELETE("/employees/:id", handlers.DeleteEmployee)
		protected.POST("/employees/:id/delete", handlers.DeleteEmployeeForm)

		protected.GET("/department", handlers.DepartmentPageHandler)
		protected.POST("/department", handlers.CreatedepartmentHandler)
		protected.GET("/department/:id", handlers.DepartmentHandler)
		protected.POST("/department/:id/add_employee", handlers.AssignEmployeeToDepartment)
		protected.POST("/department/:id/remove_employee", handlers.DeleteEmployeeFromDepartment)
		protected.POST("/department/:id/delete", handlers.DeleteDepartment)

		protected.GET("/overview", handlers.OverviewHandler)
		protected.GET("/api/overview", handlers.OverviewDataHandler)

		protected.GET("/logout", handlers.Logout)
	}

	r.Run(":8000")
}