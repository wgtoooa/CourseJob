package schedule

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"errors"
	"strings"
)

type PlanUpsertInput struct {
	Course       int
	Subject      string
	PlannedPairs int
}

func (s *Service) GetPlanByCourse(ctx context.Context, course int) ([]domain.PlanItem, error) {
	if course < 1 || course > 4 {
		return nil, errors.New("course must be an integer from 1 to 4")
	}

	var items []domain.PlanItem
	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		var repoErr error
		items, repoErr = repo.Plan().GetPlanByCourse(ctx, course)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	if items == nil {
		return []domain.PlanItem{}, nil
	}
	return items, nil
}

func (s *Service) UpsertPlanItem(ctx context.Context, input PlanUpsertInput) error {
	if input.Course < 1 || input.Course > 4 {
		return errors.New("course must be an integer from 1 to 4")
	}
	if input.PlannedPairs < 0 {
		return errors.New("planned_pairs must be non-negative")
	}
	subject := strings.Join(strings.Fields(strings.TrimSpace(input.Subject)), " ")
	if subject == "" {
		return errors.New("subject is required")
	}

	return s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		return repo.Plan().UpsertPlanItem(ctx, &domain.PlanItem{
			Course:       input.Course,
			Subject:      subject,
			PlannedPairs: input.PlannedPairs,
		})
	})
}
