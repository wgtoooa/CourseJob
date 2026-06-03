package attendance

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"errors"
	"math"
	"strings"
	"time"
)

var (
	ErrStudentAccessDenied    = errors.New("student access denied")
	ErrStudentProfileNotFound = errors.New("student profile not found")
)

type StudentAttendanceFilter struct {
	FromDate *time.Time
	ToDate   *time.Time
}

func (s *AttendanceService) GetMyAttendance(
	ctx context.Context,
	userID int64,
	role string,
	filter StudentAttendanceFilter,
) (*domain.StudentAttendanceReport, error) {
	if userID <= 0 {
		return nil, ErrStudentAccessDenied
	}
	if strings.ToLower(strings.TrimSpace(role)) != domain.RoleStudent {
		return nil, ErrStudentAccessDenied
	}

	report := &domain.StudentAttendanceReport{}
	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		user, err := repo.Users().GetByID(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil || !user.IsActive || user.StudentID == nil || *user.StudentID <= 0 {
			return ErrStudentProfileNotFound
		}

		records, err := repo.Report().ListStudentAttendance(ctx, *user.StudentID, filter.FromDate, filter.ToDate)
		if err != nil {
			return err
		}

		report.StudentID = *user.StudentID
		report.Records = records
		return nil
	})
	if err != nil {
		return nil, err
	}

	report.TotalSessions = len(report.Records)
	for _, rec := range report.Records {
		if rec.Present {
			report.PresentCount++
		}
	}
	report.AbsentCount = report.TotalSessions - report.PresentCount
	if report.TotalSessions > 0 {
		value := (float64(report.PresentCount) * 100) / float64(report.TotalSessions)
		report.AttendancePercent = math.Round(value*100) / 100
	}

	return report, nil
}
