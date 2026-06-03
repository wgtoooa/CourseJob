CREATE TABLE student (
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL,
    course SMALLINT NOT NULL CHECK (course BETWEEN 1 AND 4),
    group_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    card_uid_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE teachers (
    id BIGSERIAL PRIMARY KEY,
    full_name TEXT NOT NULL UNIQUE,
    post TEXT NOT NULL
);

CREATE TABLE attendance_session (
    id BIGSERIAL PRIMARY KEY,
    room TEXT NOT NULL,
    source TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ NOT NULL,
    data DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (finished_at >= started_at)
);

CREATE TABLE attendance_event (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES attendance_session(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES student(id) ON DELETE RESTRICT,
    card_uid_hash TEXT NOT NULL,
    scanned_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE attendance_report_row (
   id BIGSERIAL PRIMARY KEY,
   session_id BIGINT NOT NULL REFERENCES attendance_session(id) ON DELETE CASCADE,
   student_id BIGINT NOT NULL REFERENCES student(id) ON DELETE RESTRICT,
   present BOOLEAN NOT NULL,
   scanned_at TIMESTAMPTZ,
   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),


   UNIQUE (session_id, student_id)
);
CREATE INDEX idx_student_course_group_name
    ON student(course, group_name);

CREATE INDEX idx_attendance_session_room
    ON attendance_session(room);

CREATE INDEX idx_attendance_session_source
    ON attendance_session(source);

CREATE INDEX idx_attendance_session_started_at
    ON attendance_session(started_at);

CREATE INDEX idx_attendance_event_session_id
    ON attendance_event(session_id);

CREATE INDEX idx_attendance_event_student_id
    ON attendance_event(student_id);

CREATE INDEX idx_attendance_event_scanned_at
    ON attendance_event(scanned_at);

CREATE UNIQUE INDEX uq_attendance_session_room_time_source_data
    ON attendance_session(room, started_at, finished_at, source, data);

CREATE UNIQUE INDEX uq_attendance_event_session_student
    ON attendance_event(session_id, student_id);

