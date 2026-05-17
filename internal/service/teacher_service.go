package service

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"strings"
)

type TeachersExport struct {
	PostFullName string
}

func (s *AttendanceService) GetTeachers(ctx context.Context) ([]TeachersExport, error) {
	var teachers []domain.Teacher

	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		var err error

		teachers, err = repo.Teachers().GetTeachers(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	result := make([]TeachersExport, 0, len(teachers))
	for _, teacher := range teachers {
		data := strings.ToLower(strings.Join(strings.Fields(teacher.Post+" "+teacher.FullName), ""))
		result = append(result, TeachersExport{PostFullName: data})
	}

	return result, nil
}
