package service

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
)

type ScheduleImportInput struct {
	WeekSchedule domain.WeekSchedule
}

func (s *AttendanceService) ScheduleImport(ctx context.Context, input ScheduleImportInput) error {
	err := s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		if err := repo.Schedule().UpSetWeekSchedule(ctx, &input.WeekSchedule); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
