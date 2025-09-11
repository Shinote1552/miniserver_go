// postgres/querier.go
package postgres

import (
	"context"
	"database/sql"
)

// SQLQuerier реализация Querier для sql.DB (обычное соединение)
type SQLQuerier struct {
	db *sql.DB
}

// QueryRowContext выполняет запрос, возвращающий не более одной строки
func (q *SQLQuerier) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return q.db.QueryRowContext(ctx, query, args...)
}

// QueryContext выполняет запрос, возвращающий multiple строки
func (q *SQLQuerier) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return q.db.QueryContext(ctx, query, args...)
}

// ExecContext выполняет запрос без возвращения строк
func (q *SQLQuerier) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return q.db.ExecContext(ctx, query, args...)
}

// TxQuerier реализация Querier для sql.Tx (транзакция)
type TxQuerier struct {
	tx *sql.Tx
}

// QueryRowContext выполняет запрос в транзакции, возвращающий не более одной строки
func (q *TxQuerier) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return q.tx.QueryRowContext(ctx, query, args...)
}

// QueryContext выполняет запрос в транзакции, возвращающий multiple строки
func (q *TxQuerier) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return q.tx.QueryContext(ctx, query, args...)
}

// ExecContext выполняет запрос в транзакции без возвращения строк
func (q *TxQuerier) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return q.tx.ExecContext(ctx, query, args...)
}
