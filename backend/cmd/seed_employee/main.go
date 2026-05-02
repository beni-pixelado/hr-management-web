package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Employee struct {
	FullName string
	Position string
	Email    string
}

func main() {
	db, err := sql.Open("sqlite3", "data/users.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	positions := []string{
		"Backend Developer",
		"Frontend Developer",
		"DevOps Engineer",
		"QA Engineer",
		"Product Manager",
	}

	var employees []Employee

	// gera 50 funcionários
	for i := 1; i <= 50; i++ {
		pos := positions[i%len(positions)]

		employees = append(employees, Employee{
			FullName: fmt.Sprintf("Employee %d", i),
			Email:    fmt.Sprintf("employee%d@company.com", i),
			Position: pos,
		})
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO employees (full_name, position, email)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, e := range employees {
		_, err := stmt.Exec(e.FullName, e.Position, e.Email)
		if err != nil {
			tx.Rollback()
			log.Fatal(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("50 employees added successfully!")
}