# CourseJob Backend

Backend service for attendance and schedule data.

## What this service does

- Registers students (`POST /api/v1/student`)
- Saves attendance sessions and scans (`POST /api/v1/attendance/sessions`)
- Imports schedule from parser (`POST /api/v1/schedule`)
- Returns schedule by course with server-side filters (`GET /api/v1/schedule`)
- Stores and returns plan values by course+subject (`GET/PUT /api/v1/plan`)
- Returns catalogs:
  - teachers (`GET /api/v1/teachers`)
  - subjects (`GET /api/v1/subjects`)
  - rooms (`GET /api/v1/rooms`)

## Stack

- Go 1.25
- PostgreSQL
- `github.com/jackc/pgx/v5`
- `github.com/go-chi/chi/v5`
- `github.com/golang-migrate/migrate/v4`

## Project structure

```text
cmd/server                 server entry point
internal/config            env config loader
internal/domain            domain models
internal/service           business logic
internal/storage/postgres  repositories + tx manager
internal/transport/http    handlers/router/dto/validators
migrations                 SQL migrations
docs                       API contract notes
```

## Configuration

The app loads `.env` automatically (if present).

You can configure DB in two ways:

1. `DATABASE_URL` (preferred)
2. Split vars: `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_NAME`

Example `.env`:

```env
HTTP_ADDR=:8080

DATABASE_URL=postgres://postgres:postgres@localhost:5432/coursejob?sslmode=disable

# If DATABASE_URL is empty, these are required:
# DATABASE_USER=postgres
# DATABASE_PASSWORD=postgres
# DATABASE_HOST=localhost
# DATABASE_PORT=5432
# DATABASE_NAME=coursejob
```

## Run locally

1. Start PostgreSQL.
2. Create database `coursejob`.
3. Run:

```bash
go run ./cmd/server
```

Server starts on `HTTP_ADDR` (default `:8080`).

## Docker

Build:

```bash
docker build -t coursejob:local .
```

Run:

```bash
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:postgres@host.docker.internal:5432/coursejob?sslmode=disable" \
  coursejob:local
```

## Migrations

Migrations are executed automatically on startup from `./migrations`.

- `000001_attendance.*`:
  - `student`
  - `teachers`
  - `attendance_session`
  - `attendance_event`
- `000002_schedule.*`:
  - `schedule_week`
  - `schedule_group`
  - `schedule_lesson`
  - `subject_catalog`
  - `room_catalog`
  - `plan_item`

### Migration troubleshooting

If startup fails with dirty/version errors, fix `schema_migrations` manually.

Typical cases:

- `Dirty database version ...`
- `schema_migrations points to version 0 ...`

For a clean local DB, easiest path is recreate DB and start again.

If you need manual repair in existing DB:

```sql
SELECT * FROM schema_migrations;
```

Then set a valid version and clean state:

```sql
UPDATE schema_migrations SET version = 2, dirty = FALSE;
```

Use version that matches your real schema state.

## CORS

Allowed origins:

- `http://localhost:5173`
- `http://localhost:4173`
- `https://yaroslavka123.github.io`
- `https://rfict.up.railway.app`

Allowed methods: `GET, POST, PUT, OPTIONS`  
Allowed headers: `Content-Type, Accept`

## API

Base URL example: `http://localhost:8080`

### Health

#### GET `/health/live`

Response `200`:

```json
{
  "status": "ok"
}
```

#### GET `/health/ready`

Response `200`:

```json
{
  "status": "ok",
  "db": "connected"
}
```

Can return `503` if DB is unavailable or service is shutting down.

### Students

#### POST `/api/v1/student`

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

Validation:

- `course`: `1..4`
- `card_uid`: regex `^[A-F0-9]{4,7}$`
- `email`: valid email

Normalization before validation:

- `card_uid`: upper + trim
- `group_name`: trim
- `email`: lower + trim

Success `201`:

```json
{
  "status": "created"
}
```

Duplicate card UID: `409`

### Attendance sessions

