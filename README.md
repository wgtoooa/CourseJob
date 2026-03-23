# Course Project Roadmap

_Last updated: 2026-03-22 11:19_

## Current State
We already have:
- Database migrations
- Domain models
- DTO (transport layer)
- Project structure

---

## Next Steps
### ~~0. env file~~ +
### 1. Service Layer
Core business logic:
- ~~Accept batch request~~
- ~~Create attendance session~~
- ~~Iterate over scans~~
- ~~Find student by card_uid~~
- ~~Save attendance_event~~
- ~~Collect not_found_cards~~

---

### 2. HTTP Handler
Endpoint:
POST /api/v1/attendance/sessions

Responsibilities:
- ~~Parse JSON (DTO)~~
- ~~Call service~~
- ~~Return response~~

---

### 3. Wiring (main/router)
- ~~Initialize DB~~
- ~~Initialize repositories~~
- ~~Initialize services~~
- ~~Initialize handlers~~
- ~~Register routes~~

---

### 4. Manual Testing
Use:
- curl
- Postman

Test cases:
- All cards valid
- Some cards not found
- Empty scans
- Invalid JSON

---

### 5. Seed Data
Add test students with:
- card_uid
- names, course, group

---

### 6. Error Handling
- 400 → bad request / invalid JSON
- 500 → DB errors
- validation (empty fields)

---

### 7. Transactions
Wrap session + events creation in transaction

---

### 8. First Read Endpoint
Example:
- Get events by session
- Get attendance by student
- Get attendance by room/date

---

### 9. Future Improvements
- Authentication
- Roles
- Reports
- Student import
- Multi-device support
