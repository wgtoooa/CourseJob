package postgres

import (
	"CourseJob/internal/domain"
	"context"
)

type AttendanceSessionRepository struct {
	db DBTX
}

func NewAttendanceSessionRepository(db DBTX) *AttendanceSessionRepository {
	return &AttendanceSessionRepository{db: db}
}

func (repo *AttendanceSessionRepository) Create(ctx context.Context, session *domain.AttendanceSession) error {
	const query = `
					insert into attendance_session(room,source,started_at,finished_at) values
					($1, $2, $3,$4)
					returning id,created_at;`
	err := repo.db.QueryRow(ctx,
		query,
		session.Room,
		session.Source,
		session.StartedAt,
		session.FinishedAt).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
