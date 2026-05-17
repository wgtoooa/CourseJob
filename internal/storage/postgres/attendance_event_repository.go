package postgres

import (
	"CourseJob/internal/domain"
	"context"
)

type AttendanceEventRepository struct {
	db DBTX
}

func NewAttendanceEventRepository(db DBTX) *AttendanceEventRepository {
	return &AttendanceEventRepository{db}
}

func (repo *AttendanceEventRepository) Create(ctx context.Context, event *domain.AttendanceEvent) error {
	const query = `
		INSERT INTO attendance_event (session_id, student_id, card_uid, scanned_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	err := repo.db.QueryRow(
		ctx,
		query,
		event.SessionID,
		event.StudentID,
		event.CardUID,
		event.ScannedAt,
	).Scan(&event.ID, &event.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}
