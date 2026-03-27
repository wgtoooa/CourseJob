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

func (repo *StudentRepository) CreateStudent(ctx context.Context, st *domain.Student) error {
	query := `
INSERT INTO student (full_name,course,group_name,card_uid,created_at)
VALUES ($1,$2,$3,$4,$5)`
	_, err := repo.db.Exec(ctx, query,
		st.FullName,
		st.Course,
		st.GroupName,
		st.CardUID,
		st.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
