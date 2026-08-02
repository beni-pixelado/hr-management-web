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
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "All fields are required"})
		return
	}

	var existingUser User
	if err := DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "User already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"error": "Internal error"})
		return
	}

	newUser := User{Username: username, Password: string(hashedPassword), Email: email}
	if err := DB.Create(&newUser).Error; err != nil {
		log.Println("Error creating user:", err)
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{"error": "Error creating account"})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{"success": "Account created successfully! Log in."})
}

func Logout(c *gin.Context) {
	if err := auth.DestroySession(c); err != nil {
		log.Println("Error destroying session:", err)
	}

	c.Redirect(http.StatusFound, "/login")
}

func Login(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if username == "" || email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "All fields are required"})
		return
	}

	var user User
	if err := DB.Where("username = ? AND email = ?", username, email).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Incorrect username, email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Incorrect username, email or password"})
		return
	}

	if err := auth.CreateSession(c, user.ID); err != nil {
		log.Println("Error creating the session:", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{"error": "Internal error creating session"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	var user User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	token := uuid.NewString()
	resetLink := fmt.Sprintf("https://meusite.com/reset-password?token=%s", token)
	user.ResetToken = token
	user.ResetExpires = time.Now().Add(1 * time.Hour).Unix()

	if err := DB.Save(&user).Error; err != nil {
		log.Println("Error saving reset token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	params := &resend.SendEmailRequest{
		From:    "HR Management <onboarding@resend.dev>",
		To:      []string{user.Email},
		Subject: "Password recovery",
		Html: fmt.Sprintf(`
        <h2>Password recovery</h2>
        <p>Click the button below:</p>

        <a href="%s">
            Reset password
        </a>
    `, resetLink),
	}

	if _, err := client.Emails.Send(params); err != nil {
		log.Println("Error sending recovery email:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending recovery email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recovery email sent"})
}
