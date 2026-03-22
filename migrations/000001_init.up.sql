CREATE TABLE IF NOT EXISTS students(
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    course SMALLINT NOT NULL CHECK(course > 0 AND course < 5),
    group_name TEXT NOT NULL,
    card_uid TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ not null default NOW()
    );

CREATE TABLE IF NOT EXISTS attendance_sessions (
    id BIGSERIAL PRIMARY KEY,
    room TEXT NOT NULL,
    source TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (finished_at IS NULL OR finished_at >= started_at)
    );


CREATE TABLE IF NOT EXISTS attendance_events (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES attendance_sessions(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    card_uid TEXT NOT NULL,
    scanned_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_students_group_name
    ON students(group_name);

CREATE INDEX IF NOT EXISTS idx_students_course
    ON students(course);

CREATE INDEX IF NOT EXISTS idx_students_card_uid
    ON students(card_uid);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_room
    ON attendance_sessions(room);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_source
    ON attendance_sessions(source);

CREATE INDEX IF NOT EXISTS idx_attendance_sessions_started_at
    ON attendance_sessions(started_at);

CREATE INDEX IF NOT EXISTS idx_attendance_events_session_id
    ON attendance_events(session_id);

CREATE INDEX IF NOT EXISTS idx_attendance_events_student_id
    ON attendance_events(student_id);

CREATE INDEX IF NOT EXISTS idx_attendance_events_scanned_at
    ON attendance_events(scanned_at);