package handlers

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
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

	UserID         uint `gorm:"not null;index" json:"user_id"`
	OrganizationID uint `gorm:"index;default:0" json:"organization_id"`

	FullName     string `json:"full_name" gorm:"not null"`
	Email        string `json:"email" gorm:"not null"`
	Position     string `json:"position" gorm:"not null"`
	Description  string `json:"description" gorm:"type:text"`
	Status       string `json:"status" gorm:"not null;default:'pending'"`
	HireDate     string    `json:"hire_date"`
	Photo        string    `json:"photo"`
	DepartmentID uint      `json:"department_id"`
	Absences     uint      `json:"absences" gorm:"not null;default:0"`
	RejectedAt   *time.Time `json:"rejected_at"`
}

type Absence struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UserID         uint      `gorm:"not null;index"`
	EmployeeID     uint      `gorm:"not null;index"`
	OrganizationID uint      `gorm:"index;default:0"`
}

type Termination struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	OrganizationID uint      `gorm:"index;default:0" json:"organization_id"`
	EmployeeID     uint      `json:"employee_id"`
	EmployeeName   string    `json:"employee_name"`
	Reason         string    `json:"reason"`
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
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	orgID := actor.OrganizationID

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

	DB.
		Model(&Employee{}).
		Where("organization_id = ?", orgID).
		Count(&totalAll)

	DB.
		Model(&Employee{}).
		Where("organization_id = ? AND status = ?", orgID, "pending").
		Count(&totalInterviewing)

	DB.
		Model(&Employee{}).
		Where("organization_id = ? AND status = ?", orgID, "contractors").
		Count(&totalHired)

	DB.
		Model(&Employee{}).
		Where("organization_id = ? AND status = ?", orgID, "rejected").
		Count(&totalRejected)

	countQuery := DB.
		Model(&Employee{}).
		Where("organization_id = ?", orgID)
	if statusFilter != "all" {
		countQuery = countQuery.Where("status = ?", statusFilter)
	}
	countQuery.Count(&totalFiltered)

	listQuery := DB.Where("organization_id = ?", orgID)
	if statusFilter != "all" {
		listQuery = listQuery.Where("status = ?", statusFilter)
	}
	listQuery.
		Limit(limit).
		Offset(offset).
		Find(&employees)

	totalPages := int(math.Ceil(float64(totalFiltered) / float64(limit)))

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
		"user":              actor,
		"canEdit":           actor.CanEditEmployees(),
	})
}

func DownloadEmployeesCSV(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	statusFilter := c.DefaultQuery("status", "all")
	switch statusFilter {
	case "interviewing":
		statusFilter = "pending"
	case "hired":
		statusFilter = "contractors"
	}

	query := DB.
		Model(&Employee{}).
		Where("organization_id = ?", actor.OrganizationID)
	if statusFilter != "all" {
		query = query.Where("status = ?", statusFilter)
	}

	var employees []Employee
	if err := query.Order("full_name").Find(&employees).Error; err != nil {
		slog.Error("Error fetching employees for CSV", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error exporting employees"})
		return
	}

	fileName := "employees.csv"
	if statusFilter != "all" {
		fileName = "employees_" + statusFilter + ".csv"
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename="+fileName)

	writer := csv.NewWriter(c.Writer)
	writer.Write([]string{"ID", "Full Name", "Email", "Position", "Description", "Status", "Hiring Date", "Absences", "Created At"})
	for _, emp := range employees {
		writer.Write([]string{
			strconv.FormatUint(uint64(emp.ID), 10),
			emp.FullName,
			emp.Email,
			emp.Position,
			emp.Description,
			emp.Status,
			emp.HireDate,
			strconv.FormatUint(uint64(emp.Absences), 10),
			emp.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	writer.Flush()
}

func CreateEmployee(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanEditEmployees() {
		abortUnauthorized(c)
		return
	}

	fullName := c.PostForm("full_name")
	email := c.PostForm("email")
	position := c.PostForm("position")
	description := c.PostForm("description")

	if fullName == "" || email == "" || position == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields (name, email, position) are required"})
		return
	}

	employee := Employee{
		UserID:         actor.ID,
		OrganizationID: actor.OrganizationID,
		FullName:       fullName,
		Email:          email,
		Position:       position,
		Description:    description,
		Status:         "pending",
	}

	file, err := c.FormFile("photo")
	if err == nil && file != nil {

		photoURL, saveErr := saveUploadedImage(c, file)
		if saveErr != nil {
			slog.Error("Error saving image", "error", saveErr)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Upload error: %v", saveErr),
			})
			return
		}
		employee.Photo = photoURL
	} else if err != http.ErrMissingFile {

		slog.Error("Error processing upload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Error processing the uploaded file",
		})
		return
	}

	if err := DB.Create(&employee).Error; err != nil {
		slog.Error("Error creating employee", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error saving employee to the database",
		})
		return
	}

	slog.Info("New employee added", "name", fullName, "photo", employee.Photo)

	c.Redirect(http.StatusFound, "/employees")
}

