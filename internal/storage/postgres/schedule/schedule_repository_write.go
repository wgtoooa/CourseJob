package schedule

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *ScheduleRepository) UpSetLessons(ctx context.Context, lessons *domain.ScheduleLesson) error {
	_ = ctx
	if lessons == nil {
		return errors.New("lesson is nil")
	}

	return errors.New("standalone lesson upsert is not supported; use UpSetWeekSchedule")
}

func (s *ScheduleRepository) UpSetGroup(ctx context.Context, group *domain.ScheduleGroup) error {
	if group == nil {
		return errors.New("group is nil")
	}

	const query = `
INSERT INTO schedule_group (id, name, specialty, department, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  specialty = EXCLUDED.specialty,
  department = EXCLUDED.department,
  updated_at = NOW()`

	_, err := s.DB.Exec(
		ctx,
		query,
		group.ID,
		group.Name,
		group.Specialty,
		group.Department,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *ScheduleRepository) UpSetWeekSchedule(ctx context.Context, weekSchedule *domain.WeekSchedule) error {
	return s.syncWeekSchedule(ctx, weekSchedule)
}

func (s *ScheduleRepository) syncWeekSchedule(
	ctx context.Context,
	weekSchedule *domain.WeekSchedule,
) error {
	if weekSchedule == nil {
		return errors.New("week schedule is nil")
	}

	generatedAt, err := parseGeneratedAt(weekSchedule.GeneratedAt)
	if err != nil {
		return err
	}

	weekID, err := s.upsertWeek(ctx, weekSchedule, generatedAt)
	if err != nil {
		return err
	}

	groupLookup, groupNameByID, err := s.upsertGroupsAndBuildLookup(ctx, weekSchedule.Groups)
	if err != nil {
		return err
	}

	if err = s.upsertLessonDictionaries(ctx, weekSchedule.Lessons); err != nil {
		return err
	}

	if err = s.replaceWeekLessonsFull(ctx, weekID, weekSchedule.Lessons, groupLookup, groupNameByID); err != nil {
		return err
	}

	return nil
}

func parseGeneratedAt(value string) (time.Time, error) {
	generatedAt, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse generated_at: %w", err)
	}

	return generatedAt, nil
}

func (s *ScheduleRepository) upsertWeek(
	ctx context.Context,
	weekSchedule *domain.WeekSchedule,
	generatedAt time.Time,
) (int64, error) {
	const upsertWeekQuery = `
INSERT INTO schedule_week (name, generated_at, course, semester, week_number, date_range, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (course, semester, week_number)
DO UPDATE SET
  name = EXCLUDED.name,
  generated_at = EXCLUDED.generated_at,
  date_range = EXCLUDED.date_range,
  updated_at = NOW()
RETURNING id`

	var weekID int64
	if err := s.DB.QueryRow(
		ctx,
		upsertWeekQuery,
		weekSchedule.Name,
		generatedAt,
		weekSchedule.Course,
		weekSchedule.Semester,
		weekSchedule.WeekNumber,
		weekSchedule.DateRange,
	).Scan(&weekID); err != nil {
		return 0, err
	}

	return weekID, nil
}

func (s *ScheduleRepository) upsertGroupsAndBuildLookup(
	ctx context.Context,
	groups []domain.ScheduleGroup,
) (map[string]string, map[string]string, error) {
	groupLookup := make(map[string]string, len(groups)*2)
	groupNameByID := make(map[string]string, len(groups))
	for _, group := range groups {
		groupCopy := group
		if err := s.UpSetGroup(ctx, &groupCopy); err != nil {
			return nil, nil, err
		}
		if normalizedName := normalizeLookupKey(group.Name); normalizedName != "" {
			groupLookup[normalizedName] = group.ID
		}
		if normalizedID := normalizeLookupKey(group.ID); normalizedID != "" {
			groupLookup[normalizedID] = group.ID
		}
		if group.ID != "" {
			groupNameByID[group.ID] = group.Name
		}
	}

	return groupLookup, groupNameByID, nil
}

