package domain

type WeekSchedule struct {
	Name        string
	GeneratedAt string
	Course      int
	Semester    int
	WeekNumber  int
	DateRange   string
	Groups      []ScheduleGroup
	Lessons     []ScheduleLesson
}

type ScheduleGroup struct {
	ID         string
	Name       string
	Specialty  string
	Department string
}

type ScheduleLesson struct {
	Day           string
	DayNumber     int
	Date          *string
	Pair          int
	Duration      int
	Time          string
	Group         string
	Type          string
	Subject       string
	Teacher       *string
	Room          *string
	Subgroup      *string
	Frequency     *string
	PeriodStart   *string
	PeriodEnd     *string
	Comment       *string
	Cancelled     bool
	WeekNumber    int
	GoogleSheetID *string
}

type ScheduleCourse struct {
	Course      int
	GeneratedAt string
	Groups      []ScheduleGroupSummary
	Weeks       []ScheduleWeekExport
}

type ScheduleFilters struct {
	Week       int
	Group      string
	Day        string
	LessonType string
	Teacher    string
	Subject    string
}

type ScheduleGroupSummary struct {
	ID    string
	Name  string
	Count int
}

type ScheduleWeekExport struct {
	Name        string
	GeneratedAt string
	Course      int
	Semester    int
	WeekNumber  int
	DateRange   string
	Groups      []ScheduleGroupSummary
	Lessons     []ScheduleLesson
}
