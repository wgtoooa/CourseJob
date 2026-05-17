package domain

import "time"

type AttendanceEvent struct {
	ID        int64
	SessionID int64
	StudentID int64
	CardUID   string
	ScannedAt time.Time
	CreatedAt time.Time
}
