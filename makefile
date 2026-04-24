.PHONY: test

run:
	go run ./backend/cmd/server

build:
	go build -o app backend/main.go

start:
	./app

setup:
	go mod tidy

test:
	go run ./backend/cmd/seed_users