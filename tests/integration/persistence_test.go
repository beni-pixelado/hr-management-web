package integration

import (
	"fmt"
	"hr-management-web/backend/handlers"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)


func OpenPersistentDatabase(t *testing.T, filename string) (*gorm.DB, func()) {
	t.Helper()


	tmpDir, err := ioutil.TempDir("", "hr-test-*")
	if err != nil {
		t.Fatalf("falha ao criar diretório temporário: %v", err)
	}

	dbPath := filepath.Join(tmpDir, filename)


	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("falha ao abrir banco SQLite em arquivo: %v", err)
	}


	if err := db.AutoMigrate(&handlers.User{}, &handlers.Employee{}); err != nil {
		t.Fatalf("falha na migração: %v", err)
	}


	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestUserCreation(t *testing.T) {

	db := OpenMainDatabase(t)
	router := SetupTestRouter(db, AbsTemplateGlob(t))


	now := time.Now().UnixNano()

	t.Logf("Fase 1: Criando 60 usuários únicos (sufixo %d)", now)
	created := make([]string, 0, 60)
	for i := 1; i <= 60; i++ {
		username := fmt.Sprintf("perm%d_%03d", now, i)
		email := fmt.Sprintf("%s@example.com", username)
		password := "password123"

		values := url.Values{}
		values.Set("username", username)
		values.Set("email", email)
		values.Set("password", password)

		rr := PostForm(t, router, "/register", values)
		if rr.Code != 200 {
			t.Fatalf("falha no cadastro de %s: status %d, corpo=%s", username, rr.Code, rr.Body.String())
		}

		created = append(created, username)
	}

	t.Logf("Fase 1 concluida: %d usuários criados", len(created))


	for _, username := range created {
		var u handlers.User
		if err := db.Where("username = ?", username).First(&u).Error; err != nil {
			t.Fatalf("usuário %s não encontrado após criação: %v", username, err)
		}
	}

	t.Logf("Fase 2 concluida: Todos os %d usuários recém-criados confirmados no banco SQLite", len(created))
}
