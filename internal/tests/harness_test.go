package tests

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"hr-management-web/backend/handlers"
	"hr-management-web/internal/auth"
	"hr-management-web/internal/csrf"
	"hr-management-web/internal/middleware"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	testDB    *gorm.DB
	serverURL string
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Setenv("SESSION_SECRET", "test-secret-that-is-long-enough-32-chars!")

	initDB()
	auth.InitSessionStore()

	srv := httptest.NewServer(newRouter())
	serverURL = srv.URL
	defer srv.Close()

	code := m.Run()
	os.Exit(code)
}

func initDB() {
	var err error
	testDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to open test db:", err)
	}
	sqlDB, _ := testDB.DB()
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := testDB.AutoMigrate(
		&handlers.Organization{},
		&handlers.User{},
		&handlers.Employee{},
		&handlers.Department{},
		&handlers.Report{},
		&handlers.Absence{},
		&handlers.Termination{},
		&handlers.PasswordResetToken{},
	); err != nil {
		log.Fatal("failed to migrate test db:", err)
	}

	handlers.DB = testDB
	middleware.DB = testDB
}

// newRouter builds a gin engine mirroring the production routing for the flows
// under test, including CSRF and the auth middleware chain.
func newRouter() *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.SetFuncMap(template.FuncMap{
		"lower":     strings.ToLower,
		"add":       func(a, b int) int { return a + b },
		"csrfField": csrf.Field,
	})
	r.Use(csrf.Protect())
	r.LoadHTMLGlob("../../backend/templates/*")

	r.GET("/login", func(c *gin.Context) { c.HTML(http.StatusOK, "login.html", nil) })
	r.GET("/register", func(c *gin.Context) { c.HTML(http.StatusOK, "register.html", nil) })
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	protected := r.Group("/")
	protected.Use(middleware.RequireAuth, middleware.LoadUser, middleware.BlockViewerWrites)
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			user, ok := handlers.GetCurrentUser(c)
			if !ok {
				c.Redirect(http.StatusFound, "/login")
				return
			}
			if user.Role == handlers.RoleRecruit {
				c.Redirect(http.StatusFound, "/employees")
				return
			}
			c.HTML(http.StatusOK, "dashboard.html", gin.H{"user": user})
		})
		protected.GET("/employees", handlers.GetEmployees)
		protected.POST("/employees", handlers.CreateEmployee)
		protected.GET("/employees/:id/edit", handlers.EditEmployeePage)
		protected.POST("/employees/:id/edit", handlers.UpdateEmployee)
		protected.POST("/employees/:id/status", handlers.UpdateEmployeeStatus)
		protected.DELETE("/employees/:id", handlers.DeleteEmployee)
		protected.POST("/employees/:id/delete", handlers.DeleteEmployeeForm)

		protected.GET("/department", handlers.DepartmentPageHandler)
		protected.POST("/department", handlers.CreatedepartmentHandler)
		protected.GET("/department/:id", handlers.DepartmentHandler)
		protected.POST("/department/:id/add_employee", handlers.AssignEmployeeToDepartment)
		protected.POST("/department/:id/remove_employee", handlers.DeleteEmployeeFromDepartment)
		protected.POST("/department/:id/delete", handlers.DeleteDepartment)

		protected.GET("/team", handlers.TeamPageHandler)
		protected.POST("/team/invite", handlers.InviteMemberHandler)
		protected.POST("/team/:id/role", handlers.ChangeRoleHandler)
		protected.POST("/team/:id/remove", handlers.RemoveMemberHandler)
		protected.POST("/team/:id/transfer", handlers.TransferOwnershipHandler)

		protected.GET("/logout", handlers.Logout)
	}

	return r
}

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
}

var csrfRe = regexp.MustCompile(`name="_csrf" value="([^"]+)"`)

// csrfToken fetches path and returns the CSRF token embedded in the page.
func csrfToken(t *testing.T, client *http.Client, path string) string {
	t.Helper()
	resp, err := client.Get(serverURL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	m := csrfRe.FindStringSubmatch(buf.String())
	if len(m) != 2 {
		t.Fatalf("no csrf token found in %s", path)
	}
	return m[1]
}

// post sends a form POST (multipart/form-data) with the CSRF token included.
func post(t *testing.T, client *http.Client, path string, form url.Values) *http.Response {
	t.Helper()
	if form == nil {
		form = url.Values{}
	}
	if form.Get("_csrf") == "" {
		form.Set("_csrf", csrfToken(t, client, "/login"))
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, vs := range form {
		for _, v := range vs {
			if err := writer.WriteField(k, v); err != nil {
				t.Fatalf("write field: %v", err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, serverURL+path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

// registerAndLogin creates a brand-new organization + owner and returns an
// authenticated client, along with the created user.
func registerAndLogin(t *testing.T, suffix string) (*http.Client, handlers.User) {
	t.Helper()
	client := newClient()

	email := "owner_" + suffix + "@example.com"
	form := url.Values{
		"username": []string{"owner_" + suffix},
		"password": []string{"Secret123!"},
		"email":    {email},
	}
	resp := post(t, client, "/register", form)
	resp.Body.Close()

	login := url.Values{
		"username": []string{"owner_" + suffix},
		"email":    {email},
		"password": []string{"Secret123!"},
	}
	resp = post(t, client, "/login", login)
	resp.Body.Close()

	var user handlers.User
	if err := testDB.Where("email = ?", email).First(&user).Error; err != nil {
		t.Fatalf("load user: %v", err)
	}
	return client, user
}

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return string(h)
}