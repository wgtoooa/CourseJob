package dto

type StudentAttendanceSummary struct {
	TotalSessions     int     `json:"total_sessions"`
	PresentCount      int     `json:"present_count"`
	AbsentCount       int     `json:"absent_count"`
	AttendancePercent float64 `json:"attendance_percent"`
}

type StudentAttendanceRecordResponse struct {
	SessionID  int64   `json:"session_id"`
	Date       string  `json:"date"`
	Room       string  `json:"room"`
	Source     string  `json:"source"`
	StartedAt  string  `json:"started_at"`
	FinishedAt string  `json:"finished_at"`
	Present    bool    `json:"present"`
	ScannedAt  *string `json:"scanned_at,omitempty"`
	Subject    *string `json:"subject,omitempty"`
	Group      *string `json:"group,omitempty"`
	Pair       *int    `json:"pair,omitempty"`
	LessonType *string `json:"lesson_type,omitempty"`
	Teacher    *string `json:"teacher,omitempty"`
}

type StudentAttendanceResponse struct {
	Status    string                            `json:"status"`
	StudentID int64                             `json:"student_id"`
	Summary   StudentAttendanceSummary          `json:"summary"`
	Records   []StudentAttendanceRecordResponse `json:"records"`
}
