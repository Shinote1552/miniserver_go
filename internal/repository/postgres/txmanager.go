package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type keyTxType int

const (
	keyTxValue keyTxType = iota
)

var (
	ErrNoTransaction = errors.New("no transaction in context")
)

// SQLTxManager реализация для SQL базы данных
type SQLTxManager struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) *SQLTxManager {
	return &SQLTxManager{
		db: db,
	}
}

// Если opts == nil, применяется по умолчанию уровень ReadCommitted
func (tm *SQLTxManager) WithTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error {
	if opts == nil {
		opts = &sql.TxOptions{
			Isolation: sql.LevelSerializable,
			ReadOnly:  false,
		}
	}

	tx, err := tm.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	ctx = context.WithValue(ctx, keyTxValue, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	err = fn(ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("rollback error: %w, original error: %w", rollbackErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (tm *SQLTxManager) GetQuerier(ctx context.Context) (Querier, error) {
	tx, err := tm.getTx(ctx)
	if err != nil {
		if errors.Is(err, ErrNoTransaction) {
			// Возвращаем Querier для основного соединения
			return &SQLQuerier{db: tm.db}, nil
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	// Возвращаем Querier для транзакции
	return &TxQuerier{tx: tx}, nil
}

// GetTx извлекает транзакцию из контекста
func (tm *SQLTxManager) getTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value(keyTxValue).(*sql.Tx)
	if !ok {
		return nil, ErrNoTransaction
	}
	return tx, nil
}
