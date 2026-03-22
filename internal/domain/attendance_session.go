package domain

import "time"

type AttendanceSession struct {
	ID         int64
	Room       string
	Source     string
	StartedAt  time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time
}
