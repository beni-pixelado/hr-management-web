package integration

import (
	"encoding/csv"
	"hr-management-web/backend/handlers"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)


func OpenTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("falha ao abrir banco de dados de teste: %v", err)
	}

	if err := db.AutoMigrate(&handlers.User{}, &handlers.Employee{}); err != nil {
		t.Fatalf("falha na migração de teste: %v", err)
	}

	return db
}


func OpenMainDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := "/workspaces/hr-management-web/data/users.db"
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("falha ao abrir banco principal: %v", err)
	}

	if err := db.AutoMigrate(&handlers.User{}, &handlers.Employee{}); err != nil {
		t.Fatalf("falha na migração do banco principal: %v", err)
	}

	return db
}


func AbsTemplateGlob(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("falha ao obter informações do arquivo")
	}

	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	path := filepath.Join(projectRoot, "backend", "templates", "*")
	return path
}


func LoadRegistrationRecords(t *testing.T) [][]string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("falha ao obter informações do arquivo")
	}

	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	path := filepath.Join(projectRoot, "tests", "loads", "registration-load.csv")

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("falha ao abrir arquivo de carga: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("falha ao ler CSV de carga: %v", err)
	}

	if len(records) <= 1 {
		t.Fatalf("arquivo de carga deve conter pelo menos 1 registro; recebeu %d", len(records)-1)
	}

	return records[1:]
}


func PostForm(t *testing.T, router http.Handler, path string, values url.Values) *httptest.ResponseRecorder {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
	if err != nil {
		t.Fatalf("falha ao criar requisição: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}


func SetupTestRouter(db *gorm.DB, templatesGlob string) *gin.Engine {
	handlers.DB = db

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"lower": strings.ToLower,
		"add":   func(a, b int) int { return a + b },
	})

	if templatesGlob != "" {
		r.LoadHTMLGlob(templatesGlob)
	}

	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)

	return r
}
