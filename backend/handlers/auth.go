package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"hr-management-web/internal/auth"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique;not null"`
	Password     string `json:"password"`
	Email        string `json:"email" gorm:"unique;not null"`
	Photo        string `json:"photo"`
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
	email := strings.TrimSpace(c.PostForm("email"))
	if email == "" {
		c.HTML(http.StatusBadRequest, "recover.html", gin.H{"error": "Email is required"})
		return
	}

	var user User
	if err := DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.HTML(http.StatusOK, "recover.html", gin.H{"error": "If this email exists, a recovery link was sent"})
		return
	}

	token := uuid.NewString()
	expiresAt := time.Now().Add(1 * time.Hour).Unix()
	resetRecord := createPasswordResetToken(user.Email, token, expiresAt)

	if err := DB.Create(resetRecord).Error; err != nil {
		log.Println("Error saving reset code:", err)
		c.HTML(http.StatusInternalServerError, "recover.html", gin.H{"error": "Internal error, try again"})
		return
	}

	baseURL := strings.TrimRight(os.Getenv("SITE_URL"), "/")
	if baseURL == "" {
		baseURL = "http://" + c.Request.Host
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
</head>
<body style="margin:0;padding:0;background:#f4f6f9;font-family:Arial,Helvetica,sans-serif;">

    <table width="100%%" cellpadding="0" cellspacing="0" style="padding:40px 0;">
        <tr>
            <td align="center">

                <table width="500" cellpadding="0" cellspacing="0"
                    style="background:#ffffff;border-radius:12px;padding:40px;box-shadow:0 4px 12px rgba(0,0,0,.08);">

                    <tr>
                        <td align="center">
                            <h1 style="margin:0;color:#2563eb;">
                                Staffio
                            </h1>

                            <h2 style="margin:25px 0 10px;color:#222;">
                                Password Recovery
                            </h2>

                            <p style="color:#555;line-height:1.6;">
                                We received a request to reset your password.
                                Click the button below to create a new password.
                            </p>

                            <a href="%s"
                               style="display:inline-block;
                                      margin:30px 0;
                                      padding:14px 28px;
                                      background:#2563eb;
                                      color:#ffffff;
                                      text-decoration:none;
                                      border-radius:8px;
                                      font-weight:bold;">
                                Reset Password
                            </a>

                            <p style="font-size:14px;color:#777;line-height:1.5;">
                                This link will expire in <strong>15 minutes</strong>.
                            </p>

                            <hr style="border:none;border-top:1px solid #eee;margin:30px 0;">

                            <p style="font-size:13px;color:#888;line-height:1.5;">
                                If you didn't request a password reset, you can safely ignore this email.
                                Your account will remain secure.
                            </p>

                        </td>
                    </tr>

                </table>

            </td>
        </tr>
    </table>

</body>
</html>
`, resetLink)

	if err := sendEmail(user.Email, "Password recovery", html); err != nil {
		log.Println("Error sending recovery email:", err)
		c.HTML(http.StatusInternalServerError, "recover.html", gin.H{"error": "Error sending recovery email"})
		return
	}

	c.HTML(http.StatusOK, "recover.html", gin.H{"success": "Recovery code sent to your email"})
}

func sendEmail(to, subject, html string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = user
	}
	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "Staffio"
	}

	auth := smtp.PlainAuth("", user, password, host)

	message := fmt.Sprintf(
		"From: %s <%s>\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s",
		fromName, from, to, subject, html,
	)

	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(message))
}

func ResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "Invalid or missing token"})
		return
	}

	var resetRecord PasswordResetToken
	if err := DB.Where("token = ?", token).First(&resetRecord).Error; err != nil {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "Invalid or expired link"})
		return
	}

	if resetRecord.ExpiresAt < time.Now().Unix() {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "This link has expired, request a new one"})
		return
	}

	c.HTML(http.StatusOK, "reset-password.html", gin.H{"Token": token})
}

func ResetPassword(c *gin.Context) {
	token := c.PostForm("token")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	if token == "" {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "Invalid or missing token"})
		return
	}

	if newPassword == "" || newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "Passwords do not match", "Token": token})
		return
	}

	if len(newPassword) < 6 {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "Password must be at least 6 characters", "Token": token})
		return
	}

	var resetRecord PasswordResetToken
	if err := DB.Where("token = ?", token).First(&resetRecord).Error; err != nil {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "Invalid or expired link"})
		return
	}

	if resetRecord.ExpiresAt < time.Now().Unix() {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "This link has expired, request a new one"})
		return
	}

	var user User
	if err := DB.Where("email = ?", resetRecord.Email).First(&user).Error; err != nil {
		c.HTML(http.StatusBadRequest, "reset-password.html", gin.H{"error": "Invalid or expired link"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error hashing password:", err)
		c.HTML(http.StatusInternalServerError, "reset-password.html", gin.H{"error": "Internal error, try again"})
		return
	}

	user.Password = string(hashedPassword)
	if err := DB.Save(&user).Error; err != nil {
		log.Println("Error saving new password:", err)
		c.HTML(http.StatusInternalServerError, "reset-password.html", gin.H{"error": "Internal error, try again"})
		return
	}

	if err := DB.Delete(&resetRecord).Error; err != nil {
		log.Println("Error deleting reset record:", err)
	}

	c.Redirect(http.StatusFound, "/login")
}

func RecoverAccountPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "recover.html", nil)
}
