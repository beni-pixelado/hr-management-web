package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"os"

	"hr-management-web/backend/handlers"
	"hr-management-web/internal/auth"
	"hr-management-web/internal/middleware"
	"hr-management-web/backend/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}

	DB = db

	log.Println("Banco conectado!")
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Erro ao carregar .env")
	}

	database.Connect()

	auth.InitSessionStore()

	handlers.DB = database.DB
db := database.DB

	if err := db.AutoMigrate(&handlers.User{}); err != nil {
		log.Fatal("Migration failed:", err)
	}
	if err := db.AutoMigrate(&handlers.Employee{}); err != nil {
		log.Fatal("Employee migration failure:", err)
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

	r.GET("/debug/cookie", func(c *gin.Context) {
		cookie, err := c.Cookie("hr_session")
		if err != nil {
			c.String(http.StatusOK, "COOKIE NÃO ENCONTRADO: %v", err)
			return
		}

		authenticated, userID := auth.IsAuthenticated(c)

		c.String(http.StatusOK,
			"Cookie encontrado!\n"+
				"Valor: %s\n"+
				"Autenticado: %v\n"+
				"UserID: %d\n"+
				"Headers recebidos: %v",
			cookie,
			authenticated,
			userID,
			c.Request.Header,
		)
	})


	protected := r.Group("/")
	protected.Use(middleware.RequireAuth)
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			search := strings.TrimSpace(c.DefaultQuery("search", ""))
			var employees []handlers.Employee
			var query = db
			log.Printf("the search carried out was '%s'", search)
			if search != "" {
				query = query.Where(
					"full_name LIKE ? OR email LIKE ?",
					"%"+search+"%",
					"%"+search+"%",
				)
			}
			showAll := c.Query("all") == "true"
			showOff := c.Query("all") == "false"

			if  showOff {
				query = query.Limit(1000)
			}

			if !showAll {
				query = query.Limit(20)
			}
			

			var totalEmployees int64
			db.Model(&handlers.Employee{}).Count(&totalEmployees)

			query.Find(&employees)

			log.Printf("results: %d funcionários encontrados", len(employees))

			c.HTML(http.StatusOK, "dashboard.html", gin.H{
				"employees": employees,
				"search":    search,
				"showAll": showAll,
				"showOff": showOff,
				"totalEmployees": totalEmployees,
				"debug_msg": fmt.Sprintf(
					"Busca por '%s' retornou %d resultados",
					search,
					len(employees),
				),
			})
		
	
		r.GET("/badge/:id", handlers.BadgeHandler)  
		c.HTML(http.StatusOK, "id-card.html", nil)
	
		

		protected.GET("/employees", handlers.GetEmployees)
		protected.POST("/employees", handlers.CreateEmployee)
		protected.POST("/employees/:id/status", handlers.UpdateEmployeeStatus)
		protected.DELETE("/employees/:id", handlers.DeleteEmployee)
		protected.GET("/logout", handlers.Logout)
		})

	r.Run(":8000")
}}


