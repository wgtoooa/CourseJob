package postgres

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

type StudentRepository struct {
	db DBTX
}

func NewStudentRepository(db DBTX) *StudentRepository {
	return &StudentRepository{db: db}
}

func (repo *StudentRepository) GetByCardUID(ctx context.Context, UID string) (*domain.Student, error) {
	const query = `
SELECT id,full_name,course,group_name,card_uid,created_at
from student
where card_uid = $1`

	var students domain.Student
	err := repo.db.QueryRow(ctx, query, UID).Scan(
		&students.ID,
		&students.FullName,
		&students.Course,
		&students.GroupName,
		&students.CardUID,
		&students.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &students, nil
}
