package http

import (
	"CourseJob/internal/transport/http/dto"
	"errors"
	"regexp"
	"strings"
)

func ValidatorSession(session *dto.AttendanceSessionRequest) error {
	var cardUIDRegex = regexp.MustCompile(`^[A-F0-9]{4,7}$`)
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
	if session.Scans == nil {
		return errors.New("scans is empty")
	}
	if len(session.Scans) == 0 {
		return errors.New("scans is empty")
	}
	for _, scan := range session.Scans {
		if !cardUIDRegex.MatchString(scan.CardUID) {
			return errors.New("invalid card_uid format")
		}
	}
	return nil
}

func NormalizeSessionRequest(req *dto.AttendanceSessionRequest) {
	req.Room = strings.ToLower(strings.TrimSpace(req.Room))
	req.Source = strings.ToLower(strings.TrimSpace(req.Source))

	for i := range req.Scans {
		req.Scans[i].CardUID = strings.ToUpper(strings.TrimSpace(req.Scans[i].CardUID))
	}
}
