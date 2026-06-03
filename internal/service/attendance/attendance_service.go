package attendance

import (
	authA "CourseJob/internal/auth"
	"CourseJob/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"time"

	"CourseJob/internal/domain"
)

type AttendanceService struct {
	transactor postgres.Transactor
	auth       authA.Hasher
}

func NewAttendanceService(
	transactor postgres.Transactor,
	auth authA.Hasher,
) *AttendanceService {
	return &AttendanceService{
		transactor: transactor,
		auth:       auth,
	}
}

type AttendanceInput struct {
	Room       string
	Source     string
	StartedAt  time.Time
	FinishedAt time.Time
	Data       time.Time
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
	input AttendanceInput,
) (*ProcessAttendanceResult, error) {

	var result *ProcessAttendanceResult
	err := s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		session := &domain.AttendanceSession{
			Room:       input.Room,
			Source:     input.Source,
			StartedAt:  input.StartedAt,
			FinishedAt: input.FinishedAt,
			Data:       input.Data,
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
			normalizedCardUID, ok := authA.NormalizeAndValidateCardUID(scan.CardUID)
			if !ok {
				return errors.New("invalid card uid format")
			}

			cardHash, err := s.auth.Hash(normalizedCardUID)
			if err != nil {
				return err
			}
			student, err := repo.Students().GetByCardUID(ctx, cardHash)
			if err != nil {
				return err
			}

			if student == nil {
				result.NotFoundCards = append(result.NotFoundCards, normalizedCardUID)
				continue
			}

			event := &domain.AttendanceEvent{
				SessionID: session.ID,
				StudentID: student.ID,
				CardUID:   cardHash,
				ScannedAt: scan.ScannedAt,
			}

			inserted, err := repo.Events().Create(ctx, event)
			if err != nil {
				return err
			}

			if inserted {
				result.SavedEvents++
			}
		}
		if err := repo.Report().BuildSessionReport(ctx, session.ID); err != nil {
			return fmt.Errorf("failed to report session report: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil

}
