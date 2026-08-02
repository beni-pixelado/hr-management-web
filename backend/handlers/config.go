package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func ConfigPageHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "config.html", nil)
}

func AccountPageHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "account.html", gin.H{
		"User": user,
	})
}

func UpdateProfileHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	username := c.PostForm("username")
	email := c.PostForm("email")

	loadUser := func() User {
		var user User
		DB.Where("id = ?", userID).First(&user)
		return user
	}

	if username == "" || email == "" {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":    loadUser(),
			"error":   "All fields are mandatory",
			"profile": true,
		})
		return
	}

	var existing User
	if err := DB.
		Where("(username = ? OR email = ?) AND id != ?", username, email, userID).
		First(&existing).Error; err == nil {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":    loadUser(),
			"error":   "Username or email is already in use",
			"profile": true,
		})
		return
	}

	if err := DB.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"username": username,
		"email":    email,
	}).Error; err != nil {
		log.Println("Error updating profile:", err)
		c.HTML(http.StatusInternalServerError, "account.html", gin.H{
			"User":    loadUser(),
			"error":   "Error updating profile",
			"profile": true,
		})
		return
	}

	var user User
	DB.Where("id = ?", userID).First(&user)

	c.HTML(http.StatusOK, "account.html", gin.H{
		"User":    user,
		"success": "Profile updated successfully!",
		"profile": true,
	})
}

func ChangePasswordHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	currentPassword := c.PostForm("current_password")
	newPassword := c.PostForm("new_password")
	confirmPassword := c.PostForm("confirm_password")

	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if newPassword == "" {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":     user,
			"error":    "The new password cannot be empty",
			"password": true,
		})
		return
	}

	if newPassword != confirmPassword {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":     user,
			"error":    "Passwords don't match",
			"password": true,
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":     user,
			"error":    "Incorrect current password",
			"password": true,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Println("Error generating hash:", err)
		c.HTML(http.StatusInternalServerError, "account.html", gin.H{
			"User":     user,
			"error":    "Internal error when changing password",
			"password": true,
		})
		return
	}

	if err := DB.Model(&User{}).Where("id = ?", userID).Update("password", string(hashedPassword)).Error; err != nil {
		log.Println("Error changing password:", err)
		c.HTML(http.StatusInternalServerError, "account.html", gin.H{
			"User":     user,
			"error":    "Error changing password",
			"password": true,
		})
		return
	}

	c.HTML(http.StatusOK, "account.html", gin.H{
		"User":     user,
		"success":  "Password changed successfully!",
		"password": true,
	})
}
