package txmanager

import (
	"context"
	"database/sql"
	"fmt"
)

type keyTxType int

const (
	keyTxValue keyTxType = iota
)

type TxManager struct {
	conn *sql.DB
}

func New(conn *sql.DB) *TxManager {
	return &TxManager{conn: conn}
}

func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, ok := ctx.Value(keyTxValue).(*sql.Tx)
	if ok {
		return fn(ctx)
	}

	tx, err = m.conn.BeginTx(ctx, &sql.TxOptions{
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

		// Обрабатываем ошибки
		if err != nil {
			_ = tx.Rollback()
			return
		}

		// Коммитим если нет ошибок
		if commitErr := tx.Commit(); commitErr != nil {
			err = fmt.Errorf("failed to commit transaction: %w", commitErr)
		}
	}()

	err = fn(ctx)
	return
}
