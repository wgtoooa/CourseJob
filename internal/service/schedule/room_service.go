package schedule

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
)

type RoomsExport struct {
	Name string `json:"name"`
}

func (s *Service) GetRooms(ctx context.Context) ([]RoomsExport, error) {
	var rooms []domain.Room

	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		var err error

		rooms, err = repo.Rooms().GetRooms(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	result := make([]RoomsExport, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, RoomsExport{Name: room.Name})
	}

	return result, nil
}
