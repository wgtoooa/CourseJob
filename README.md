# CourseJob

Backend service for registering attendance sessions and student card scans.

## Stack

- Go
- PostgreSQL
- `pgx`
- `chi`

## Project Structure

```text
cmd/server                 application entry point
internal/config            environment configuration
internal/domain            domain models
internal/service           business logic
internal/storage/postgres  postgres repositories and transaction manager
internal/transport/http    HTTP handlers, router, DTO, validation
migrations                 SQL migrations
```

## Features

- create attendance session
- save attendance scan events
- search student by card UID
- return list of unknown cards
- transactional write of session and events
- health check endpoints

## Environment Variables

Create `.env` file in the project root:

```env
HTTP_ADDR=:8080

DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=coursejob
```

## Run

1. Start PostgreSQL.
2. Create database.
3. Apply migrations from [migrations/000001_init.up.sql](D:/Projects/CourseJob/migrations/000001_init.up.sql).
4. Run the server:

```bash
go run ./cmd/server
```

## HTTP API

### Health Check

```http
GET /ping
GET /db/ping
```

### Create Attendance Session

```http
POST /api/v1/attendance/sessions
Content-Type: application/json
```

Request body:

```json
{
  "room": "A-101",
  "source": "scanner-1",
  "started_at": "2026-03-26T09:00:00Z",
  "finished_at": "2026-03-26T10:30:00Z",
  "scans": [
    {
      "card_uid": "04AABB1122",
      "scanned_at": "2026-03-26T09:10:00Z"
    },
    {
      "card_uid": "04CCDD3344",
      "scanned_at": "2026-03-26T09:12:00Z"
    }
  ]
}
```

Successful response:

```json
{
  "status": "created",
  "data": {
    "session_id": 1,
    "saved_events": 2,
    "not_found_cards": []
  }
}
```

## Notes

- `finished_at` is required.
- request is rejected if `scans` is empty.
- all writes are performed inside one transaction.

## Next Improvements

- add tests
- add structured error handling
- add read endpoints
- add logging/middleware
- add seed script or fixtures
