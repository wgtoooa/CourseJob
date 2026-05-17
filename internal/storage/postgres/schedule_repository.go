package postgres

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"sort"
	"strings"
	"time"
)

type ScheduleRepository struct {
	DB DBTX
}

func NewScheduleRepository(db DBTX) *ScheduleRepository {
	return &ScheduleRepository{
		DB: db,
	}
}

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
	if weekSchedule == nil {
		return errors.New("week schedule is nil")
	}

	generatedAt, err := time.Parse(time.RFC3339, weekSchedule.GeneratedAt)
	if err != nil {
		return fmt.Errorf("parse generated_at: %w", err)
	}

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
	if err = s.DB.QueryRow(
		ctx,
		upsertWeekQuery,
		weekSchedule.Name,
		generatedAt,
		weekSchedule.Course,
		weekSchedule.Semester,
		weekSchedule.WeekNumber,
		weekSchedule.DateRange,
	).Scan(&weekID); err != nil {
		return err
	}

	groupLookup := make(map[string]string, len(weekSchedule.Groups)*2)
	groupNameByID := make(map[string]string, len(weekSchedule.Groups))
	for _, group := range weekSchedule.Groups {
		groupCopy := group
		if err = s.UpSetGroup(ctx, &groupCopy); err != nil {
			return err
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

	teacherRepo := NewTeacherRepository(s.DB)
	subjectRepo := NewSubjectRepository(s.DB)
	roomRepo := NewRoomRepository(s.DB)
	seenTeachers := make(map[string]struct{})
	seenSubjects := make(map[string]struct{})
	seenRooms := make(map[string]struct{})
	for _, lesson := range weekSchedule.Lessons {
		subjectName := normalizeSubjectValue(lesson.Subject)
		if subjectName != "" {
			subjectKey := normalizeLookupKey(subjectName)
			if _, exists := seenSubjects[subjectKey]; !exists {
				if err = subjectRepo.UpSetSubject(ctx, &domain.Subject{Name: subjectName}); err != nil {
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
					if err = roomRepo.UpSetRoom(ctx, &domain.Room{Name: roomName}); err != nil {
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

		if err = teacherRepo.UpSetTeacher(ctx, &domain.Teacher{
			Post:     post,
			FullName: teacherFullName,
		}); err != nil {
			return err
		}
		seenTeachers[teacherKey] = struct{}{}
	}

	const deleteLessonsQuery = `DELETE FROM schedule_lesson WHERE week_id = $1`
	if _, err = s.DB.Exec(ctx, deleteLessonsQuery, weekID); err != nil {
		return err
	}

	if len(weekSchedule.Lessons) == 0 {
		return nil
	}

	lessonRows := buildLessonRows(weekID, weekSchedule.Lessons, groupLookup, groupNameByID)
	_, err = s.DB.CopyFrom(
		ctx,
		pgx.Identifier{"schedule_lesson"},
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
		pgx.CopyFromRows(lessonRows),
	)
	if err != nil {
		return err
	}

	return nil
}

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
	weekQuery += `
ORDER BY week_number ASC, generated_at DESC`

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
	lessonQuery += `
ORDER BY week_id ASC, day_number ASC, pair ASC, id ASC`

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

func nullableString(value *string) any {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return trimmed
}

func buildLessonRows(
	weekID int64,
	lessons []domain.ScheduleLesson,
	groupLookup map[string]string,
	groupNameByID map[string]string,
) [][]any {
	rows := make([][]any, 0, len(lessons))
	for _, lesson := range lessons {
		groupName := strings.TrimSpace(lesson.Group)
		groupID := any(nil)
		if resolvedGroupID, ok := groupLookup[normalizeLookupKey(groupName)]; ok {
			groupID = resolvedGroupID
			if canonicalGroupName, exists := groupNameByID[resolvedGroupID]; exists && strings.TrimSpace(canonicalGroupName) != "" {
				groupName = strings.TrimSpace(canonicalGroupName)
			}
		}

		rows = append(rows, []any{
			weekID,
			lesson.Day,
			lesson.DayNumber,
			nullableString(lesson.Date),
			lesson.Pair,
			lesson.Duration,
			lesson.Time,
			groupID,
			groupName,
			lesson.Type,
			lesson.Subject,
			nullableString(lesson.Teacher),
			nullableString(lesson.Room),
			nullableString(lesson.Subgroup),
			nullableString(lesson.Frequency),
			nullableString(lesson.PeriodStart),
			nullableString(lesson.PeriodEnd),
			nullableString(lesson.Comment),
			lesson.Cancelled,
			nullableString(lesson.GoogleSheetID),
		})
	}
	return rows
}

func normalizeLookupKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizedOptionalString(value *string) (string, bool) {
	if value == nil {
		return "", false
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
}

func splitPostFullName(raw string) (post string, fullName string) {
	for i := 0; i < len(raw); i++ {
		if raw[i] == '.' {
			return strings.TrimSpace(raw[:i]), strings.TrimSpace(raw[i+1:])
		}
	}
	return "", strings.TrimSpace(raw)
}

func normalizeSubjectValue(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func normalizeRoomValue(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
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
