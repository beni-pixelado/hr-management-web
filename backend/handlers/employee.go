package handlers

import (
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hr-management-web/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Employee struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	UserID uint `gorm:"not null;index"`
	User   User `gorm:"constraint:OnDelete:CASCADE;"`

	FullName     string `json:"full_name" gorm:"not null"`
	Email        string `json:"email" gorm:"not null"`
	Position     string `json:"position" gorm:"not null"`
	Description  string `json:"description" gorm:"type:text"`
	Status       string `json:"status" gorm:"not null;default:'pending'"`
	HireDate     string `json:"hire_date"`
	Photo        string `json:"photo"`
	DepartmentID uint   `json:"department_id"`
}

const MaxFileSize = 5 * 1024 * 1024

var allowedMimeTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
}

func saveUploadedImage(c *gin.Context, file *multipart.FileHeader) (string, error) {

	if file.Size > MaxFileSize {
		return "", fmt.Errorf("file too large (max 5MB, received %.2fMB)",
			float64(file.Size)/(1024*1024))
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}
	defer src.Close()

	buffer := make([]byte, 512)
	n, err := src.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error detecting file type: %v", err)
	}

	mimeType := http.DetectContentType(buffer[:n])

	isAllowed := false
	for _, allowed := range allowedMimeTypes {
		if strings.HasPrefix(mimeType, allowed) || mimeType == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("file type not allowed: %s (accepted: JPG, PNG, GIF, WebP)",
			mimeType)
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("error preparing upload: %v", err)
	}

	secureURL, err := storage.Upload(c.Request.Context(), src, uuid.New().String())
	if err != nil {
		return "", fmt.Errorf("error uploading to Cloudinary: %v", err)
	}

	return secureURL, nil
}

func GetEmployees(c *gin.Context) {
	var employees []Employee

	pageStr := c.DefaultQuery("page", "1")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	status := c.DefaultQuery("status", "all")

	statusFilter := status
	currentLabel := ""
	switch status {
	case "interviewing":
		statusFilter = "pending"
		currentLabel = "in interview"
	case "hired":
		statusFilter = "contractors"
		currentLabel = "hired"
	case "rejected":
		statusFilter = "rejected"
		currentLabel = "rejected"
	}

	limit := 20
	offset := (page - 1) * limit

	var totalAll int64
	var totalInterviewing int64
	var totalHired int64
	var totalRejected int64
	var totalFiltered int64
	userID := GetCurrentUserID(c)

	DB.
		Model(&Employee{}).
		Where("user_id = ?", userID).
		Count(&totalAll)

	DB.
		Model(&Employee{}).
		Where("user_id = ? AND status = ?", userID, "pending").
		Count(&totalInterviewing)

	DB.
		Model(&Employee{}).
		Where("user_id = ? AND status = ?", userID, "contractors").
		Count(&totalHired)

	DB.
		Model(&Employee{}).
		Where("user_id = ? AND status = ?", userID, "rejected").
		Count(&totalRejected)

	countQuery := DB.
		Model(&Employee{}).
		Where("user_id = ?", userID)
	if statusFilter != "all" {
		countQuery = countQuery.Where("status = ?", statusFilter)
	}
	countQuery.Count(&totalFiltered)

	listQuery := DB.Where("user_id = ?", userID)
	if statusFilter != "all" {
		listQuery = listQuery.Where("status = ?", statusFilter)
	}
	listQuery.
		Limit(limit).
		Offset(offset).
		Find(&employees)

	totalPages := int(math.Ceil(float64(totalFiltered) / float64(limit)))

	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "employees.html", gin.H{
		"employees":         employees,
		"currentPage":       page,
		"totalPages":        totalPages,
		"totalEmployees":    totalAll,
		"currentCount":      totalFiltered,
		"totalInterviewing": totalInterviewing,
		"totalHired":        totalHired,
		"totalRejected":     totalRejected,
		"currentStatus":     status,
		"currentLabel":      currentLabel,
		"prevPage":          page - 1,
		"nextPage":          page + 1,
		"user":              user,
	})
}

func CreateEmployee(c *gin.Context) {
	fullName := c.PostForm("full_name")
	email := c.PostForm("email")
	position := c.PostForm("position")
	description := c.PostForm("description")

	if fullName == "" || email == "" || position == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields (name, email, position) are required"})
		return
	}

	employee := Employee{
		UserID:      GetCurrentUserID(c),
		FullName:    fullName,
		Email:       email,
		Position:    position,
		Description: description,
		Status:      "pending",
	}

	file, err := c.FormFile("photo")
	if err == nil && file != nil {

		photoURL, saveErr := saveUploadedImage(c, file)
		if saveErr != nil {
			log.Printf("Error saving image: %v\n", saveErr)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Upload error: %v", saveErr),
			})
			return
		}
		employee.Photo = photoURL
	} else if err != http.ErrMissingFile {

		log.Printf("Error processing upload: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error processing the uploaded file",
		})
		return
	}

	if err := DB.Create(&employee).Error; err != nil {
		log.Println("Error creating employee:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error saving employee to the database",
		})
		return
	}

	log.Printf("New employee added: %s (Photo: %s)\n", fullName, employee.Photo)

	c.Redirect(http.StatusFound, "/employees")
}