func MarkAbsence(c *gin.Context) {
	id := c.Param("id")
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !actor.CanEditEmployees() {
		abortUnauthorized(c)
		return
	}
	orgID := actor.OrganizationID

	err := DB.Transaction(func(tx *gorm.DB) error {
		var employee Employee
		if err := tx.
			Where("id = ? AND organization_id = ?", id, orgID).
			First(&employee).Error; err != nil {
			return err
		}

		if err := tx.Create(&Absence{
			UserID:         actor.ID,
			EmployeeID:     employee.ID,
			OrganizationID: orgID,
		}).Error; err != nil {
			return err
		}

		return tx.
			Model(&Employee{}).
			Where("id = ? AND organization_id = ?", id, orgID).
			UpdateColumn("absences", gorm.Expr("absences + 1")).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
			return
		}
		slog.Error("Error marking absence", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marking absence"})
		return
	}

	slog.Info("Absence marked for employee", "id", id)

	c.JSON(http.StatusOK, gin.H{"message": "Absence marked"})
}

func UpdateEmployeeStatus(c *gin.Context) {
	id := c.Param("id")
	newStatus := c.PostForm("status")
	hireDate := c.PostForm("hire_date")

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanEditEmployees() {
		abortUnauthorized(c)
		return
	}

	if newStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	// Recruit may not reject/fire candidates whose email belongs to a member
	// with a higher role than them.
	if actor.IsRecruit() && (newStatus == "rejected") {
		blocked, _ := employeeLinksToSuperior(id, actor.OrganizationID, actor)
		if blocked {
			abortUnauthorized(c)
			return
		}
	}

	updates := map[string]interface{}{
		"status":    newStatus,
		"hire_date": hireDate,
	}
	now := time.Now()
	if newStatus == "rejected" {
		updates["rejected_at"] = now
	} else if newStatus != "rejected" {
		updates["rejected_at"] = nil
	}

	var employee Employee
	if err := DB.Model(&employee).Where("id = ? AND organization_id = ?", id, actor.OrganizationID).Updates(updates).Error; err != nil {
		slog.Error("Error updating employee", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating employee"})
		return
	}

	if newStatus == "rejected" {
		var emp Employee
		if err := DB.Where("id = ? AND organization_id = ?", id, actor.OrganizationID).First(&emp).Error; err == nil {
			DB.Create(&Termination{
				OrganizationID: actor.OrganizationID,
				EmployeeID:     emp.ID,
				EmployeeName:   emp.FullName,
				Reason:         "rejected",
			})
		}
	}

	c.Redirect(http.StatusFound, "/employees")
}

func DeleteEmployee(c *gin.Context) {
	id := c.Param("id")

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if !actor.CanEditEmployees() {
		abortUnauthorized(c)
		return
	}
	if actor.IsRecruit() {
		if blocked, _ := employeeLinksToSuperior(id, actor.OrganizationID, actor); blocked {
			abortUnauthorized(c)
			return
		}
	}

	var employee Employee
	if err := DB.
		Where("id = ? AND organization_id = ?", id, actor.OrganizationID).
		First(&employee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	if err := DB.Delete(&employee).Error; err != nil {
		slog.Error("Error deleting employee", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting employee"})
		return
	}

	DB.Create(&Termination{
		OrganizationID: actor.OrganizationID,
		EmployeeID:     employee.ID,
		EmployeeName:   employee.FullName,
		Reason:         "deleted",
	})

	if err := storage.Destroy(c.Request.Context(), employee.Photo); err != nil {
		slog.Error("Error deleting photo on Cloudinary", "error", err)
	}

	slog.Info("Employee deleted", "id", id)

	c.JSON(http.StatusOK, gin.H{"message": "Employee deleted successfully"})
}

func DeleteEmployeeForm(c *gin.Context) {
	id := c.Param("id")

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanEditEmployees() {
		abortUnauthorized(c)
		return
	}
	if actor.IsRecruit() {
		if blocked, _ := employeeLinksToSuperior(id, actor.OrganizationID, actor); blocked {
			abortUnauthorized(c)
			return
		}
	}

	var employee Employee
	if err := DB.
		Where("id = ? AND organization_id = ?", id, actor.OrganizationID).
		First(&employee).Error; err != nil {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	if err := DB.Delete(&employee).Error; err != nil {
		slog.Error("Error deleting employee (form)", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting employee"})
		return
	}

	DB.Create(&Termination{
		OrganizationID: actor.OrganizationID,
		EmployeeID:     employee.ID,
		EmployeeName:   employee.FullName,
		Reason:         "deleted",
	})

	if err := storage.Destroy(c.Request.Context(), employee.Photo); err != nil {
		slog.Error("Error deleting photo on Cloudinary", "error", err)
	}

	slog.Info("Employee deleted via form", "id", id)

	c.Redirect(http.StatusFound, "/employees")
}

func EditEmployeePage(c *gin.Context) {
	id := c.Param("id")
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanEditEmployees() {
		abortUnauthorized(c)
		return
	}

	var employee Employee
	if err := DB.
		Where("id = ? AND organization_id = ?", id, actor.OrganizationID).
		First(&employee).Error; err != nil {
		c.Redirect(http.StatusFound, "/employees")
		return
	}

	c.HTML(http.StatusOK, "employee-edit.html", gin.H{
		"Employee": employee,
		"user":     actor,
		"canEdit":  true,
	})
}

func UpdateEmployee(c *gin.Context) {
	id := c.Param("id")
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	if !actor.CanEditEmployees() {
		abortUnauthorized(c)
		return
	}

	var employee Employee
	if err := DB.
		Where("id = ? AND organization_id = ?", id, actor.OrganizationID).
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
			slog.Error("Error saving image", "error", saveErr)
			c.Redirect(http.StatusFound, "/employees/"+id+"/edit")
			return
		}
		updates["photo"] = photoURL
		photoReplaced = true
	} else if err != http.ErrMissingFile {
		slog.Error("Error processing upload", "error", err)
		c.Redirect(http.StatusFound, "/employees/"+id+"/edit")
		return
	}

	if err := DB.
		Model(&Employee{}).
		Where("id = ? AND organization_id = ?", id, actor.OrganizationID).
		Updates(updates).Error; err != nil {
		slog.Error("Error updating employee", "error", err)
		c.Redirect(http.StatusFound, "/employees/"+id+"/edit")
		return
	}

	if photoReplaced {
		if err := storage.Destroy(c.Request.Context(), employee.Photo); err != nil {
			slog.Error("Error deleting old photo on Cloudinary", "error", err)
		}
	}

	slog.Info("Employee updated", "id", id)

	c.Redirect(http.StatusFound, "/employees")
}

func GetEmployeesAPI(c *gin.Context) {
	actor, ok := GetCurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	search := c.DefaultQuery("search", "")
	status := c.DefaultQuery("status", "all")

	query := DB.
		Model(&Employee{}).
		Where("organization_id = ?", actor.OrganizationID)

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

	actor, ok := GetCurrentUser(c)
	if !ok {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var employee Employee

	if err := DB.
		Where("id = ? AND organization_id = ?", id, actor.OrganizationID).
		First(&employee).Error; err != nil {
		c.String(404, "Employee not found")
		return
	}

	c.HTML(200, "id-card.html", employee)
}
