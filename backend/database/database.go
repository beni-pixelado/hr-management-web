package database

import (
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := os.Getenv("DATABASE_URL")

	connConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Error parsing database config:", err)
	}

	connConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	sqlDB := stdlib.OpenDB(*connConfig)

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})

	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	DB = db
}
