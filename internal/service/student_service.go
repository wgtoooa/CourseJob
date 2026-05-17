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
	Email     string
	CardUID   string
	CreatedAt time.Time
}

func (s *AttendanceService) CreateStudent(ctx context.Context, student *StudentInput) error {
	if student == nil {
		return errors.New("student is nil")
	}

	return s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		existing, err := s.findStudentByCardUID(ctx, repo, student.CardUID)
		if err != nil {
			return err
		}
		if existing != nil {
			return ErrStudentExists
		}

		cardHash, err := s.auth.Hash(student.CardUID)
		if err != nil {
			return err
		}

		st := &domain.Student{
			FullName:  student.FullName,
			Course:    student.Course,
			GroupName: student.GroupName,
			Email:     student.Email,
			CardUID:   cardHash,
			CreatedAt: student.CreatedAt,
		}
		if st.CreatedAt.IsZero() {
			st.CreatedAt = time.Now().UTC()
		}

		return repo.Students().CreateStudent(ctx, st)
	})
}
