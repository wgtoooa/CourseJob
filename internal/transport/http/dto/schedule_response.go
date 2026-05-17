package dto

type CourseScheduleResponse struct {
	Course      int                       `json:"course"`
	GeneratedAt string                    `json:"generated_at"`
	Groups      []ScheduleGroupSummaryDTO `json:"groups"`
	Weeks       []WeekScheduleResponse    `json:"weeks"`
}

type ScheduleGroupSummaryDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type WeekScheduleResponse struct {
	Name        string                    `json:"name"`
	GeneratedAt string                    `json:"generated_at"`
	Course      int                       `json:"course"`
	Semester    int                       `json:"semester"`
	WeekNumber  int                       `json:"week_number"`
	DateRange   string                    `json:"date_range"`
	Groups      []ScheduleGroupSummaryDTO `json:"groups"`
	Lessons     []ScheduleLessonResponse  `json:"lessons"`
}

type ScheduleLessonResponse struct {
	Day           string  `json:"day"`
	DayNumber     int     `json:"day_number"`
	Date          *string `json:"date"`
	Pair          int     `json:"pair"`
	Duration      int     `json:"duration"`
	Time          string  `json:"time"`
	Group         string  `json:"group"`
	Type          string  `json:"type"`
	Subject       string  `json:"subject"`
	Teacher       *string `json:"teacher"`
	Room          *string `json:"room"`
	Subgroup      *string `json:"subgroup"`
	Frequency     *string `json:"frequency"`
	PeriodStart   *string `json:"period_start"`
	PeriodEnd     *string `json:"period_end"`
	Comment       *string `json:"comment"`
	Cancelled     bool    `json:"cancelled"`
	WeekNumber    int     `json:"week_number"`
	GoogleSheetID *string `json:"google_sheet_id"`
}
