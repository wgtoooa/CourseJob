package validator

import (
	authA "CourseJob/internal/auth"
	"CourseJob/internal/transport/http/dto"
	"errors"
	"strings"
)

func ValidatorSession(session *dto.AttendanceSessionRequest) error {

	if session == nil {
		return errors.New("session is nil")
	}
	if session.Room == "" {
		return errors.New("room is empty")
	}
	if session.Source == "" {
		return errors.New("source is empty")
	}
	if session.FinishedAt.IsZero() {
		return errors.New("finished_at is zero")
	}
	if session.StartedAt.IsZero() {
		return errors.New("started_at is zero")
	}
	if session.Data.IsZero() {
		return errors.New("data is zero")
	}
	if session.FinishedAt.Before(session.StartedAt) {
		return errors.New("finished_at must be greater than or equal to started_at")
	}
	if session.Scans == nil {
		return errors.New("scans is empty")
	}
	if len(session.Scans) == 0 {
		return errors.New("scans is empty")
	}
	for _, scan := range session.Scans {
		if !validUID(scan.CardUID) {
			return errors.New("invalid card_uid format")
		}
		if scan.ScannedAt.IsZero() {
			return errors.New("scanned_at is zero")
		}
	}
	return nil
}

func NormalizeSessionRequest(req *dto.AttendanceSessionRequest) {
	req.Room = strings.TrimSpace(req.Room)
	req.Source = strings.ToLower(strings.TrimSpace(req.Source))

	for i := range req.Scans {
		req.Scans[i].CardUID = authA.NormalizeCardUID(req.Scans[i].CardUID)
	}
}
