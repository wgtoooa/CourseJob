package postgres

import (
	"CourseJob/internal/domain"
	"context"
)

type AttendanceSessionRepository struct {
	db *DB
}

func NewAttendanceSessionRepository(db *DB) *AttendanceSessionRepository {
	return &AttendanceSessionRepository{db: db}
}

func (repo *AttendanceSessionRepository) Create(ctx context.Context, session *domain.AttendanceSession) error {
	const query = `
					insert into attendance_session(room,source,created_at,finished_at) values
					($1, $2, $3,$4)
					returning id created_at;`
	err := repo.db.Pool.QueryRow(ctx,
		query,
		session.Room,
		session.Source,
		session.CreatedAt,
		session.FinishedAt).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
