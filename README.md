# CourseJob Backend (MVP)

Бэкенд-сервис для:
- импорта и чтения расписания;
- учета посещаемости по сканам карт;
- хранения плановых пар по предметам;
- выдачи справочников (преподаватели, предметы, аудитории).

## Что уже реализовано

- `POST /api/v1/student` — создание одного или нескольких студентов.
- `POST /api/v1/attendance/sessions` — прием сессии посещаемости со сканами.
- `POST /api/v1/schedule` — импорт чанка расписания.
- `GET /api/v1/schedule` — выдача расписания по курсу с фильтрами.
- `GET /api/v1/sse/schedule` — SSE-стрим обновлений расписания.
- `GET /api/v1/plan` и `PUT /api/v1/plan` — чтение/апсерт плана.
- `GET /api/v1/teachers`, `GET /api/v1/subjects`, `GET /api/v1/rooms` — справочники.
- `GET /health/live`, `GET /health/ready` — health checks.

## Технологии

- Go `1.25`
- PostgreSQL
- Router: `github.com/go-chi/chi/v5`
- DB driver: `github.com/jackc/pgx/v5`
- Миграции: `github.com/golang-migrate/migrate/v4`

## Структура проекта

```text
cmd/server                 Точка входа
internal/config            Загрузка конфигурации
internal/domain            Доменные модели
internal/service           Бизнес-логика
internal/storage/postgres  Репозитории + транзакции
internal/transport/http    Роутер, хендлеры, DTO, валидации
migrations                 SQL-миграции
docs                       Дополнительные контракты/примеры
```

## Конфигурация

Приложение пытается читать `.env` автоматически.

### Переменные окружения

- `HTTP_ADDR` — адрес HTTP-сервера (по умолчанию `:8080`).
- `DATABASE_URL` — приоритетный способ подключения к БД.

Если `DATABASE_URL` не задан, нужны:
- `DATABASE_USER`
- `DATABASE_PASSWORD`
- `DATABASE_HOST`
- `DATABASE_PORT`
- `DATABASE_NAME`

Пример `.env`:

```env
HTTP_ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/coursejob?sslmode=disable
```

## Запуск локально

1. Поднять PostgreSQL.
2. Создать БД `coursejob`.
3. Запустить:

```bash
go run ./cmd/server
```

Сервис стартует на `HTTP_ADDR` (по умолчанию `:8080`).

## Docker

Сборка:

```bash
docker build -t coursejob:local .
```

Запуск:

```bash
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:postgres@host.docker.internal:5432/coursejob?sslmode=disable" \
  coursejob:local
```

## Миграции

Миграции запускаются автоматически при старте сервера.

- `000001_attendance.*`:
  - `student`
  - `teachers`
  - `attendance_session`
  - `attendance_event`
  - `attendance_report_row`
- `000002_schedule.*`:
  - `schedule_week`
  - `schedule_group`
  - `schedule_lesson`
  - `subject_catalog`
  - `room_catalog`
  - `plan_item`

### Если миграции "залипли" (dirty)

Проверка:

```sql
SELECT * FROM schema_migrations;
```

Ручная фиксация (если уверены в состоянии схемы):

```sql
UPDATE schema_migrations SET version = 2, dirty = FALSE;
```

## CORS

В проекте есть middleware `CORS()`, но в роутере оно сейчас отключено:

```go
// r.Use(CORS())
```

Если фронт работает с другого origin, включите middleware в `internal/transport/http/router.go`.

## Общие правила API

- Формат: `application/json`.
- Все ответы — JSON.
- Для POST/PUT с телом в текущем MVP лимит тела: около `1 MB`.
- Времена — RFC3339 (`2026-05-12T09:00:00Z`).
- Поле даты в посещаемости `data`: формат `YYYY-MM-DD`.
- Во многих error-ответах используется `{"status":"error","error":"..."}`.
- В ряде endpoint’ов success-форматы отличаются (это текущая реализация MVP).

