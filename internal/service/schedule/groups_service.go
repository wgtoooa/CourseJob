package schedule

import (
	"CourseJob/internal/domain"
	"context"
)

type GroupExport struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (s *Service) GetGroupsByCourse(ctx context.Context, course int) ([]GroupExport, error) {
	data, err := s.GetScheduleByCourse(ctx, course, domain.ScheduleFilters{})
	if err != nil {
		return nil, err
	}

	result := make([]GroupExport, 0, len(data.Groups))
	for _, group := range data.Groups {
		result = append(result, GroupExport{
			ID:    group.ID,
			Name:  group.Name,
			Count: group.Count,
		})
	}

	return result, nil
}
