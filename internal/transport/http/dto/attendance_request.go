package dto

import "time"

type AttendanceScanRequest struct {
	CardUID   string    `json:"card_uid"`
	ScannedAt time.Time `json:"scanned_at"`
}

type AttendanceSessionRequest struct {
	Room       string                  `json:"room"`
	Source     string                  `json:"source"`
	StartedAt  time.Time               `json:"started_at"`
	FinishedAt time.Time               `json:"finished_at"`
	Scans      []AttendanceScanRequest `json:"scans"`
}
