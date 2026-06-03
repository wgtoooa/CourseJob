package schedule

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *ScheduleRepository) GetCourseSchedule(ctx context.Context, course int, filters domain.ScheduleFilters) (*domain.ScheduleCourse, error) {
	if course <= 0 {
		return nil, errors.New("course must be positive")
	}

	weekQuery := `
SELECT id, name, generated_at, course, semester, week_number, COALESCE(date_range, '')
FROM schedule_week
WHERE course = $1`
	weekArgs := []any{course}
	if filters.Week > 0 {
		weekArgs = append(weekArgs, filters.Week)
		weekQuery += fmt.Sprintf(" AND week_number = $%d", len(weekArgs))
	}
	weekRows, err := s.DB.Query(ctx, weekQuery, weekArgs...)
	if err != nil {
		return nil, err
	}
	defer weekRows.Close()

	out := &domain.ScheduleCourse{
		Course: course,
		Groups: []domain.ScheduleGroupSummary{},
		Weeks:  []domain.ScheduleWeekExport{},
	}

	var maxGeneratedAt time.Time
	hasGeneratedAt := false
	weekIDs := make([]int64, 0)
	weekIndexByID := make(map[int64]int)

	for weekRows.Next() {
		var (
			weekID      int64
			name        string
			generatedAt time.Time
			weekCourse  int
			semester    int
			weekNumber  int
			dateRange   string
		)

		if err = weekRows.Scan(&weekID, &name, &generatedAt, &weekCourse, &semester, &weekNumber, &dateRange); err != nil {
			return nil, err
		}

		if !hasGeneratedAt || generatedAt.After(maxGeneratedAt) {
			hasGeneratedAt = true
			maxGeneratedAt = generatedAt
		}

		weekIDs = append(weekIDs, weekID)
		weekIndexByID[weekID] = len(out.Weeks)
		out.Weeks = append(out.Weeks, domain.ScheduleWeekExport{
			Name:        name,
			GeneratedAt: generatedAt.UTC().Format(time.RFC3339),
			Course:      weekCourse,
			Semester:    semester,
			WeekNumber:  weekNumber,
			DateRange:   dateRange,
			Groups:      []domain.ScheduleGroupSummary{},
			Lessons:     []domain.ScheduleLesson{},
		})
	}
	if err = weekRows.Err(); err != nil {
		return nil, err
	}

	if hasGeneratedAt {
		out.GeneratedAt = maxGeneratedAt.UTC().Format(time.RFC3339)
	}
	if len(weekIDs) == 0 {
		return out, nil
	}

	groupCounts, err := s.getStudentGroupCounts(ctx, weekIDs)
	if err != nil {
		return nil, err
	}

	lessonQuery := `
SELECT
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
FROM schedule_lesson
WHERE week_id = ANY($1::bigint[])`

	lessonArgs := []any{weekIDs}
	if day := strings.TrimSpace(filters.Day); day != "" {
		lessonArgs = append(lessonArgs, strings.ToLower(day))
		lessonQuery += fmt.Sprintf(" AND lower(trim(day)) = $%d", len(lessonArgs))
	}
	if lessonType := strings.TrimSpace(filters.LessonType); lessonType != "" {
		lessonArgs = append(lessonArgs, strings.ToLower(lessonType))
		lessonQuery += fmt.Sprintf(" AND lower(trim(type)) = $%d", len(lessonArgs))
	}
	if group := strings.TrimSpace(filters.Group); group != "" {
		lessonArgs = append(lessonArgs, "%"+strings.ToLower(group)+"%")
		lessonQuery += fmt.Sprintf(
			" AND (lower(coalesce(group_id, '')) LIKE $%d OR lower(group_name) LIKE $%d)",
			len(lessonArgs), len(lessonArgs),
		)
	}
	if teacher := strings.TrimSpace(filters.Teacher); teacher != "" {
		lessonArgs = append(lessonArgs, "%"+strings.ToLower(teacher)+"%")
		lessonQuery += fmt.Sprintf(" AND lower(coalesce(teacher, '')) LIKE $%d", len(lessonArgs))
	}
	if subject := strings.TrimSpace(filters.Subject); subject != "" {
		lessonArgs = append(lessonArgs, "%"+strings.ToLower(subject)+"%")
		lessonQuery += fmt.Sprintf(" AND lower(subject) LIKE $%d", len(lessonArgs))
	}
	lessonRows, err := s.DB.Query(ctx, lessonQuery, lessonArgs...)
	if err != nil {
		return nil, err
	}
	defer lessonRows.Close()

	courseGroupByKey := make(map[string]domain.ScheduleGroupSummary)
	weekGroupByWeekID := make(map[int64]map[string]domain.ScheduleGroupSummary, len(weekIDs))

	for lessonRows.Next() {
		var (
			weekID        int64
			day           string
			dayNumber     int
			lessonDate    *string
			pair          int
			duration      int
			lessonTime    string
			groupID       *string
			groupName     string
			lessonType    string
			subject       string
			teacher       *string
			room          *string
			subgroup      *string
			frequency     *string
			periodStart   *string
			periodEnd     *string
			comment       *string
			cancelled     bool
			googleSheetID *string
		)

		if err = lessonRows.Scan(
			&weekID,
			&day,
			&dayNumber,
			&lessonDate,
			&pair,
			&duration,
			&lessonTime,
			&groupID,
			&groupName,
			&lessonType,
			&subject,
			&teacher,
			&room,
			&subgroup,
			&frequency,
			&periodStart,
			&periodEnd,
			&comment,
			&cancelled,
			&googleSheetID,
		); err != nil {
			return nil, err
		}

		weekIndex, exists := weekIndexByID[weekID]
		if !exists {
			continue
		}

		groupIDValue := strings.TrimSpace(derefString(groupID))
		groupNameValue := strings.TrimSpace(groupName)
		groupForLesson := groupNameValue
		if groupIDValue != "" {
			groupForLesson = groupIDValue
		}
		if groupForLesson == "" {
			groupForLesson = groupNameValue
		}

		out.Weeks[weekIndex].Lessons = append(out.Weeks[weekIndex].Lessons, domain.ScheduleLesson{
			Day:           day,
			DayNumber:     dayNumber,
			Date:          lessonDate,
			Pair:          pair,
			Duration:      duration,
			Time:          lessonTime,
			Group:         groupForLesson,
			Type:          lessonType,
			Subject:       subject,
			Teacher:       teacher,
			Room:          room,
			Subgroup:      subgroup,
			Frequency:     frequency,
			PeriodStart:   periodStart,
			PeriodEnd:     periodEnd,
			Comment:       comment,
			Cancelled:     cancelled,
			WeekNumber:    out.Weeks[weekIndex].WeekNumber,
			GoogleSheetID: googleSheetID,
		})

		groupSummary, hasGroup := buildGroupSummary(groupIDValue, groupNameValue, groupCounts)
		if !hasGroup {
			continue
		}

		groupKey := normalizeLookupKey(groupSummary.ID)
		if weekGroupByWeekID[weekID] == nil {
			weekGroupByWeekID[weekID] = make(map[string]domain.ScheduleGroupSummary)
		}
		weekGroupByWeekID[weekID][groupKey] = groupSummary
		courseGroupByKey[groupKey] = groupSummary
	}
	if err = lessonRows.Err(); err != nil {
		return nil, err
	}

	out.Groups = mapToSortedGroups(courseGroupByKey)
	for weekID, weekMap := range weekGroupByWeekID {
		weekIndex, exists := weekIndexByID[weekID]
		if !exists {
			continue
		}
		out.Weeks[weekIndex].Groups = mapToSortedGroups(weekMap)
	}
	out.Weeks = filterWeeksWithLessons(out.Weeks)
	if len(out.Weeks) == 0 {
		out.Groups = []domain.ScheduleGroupSummary{}
	}

	return out, nil
}

