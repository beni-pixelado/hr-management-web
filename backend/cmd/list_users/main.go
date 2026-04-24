package main

import (
	"fmt"
	"hr-management-web/backend/handlers"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("/workspaces/hr-management-web/data/users.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Erro ao conectar ao banco de dados:", err)
	}

	var users []handlers.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatal("Erro ao buscar usuários:", err)
	}

	fmt.Printf("Total de usuários: %d\n", len(users))
	fmt.Println("---")
	for i, u := range users {
		fmt.Printf("%d. %s (%s) - ID:%d\n", i+1, u.Username, u.Email, u.ID)
	}
}
