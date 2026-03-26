package dto

import "time"

type AttendanceScanRequest struct {
	CardUID   string    `json:"card_uid"`
	ScannedAt time.Time `json:"scanned_at" validate:"required"`
}

type AttendanceSessionRequest struct {
	Room       string                  `json:"room"`
	Source     string                  `json:"source"`
	StartedAt  time.Time               `json:"started_at" validate:"required"`
	FinishedAt time.Time               `json:"finished_at" validate:"required"`
	Scans      []AttendanceScanRequest `json:"scans" validate:"required"`
}
