package integration

import (
	"fmt"
	"hr-management-web/backend/handlers"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestRegisterLoadFromCSV(t *testing.T) {
	db := OpenMainDatabase(t)
	router := SetupTestRouter(db, AbsTemplateGlob(t))

	records := LoadRegistrationRecords(t)
	now := time.Now().UnixNano()
	var createdUsernames []string

	for index, record := range records {
		index := index
		record := record
		if len(record) != 3 {
			t.Fatalf("linha %d do CSV deve ter 3 colunas, recebeu %d", index+2, len(record))
		}

		t.Run("register_user_"+record[0], func(t *testing.T) {
			values := url.Values{}

			username := fmt.Sprintf("%s_%d", record[0], now)
			email := fmt.Sprintf("%s_%d@%s", record[1], now, strings.Split(record[1], "@")[1])
			values.Set("username", username)
			values.Set("email", email)
			values.Set("password", record[2])

			rr := PostForm(t, router, "/register", values)

			if rr.Code != 200 {
				t.Fatalf("esperado status 200 para cadastro; recebeu %d, corpo=%q", rr.Code, rr.Body.String())
			}

			if !strings.Contains(rr.Body.String(), "Conta criada com sucesso") {
				t.Fatalf("mensagem de sucesso não encontrada em resposta: %s", rr.Body.String())
			}

			var user handlers.User
			result := db.Where("username = ?", username).First(&user)
			if result.Error != nil {
				t.Fatalf("usuário %q não foi encontrado no banco de dados: %v", username, result.Error)
			}

			if user.Email != email {
				t.Fatalf("email do usuário %q não corresponde. Esperado: %q, recebido: %q", username, email, user.Email)
			}

			if user.Username != username {
				t.Fatalf("username não corresponde. Esperado: %q, recebido: %q", username, user.Username)
			}

			createdUsernames = append(createdUsernames, username)
		})
	}


	for _, u := range createdUsernames {
		var dbUser handlers.User
		if err := db.Where("username = ?", u).First(&dbUser).Error; err != nil {
			t.Fatalf("usuário %q não encontrado no banco principal: %v", u, err)
		}
	}

	t.Logf("%d usuários confirmados no banco principal", len(createdUsernames))
}

func TestUserPersistenceInDatabase(t *testing.T) {
	db := OpenMainDatabase(t)
	router := SetupTestRouter(db, AbsTemplateGlob(t))


	for i := 1; i <= 60; i++ {
		username := fmt.Sprintf("itest%03d", i)
		email := fmt.Sprintf("%s@example.test", username)
		password := "password123"


		var existing handlers.User
		if err := db.Where("username = ?", username).First(&existing).Error; err == nil {
			continue
		}

		values := url.Values{}
		values.Set("username", username)
		values.Set("email", email)
		values.Set("password", password)

		rr := PostForm(t, router, "/register", values)
		if rr.Code != 200 {
			t.Fatalf("falha no cadastro de %s: status %d", username, rr.Code)
		}
	}


	for _, sample := range []string{"itest001", "itest030", "itest060"} {
		var u handlers.User
		if err := db.Where("username = ?", sample).First(&u).Error; err != nil {
			t.Fatalf("usuário amostra %s não encontrado: %v", sample, err)
		}
	}

	t.Logf("60 usuários (ou os que faltavam) foram criados no banco principal com sucesso")
}
