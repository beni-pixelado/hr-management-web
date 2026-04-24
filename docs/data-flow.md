# Data Flow

## Candidate Creation

POST /employees

- Validate input
- Upload photo
- Save to DB
- Default status: pending

## Status Update

POST /employees/:id/status

- Validate status
- Update database
- Redirect

## Dashboard

GET /dashboard

- Count employees by status
- Render template