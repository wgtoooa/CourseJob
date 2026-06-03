package schedule

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"errors"
)

func (s *Service) FindLessonBySession(ctx context.Context, sessionID int64) (*domain.ScheduleLesson, error) {
	if sessionID <= 0 {
		return nil, errors.New("sessionID must be positive")
	}

	var lesson *domain.ScheduleLesson
	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		var repoErr error
		lesson, repoErr = repo.Schedule().FindLessonBySession(ctx, sessionID)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	return lesson, nil
}
