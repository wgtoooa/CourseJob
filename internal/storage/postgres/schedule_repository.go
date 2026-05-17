package postgres

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"fmt"
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
INSERT INTO schedule_week (name, generated_at, course, semester, week_number, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (generated_at, course, semester, week_number)
DO UPDATE SET
  name = EXCLUDED.name,
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
	).Scan(&weekID); err != nil {
		return err
	}

	groupByName := make(map[string]string, len(weekSchedule.Groups))
	for _, group := range weekSchedule.Groups {
		groupCopy := group
		if err = s.UpSetGroup(ctx, &groupCopy); err != nil {
			return err
		}
		if normalizedName := normalizeLookupKey(group.Name); normalizedName != "" {
			groupByName[normalizedName] = group.ID
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

		post, teacherFullName := splitTeacherByFields(teacherRaw)
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

	const insertLessonQuery = `
INSERT INTO schedule_lesson (
  week_id, day, day_number, pair, duration, lesson_time, group_id, group_name, type, subject,
  teacher, room, subgroup, frequency, period_start, period_end, comment, cancelled
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
  $11, $12, $13, $14, $15, $16, $17, $18
)`

	for _, lesson := range weekSchedule.Lessons {
		groupName := strings.TrimSpace(lesson.Group)
		groupID := any(nil)
		if resolvedGroupID, ok := groupByName[normalizeLookupKey(groupName)]; ok {
			groupID = resolvedGroupID
		}

		if _, err = s.DB.Exec(
			ctx,
			insertLessonQuery,
			weekID,
			lesson.Day,
			lesson.DayNumber,
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
		); err != nil {
			return err
		}
	}

	return nil
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

func splitTeacherByFields(raw string) (post string, fullName string) {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return "", parts[0]
	}

	return parts[0], strings.Join(parts[1:], " ")
}

func normalizeSubjectValue(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func normalizeRoomValue(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
