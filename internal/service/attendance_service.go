package service

import (
	"CourseJob/internal/storage/postgres"
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
	transactor postgres.Transactor
	event      AttendanceEventCreator
	session    AttendanceSessionCreator
	student    StudentGetter
}

func NewAttendanceService(
	transactor postgres.Transactor,
	event AttendanceEventCreator,
	studentGetter StudentGetter,
	sessionCreator AttendanceSessionCreator,
) *AttendanceService {
	return &AttendanceService{
		transactor: transactor,
		event:      event,
		student:    studentGetter,
		session:    sessionCreator,
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

	var result *ProcessAttendanceResult
	err := s.transactor.WithinTransaction(ctx, func(repo postgres.Repositories) error {
		session := &domain.AttendanceSession{
			Room:       input.Room,
			Source:     input.Source,
			StartedAt:  input.StartedAt,
			FinishedAt: input.FinishedAt,
		}
		if err := repo.Sessions().Create(ctx, session); err != nil {
			return err
		}

		result = &ProcessAttendanceResult{
			SessionID:     session.ID,
			SavedEvents:   0,
			NotFoundCards: []string{},
		}

		for _, scan := range input.Scans {
			student, err := repo.Students().GetByCardUID(ctx, scan.CardUID)
			if err != nil {
				return err
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

			if err := repo.Events().Create(ctx, event); err != nil {
				return err
			}

			result.SavedEvents++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil

}
