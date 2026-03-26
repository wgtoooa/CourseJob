package postgres

import (
	"context"
	"fmt"
)

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(repo Repository) error) error
}

func (txm *TxManager) WithinTransaction(ctx context.Context, fn func(repo Repository) error) (err error) {
	tx, err := txm.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("rollback failed: %v: %w", rbErr, err)
			}
		}
	}()

	repo := NewRepositories(tx)

	if err = fn(repo); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
