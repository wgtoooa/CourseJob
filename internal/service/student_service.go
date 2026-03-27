package service

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"errors"
	"time"
)

type StudentInput struct {
	FullName  string
	Course    int
	GroupName string
	CardUID   string
	CreatedAt time.Time
}

func (s *AttendanceService) CreateStudent(ctx context.Context, student *StudentInput) (err error) {
	if student == nil {
		return errors.New("student is nil")
	}

	st := &domain.Student{
		FullName:  student.FullName,
		Course:    student.Course,
		GroupName: student.GroupName,
		CardUID:   student.CardUID,
		CreatedAt: student.CreatedAt,
	}
	if st.CreatedAt.IsZero() {
		st.CreatedAt = time.Now().UTC()
	}

	return s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		return repo.Students().CreateStudent(ctx, st)
	})
}
