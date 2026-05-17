package postgres

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"strings"
)

type SubjectRepository struct {
	db DBTX
}

func NewSubjectRepository(db DBTX) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (s *SubjectRepository) UpSetSubject(ctx context.Context, subject *domain.Subject) error {
	if subject == nil {
		return errors.New("subject is nil")
	}

	name := normalizeSubjectName(subject.Name)
	if name == "" {
		return errors.New("subject name is empty")
	}

	if err := s.ensureSubjectCatalogTable(ctx); err != nil {
		return err
	}

	const query = `
INSERT INTO subject_catalog(name)
VALUES ($1)
ON CONFLICT (name) DO NOTHING`

	_, err := s.db.Exec(ctx, query, name)
	return err
}

func (s *SubjectRepository) GetSubjects(ctx context.Context) ([]domain.Subject, error) {
	if err := s.ensureSubjectCatalogTable(ctx); err != nil {
		return nil, err
	}

	const query = `SELECT id, name FROM subject_catalog ORDER BY name`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []domain.Subject
	for rows.Next() {
		var subject domain.Subject
		if err = rows.Scan(&subject.ID, &subject.Name); err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subjects, nil
}

func (s *SubjectRepository) ensureSubjectCatalogTable(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS subject_catalog (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
)`

	_, err := s.db.Exec(ctx, query)
	return err
}

func normalizeSubjectName(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
