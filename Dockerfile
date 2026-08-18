# ---- Build stage ----
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# The server binary does not import the cgo SQLite driver, so a pure-Go,
# static build is safe and produces a small deployable artifact.
RUN CGO_ENABLED=0 go build -o server ./backend/cmd/server

# ---- Runtime stage ----
FROM alpine:3.20
WORKDIR /app

COPY --from=builder /app/server ./server
COPY backend/templates ./backend/templates
COPY frontend ./frontend
COPY robots.txt ./

ENV PORT=8000
EXPOSE 8000

CMD ["./server"]