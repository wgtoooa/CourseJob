# CourseJob

Backend service for attendance tracking:
- student registration
- attendance session ingest (scanner scans)
- schedule import
- teacher list endpoint
- subject list endpoint
- room list endpoint

## Stack

- Go
- PostgreSQL
- `pgx/v5`
- `chi/v5`
- `golang-migrate`

## Project Structure

```text
cmd/server                 server entry point
internal/config            environment configuration
internal/domain            domain models
internal/service           business logic
internal/storage/postgres  repositories and transaction manager
internal/transport/http    handlers, router, DTO, validation
migrations                 SQL migrations
docs                       additional API notes
```

## Environment Variables

Create `.env` in project root:

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
2. Create database `coursejob`.
3. Run server:

```bash
go run ./cmd/server
```

On startup service runs migrations automatically from `./migrations`.

## Migrations

Current migration files:
- `000001_init.*` - base tables (`student`, `teacher`, `attendance_session`, `attendance_event`)
- `000002_schedule.*` - schedule tables (`schedule_week`, `schedule_group`, `schedule_lesson`)

## HTTP API

Base URL example: `http://localhost:8080`

### Health

- `GET /health/live`
- `GET /health/ready`

`/health/ready` returns `503` when DB is unavailable or server is shutting down.

### Create Student

`POST /api/v1/students`

Request:

```json
{
  "full_name": "Ivan Ivanov",
  "course": 2,
  "group_name": "53",
  "email": "ivanov@example.com",
  "card_uid": "A1B2C3",
  "created_at": "2026-05-12T10:00:00Z"
}
```

Rules:
- `course` in range `1..4`
- `card_uid` format `[A-F0-9]{4,7}`
- `email` required and validated

Success response:

```json
{
  "status": "created"
}
```

### Create Attendance Session

`POST /api/v1/attendance/sessions`

Request:

```json
{
  "room": "A-101",
  "source": "scanner-1",
  "started_at": "2026-03-26T09:00:00Z",
  "finished_at": "2026-03-26T10:30:00Z",
  "scans": [
    {
      "card_uid": "04AA",
      "scanned_at": "2026-03-26T09:10:00Z"
    },
    {
      "card_uid": "04CCDD3",
      "scanned_at": "2026-03-26T09:12:00Z"
    }
  ]
}
```

Success response:

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

Notes:
- `finished_at` is required
- `scans` must not be empty
- entire write is transactional

### Import Schedule

`POST /api/v1/schedule`

Request shape:

```json
{
  "name": "3 course, week 14",
  "generated_at": "2026-05-16T08:00:00Z",
  "course": 3,
  "semester": 2,
  "week_number": 14,
  "groups": [
    {
      "id": "53",
      "name": "53",
      "specialty": "Software Engineering",
      "department": "CS"
    }
  ],
  "lessons": [
    {
      "day": "Monday",
      "day_number": 1,
      "pair": 1,
      "duration": 90,
      "time": "09:00",
      "group": "53",
      "type": "lecture",
      "subject": "Databases",
      "teacher": "Ivanov I.I.",
      "room": "A-201",
      "subgroup": null,
      "frequency": null,
      "period_start": null,
      "period_end": null,
      "comment": null,
      "cancelled": false
    }
  ]
}
```

Success response:

```json
{
  "status": "success"
}
```

### Get Teachers

`GET /api/v1/teachers`

Success response:

```json
{
  "status": "success",
  "teachers": [
    {
      "PostFullName": "ст.пр. БолотькоТ.П."
    }
  ]
}
```

### Get Subjects

`GET /api/v1/subjects`

Success response:

```json
{
  "status": "success",
  "subjects": [
    {
      "Name": "Базы данных"
    }
  ]
}
```

### Get Rooms

`GET /api/v1/rooms`

Success response:

```json
{
  "status": "success",
  "rooms": [
    {
      "Name": "A-201"
    }
  ]
}
```

## Notes

- Student card UID is stored in DB as HMAC-SHA256 hash string.
- `go test ./...` currently has no test files yet.
