package schedule

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

type StudentRepository struct {
	db contract.DBTX
}

func NewStudentRepository(db contract.DBTX) *StudentRepository {
	return &StudentRepository{db: db}
}

func (repo *StudentRepository) GetByCardUID(ctx context.Context, UID string) (*domain.Student, error) {
	const query = `
SELECT id,full_name,course,group_name,email,card_uid_hash,created_at
from student
where card_uid_hash = $1`

	var students domain.Student
	err := repo.db.QueryRow(ctx, query, UID).Scan(
		&students.ID,
		&students.FullName,
		&students.Course,
		&students.GroupName,
		&students.Email,
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
INSERT INTO student (full_name,course,group_name,email,card_uid_hash,created_at)
VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := repo.db.Exec(ctx, query,
		st.FullName,
		st.Course,
		st.GroupName,
		st.Email,
		st.CardUID,
		st.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (repo *StudentRepository) CreateStudents(ctx context.Context, students []domain.Student) ([]string, error) {
	if len(students) == 0 {
		return []string{}, nil
	}

	fullNames := make([]string, 0, len(students))
	courses := make([]int16, 0, len(students))
	groupNames := make([]string, 0, len(students))
	emails := make([]string, 0, len(students))
	cardUIDs := make([]string, 0, len(students))
	createdAts := make([]time.Time, 0, len(students))

	for i := range students {
		fullNames = append(fullNames, students[i].FullName)
		courses = append(courses, int16(students[i].Course))
		groupNames = append(groupNames, students[i].GroupName)
		emails = append(emails, students[i].Email)
		cardUIDs = append(cardUIDs, students[i].CardUID)
		createdAts = append(createdAts, students[i].CreatedAt)
	}

	const query = `
INSERT INTO student (full_name, course, group_name, email, card_uid_hash, created_at)
SELECT *
FROM unnest(
  $1::text[],
  $2::smallint[],
  $3::text[],
  $4::text[],
  $5::text[],
  $6::timestamptz[]
)
ON CONFLICT (card_uid_hash) DO NOTHING
RETURNING card_uid_hash`

	rows, err := repo.db.Query(ctx, query, fullNames, courses, groupNames, emails, cardUIDs, createdAts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	insertedCardUIDs := make([]string, 0, len(students))
	for rows.Next() {
		var cardUID string
		if err = rows.Scan(&cardUID); err != nil {
			return nil, err
		}
		insertedCardUIDs = append(insertedCardUIDs, cardUID)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return insertedCardUIDs, nil
}
