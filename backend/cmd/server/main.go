package main






import (
	"hr-management-web/backend/handlers"
	"log"
	"net/http"
	"strings"

	"html/template"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)



func connectDatabase() *gorm.DB {

	db, err := gorm.Open(sqlite.Open("/workspaces/hr-management-web/data/users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	return db
}

func main() {

	db := connectDatabase()
	handlers.DB = db



	err := db.AutoMigrate(&handlers.User{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}


	err = db.AutoMigrate(&handlers.Employee{})
	if err != nil {
		log.Fatal("Employee migration failure:", err)
	}

	r := gin.Default() // Cria instância do Gin

	r.SetFuncMap(template.FuncMap{
		"lower": strings.ToLower,
		"add":   func(a, b int) int { return a + b },
	})


	r.LoadHTMLGlob("backend/templates/*")


	r.Static("/css", "frontend/css")
	r.Static("/uploads", "./uploads")




	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})


	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})


	r.GET("/dashboard", func(c *gin.Context) {
	var employees []handlers.Employee
	db.Find(&employees)

		c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"employees": employees,
	})
})
	r.GET("/employees", handlers.GetEmployees)

	r.POST("/register", handlers.Register)

	r.POST("/login", handlers.Login)

	r.POST("/employees", handlers.CreateEmployee)

	r.POST("/employees/:id/status", handlers.UpdateEmployeeStatus)

	
	r.DELETE("/employees/:id", handlers.DeleteEmployee)


	log.Println("Servidor iniciando em http://localhost:8000")
	r.Run(":8000")
}
