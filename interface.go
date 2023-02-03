//go:generate mockgen -source $GOFILE -destination ./mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package exql

import (
	"context"
	"database/sql"
)

// Executor is an abstraction of both sql.DB/sql.Tx
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Prepare(stmt string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, stmt string) (*sql.Stmt, error)
}

type Model interface {
	TableName() string
}

type ModelUpdate interface {
	UpdateTableName() string
}
