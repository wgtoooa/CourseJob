package dto

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AttendanceScanRequest struct {
	CardUID   string    `json:"card_uid"`
	ScannedAt time.Time `json:"scanned_at" validate:"required"`
}

type AttendanceSessionRequest struct {
	Room       string                  `json:"room"`
	Source     string                  `json:"source"`
	StartedAt  time.Time               `json:"started_at" validate:"required"`
	FinishedAt time.Time               `json:"finished_at" validate:"required"`
	Data       time.Time               `json:"data" validate:"required"`
	Scans      []AttendanceScanRequest `json:"scans" validate:"required"`
}

func (r *AttendanceSessionRequest) UnmarshalJSON(payload []byte) error {
	type rawAttendanceSessionRequest struct {
		Room       string                  `json:"room"`
		Source     string                  `json:"source"`
		StartedAt  time.Time               `json:"started_at"`
		FinishedAt time.Time               `json:"finished_at"`
		Data       string                  `json:"data"`
		Scans      []AttendanceScanRequest `json:"scans"`
	}

	var raw rawAttendanceSessionRequest
	if err := json.Unmarshal(payload, &raw); err != nil {
		return err
	}

	parsedDate, err := parseDateOnly(raw.Data)
	if err != nil {
		return fmt.Errorf("invalid data format: %w", err)
	}

	r.Room = raw.Room
	r.Source = raw.Source
	r.StartedAt = raw.StartedAt
	r.FinishedAt = raw.FinishedAt
	r.Data = parsedDate
	r.Scans = raw.Scans
	return nil
}

func parseDateOnly(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, nil
	}

	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("expected YYYY-MM-DD")
	}

	return parsed, nil
}