func UpdateEmployeeStatus(c *gin.Context) {
	id := c.Param("id")
	newStatus := c.PostForm("status")
	hireDate := c.PostForm("hire_date")

	if newStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	var employee Employee
	if err := DB.Model(&employee).Where("id = ? AND user_id = ?", id, GetCurrentUserID(c)).Updates(map[string]interface{}{
		"status":    newStatus,
		"hire_date": hireDate,
	}).Error; err != nil {
		log.Println("Error updating employee:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating employee"})
		return
	}

	c.Redirect(http.StatusFound, "/employees")
}

func DeleteEmployee(c *gin.Context) {
	id := c.Param("id")

	var employee Employee
	if err := DB.
		Where("id = ? AND user_id = ?", id, GetCurrentUserID(c)).
		First(&employee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	if err := DB.Delete(&employee).Error; err != nil {
		log.Println("Error deleting employee:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting employee"})
		return
	}

	if err := storage.Destroy(c.Request.Context(), employee.Photo); err != nil {
		log.Printf("Error deleting photo on Cloudinary: %v\n", err)
	}

	log.Printf("Employee %s deleted\n", id)

	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
}

func DeleteEmployeeForm(c *gin.Context) {
	id := c.Param("id")

	var employee Employee
	if err := DB.
		Where("id = ? AND user_id = ?", id, GetCurrentUserID(c)).
		First(&employee).Error; err != nil {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	if err := DB.Delete(&employee).Error; err != nil {
		log.Println("Error deleting employee (form):", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting employee"})
		return
	}

	if err := storage.Destroy(c.Request.Context(), employee.Photo); err != nil {
		log.Printf("Error deleting photo on Cloudinary: %v\n", err)
	}

	log.Printf("Employee %s deleted via form\n", id)

	c.Redirect(http.StatusFound, "/employees")
}

func EditEmployeePage(c *gin.Context) {
	id := c.Param("id")
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var employee Employee
	if err := DB.
		Where("id = ? AND user_id = ?", id, userID).
		First(&employee).Error; err != nil {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	var user User
	if err := DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	c.HTML(http.StatusOK, "employee-edit.html", gin.H{
		"Employee": employee,
		"user":     user,
	})
}

func UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var employee Employee
	if err := DB.
		Where("id = ? AND user_id = ?", id, userID).
		First(&employee).Error; err != nil {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	fullName := c.PostForm("full_name")
	email := c.PostForm("email")
	position := c.PostForm("position")
	description := c.PostForm("description")

	if fullName == "" || email == "" || position == "" {
		c.Redirect(http.StatusFound, "/employees/"+id+"/edit")
		return
	}

	updates := map[string]interface{}{
		"full_name":   fullName,
		"email":       email,
		"position":    position,
		"description": description,
	}

	photoReplaced := false

	file, err := c.FormFile("photo")
	if err == nil {
		photoURL, saveErr := saveUploadedImage(c, file)
		if saveErr != nil {
			log.Printf("Error saving image: %v\n", saveErr)
			c.Redirect(http.StatusFound, "/employees/"+id+"/edit")
			return
		}
		updates["photo"] = photoURL
		photoReplaced = true
	} else if err != http.ErrMissingFile {
		log.Printf("Error processing upload: %v\n", err)
		c.Redirect(http.StatusFound, "/employees/"+id+"/edit")
		return
	}

	if err := DB.
		Model(&Employee{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error; err != nil {
		log.Println("Error updating employee:", err)
		c.Redirect(http.StatusFound, "/employees/"+id+"/edit")
		return
	}

	if photoReplaced {
		if err := storage.Destroy(c.Request.Context(), employee.Photo); err != nil {
			log.Printf("Error deleting old photo on Cloudinary: %v\n", err)
		}
	}

	log.Printf("Employee %s updated\n", id)

	c.Redirect(http.StatusFound, "/employees")
}

func GetEmployeesAPI(c *gin.Context) {

	search := c.DefaultQuery("search", "")
	status := c.DefaultQuery("status", "all")

	query := DB.
		Model(&Employee{}).
		Where("user_id = ?", GetCurrentUserID(c))

	if search != "" {
		query = query.Where("full_name LIKE ? OR email LIKE ?",
			"%"+search+"%",
			"%"+search+"%")
	}

	if status != "all" {
		query = query.Where("status = ?", status)
	}

	var employees []Employee
	if err := query.Find(&employees).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching employees",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"employees": employees,
		"total":     len(employees),
	})
}

func BadgeHandler(c *gin.Context) {
	id := c.Param("id")

	var employee Employee

	if err := DB.
		Where("id = ? AND user_id = ?", id, GetCurrentUserID(c)).
		First(&employee).Error; err != nil {
		c.String(404, "Employee not found")
		return
	}

	c.HTML(200, "id-card.html", employee)
}
