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
	kill -9 $$(lsof -t -i:8000) || true

dev:
	@echo " Killing process on port 8000..."
	@kill -9 $$(lsof -t -i:8000) || true
	@echo " Starting Air..."
	air

k6-dashboard:
	k6 run internal/tests/k6/dashboard.js

k6-department:
	BASE_URL=http://localhost:8000 ACCOUNTS=20 ITERATIONS=20 SLEEP_MS=100 k6 run internal/tests/k6/department.js

k6-create:
	BASE_URL=http://localhost:8000 ACCOUNTS=20 EMPLOYEES_PER_ACCOUNT=10 SLEEP_MS=100 k6 run internal/tests/k6/create-employee.js

go-create:
	BASE_URL=http://localhost:8000 go test ./... -run TestCreateEmployees -v