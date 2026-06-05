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
		"Software Engineer",
	"Frontend Developer",
	"Backend Developer",
	"Full Stack Developer",
	"DevOps Engineer",
	"Cloud Engineer",
	"Data Scientist",
	"Data Analyst",
	"Machine Learning Engineer",
	"AI Engineer",
	"Cybersecurity Analyst",
	"Security Engineer",
	"Systems Administrator",
	"Network Engineer",
	"IT Support Specialist",
	"Hardware Engineer",
	"Embedded Systems Engineer",
	"Firmware Engineer",
	"Robotics Engineer",
	"QA Engineer",
	"Product Manager",
	"Technical Product Manager",
	"UI/UX Designer",
	"Interaction Designer",
	"Game Developer",
	"Game Designer",
	"AR/VR Developer",
	"Blockchain Developer",
	"Mobile App Developer",
	"iOS Developer",
	"Android Developer",
	"Site Reliability Engineer",
	"Marketing Manager",
	"Digital Marketing Specialist",
	"SEO Specialist",
	"Content Marketer",
	"Growth Hacker",
	"Social Media Manager",
	"Influencer Marketing Manager",
	"Brand Strategist",
	"Performance Marketing Specialist",
	"PPC Specialist",
	"Copywriter",
	"Content Strategist",
	"Email Marketing Specialist",
	"Affiliate Marketing Manager",
	"Media Buyer",
	"Creative Director",
	"Influencer",
	"Technical Writer",
	}

	var employees []Employee

	
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