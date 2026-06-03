package attendance

import (
	authA "CourseJob/internal/auth"
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"errors"
	"fmt"
	"time"
)

type StudentInput struct {
	FullName  string
	Course    int
	GroupName string
	Email     string
	CardUID   string
	CreatedAt *time.Time
}

var ErrStudentExists = errors.New("student with this card_uid already exists")

func (s *AttendanceService) CreateStudent(ctx context.Context, student *StudentInput) error {
	if student == nil {
		return errors.New("student is nil")
	}

	return s.CreateStudents(ctx, []StudentInput{*student})
}

func (s *AttendanceService) CreateStudents(ctx context.Context, students []StudentInput) error {
	if len(students) == 0 {
		return errors.New("students is empty")
	}

	return s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		seenCardUIDs := make(map[string]struct{}, len(students))
		payload := make([]domain.Student, 0, len(students))
		payloadCardUIDs := make([]string, 0, len(students))

		for i := range students {
			prepared, err := s.prepareStudentForInsert(&students[i], seenCardUIDs)
			if err != nil {
				return fmt.Errorf("student[%d]: %w", i, err)
			}
			payload = append(payload, *prepared)
			payloadCardUIDs = append(payloadCardUIDs, prepared.CardUID)
		}

		insertedCardUIDs, err := repo.Students().CreateStudents(ctx, payload)
		if err != nil {
			return err
		}

		insertedSet := make(map[string]struct{}, len(insertedCardUIDs))
		for _, cardUID := range insertedCardUIDs {
			insertedSet[cardUID] = struct{}{}
		}

		if len(insertedSet) != len(payloadCardUIDs) {
			for i, cardUID := range payloadCardUIDs {
				if _, ok := insertedSet[cardUID]; !ok {
					return fmt.Errorf("student[%d]: %w", i, ErrStudentExists)
				}
			}
		}
		return nil
	})
}

func (s *AttendanceService) prepareStudentForInsert(
	student *StudentInput,
	seenCardUIDs map[string]struct{},
) (*domain.Student, error) {
	if student == nil {
		return nil, errors.New("student is nil")
	}

	normalizedCardUID, ok := authA.NormalizeAndValidateCardUID(student.CardUID)
	if !ok {
		return nil, errors.New("invalid card uid format")
	}

	if _, exists := seenCardUIDs[normalizedCardUID]; exists {
		return nil, ErrStudentExists
	}

	seenCardUIDs[normalizedCardUID] = struct{}{}

	cardHash, err := s.auth.Hash(normalizedCardUID)
	if err != nil {
		return nil, err
	}

	var createdAt time.Time
	if student.CreatedAt != nil {
		createdAt = *student.CreatedAt
	}
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &domain.Student{
		FullName:  student.FullName,
		Course:    student.Course,
		GroupName: student.GroupName,
		Email:     student.Email,
		CardUID:   cardHash,
		CreatedAt: createdAt,
	}, nil
}