func (s *ScheduleRepository) upsertLessonDictionaries(ctx context.Context, lessons []domain.ScheduleLesson) error {
	teacherRepo := NewTeacherRepository(s.DB)
	subjectRepo := NewSubjectRepository(s.DB)
	roomRepo := NewRoomRepository(s.DB)
	seenTeachers := make(map[string]struct{})
	seenSubjects := make(map[string]struct{})
	seenRooms := make(map[string]struct{})
	for _, lesson := range lessons {
		subjectName := normalizeSubjectValue(lesson.Subject)
		if subjectName != "" {
			subjectKey := normalizeLookupKey(subjectName)
			if _, exists := seenSubjects[subjectKey]; !exists {
				if err := subjectRepo.UpSetSubject(ctx, &domain.Subject{Name: subjectName}); err != nil {
					return err
				}
				seenSubjects[subjectKey] = struct{}{}
			}
		}

		if roomRaw, ok := normalizedOptionalString(lesson.Room); ok {
			roomName := normalizeRoomValue(roomRaw)
			if roomName != "" {
				roomKey := normalizeLookupKey(roomName)
				if _, exists := seenRooms[roomKey]; !exists {
					if err := roomRepo.UpSetRoom(ctx, &domain.Room{Name: roomName}); err != nil {
						return err
					}
					seenRooms[roomKey] = struct{}{}
				}
			}
		}

		teacherRaw, ok := normalizedOptionalString(lesson.Teacher)
		if !ok {
			continue
		}

		post, teacherFullName := splitPostFullName(teacherRaw)
		if teacherFullName == "" {
			continue
		}

		teacherKey := normalizeLookupKey(post + " " + teacherFullName)
		if _, exists := seenTeachers[teacherKey]; exists {
			continue
		}

		if err := teacherRepo.UpSetTeacher(ctx, &domain.Teacher{
			Post:     post,
			FullName: teacherFullName,
		}); err != nil {
			return err
		}
		seenTeachers[teacherKey] = struct{}{}
	}

	return nil
}

func (s *ScheduleRepository) replaceWeekLessonsFull(
	ctx context.Context,
	weekID int64,
	lessons []domain.ScheduleLesson,
	groupLookup map[string]string,
	groupNameByID map[string]string,
) error {
	if len(lessons) == 0 {
		return nil
	}

	lessonRows := buildLessonRows(weekID, lessons, groupLookup, groupNameByID)
	lessonRows = deduplicateLessonRows(lessonRows)
	if err := s.mergeWeekLessonsBulk(ctx, lessonRows); err != nil {
		return fmt.Errorf("bulk merge lessons: %w", err)
	}

	return nil
}

