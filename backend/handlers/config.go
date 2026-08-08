package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"hr-management-web/internal/auth"
	"hr-management-web/internal/storage"
)

func ConfigPageHandler(c *gin.Context) {
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

	c.HTML(http.StatusOK, "config.html", gin.H{
		"user": user,
	})
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
		"user": user,
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

func DeleteAccountHandler(c *gin.Context) {
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	currentPassword := c.PostForm("current_password")

	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":   user,
			"error":  "Incorrect password, account was not deleted",
			"delete": true,
		})
		return
	}

	if err := DB.Delete(&user).Error; err != nil {
		log.Println("Error deleting account:", err)
		c.HTML(http.StatusInternalServerError, "account.html", gin.H{
			"User":   user,
			"error":  "Error deleting account",
			"delete": true,
		})
		return
	}

	if err := auth.DestroySession(c); err != nil {
		log.Println("Error destroying session after account deletion:", err)
	}

	c.Redirect(http.StatusFound, "/register")
}

func DevicePageHandler(c *gin.Context) {
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

	ua := c.Request.UserAgent()
	c.HTML(http.StatusOK, "device.html", gin.H{
		"Browser":  detectBrowser(ua),
		"Platform": detectPlatform(ua),
		"user":     user,
	})
}

func detectBrowser(ua string) string {
	switch {
	case strings.Contains(ua, "Edg/"):
		return "Edge"
	case strings.Contains(ua, "Chrome/"):
		return "Chrome"
	case strings.Contains(ua, "Firefox/"):
		return "Firefox"
	case strings.Contains(ua, "Safari/"):
		return "Safari"
	default:
		return "Unknown"
	}
}

func detectPlatform(ua string) string {
	switch {
	case strings.Contains(ua, "Windows"):
		return "Windows"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "Mac OS X"):
		return "macOS/iOS"
	case strings.Contains(ua, "Android"):
		return "Android"
	case strings.Contains(ua, "Linux"):
		return "Linux"
	default:
		return "Unknown"
	}
}

func UpdateProfilePhotoHandler(c *gin.Context) {
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

	file, err := c.FormFile("photo")
	if err == http.ErrMissingFile {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":  user,
			"error": "No photo selected",
		})
		return
	}
	if err != nil {
		log.Println("Error processing upload:", err)
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":  user,
			"error": "Error processing the uploaded file",
		})
		return
	}

	photoURL, saveErr := saveUploadedImage(c, file)
	if saveErr != nil {
		c.HTML(http.StatusBadRequest, "account.html", gin.H{
			"User":  user,
			"error": fmt.Sprintf("Upload error: %v", saveErr),
		})
		return
	}

	if err := DB.Model(&User{}).Where("id = ?", userID).Update("photo", photoURL).Error; err != nil {
		log.Println("Error updating profile photo:", err)
		c.HTML(http.StatusInternalServerError, "account.html", gin.H{
			"User":  user,
			"error": "Error updating profile photo",
		})
		return
	}

	if err := storage.Destroy(c.Request.Context(), user.Photo); err != nil {
		log.Printf("Error deleting old profile photo on Cloudinary: %v\n", err)
	}

	user.Photo = photoURL

	c.HTML(http.StatusOK, "account.html", gin.H{
		"User":    user,
		"success": "Profile photo updated successfully!",
	})
}
