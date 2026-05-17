CREATE TABLE schedule_week (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL,
    course SMALLINT NOT NULL,
    semester SMALLINT NOT NULL,
    week_number SMALLINT NOT NULL,
    date_range TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course, semester, week_number)
);

CREATE TABLE schedule_group (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    specialty TEXT NOT NULL,
    department TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE subject_catalog (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE room_catalog (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE plan_item (
    id BIGSERIAL PRIMARY KEY,
    course SMALLINT NOT NULL CHECK (course BETWEEN 1 AND 4),
    subject TEXT NOT NULL,
    subject_key TEXT NOT NULL,
    planned_pairs INT NOT NULL CHECK (planned_pairs >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course, subject_key)
);

CREATE TABLE schedule_lesson (
    id BIGSERIAL PRIMARY KEY,
    week_id BIGINT NOT NULL REFERENCES schedule_week(id) ON DELETE CASCADE,
    day TEXT NOT NULL,
    day_number SMALLINT NOT NULL,
    lesson_date TEXT,
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
    google_sheet_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schedule_week_generated_at ON schedule_week(generated_at);
CREATE INDEX idx_schedule_week_course_week_number ON schedule_week(course, week_number);
CREATE INDEX idx_schedule_lesson_week_id ON schedule_lesson(week_id);
CREATE INDEX idx_schedule_lesson_day_pair ON schedule_lesson(day_number, pair);
CREATE INDEX idx_schedule_lesson_group_name ON schedule_lesson(group_name);
CREATE INDEX idx_schedule_lesson_subject ON schedule_lesson(subject);
CREATE INDEX idx_schedule_lesson_teacher ON schedule_lesson(teacher);
CREATE INDEX idx_schedule_lesson_room ON schedule_lesson(room);
CREATE INDEX idx_plan_item_course ON plan_item(course);
