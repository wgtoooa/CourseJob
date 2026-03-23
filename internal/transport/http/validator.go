package http

import (
	"CourseJob/internal/transport/http/dto"
	"errors"
	"strings"
)

func ValidatorSession(session *dto.AttendanceSessionRequest) error {
	if session == nil {
		return errors.New("session is nil")
	}
	if session.Room == "" || strings.TrimSpace(session.Room) == "" {
		return errors.New("room is empty")
	}
	if session.Source == "" || strings.TrimSpace(session.Source) == "" {
		return errors.New("source is empty")
	}
	if session.FinishedAt.IsZero() {
		return errors.New("finished_at is zero")
	}
	if session.StartedAt.IsZero() {
		return errors.New("started_at is zero")
	}
	if session.Scans == nil {
		return errors.New("scans is empty")
	}
	return nil
}