func (s *ScheduleRepository) getStudentGroupCounts(ctx context.Context, weekIDs []int64) (map[string]int, error) {
	if len(weekIDs) == 0 {
		return map[string]int{}, nil
	}

	const query = `
WITH relevant_groups AS (
  SELECT DISTINCT lower(trim(group_name)) AS group_key
  FROM schedule_lesson
  WHERE week_id = ANY($1::bigint[])
)
SELECT rg.group_key, count(st.id) AS students_count
FROM relevant_groups rg
LEFT JOIN student st ON lower(trim(st.group_name)) = rg.group_key
GROUP BY rg.group_key`

	rows, err := s.DB.Query(ctx, query, weekIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]int)
	for rows.Next() {
		var (
			groupKey string
			count    int
		)
		if err = rows.Scan(&groupKey, &count); err != nil {
			return nil, err
		}
		out[groupKey] = count
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func buildGroupSummary(groupID string, groupName string, studentGroupCounts map[string]int) (domain.ScheduleGroupSummary, bool) {
	id := strings.TrimSpace(groupID)
	name := strings.TrimSpace(groupName)
	if id == "" && name == "" {
		return domain.ScheduleGroupSummary{}, false
	}
	if id == "" {
		id = name
	}
	if name == "" {
		name = id
	}

	count := 0
	if idKey := normalizeLookupKey(id); idKey != "" {
		if value, exists := studentGroupCounts[idKey]; exists {
			count = value
		}
	}
	if count == 0 {
		if nameKey := normalizeLookupKey(name); nameKey != "" {
			if value, exists := studentGroupCounts[nameKey]; exists {
				count = value
			}
		}
	}

	return domain.ScheduleGroupSummary{
		ID:    id,
		Name:  name,
		Count: count,
	}, true
}

func mapToSortedGroups(groups map[string]domain.ScheduleGroupSummary) []domain.ScheduleGroupSummary {
	out := make([]domain.ScheduleGroupSummary, 0, len(groups))
	for _, group := range groups {
		out = append(out, group)
	}
	sort.Slice(out, func(i, j int) bool {
		leftName := strings.ToLower(strings.TrimSpace(out[i].Name))
		rightName := strings.ToLower(strings.TrimSpace(out[j].Name))
		if leftName == rightName {
			return strings.ToLower(strings.TrimSpace(out[i].ID)) < strings.ToLower(strings.TrimSpace(out[j].ID))
		}
		return leftName < rightName
	})

	return out
}

func filterWeeksWithLessons(weeks []domain.ScheduleWeekExport) []domain.ScheduleWeekExport {
	if len(weeks) == 0 {
		return weeks
	}

	filtered := make([]domain.ScheduleWeekExport, 0, len(weeks))
	for _, week := range weeks {
		if len(week.Lessons) == 0 {
			continue
		}
		filtered = append(filtered, week)
	}

	return filtered
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *ScheduleRepository) FindLessonBySession(ctx context.Context, sessionID int64) (*domain.ScheduleLesson, error) {
	if sessionID <= 0 {
		return nil, errors.New("sessionID must be positive")
	}

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
SELECT
  sl.day,
  sl.day_number,
  NULLIF(trim(sl.lesson_date), '') AS lesson_date,
  sl.pair,
  sl.duration,
  sl.lesson_time,
  COALESCE(NULLIF(trim(sl.group_id), ''), sl.group_name) AS lesson_group,
  sl.type,
  sl.subject,
  sl.teacher,
  sl.room,
  sl.subgroup,
  sl.frequency,
  sl.period_start,
  sl.period_end,
  sl.comment,
  sl.cancelled,
  sw.week_number,
  sl.google_sheet_id
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
ORDER BY
  sw.generated_at DESC,
  sl.id ASC
LIMIT 1`

	var (
		lessonDate    *string
		lessonTeacher *string
		lessonRoom    *string
		subgroup      *string
		frequency     *string
		periodStart   *string
		periodEnd     *string
		comment       *string
		googleSheetID *string
		lesson        domain.ScheduleLesson
	)

	err := s.DB.QueryRow(ctx, query, sessionID).Scan(
		&lesson.Day,
		&lesson.DayNumber,
		&lessonDate,
		&lesson.Pair,
		&lesson.Duration,
		&lesson.Time,
		&lesson.Group,
		&lesson.Type,
		&lesson.Subject,
		&lessonTeacher,
		&lessonRoom,
		&subgroup,
		&frequency,
		&periodStart,
		&periodEnd,
		&comment,
		&lesson.Cancelled,
		&lesson.WeekNumber,
		&googleSheetID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	lesson.Date = lessonDate
	lesson.Teacher = lessonTeacher
	lesson.Room = lessonRoom
	lesson.Subgroup = subgroup
	lesson.Frequency = frequency
	lesson.PeriodStart = periodStart
	lesson.PeriodEnd = periodEnd
	lesson.Comment = comment
	lesson.GoogleSheetID = googleSheetID

	return &lesson, nil
}
