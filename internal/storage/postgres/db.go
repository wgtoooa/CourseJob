package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewPool(ctx context.Context, dataBaseURL string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dataBaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool err:%s", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database:%s", err)
	}
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}
