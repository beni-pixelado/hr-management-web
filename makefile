.PHONY: test

run:
	go run ./backend/cmd/server

build:
	go build -o app backend/main.go

start:
	./app

setup:
	go mod tidy

test-user:
	go run ./backend/cmd/seed_users

test-employee:
	go run backend/cmd/seed_employee/main.go

kill:
	kill 2351 || true

dev:
	@echo " Killing process on port 8000..."
	@kill -9 $$(lsof -t -i:8000) || true
	@echo " Starting Air..."
	air