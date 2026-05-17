CREATE TABLE IF NOT EXISTS schedule_week (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL,
    course SMALLINT NOT NULL,
    semester SMALLINT NOT NULL,
    week_number SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (generated_at, course, semester, week_number)
);

CREATE TABLE IF NOT EXISTS schedule_group (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    specialty TEXT NOT NULL,
    department TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subject_catalog (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS room_catalog (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS schedule_lesson (
    id BIGSERIAL PRIMARY KEY,
    week_id BIGINT NOT NULL REFERENCES schedule_week(id) ON DELETE CASCADE,
    day TEXT NOT NULL,
    day_number SMALLINT NOT NULL,
    pair SMALLINT NOT NULL,
    duration INT NOT NULL,
    lesson_time TEXT NOT NULL,
    group_id TEXT REFERENCES schedule_group(id) ON DELETE SET NULL,
    group_name TEXT NOT NULL,
    type TEXT NOT NULL,
    subject TEXT NOT NULL,
    teacher TEXT,
    room TEXT,
    subgroup TEXT,
    frequency TEXT,
    period_start TEXT,
    period_end TEXT,
    comment TEXT,
    cancelled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Compatibility for databases where schedule_lesson existed without week_id.
ALTER TABLE schedule_lesson
    ADD COLUMN IF NOT EXISTS week_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_schedule_week_generated_at ON schedule_week (generated_at);
CREATE INDEX IF NOT EXISTS idx_schedule_lesson_week_id ON schedule_lesson (week_id);
CREATE INDEX IF NOT EXISTS idx_schedule_lesson_day_pair ON schedule_lesson (day_number, pair);
