package handlers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/resend/resend-go/v2"

	"hr-management-web/internal/auth"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique;not null"`
	Password     string `json:"password"`
	Email        string `json:"email" gorm:"unique;not null"`
	ResetToken   string `json:"reset_token,omitempty" gorm:"column:reset_token"`
	ResetExpires int64  `json:"reset_expires,omitempty" gorm:"column:reset_expires"`
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

	if err := auth.CreateSession(c, user.ID); err != nil {
		log.Println(" ERRO creating the session:", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Erro interno ao criar sessão"})
		return
	}

	log.Printf(" Session has been created for UserID: %d | Username: %s", user.ID, user.Username)

	c.Redirect(http.StatusFound, "/dashboard")
}

func GetCurrentUserID(c *gin.Context) uint {
	_, userID := auth.IsAuthenticated(c)
	return uint(userID)
}

func RecoverAccount(c *gin.Context) {
	email := c.PostForm("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "E-mail é obrigatório"})
		return
	}

	var user User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	token := uuid.NewString()
	resetLink := fmt.Sprintf("https://meusite.com/reset-password?token=%s", token)
	user.ResetToken = token
	user.ResetExpires = time.Now().Add(1 * time.Hour).Unix()

	if err := DB.Save(&user).Error; err != nil {
		log.Println("Erro ao salvar token de reset:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno"})
		return
	}

	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	params := &resend.SendEmailRequest{
		From:    "HR Management <onboarding@resend.dev>",
		To:      []string{user.Email},
		Subject: "Recuperação de senha",
		Html: fmt.Sprintf(`
        <h2>Recuperação de senha</h2>
        <p>Clique no botão abaixo:</p>

        <a href="%s">
            Recuperar senha
        </a>
    `, resetLink),
	}

	if _, err := client.Emails.Send(params); err != nil {
		log.Println("Erro ao enviar e-mail de recuperação:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar e-mail de recuperação"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "E-mail de recuperação enviado"})
}
