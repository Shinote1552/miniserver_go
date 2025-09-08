package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type keyTxType int

const (
	keyTxValue keyTxType = iota
)

func (p *PostgresStorage) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, ok := ctx.Value(keyTxValue).(*sql.Tx)
	if ok {
		return fn(ctx)
	}

	tx, err = p.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	ctx = context.WithValue(ctx, keyTxValue, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}

		if err != nil {
			_ = tx.Rollback()
			return
		}

		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		}
	}()

	err = fn(ctx)
	return
}
