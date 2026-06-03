package attendance

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"context"
)

type AttendanceSessionRepository struct {
	db contract.DBTX
}

func NewAttendanceSessionRepository(db contract.DBTX) *AttendanceSessionRepository {
	return &AttendanceSessionRepository{db: db}
}

func (repo *AttendanceSessionRepository) Create(ctx context.Context, session *domain.AttendanceSession) error {
	const query = `
		WITH inserted AS (
			INSERT INTO attendance_session (room, source, started_at, finished_at, data)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (room, source, started_at, finished_at, data) DO NOTHING
			RETURNING id, created_at
		)
		SELECT id, created_at
		FROM inserted
		UNION ALL
		SELECT id, created_at
		FROM attendance_session
		WHERE room = $1
		  AND source = $2
		  AND started_at = $3
		  AND finished_at = $4
		  AND data = $5
		LIMIT 1;`
	err := repo.db.QueryRow(ctx,
		query,
		session.Room,
		session.Source,
		session.StartedAt,
		session.FinishedAt,
		session.Data).Scan(&session.ID, &session.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
