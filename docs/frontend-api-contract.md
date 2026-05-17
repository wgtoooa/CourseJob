# CourseJob Frontend API Contract

## 1) Common Requirements

- Base URL is environment-specific (example: `http://localhost:8080`).
- Content type: `application/json; charset=utf-8`.
- All responses are JSON.
- Top-level response fields:
  - success: `"status": "ok"`, `"status": "created"`, or `"status": "success"`
  - error: `"status": "error"` and `"error": "<message>"`
- Datetime fields use RFC3339 (example: `2026-05-12T09:00:00Z`).

## 2) Full Endpoint List

- `GET /health/live`
- `GET /health/ready`
- `POST /api/v1/student`
- `POST /api/v1/attendance/sessions`
- `POST /api/v1/schedule`
- `GET /api/v1/teachers`
- `GET /api/v1/subjects`
- `GET /api/v1/rooms`

## 3) Health Endpoints

### GET `/health/live`

Purpose: liveness probe.

Success `200`:

```json
{
  "status": "ok"
}
```

### GET `/health/ready`

Purpose: readiness probe (checks in-memory `ready` flag and PostgreSQL reachability).

Success `200`:

```json
{
  "status": "ok",
  "db": "connected"
}
```

Unavailable `503`:

```json
{
  "status": "error",
  "error": "service is shutting down"
}
```

or

```json
{
  "status": "error",
  "error": "database is unavailable"
}
```

## 4) Students

### POST `/api/v1/students`

Body size limit: ~1 MB.

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

Validation rules:

- `full_name` is required.
- `course` must be from `1` to `4`.
- `group_name` is required.
- `email` is required and must be valid.
- `card_uid` format: only `[A-F0-9]`, length `4..7`.
- Before validation, server normalizes:
  - `card_uid` -> uppercase + trim
  - `group_name` -> trim
  - `email` -> lowercase + trim

Success `201`:

```json
{
  "status": "created"
}
```

Errors:

- `400` invalid request body / validation errors.
- `409` student with the same `card_uid` already exists.
- `500` internal error.

## 5) Attendance Sessions

### POST `/api/v1/attendance/sessions`

Body size limit: ~1 MB.

Request:

```json
{
  "room": "117",
  "source": "rfid-gate-1",
  "started_at": "2026-05-12T09:00:00Z",
  "finished_at": "2026-05-12T10:20:00Z",
  "scans": [
    {
      "card_uid": "A1B2C3",
      "scanned_at": "2026-05-12T09:10:00Z"
    }
  ]
}
```

Validation rules:

- `room` is required.
- `source` is required.
- `started_at` is required.
- `finished_at` is required.
- `finished_at >= started_at`.
- `scans` must not be empty.
- For each scan:
  - `card_uid` format `[A-F0-9]{4,7}`
  - `scanned_at` is required.
- Before validation, server normalizes:
  - `room` -> lowercase + trim
  - `source` -> lowercase + trim
  - `scans[].card_uid` -> uppercase + trim

Success `201`:

```json
{
  "status": "created",
  "data": {
    "session_id": 123,
    "saved_events": 1,
    "not_found_cards": []
  }
}
```

Errors:

- `400` invalid request body / validation errors.
- `500` internal error.

## 6) Schedule Import

### POST `/api/v1/schedule`

Body size limit: ~1 MB.

Request (trimmed shape):

```json
{
  "name": "11.05-16.05(14-я неделя)",
  "generated_at": "2026-05-16T08:00:00Z",
  "course": 3,
  "semester": 6,
  "week_number": 14,
  "groups": [
    {
      "id": "53",
      "name": "группа 53",
      "specialty": "Software Engineering",
      "department": "CS"
    }
  ],
  "lessons": [
    {
      "day": "Пн",
      "day_number": 1,
      "pair": 1,
      "duration": 1,
      "time": "09:00 - 10:25",
      "group": "1",
      "type": "lecture",
      "subject": "Интелектуальный анализ данных",
      "teacher": "доц. ЯцковН.Н.",
      "room": "115",
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

Success `200`:

```json
{
  "status": "success"
}
```

Errors:

- `400` invalid request body.
- `500` failed import.

## 7) Teachers

### GET `/api/v1/teachers`

Success `200`:

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

Notes:

- `PostFullName` is generated from `post + full_name`.
- The output is normalized with a single space between parts.

## 8) Frontend Notes

- Always send `Content-Type: application/json`.
- For `400`, show the message from `error` to the user/operator.
- For `409` on students, show clear duplicate `card_uid` message.

## 9) Subjects

### GET `/api/v1/subjects`

Success `200`:

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

## 10) Rooms

### GET `/api/v1/rooms`

Success `200`:

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
