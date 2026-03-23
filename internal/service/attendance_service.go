package service

import (
	"context"
	"time"

	"CourseJob/internal/domain"
)

type StudentGetter interface {
	GetByCardUID(ctx context.Context, cardUID string) (*domain.Student, error)
}

type AttendanceSessionCreator interface {
	Create(ctx context.Context, session *domain.AttendanceSession) error
}

type AttendanceEventCreator interface {
	Create(ctx context.Context, event *domain.AttendanceEvent) error
}

type AttendanceService struct {
	studentRepo StudentGetter
	sessionRepo AttendanceSessionCreator
	eventRepo   AttendanceEventCreator
}

func NewAttendanceService(
	studentRepo StudentGetter,
	sessionRepo AttendanceSessionCreator,
	eventRepo AttendanceEventCreator,
) *AttendanceService {
	return &AttendanceService{
		studentRepo: studentRepo,
		sessionRepo: sessionRepo,
		eventRepo:   eventRepo,
	}
}

type ProcessAttendanceInput struct {
	Room       string
	Source     string
	StartedAt  time.Time
	FinishedAt *time.Time
	Scans      []ProcessAttendanceScanInput
}

type ProcessAttendanceScanInput struct {
	CardUID   string
	ScannedAt time.Time
}

type ProcessAttendanceResult struct {
	SessionID     int64
	SavedEvents   int
	NotFoundCards []string
}

func (s *AttendanceService) ProcessAttendance(
	ctx context.Context,
	input ProcessAttendanceInput,
) (*ProcessAttendanceResult, error) {
	session := &domain.AttendanceSession{
		Room:       input.Room,
		Source:     input.Source,
		StartedAt:  input.StartedAt,
		FinishedAt: input.FinishedAt,
	}
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	result := &ProcessAttendanceResult{
		SessionID:     session.ID,
		SavedEvents:   0,
		NotFoundCards: []string{},
	}

	for _, scan := range input.Scans {
		student, err := s.studentRepo.GetByCardUID(ctx, scan.CardUID)
		if err != nil {
			return nil, err
		}

		if student == nil {
			result.NotFoundCards = append(result.NotFoundCards, scan.CardUID)
			continue
		}

		event := &domain.AttendanceEvent{
			SessionID: session.ID,
			StudentID: student.ID,
			CardUID:   scan.CardUID,
			ScannedAt: scan.ScannedAt,
		}

		if err := s.eventRepo.Create(ctx, event); err != nil {
			return nil, err
		}

		result.SavedEvents++
	}

	return result, nil
}
