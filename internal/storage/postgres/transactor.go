package postgres

import "context"

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(repo Repositories) error) error
}

func (txm *TxManager) WithinTransaction(ctx context.Context, fn func(rep Repositories) error) error {
	tx, err := txm.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	repo := NewRepository(tx)

	if err = fn(repo); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
