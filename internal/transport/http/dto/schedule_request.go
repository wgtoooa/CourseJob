package dto

type WeekScheduleRequest struct {
	Name        string                  `json:"name"`
	GeneratedAt string                  `json:"generated_at"`
	Course      int                     `json:"course"`
	Semester    int                     `json:"semester"`
	WeekNumber  int                     `json:"week_number"`
	Groups      []ScheduleGroupRequest  `json:"groups"`
	Lessons     []ScheduleLessonRequest `json:"lessons"`
}

type ScheduleGroupRequest struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Specialty  string `json:"specialty"`
	Department string `json:"department"`
}

type ScheduleLessonRequest struct {
	Day         string  `json:"day"`
	DayNumber   int     `json:"day_number"`
	Pair        int     `json:"pair"`
	Duration    int     `json:"duration"`
	Time        string  `json:"time"`
	Group       string  `json:"group"`
	Type        string  `json:"type"`
	Subject     string  `json:"subject"`
	Teacher     *string `json:"teacher"`
	Room        *string `json:"room"`
	Subgroup    *string `json:"subgroup"`
	Frequency   *string `json:"frequency"`
	PeriodStart *string `json:"period_start"`
	PeriodEnd   *string `json:"period_end"`
	Comment     *string `json:"comment"`
	Cancelled   bool    `json:"cancelled"`
}
