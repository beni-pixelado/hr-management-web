# System Overview

HR Management Web is a server-rendered, monolithic web application for managing HR candidate pipelines.

## Stack

- Go (Golang)
- Gin (HTTP framework)
- SQLite (database)
- HTML Templates (server-side rendering)

## Architecture Pattern

The system follows an MVC-adjacent pattern:

- Model: SQLite via Go database/sql
- View: HTML templates (`backend/templates/`)
- Controller: Handlers (`backend/handlers/`)

## Key Characteristics

- No external dependencies
- Single binary deployment
- Ideal for small teams and internal tools