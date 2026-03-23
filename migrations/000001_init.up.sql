CREATE TABLE IF NOT EXISTS student(
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    course SMALLINT NOT NULL CHECK(course > 0 AND course < 5),
    group_name TEXT NOT NULL,
    card_uid TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ not null default NOW()
    );

CREATE TABLE IF NOT EXISTS attendance_session (
    id BIGSERIAL PRIMARY KEY,
    room TEXT NOT NULL,
    source TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (finished_at IS NULL OR finished_at >= started_at)
    );


CREATE TABLE IF NOT EXISTS attendance_event (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES attendance_session(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES student(id) ON DELETE RESTRICT,
    card_uid TEXT NOT NULL,
    scanned_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_student_group_name
    ON student(group_name);

CREATE INDEX IF NOT EXISTS idx_student_course
    ON student(course);

CREATE INDEX IF NOT EXISTS idx_student_card_uid
    ON student(card_uid);

CREATE INDEX IF NOT EXISTS idx_attendance_session_room
    ON attendance_session (room);

CREATE INDEX IF NOT EXISTS idx_attendance_session_source
    ON attendance_session (source);

CREATE INDEX IF NOT EXISTS idx_attendance_session_started_at
    ON attendance_session (started_at);

CREATE INDEX IF NOT EXISTS idx_attendance_event_session_id
    ON attendance_events(session_id);

CREATE INDEX IF NOT EXISTS idx_attendance_event_student_id
    ON attendance_events(student_id);

CREATE INDEX IF NOT EXISTS idx_attendance_event_scanned_at
    ON attendance_events(scanned_at);