package main

import (
	"fmt"
	"hr-management-web/backend/handlers"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
    "time"
)

func main() {
	db, err := gorm.Open(sqlite.Open("data/users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	if err := db.AutoMigrate(&handlers.User{}, &handlers.Employee{}); err != nil {
		log.Fatal("Migration error:", err)
	}

	fmt.Println("Creating 60 permanent users (if they don't already exist)...")

	created := 0
	skipped := 0

	for i := 1; i <= 100; i++ {
		username := fmt.Sprintf("user_%d_%d", i, time.Now().UnixNano())
		email := fmt.Sprintf("%s@example.com", username)
		password := "password123"

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("error hashing password for %s: %v", username, err)
			continue
		}

		u := handlers.User{Username: username, Email: email, Password: string(hash)}
		if err := db.Create(&u).Error; err != nil {
			log.Printf("error creating %s: %v", username, err)
			continue
		}

		created++
		fmt.Printf("[CREATED] %s\n", username)
	}

	fmt.Println("---")
	fmt.Printf("Created: %d\n", created)
	fmt.Printf("Existing (skipped): %d\n", skipped)
}
