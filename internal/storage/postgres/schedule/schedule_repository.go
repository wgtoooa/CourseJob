package schedule

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"strings"
)

type ScheduleRepository struct {
	DB contract.DBTX
}

func NewScheduleRepository(db contract.DBTX) *ScheduleRepository {
	return &ScheduleRepository{
		DB: db,
	}
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

type lessonRow struct {
	WeekID        int64
	Day           string
	DayNumber     int
	LessonDate    any
	Pair          int
	Duration      int
	LessonTime    string
	GroupID       any
	GroupName     string
	LessonType    string
	Subject       string
	Teacher       any
	Room          any
	Subgroup      any
	Frequency     any
	PeriodStart   any
	PeriodEnd     any
	Comment       any
	Cancelled     bool
	GoogleSheetID any
}

func buildLessonRows(
	weekID int64,
	lessons []domain.ScheduleLesson,
	groupLookup map[string]string,
	groupNameByID map[string]string,
) []lessonRow {
	rows := make([]lessonRow, 0, len(lessons))
	for _, lesson := range lessons {
		groupName := strings.TrimSpace(lesson.Group)
		groupID := any(nil)
		if resolvedGroupID, ok := groupLookup[normalizeLookupKey(groupName)]; ok {
			groupID = resolvedGroupID
			if canonicalGroupName, exists := groupNameByID[resolvedGroupID]; exists && strings.TrimSpace(canonicalGroupName) != "" {
				groupName = strings.TrimSpace(canonicalGroupName)
			}
		}

		rows = append(rows, lessonRow{
			WeekID:        weekID,
			Day:           lesson.Day,
			DayNumber:     lesson.DayNumber,
			LessonDate:    nullableString(lesson.Date),
			Pair:          lesson.Pair,
			Duration:      lesson.Duration,
			LessonTime:    lesson.Time,
			GroupID:       groupID,
			GroupName:     groupName,
			LessonType:    lesson.Type,
			Subject:       lesson.Subject,
			Teacher:       nullableString(lesson.Teacher),
			Room:          nullableString(lesson.Room),
			Subgroup:      nullableString(lesson.Subgroup),
			Frequency:     nullableString(lesson.Frequency),
			PeriodStart:   nullableString(lesson.PeriodStart),
			PeriodEnd:     nullableString(lesson.PeriodEnd),
			Comment:       nullableString(lesson.Comment),
			Cancelled:     lesson.Cancelled,
			GoogleSheetID: nullableString(lesson.GoogleSheetID),
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
