package attendance

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"context"
)

type AttendanceEventRepository struct {
	db contract.DBTX
}

func NewAttendanceEventRepository(db contract.DBTX) *AttendanceEventRepository {
	return &AttendanceEventRepository{db}
}

func (repo *AttendanceEventRepository) Create(ctx context.Context, event *domain.AttendanceEvent) (bool, error) {
	const query = `
		WITH inserted AS (
			INSERT INTO attendance_event (session_id, student_id, card_uid_hash, scanned_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (session_id, student_id) DO NOTHING
			RETURNING id, created_at, TRUE AS inserted
		)
		SELECT id, created_at, inserted
		FROM inserted
		UNION ALL
		SELECT id, created_at, FALSE AS inserted
		FROM attendance_event
		WHERE session_id = $1
		  AND student_id = $2
		LIMIT 1`
	var inserted bool
	err := repo.db.QueryRow(
		ctx,
		query,
		event.SessionID,
		event.StudentID,
		event.CardUID,
		event.ScannedAt,
	).Scan(&event.ID, &event.CreatedAt, &inserted)
	if err != nil {
		return false, err
	}

	return inserted, nil
}
