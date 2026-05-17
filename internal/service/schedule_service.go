package service

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"fmt"
	"log"
	"time"
)

type ScheduleImportInput struct {
	WeekSchedule domain.WeekSchedule
}

func (s *AttendanceService) ScheduleImport(ctx context.Context, input ScheduleImportInput) error {
	start := time.Now()

	counter := input.WeekSchedule.GeneratedAt

	generatedAt, err := time.Parse(time.RFC3339, counter)
	if err != nil {
		return err
	}

	err = s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		if err := repo.Schedule().UpSetWeekSchedule(ctx, &input.WeekSchedule); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	log.Printf(
		"import delay from generated_at: %v, total import time: %v",
		time.Since(generatedAt),
		time.Since(start),
	)

	return nil
}

func (s *AttendanceService) GetScheduleByCourse(ctx context.Context, course int, filters domain.ScheduleFilters) (*domain.ScheduleCourse, error) {
	if course <= 0 {
		return nil, fmt.Errorf("course must be positive")
	}

	var out *domain.ScheduleCourse
	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		var repoErr error
		out, repoErr = repo.Schedule().GetCourseSchedule(ctx, course, filters)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	if out == nil {
		return &domain.ScheduleCourse{
			Course: course,
			Groups: []domain.ScheduleGroupSummary{},
			Weeks:  []domain.ScheduleWeekExport{},
		}, nil
	}

	return out, nil
}
