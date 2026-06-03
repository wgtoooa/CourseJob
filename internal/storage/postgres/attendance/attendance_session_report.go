package attendance

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type AttendanceReportRepository struct {
	db contract.DBTX
}

func NewAttendanceReportRepository(db contract.DBTX) *AttendanceReportRepository {
	return &AttendanceReportRepository{db: db}
}

func (repo *AttendanceReportRepository) BuildSessionReport(
	ctx context.Context,
	sessionID int64,
) error {
	if sessionID <= 0 {
		return errors.New("sessionID must be positive")
	}

	lessonGroupNames, err := repo.findLessonGroupNamesBySession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("resolve lesson groups for session %d: %w", sessionID, err)
	}
	groupKeys := normalizeGroupKeys(lessonGroupNames)
	if len(groupKeys) == 0 {
		return nil
	}

	if err = repo.upsertReportRowsByGroups(ctx, sessionID, groupKeys); err != nil {
		return fmt.Errorf("upsert report rows for session %d groups %v: %w", sessionID, lessonGroupNames, err)
	}

	if err = repo.deleteRowsOutsideGroups(ctx, sessionID, groupKeys); err != nil {
		return fmt.Errorf("cleanup report rows for session %d groups %v: %w", sessionID, lessonGroupNames, err)
	}

	return nil
}

func (repo *AttendanceReportRepository) findLessonGroupNamesBySession(
	ctx context.Context,
	sessionID int64,
) ([]string, error) {
const query = `
WITH session_data AS (
    SELECT
      room AS room_value,
      to_char(data, 'YYYY-MM-DD') AS lesson_date_text,
      date_trunc('minute', started_at AT TIME ZONE 'UTC')::time AS started_time_utc,
      date_trunc('minute', started_at AT TIME ZONE 'Europe/Minsk')::time AS started_time_minsk
    FROM attendance_session
    WHERE id = $1
)
,
matched_lessons AS (
    SELECT
      sl.week_id,
      sw.generated_at,
      trim(sl.group_name) AS group_name
    FROM session_data sd
    JOIN schedule_lesson sl
      ON COALESCE(sl.room, '') = COALESCE(sd.room_value, '')
     AND (
          NULLIF(split_part(regexp_replace(sl.lesson_time, '\s+', '', 'g'), '-', 1), '')::time = sd.started_time_utc
          OR
          NULLIF(split_part(regexp_replace(sl.lesson_time, '\s+', '', 'g'), '-', 1), '')::time = sd.started_time_minsk
     )
     AND NULLIF(trim(sl.lesson_date), '') = sd.lesson_date_text
    JOIN schedule_week sw
      ON sw.id = sl.week_id
    WHERE NULLIF(trim(sl.group_name), '') IS NOT NULL
),
target_week AS (
    SELECT week_id
    FROM matched_lessons
    ORDER BY generated_at DESC, week_id ASC
    LIMIT 1
)
SELECT DISTINCT ml.group_name
FROM matched_lessons ml
JOIN target_week tw ON tw.week_id = ml.week_id
ORDER BY ml.group_name ASC`

	rows, err := repo.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groupNames := make([]string, 0)
	for rows.Next() {
		var groupName string
		if err = rows.Scan(&groupName); err != nil {
			return nil, err
		}
		groupNames = append(groupNames, groupName)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groupNames, nil
}

func (repo *AttendanceReportRepository) upsertReportRowsByGroups(
	ctx context.Context,
	sessionID int64,
	groupKeys []string,
) error {
	const query = `
INSERT INTO attendance_report_row (
	session_id,
	student_id,
	present,
	scanned_at
)
SELECT
	$1 AS session_id,
	s.id AS student_id,
	(ae.id IS NOT NULL) AS present,
	ae.scanned_at
FROM student s
LEFT JOIN attendance_event ae
	ON ae.session_id = $1
   AND ae.student_id = s.id
WHERE lower(trim(s.group_name)) = ANY($2::text[])
ON CONFLICT (session_id, student_id) DO UPDATE
SET
	present = EXCLUDED.present,
	scanned_at = EXCLUDED.scanned_at`

	_, err := repo.db.Exec(ctx, query, sessionID, groupKeys)
	return err
}

func (repo *AttendanceReportRepository) deleteRowsOutsideGroups(
	ctx context.Context,
	sessionID int64,
	groupKeys []string,
) error {
	const query = `
DELETE FROM attendance_report_row arr
WHERE arr.session_id = $1
  AND NOT EXISTS (
      SELECT 1
      FROM student s
      WHERE s.id = arr.student_id
        AND lower(trim(s.group_name)) = ANY($2::text[])
  )`

	_, err := repo.db.Exec(ctx, query, sessionID, groupKeys)
	return err
}

func normalizeGroupKeys(groupNames []string) []string {
	seen := make(map[string]struct{}, len(groupNames))
	out := make([]string, 0, len(groupNames))
	for _, groupName := range groupNames {
		groupKey := strings.ToLower(strings.TrimSpace(groupName))
		if groupKey == "" {
			continue
		}
		if _, exists := seen[groupKey]; exists {
			continue
		}
		seen[groupKey] = struct{}{}
		out = append(out, groupKey)
	}

	sort.Strings(out)
	return out
}

func (repo *AttendanceReportRepository) ListStudentAttendance(
	ctx context.Context,
	studentID int64,
	fromDate *time.Time,
	toDate *time.Time,
) ([]domain.StudentAttendanceRecord, error) {
	if studentID <= 0 {
		return nil, errors.New("studentID must be positive")
	}

	const query = `
SELECT
	arr.session_id,
	ses.data,
	ses.room,
	ses.source,
	ses.started_at,
	ses.finished_at,
	arr.present,
	arr.scanned_at
FROM attendance_report_row arr
JOIN attendance_session ses ON ses.id = arr.session_id
WHERE arr.student_id = $1
  AND ($2::date IS NULL OR ses.data >= $2::date)
  AND ($3::date IS NULL OR ses.data <= $3::date)
ORDER BY ses.data DESC, ses.started_at DESC, arr.session_id DESC`

	rows, err := repo.db.Query(ctx, query, studentID, fromDate, toDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]domain.StudentAttendanceRecord, 0)
	for rows.Next() {
		var rec domain.StudentAttendanceRecord
		var scannedAt sql.NullTime
		if err = rows.Scan(
			&rec.SessionID,
			&rec.Date,
			&rec.Room,
			&rec.Source,
			&rec.StartedAt,
			&rec.FinishedAt,
			&rec.Present,
			&scannedAt,
		); err != nil {
			return nil, err
		}

		if scannedAt.Valid {
			t := scannedAt.Time
			rec.ScannedAt = &t
		}

		records = append(records, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}
