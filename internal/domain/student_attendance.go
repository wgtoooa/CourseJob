package domain

import "time"

type StudentAttendanceRecord struct {
	SessionID  int64
	Date       time.Time
	Room       string
	Source     string
	StartedAt  time.Time
	FinishedAt time.Time
	Present    bool
	ScannedAt  *time.Time
}

type StudentAttendanceReport struct {
	StudentID         int64
	TotalSessions     int
	PresentCount      int
	AbsentCount       int
	AttendancePercent float64
	Records           []StudentAttendanceRecord
}
