//go:generate mockgen -source $GOFILE -destination ./mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package exql

import (
	"context"
	"database/sql"
)

// Executor is an abstraction of sql.DB/sql.Tx
type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Hook is an effect-free middleware for queries
type Hook interface {
	Hook(ctx context.Context, query string, args ...any)
}

type Model interface {
	TableName() string
}

type ModelUpdate interface {
	UpdateTableName() string
}
