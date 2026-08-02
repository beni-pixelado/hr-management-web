package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	println("ENTROU NO CONNECT")

	dsn := os.Getenv("DATABASE_URL")

	println("DATABASE:", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}

	DB = db
}
