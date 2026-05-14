package handlers

import (
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	
)

var DB *gorm.DB

type Employee struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	FullName string `json:"full_name" gorm:"not null"`
	Email    string `json:"email" gorm:"not null"`
	Position string `json:"position" gorm:"not null"`
	Status   string `json:"status" gorm:"not null;default:'pending'"`
	HireDate string `json:"hire_date"`
	Photo    string `json:"photo"`
}

const (
	MaxFileSize = 5 * 1024 * 1024 // 5MB
	uploadsDir  = "./uploads"
)

var allowedMimeTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
}

func saveUploadedImage(c *gin.Context, file *multipart.FileHeader) (string, error) {

	if file.Size > MaxFileSize {
		return "", fmt.Errorf("arquivo muito grande (máximo 5MB, recebido %.2fMB)",
			float64(file.Size)/(1024*1024))
	}

	ext := filepath.Ext(file.Filename)
	mimeType := mime.TypeByExtension(ext)

	if mimeType == "" {

		src, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("erro ao ler arquivo: %v", err)
		}
		defer src.Close()

		buffer := make([]byte, 512)
		if _, err := src.Read(buffer); err != nil {
			return "", fmt.Errorf("erro ao detectar tipo: %v", err)
		}
		mimeType = http.DetectContentType(buffer)
	}

	isAllowed := false
	for _, allowed := range allowedMimeTypes {
		if strings.HasPrefix(mimeType, allowed) || mimeType == allowed {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("tipo de arquivo não permitido: %s (aceitos: JPG, PNG, GIF, WebP)",
			mimeType)
	}

	if ext == "" {

		extMap := map[string]string{
			"image/jpeg": ".jpg",
			"image/png":  ".png",
			"image/gif":  ".gif",
			"image/webp": ".webp",
		}
		ext = extMap[mimeType]
	}

	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", fmt.Errorf("erro ao criar pasta de uploads: %v", err)
	}

	uniqueFilename := uuid.New().String() + ext

	filePath := filepath.Join(uploadsDir, uniqueFilename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	return uniqueFilename, nil
}

func GetEmployees(c *gin.Context) {
	var employees []Employee

	pageStr := c.DefaultQuery("page", "1")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit := 20
	offset := (page - 1) * limit

	var totalEmployees int64
	DB.Model(&Employee{}).Count(&totalEmployees)

	DB.
		Limit(limit).
		Offset(offset).
		Find(&employees)

	totalPages := int(math.Ceil(float64(totalEmployees) / float64(limit)))

	c.HTML(http.StatusOK, "employees.html", gin.H{
	"employees":      employees,
	"currentPage":    page,
	"totalPages":     totalPages,
	"totalEmployees": totalEmployees,
	"prevPage":       page - 1,
	"nextPage":       page + 1,
})
}

func CreateEmployee(c *gin.Context) {
	fullName := c.PostForm("full_name")
	email := c.PostForm("email")
	position := c.PostForm("position")

	if fullName == "" || email == "" || position == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Todos os campos (nome, email, cargo) são obrigatórios"})
		return
	}

	employee := Employee{
		FullName: fullName,
		Email:    email,
		Position: position,
		Status:   "pending",
	}

	file, err := c.FormFile("photo")
	if err == nil && file != nil {

		uniqueFilename, saveErr := saveUploadedImage(c, file)
		if saveErr != nil {
			log.Printf("Erro ao salvar imagem: %v\n", saveErr)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Erro no upload: %v", saveErr),
			})
			return
		}
		employee.Photo = uniqueFilename
	} else if err != http.ErrMissingFile {

		log.Printf("Erro ao processar upload: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Erro ao processar o arquivo enviado",
		})
		return
	}

	if err := DB.Create(&employee).Error; err != nil {
		log.Println("Erro ao criar funcionário:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao salvar funcionário no banco de dados",
		})
		return
	}

	log.Printf("Novo funcionário adicionado: %s (Foto: %s)\n", fullName, employee.Photo)

	c.Redirect(http.StatusFound, "/employees")
}

func UpdateEmployeeStatus(c *gin.Context) {
	id := c.Param("id")
	newStatus := c.PostForm("status")
	hireDate := c.PostForm("hire_date")

	if newStatus == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status é obrigatório"})
		return
	}

	var employee Employee
	if err := DB.Model(&employee).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    newStatus,
		"hire_date": hireDate,
	}).Error; err != nil {
		log.Println("Erro ao atualizar funcionário:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar funcionário"})
		return
	}

	log.Printf("Funcionário %s atualizado para status: %s\n", id, newStatus)

	c.Redirect(http.StatusFound, "/employees")
}

func DeleteEmployee(c *gin.Context) {
	id := c.Param("id")

	if err := DB.Delete(&Employee{}, id).Error; err != nil {
		log.Println("Erro ao deletar funcionário:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar funcionário"})
		return
	}

	log.Printf("Funcionário %s deletado\n", id)

	c.JSON(http.StatusOK, gin.H{"message": "Funcionário deletado com sucesso"})
}

// GetEmployeesAPI retorna lista filtrada de funcionários em JSON
func GetEmployeesAPI(c *gin.Context) {
    // Pega os parâmetros da URL
    search := c.DefaultQuery("search", "")      // ?search=joao
    status := c.DefaultQuery("status", "all")   // ?status=pending
    
    // Começa a query
    query := DB.Model(&Employee{})
    
    // Filtro por nome ou email (se tiver busca)
    if search != "" {
        query = query.Where("full_name LIKE ? OR email LIKE ?", 
            "%"+search+"%", 
            "%"+search+"%")
    }
    
    // Filtro por status (se não for "all")
    if status != "all" {
        query = query.Where("status = ?", status)
    }
    
    // Executa a busca
    var employees []Employee
    if err := query.Find(&employees).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Erro ao buscar funcionários",
        })
        return
    }
    
    // Retorna como JSON
    c.JSON(http.StatusOK, gin.H{
        "employees": employees,
        "total":     len(employees),
    })
}

func BadgeHandler(c *gin.Context) {
    id := c.Param("id")

    var employee Employee

    if err := DB.First(&employee, id).Error; err != nil {
        c.String(404, "Employee not found")
        return
    }

    c.HTML(200, "id-card.html", employee)
}