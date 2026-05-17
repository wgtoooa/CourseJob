package postgres

import (
	"CourseJob/internal/domain"
	"context"
	"errors"
	"strings"
)

type RoomRepository struct {
	db DBTX
}

func NewRoomRepository(db DBTX) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) UpSetRoom(ctx context.Context, room *domain.Room) error {
	if room == nil {
		return errors.New("room is nil")
	}

	name := normalizeRoomName(room.Name)
	if name == "" {
		return errors.New("room name is empty")
	}

	if err := r.ensureRoomCatalogTable(ctx); err != nil {
		return err
	}

	const query = `
INSERT INTO room_catalog(name)
VALUES ($1)
ON CONFLICT (name) DO NOTHING`

	_, err := r.db.Exec(ctx, query, name)
	return err
}

func (r *RoomRepository) GetRooms(ctx context.Context) ([]domain.Room, error) {
	if err := r.ensureRoomCatalogTable(ctx); err != nil {
		return nil, err
	}

	const query = `SELECT id, name FROM room_catalog ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		if err = rows.Scan(&room.ID, &room.Name); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *RoomRepository) ensureRoomCatalogTable(ctx context.Context) error {
	const query = `
CREATE TABLE IF NOT EXISTS room_catalog (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
)`

	_, err := r.db.Exec(ctx, query)
	return err
}

func normalizeRoomName(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
