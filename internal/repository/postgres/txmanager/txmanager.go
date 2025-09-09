package txmanager

import (
	"context"
	"database/sql"
	"errors"
)

type keyTxType int

const (
	keyTxValue keyTxType = iota
)

var (
	ErrNoTransaction = errors.New("no transaction in context")
)

// TxManager интерфейс для управления транзакциями
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// SQLTxManager реализация для SQL базы данных
type SQLTxManager struct {
	db *sql.DB
}

func NewSQLTxManager(db *sql.DB) *SQLTxManager {
	return &SQLTxManager{db: db}
}

func (tm *SQLTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, ok := ctx.Value(keyTxValue).(*sql.Tx)
	if ok {
		return fn(ctx)
	}

	tx, err := tm.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return err
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
			err = commitErr
		}
	}()

	err = fn(ctx)
	return err
}

// GetTx извлекает транзакцию из контекста
func GetTx(ctx context.Context) (*sql.Tx, error) {
	tx, ok := ctx.Value(keyTxValue).(*sql.Tx)
	if !ok {
		return nil, ErrNoTransaction
	}
	return tx, nil
}

// WithTx добавляет транзакцию в контекст
func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, keyTxValue, tx)
}