func (s *ScheduleRepository) mergeWeekLessonsBulk(ctx context.Context, lessonRows []lessonRow) error {
	if len(lessonRows) == 0 {
		return nil
	}

	const createTempTableQuery = `
CREATE TEMP TABLE IF NOT EXISTS schedule_lesson_import_tmp (
	week_id BIGINT NOT NULL,
	day TEXT NOT NULL,
	day_number SMALLINT NOT NULL,
	lesson_date TEXT,
	pair SMALLINT NOT NULL,
	duration INTEGER NOT NULL,
	lesson_time TEXT NOT NULL,
	group_id TEXT,
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
	cancelled BOOLEAN NOT NULL,
	google_sheet_id TEXT
) ON COMMIT DROP`

	if _, err := s.DB.Exec(ctx, createTempTableQuery); err != nil {
		return err
	}
	if _, err := s.DB.Exec(ctx, `TRUNCATE schedule_lesson_import_tmp`); err != nil {
		return err
	}

	copyRows := make([][]any, 0, len(lessonRows))
	for i := range lessonRows {
		dayNumber, err := toInt16(lessonRows[i].DayNumber, "day_number")
		if err != nil {
			return fmt.Errorf("lesson[%d]: %w", i, err)
		}
		pair, err := toInt16(lessonRows[i].Pair, "pair")
		if err != nil {
			return fmt.Errorf("lesson[%d]: %w", i, err)
		}
		duration, err := toInt32(lessonRows[i].Duration, "duration")
		if err != nil {
			return fmt.Errorf("lesson[%d]: %w", i, err)
		}

		copyRows = append(copyRows, []any{
			lessonRows[i].WeekID,
			lessonRows[i].Day,
			dayNumber,
			lessonRows[i].LessonDate,
			pair,
			duration,
			lessonRows[i].LessonTime,
			lessonRows[i].GroupID,
			lessonRows[i].GroupName,
			lessonRows[i].LessonType,
			lessonRows[i].Subject,
			lessonRows[i].Teacher,
			lessonRows[i].Room,
			lessonRows[i].Subgroup,
			lessonRows[i].Frequency,
			lessonRows[i].PeriodStart,
			lessonRows[i].PeriodEnd,
			lessonRows[i].Comment,
			lessonRows[i].Cancelled,
			lessonRows[i].GoogleSheetID,
		})
	}

	if _, err := s.DB.CopyFrom(
		ctx,
		pgx.Identifier{"schedule_lesson_import_tmp"},
		[]string{
			"week_id",
			"day",
			"day_number",
			"lesson_date",
			"pair",
			"duration",
			"lesson_time",
			"group_id",
			"group_name",
			"type",
			"subject",
			"teacher",
			"room",
			"subgroup",
			"frequency",
			"period_start",
			"period_end",
			"comment",
			"cancelled",
			"google_sheet_id",
		},
		pgx.CopyFromRows(copyRows),
	); err != nil {
		return err
	}

	const updateQuery = `
UPDATE schedule_lesson AS dst
SET
	day = src.day,
	day_number = src.day_number,
	lesson_date = src.lesson_date,
	pair = src.pair,
	duration = src.duration,
	lesson_time = src.lesson_time,
	group_id = src.group_id,
	group_name = src.group_name,
	type = src.type,
	subject = src.subject,
	teacher = src.teacher,
	room = src.room,
	subgroup = src.subgroup,
	frequency = src.frequency,
	period_start = src.period_start,
	period_end = src.period_end,
	comment = src.comment,
	cancelled = src.cancelled,
	google_sheet_id = src.google_sheet_id
FROM schedule_lesson_import_tmp AS src
WHERE dst.week_id = src.week_id
  AND dst.day_number = src.day_number
  AND dst.pair = src.pair
  AND lower(trim(dst.group_name)) = lower(trim(src.group_name))
  AND dst.subgroup IS NOT DISTINCT FROM src.subgroup
  AND dst.frequency IS NOT DISTINCT FROM src.frequency
  AND dst.lesson_date IS NOT DISTINCT FROM src.lesson_date`

	if _, err := s.DB.Exec(ctx, updateQuery); err != nil {
		return err
	}

	const insertQuery = `
INSERT INTO schedule_lesson (
	week_id,
	day,
	day_number,
	lesson_date,
	pair,
	duration,
	lesson_time,
	group_id,
	group_name,
	type,
	subject,
	teacher,
	room,
	subgroup,
	frequency,
	period_start,
	period_end,
	comment,
	cancelled,
	google_sheet_id
)
SELECT
	src.week_id,
	src.day,
	src.day_number,
	src.lesson_date,
	src.pair,
	src.duration,
	src.lesson_time,
	src.group_id,
	src.group_name,
	src.type,
	src.subject,
	src.teacher,
	src.room,
	src.subgroup,
	src.frequency,
	src.period_start,
	src.period_end,
	src.comment,
	src.cancelled,
	src.google_sheet_id
FROM schedule_lesson_import_tmp AS src
WHERE NOT EXISTS (
	SELECT 1
	FROM schedule_lesson AS dst
	WHERE dst.week_id = src.week_id
	  AND dst.day_number = src.day_number
	  AND dst.pair = src.pair
	  AND lower(trim(dst.group_name)) = lower(trim(src.group_name))
	  AND dst.subgroup IS NOT DISTINCT FROM src.subgroup
	  AND dst.frequency IS NOT DISTINCT FROM src.frequency
	  AND dst.lesson_date IS NOT DISTINCT FROM src.lesson_date
)`

	if _, err := s.DB.Exec(ctx, insertQuery); err != nil {
		return err
	}

	return nil
}

func deduplicateLessonRows(rows []lessonRow) []lessonRow {
	if len(rows) <= 1 {
		return rows
	}

	out := make([]lessonRow, 0, len(rows))
	indexByKey := make(map[string]int, len(rows))
	for _, row := range rows {
		key := lessonIdentityKey(row)
		if idx, exists := indexByKey[key]; exists {
			// Last duplicate in payload wins to match typical "latest update" behavior.
			out[idx] = row
			continue
		}

		indexByKey[key] = len(out)
		out = append(out, row)
	}

	return out
}

func lessonIdentityKey(row lessonRow) string {
	return fmt.Sprintf(
		"%d|%d|%d|%s|%s|%s|%s",
		row.WeekID,
		row.DayNumber,
		row.Pair,
		normalizeLookupKey(row.GroupName),
		normalizeOptionalIdentityPart(row.Subgroup),
		normalizeOptionalIdentityPart(row.Frequency),
		normalizeOptionalIdentityPart(row.LessonDate),
	)
}

func normalizeOptionalIdentityPart(value any) string {
	if value == nil {
		return ""
	}
	str := strings.TrimSpace(fmt.Sprint(value))
	return strings.ToLower(str)
}

func toInt16(value int, field string) (int16, error) {
	if value < math.MinInt16 || value > math.MaxInt16 {
		return 0, fmt.Errorf("%s is out of range for int16", field)
	}
	return int16(value), nil
}

func toInt32(value int, field string) (int32, error) {
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, fmt.Errorf("%s is out of range for int32", field)
	}
	return int32(value), nil
}
