package schedule

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
)

type SubjectsExport struct {
	Name string `json:"name"`
}

func (s *Service) GetSubjects(ctx context.Context) ([]SubjectsExport, error) {
	var subjects []domain.Subject

	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		var err error

		subjects, err = repo.Subjects().GetSubjects(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	result := make([]SubjectsExport, 0, len(subjects))
	for _, subject := range subjects {
		result = append(result, SubjectsExport{Name: subject.Name})
	}

	return result, nil
}
