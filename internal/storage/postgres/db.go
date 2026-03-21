package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dataNaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dataNaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool err:%s", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping daatbase:%s", err)
	}
	return pool, nil
}