#### POST `/api/v1/attendance/sessions`

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
    }
  ]
}
```

Success `201`:

```json
{
  "status": "created",
  "data": {
    "session_id": 1,
    "saved_events": 1,
    "not_found_cards": []
  }
}
```

### Schedule import

#### POST `/api/v1/schedule`

Request (trimmed):

```json
{
  "name": "14 week",
  "generated_at": "2026-05-17T19:00:00Z",
  "course": 3,
  "semester": 6,
  "week_number": 14,
  "date_range": "12.05.2026 - 18.05.2026",
  "groups": [
    {
      "id": "g-1",
      "name": "IKBO-01-23",
      "specialty": "SE",
      "department": "CS"
    }
  ],
  "lessons": [
    {
      "day": "Mon",
      "day_number": 1,
      "date": "2026-05-12",
      "pair": 1,
      "duration": 2,
      "time": "09:00-10:30",
      "group": "g-1",
      "type": "lecture",
      "subject": "Math",
      "teacher": "Ivanov I.I.",
      "room": "A-101",
      "subgroup": null,
      "frequency": null,
      "period_start": null,
      "period_end": null,
      "comment": null,
      "cancelled": false,
      "google_sheet_id": "1abc...xyz"
    }
  ]
}
```

Success `200`:

```json
{
  "ok": true,
  "status": "success",
  "course": 3,
  "week_number": 14,
  "lessons_count": 120,
  "updated_at": "2026-05-17T19:00:00Z"
}
```

### Schedule read

#### GET `/api/v1/schedule`

Required query:

- `course` (`1..4`)

Optional filters (applied on backend):

- `week` (`1..14`)
- `group` (substring, case-insensitive)
- `day` (exact match, case-insensitive)
- `type` (exact match, case-insensitive)
- `teacher` (substring, case-insensitive)
- `subject` (substring, case-insensitive)

Example:

```text
GET /api/v1/schedule?course=3&week=14&group=601&day=Пн&type=lab&teacher=Яцков&subject=БИС
```

Success `200` (shape):

```json
{
  "course": 3,
  "generated_at": "2026-05-17T19:00:00Z",
  "groups": [
    { "id": "g-1", "name": "IKBO-01-23", "count": 25 }
  ],
  "weeks": [
    {
      "name": "14 week",
      "generated_at": "2026-05-17T19:00:00Z",
      "course": 3,
      "semester": 6,
      "week_number": 14,
      "date_range": "12.05.2026 - 18.05.2026",
      "groups": [
        { "id": "g-1", "name": "IKBO-01-23", "count": 25 }
      ],
      "lessons": [
        {
          "day": "Пн",
          "day_number": 1,
          "date": "2026-05-12",
          "pair": 1,
          "duration": 2,
          "time": "09:00-10:30",
          "group": "g-1",
          "type": "lecture",
          "subject": "Math",
          "teacher": "Ivanov I.I.",
          "room": "A-101",
          "subgroup": null,
          "frequency": null,
          "period_start": null,
          "period_end": null,
          "comment": null,
          "cancelled": false,
          "week_number": 14,
          "google_sheet_id": "1abc...xyz"
        }
      ]
    }
  ]
}
```

### Plan

#### GET `/api/v1/plan?course=3`

Response `200`:

```json
[
  {
    "course": 3,
    "subject": "Mathematical Analysis",
    "planned_pairs": 24
  },
  {
    "course": 3,
    "subject": "Algebra",
    "planned_pairs": 18
  }
]
```

#### PUT `/api/v1/plan`

Request:

```json
{
  "course": 3,
  "subject": "Mathematical Analysis",
  "planned_pairs": 24
}
```

Behavior:

- Upsert by `(course, normalized subject key)`
- Subject normalization: trim + lower
- `planned_pairs` must be non-negative

Success `200`:

```json
{
  "status": "success"
}
```

### Catalogs

#### GET `/api/v1/teachers`

Response `200`:

```json
{
  "status": "success",
  "teachers": [
    {
      "PostFullName": "доц. Иванов И.И."
    }
  ]
}
```

#### GET `/api/v1/subjects`

Response `200`:

```json
{
  "status": "success",
  "subjects": [
    {
      "ID": 1,
      "Name": "Databases"
    }
  ]
}
```

#### GET `/api/v1/rooms`

Response `200`:

```json
{
  "status": "success",
  "rooms": [
    {
      "ID": 1,
      "Name": "A-201"
    }
  ]
}
```

## Dev checks

```bash
go test ./...
go vet ./...
go build ./...
```

## Notes

- Student `card_uid` is stored hashed (HMAC-SHA256), not in plain text.
- Schedule import logs two timings:
  - delay from `generated_at` to import completion
  - total backend import time
- Detailed frontend-focused contract is in:
  - `docs/frontend-api-contract.md`
