package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	return gorm.Open(
		sqlite.Open("data/users.db"),
		&gorm.Config{},
	)
}