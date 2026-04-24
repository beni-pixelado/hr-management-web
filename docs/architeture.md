# Architecture

## High-Level

Client → Gin Server → Handlers → SQLite

## Request Lifecycle

1. Client sends HTTP request
2. Gin routes request
3. Middleware executes
4. Handler processes logic
5. Data fetched/stored in DB
6. HTML rendered or redirect returned

## Design Decisions

### Server-side rendering
Simplifies architecture, no SPA needed.

### SQLite
No external DB required.

### Handler-based structure
Logic centralized per domain.