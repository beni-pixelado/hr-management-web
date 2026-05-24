package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	tables := []string{"users", "employees", "departments"}

	for _, t := range tables {
		q := fmt.Sprintf("SELECT setval(pg_get_serial_sequence('%s','id'), COALESCE((SELECT MAX(id) FROM %s), 1));", t, t)
		if err := db.Exec(q).Error; err != nil {
			log.Printf("error fixing sequence for %s: %v", t, err)
		} else {
			log.Printf("sequence fixed for %s", t)
		}
	}

	log.Println("done")
}
