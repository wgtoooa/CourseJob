package postgres

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"strings"
)

type TeacherRepository struct {
	db DBTX
}

func NewTeacherRepository(db DBTX) *TeacherRepository {
	return &TeacherRepository{db: db}
}

func (t *TeacherRepository) UpSetTeacher(ctx context.Context, teacher *domain.Teacher) error {
	if teacher == nil {
		return errors.New("teacher is nil")
	}

	fullName := strings.TrimSpace(teacher.FullName)
	if fullName == "" {
		return errors.New("teacher full_name is empty")
	}
	post := strings.TrimSpace(teacher.Post)

	query := `
		INSERT INTO teacher(full_name,post)
		VALUES ($1,$2)
		ON CONFLICT (full_name)
		DO UPDATE SET
			post = EXCLUDED.post
	`

	_, err := t.db.Exec(
		ctx,
		query,
		fullName,
		post,
	)

	if err != nil {
		return err
	}

	return nil
}
func (t *TeacherRepository) GetTeachers(ctx context.Context) ([]domain.Teacher, error) {
	query := `SELECT full_name, post FROM teacher`
	var teachers []domain.Teacher

	rows, err := t.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var teacher domain.Teacher
		if err = rows.Scan(&teacher.FullName, &teacher.Post); err != nil {
			return nil, err
		}
		teachers = append(teachers, teacher)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return teachers, nil
}
