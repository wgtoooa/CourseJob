package schedule

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"context"
	"errors"
	"strings"
)

type PlanRepository struct {
	db contract.DBTX
}

func NewPlanRepository(db contract.DBTX) *PlanRepository {
	return &PlanRepository{db: db}
}

func (p *PlanRepository) GetPlanByCourse(ctx context.Context, course int) ([]domain.PlanItem, error) {
	const query = `
SELECT course, subject, planned_pairs
FROM plan_item
WHERE course = $1
ORDER BY subject`

	rows, err := p.db.Query(ctx, query, course)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.PlanItem, 0)
	for rows.Next() {
		var item domain.PlanItem
		if err = rows.Scan(&item.Course, &item.Subject, &item.PlannedPairs); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (p *PlanRepository) UpsertPlanItem(ctx context.Context, item *domain.PlanItem) error {
	if item == nil {
		return errors.New("plan item is nil")
	}
	if item.Course < 1 || item.Course > 4 {
		return errors.New("course must be an integer from 1 to 4")
	}
	if item.PlannedPairs < 0 {
		return errors.New("planned_pairs must be non-negative")
	}

	subject := strings.Join(strings.Fields(strings.TrimSpace(item.Subject)), " ")
	if subject == "" {
		return errors.New("subject is required")
	}
	subjectKey := strings.ToLower(strings.TrimSpace(subject))

	const query = `
INSERT INTO plan_item(course, subject, subject_key, planned_pairs, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (course, subject_key)
DO UPDATE SET
  subject = EXCLUDED.subject,
  planned_pairs = EXCLUDED.planned_pairs,
  updated_at = NOW()`

	_, err := p.db.Exec(ctx, query, item.Course, subject, subjectKey, item.PlannedPairs)
	return err
}
