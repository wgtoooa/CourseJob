package domain

type WeekSchedule struct {
	Name        string
	GeneratedAt string
	Course      int
	Semester    int
	WeekNumber  int
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
	Day         string
	DayNumber   int
	Pair        int
	Duration    int
	Time        string
	Group       string
	Type        string
	Subject     string
	Teacher     *string
	Room        *string
	Subgroup    *string
	Frequency   *string
	PeriodStart *string
	PeriodEnd   *string
	Comment     *string
	Cancelled   bool
}