## Полный список endpoint’ов

- `GET /health/live`
- `GET /health/ready`
- `POST /api/v1/student`
- `POST /api/v1/attendance/sessions`
- `POST /api/v1/schedule`
- `GET /api/v1/schedule?course=N`
- `GET /api/v1/sse/schedule`
- `GET /api/v1/plan?course=N`
- `PUT /api/v1/plan`
- `GET /api/v1/teachers`
- `GET /api/v1/subjects`
- `GET /api/v1/rooms`

---

## 1) Health

### `GET /health/live`

Проверка liveness.

`200 OK`:

```json
{
  "status": "ok"
}
```

### `GET /health/ready`

Проверка readiness: флаг остановки + `db.Ping`.

`200 OK`:

```json
{
  "status": "ok",
  "db": "connected"
}
```

`503 Service Unavailable`:

```json
{
  "status": "error",
  "error": "service is shutting down"
}
```

или

```json
{
  "status": "error",
  "error": "database is unavailable"
}
```

---

## 2) Students

### `POST /api/v1/student`

Создание студента. Поддерживает:
- один объект;
- массив объектов.

Пример одного объекта:

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

Пример массива:

```json
[
  {
    "full_name": "Ivan Ivanov",
    "course": 2,
    "group_name": "53",
    "email": "ivanov@example.com",
    "card_uid": "A1B2C3"
  },
  {
    "full_name": "Petr Petrov",
    "course": 2,
    "group_name": "53",
    "email": "petrov@example.com",
    "card_uid": "A1B2C4"
  }
]
```

Валидация:
- `course` в диапазоне `1..4`;
- `card_uid` по regex `^[A-F0-9]{4,7}$`;
- `email` валидный email;
- `full_name` и `group_name` обязательны.

Нормализация перед валидацией:
- `card_uid` -> `UPPER + trim`;
- `email` -> `lower + trim`;
- `group_name`/`full_name` -> `trim`.

Успех:
- `201`:

```json
{
  "status": "created"
}
```

или для массива:

```json
{
  "status": "created",
  "created_count": 2
}
```

Ошибки:
- `400` — некорректный JSON/валидация;
- `409` — дубликат `card_uid` или `email`;
- `500` — внутренняя ошибка.

---

## 3) Attendance Sessions

### `POST /api/v1/attendance/sessions`

Сохранение сессии и сканов.

Пример:

```json
{
  "room": "117",
  "source": "rfid-gate-1",
  "data": "2026-05-12",
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

Валидация:
- `room` обязателен;
- `source` обязателен;
- `data` обязателен, формат `YYYY-MM-DD`;
- `started_at` и `finished_at` обязательны;
- `finished_at >= started_at`;
- `scans` не пустой;
- у каждого скана:
  - `card_uid` в формате `[A-F0-9]{4,7}`;
  - `scanned_at` обязателен.

Нормализация:
- `room` -> `trim`;
- `source` -> `lower + trim`;
- `scans[].card_uid` -> `UPPER + trim`.

Успех `201`:

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

Ошибки:
- `400` — некорректный JSON/валидация;
- `500` — внутренняя ошибка.

Поведение:
- сессии дедуплицируются по `(room, source, started_at, finished_at, data)`;
- события дедуплицируются по `(session_id, student_id)`;
- после обработки строится `attendance_report_row`.

---

## 4) Schedule Import

### `POST /api/v1/schedule`

Импорт чанка расписания.

Пример:

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

Минимально обязательно:
- `course > 0`
- `week_number > 0`
- `lessons` не пустой

Успех `200`:

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

Ошибки:
- `400` — невалидный payload;
- `500` — ошибка импорта.

### Важное поведение импорта (MVP)

Импорт сейчас делает **merge**, а не полную замену недели:
- существующие пары обновляются;
- новые пары добавляются;
- пары, которых нет в текущем чанке, автоматически не удаляются.

Идентификация пары для update/insert идет по:
- `week_id`
- `day_number`
- `pair`
- `group_name` (нормализованно)
- `subgroup`
- `frequency`
- `lesson_date`

Если в одном payload есть дубликаты одной и той же пары, "побеждает" последняя запись.

---

## 5) Schedule Read

### `GET /api/v1/schedule?course=N`

Возвращает расписание по курсу.

Обязательный query:
- `course` (`1..4`)

Опциональные фильтры:
- `week`
- `group` (substring, case-insensitive)
- `day` (exact, case-insensitive)
- `type` (exact, case-insensitive)
- `teacher` (substring, case-insensitive)
- `subject` (substring, case-insensitive)

Пример:

```text
GET /api/v1/schedule?course=3&week=14&group=601&day=Пн&type=lab&teacher=Яцков&subject=БИС
```

Успех `200`:

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

Ошибки:
- `400` — некорректный `course`/filters;
- `500` — ошибка чтения.

---

## 6) Schedule SSE

### `GET /api/v1/sse/schedule`

SSE-канал обновлений расписания.

Заголовки:
- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`

