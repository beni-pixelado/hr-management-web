
package handlers

import (
	"log"
	"net/http"

	"hr-management-web/internal/auth"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"password" gorm:"not null"`
	Email    string `json:"email" gorm:"not null"`
}

func Register(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	email := c.PostForm("email")

	if username == "" || password == "" || email == "" {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "Todos os campos são obrigatórios"})
		return
	}

	var existingUser User
	if err := DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "Usuário já existe"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Erro ao hash senha:", err)
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"error": "Erro interno"})
		return
	}

	newUser := User{Username: username, Password: string(hashedPassword), Email: email}
	if err := DB.Create(&newUser).Error; err != nil {
		log.Println("Erro ao criar usuário:", err)
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"error": "Erro ao criar conta"})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{"success": "Conta criada com sucesso! Faça login."})
}

func Logout(c *gin.Context) {
	if err := auth.DestroySession(c); err != nil {
		log.Println("Erro ao destruir sessão:", err)
	}

	c.Redirect(http.StatusFound, "/login")
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if username == "" || email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "Todos os campos são obrigatórios"})
		return
	}

	var user User
	if err := DB.Where("username = ? AND email = ?", username, email).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Usuário, e-mail ou senha incorretos"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Usuário, e-mail ou senha incorretos"})
		return
	}

	// Criação da sessão
	if err := auth.CreateSession(c, user.ID); err != nil {
		log.Println(" ERRO ao criar sessão:", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Erro interno ao criar sessão"})
		return
	}

	// Log de sucesso
	log.Printf(" Sessão criada para UserID: %d | Username: %s", user.ID, user.Username)

	// Redireciona para dashboard
	c.Redirect(http.StatusFound, "/dashboard")
}