При подключении:
- отправляется служебный комментарий `: connected`;
- heartbeat каждые 15 секунд (`: ping`).

Когда приходит новый импорт:
- событие `schedule_updated`;
- `data` содержит JSON:

```json
{
  "type": "schedule_updated",
  "updated_at": "2026-05-24T08:05:00Z",
  "chunk": {
    "name": "14 week",
    "generated_at": "2026-05-24T08:00:00Z",
    "course": 3,
    "semester": 6,
    "week_number": 14,
    "date_range": "12.05.2026 - 18.05.2026",
    "groups": [],
    "lessons": []
  }
}
```

---

## 7) Plan

### `GET /api/v1/plan?course=N`

Обязательный query:
- `course` (`1..4`)

Успех `200`:

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

Ошибки:
- `400` — некорректный `course`;
- `500` — ошибка чтения.

### `PUT /api/v1/plan`

Пример:

```json
{
  "course": 3,
  "subject": "Mathematical Analysis",
  "planned_pairs": 24
}
```

Поведение:
- upsert по `(course, normalized subject_key)`;
- `subject` нормализуется (`trim + collapse spaces + lower key`);
- `planned_pairs >= 0`.

Успех `200`:

```json
{
  "status": "success"
}
```

Ошибки:
- `400` — валидация;
- `500` — ошибка записи.

---

## 8) Catalogs

### `GET /api/v1/teachers`

Успех `200`:

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

### `GET /api/v1/subjects`

Успех `200`:

```json
{
  "status": "success",
  "subjects": [
    {
      "Name": "Databases"
    }
  ]
}
```

### `GET /api/v1/rooms`

Успех `200`:

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

---

## Примеры быстрых проверок (curl)

```bash
# health
curl -s http://localhost:8080/health/live
curl -s http://localhost:8080/health/ready

# schedule read
curl -s "http://localhost:8080/api/v1/schedule?course=3"

# plan
curl -s "http://localhost:8080/api/v1/plan?course=3"
curl -s -X PUT "http://localhost:8080/api/v1/plan" \
  -H "Content-Type: application/json" \
  -d '{"course":3,"subject":"Math","planned_pairs":24}'
```

## Проверки качества

```bash
go test ./...
go vet ./...
go build ./...
```

## Важные заметки MVP

- `card_uid` студента хранится в БД в виде HMAC-хеша, не в открытом виде.
- Форматы success/error-ответов пока не полностью унифицированы между endpoint’ами.
- `POST /api/v1/student` — актуальный путь в коде; путь `/api/v1/students` в старых документах считать устаревшим.
- Для фронта на другом домене нужно включить `CORS()` middleware.

## Где смотреть доп. материалы

- Примеры payload/ответов: `docs/api-endpoints-examples.json`
- Фронтовый контракт: `docs/frontend-api-contract.md`
