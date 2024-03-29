package exql

import (
	"context"
	"database/sql"
)

type stmtExecutor struct {
	ex    Executor
	stmts map[string]*sql.Stmt
}

func (e *stmtExecutor) Exec(query string, args ...any) (sql.Result, error) {
	return e.ExecContext(context.Background(), query, args...)
}

func (e *stmtExecutor) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	stmt, err := e.prepare(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.ExecContext(ctx, args...)
}

func (e *stmtExecutor) Prepare(stmt string) (*sql.Stmt, error) {
	return e.ex.Prepare(stmt)
}

func (e *stmtExecutor) PrepareContext(ctx context.Context, stmt string) (*sql.Stmt, error) {
	return e.ex.PrepareContext(ctx, stmt)
}

func (e *stmtExecutor) Query(query string, args ...any) (*sql.Rows, error) {
	return e.QueryContext(context.Background(), query, args...)
}

func (e *stmtExecutor) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	stmt, err := e.prepare(ctx, query)
	if err != nil {
		return nil, err
	}
	return stmt.QueryContext(ctx, args...)
}

func (e *stmtExecutor) QueryRow(query string, args ...any) *sql.Row {
	return e.QueryRowContext(context.Background(), query, args...)
}

func (e *stmtExecutor) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return e.ex.QueryRowContext(ctx, query, args...)
}

// StmtExecutor is the Executor that caches queries as *sql.Stmt.
// It uses the cached Stmt for the next execution if query is identical.
// They are held until Close() is called. This is useful for the case
// of executing the same query repeatedly in the for-loop.
// It may prevent errors caused by the db's connection pool.
//
// Example:
//
//	stmtExecer := exql.NewStmtExecutor(tx.Tx())
//	defer stmtExecer.Close()
//	stmtSaver := exql.NewSaver(stmtExecer)
type StmtExecutor interface {
	Executor
	// Close calls all retained *sql.Stmt and clears the buffer.
	// DON'T forget to call this on the manual use.
	Close() error
}

func (e *stmtExecutor) prepare(ctx context.Context, q string) (*sql.Stmt, error) {
	var err error
	stmt, ok := e.stmts[q]
	if !ok {
		if stmt, err = e.PrepareContext(ctx, q); err != nil {
			return nil, err
		} else {
			e.stmts[q] = stmt
		}
	}
	return stmt, nil
}

func (e *stmtExecutor) Close() error {
	var lastErr error
	for _, v := range e.stmts {
		err := v.Close()
		if err != nil {
			lastErr = err
		}
	}
	e.stmts = make(map[string]*sql.Stmt)
	return lastErr
}

func NewStmtExecutor(ex Executor) StmtExecutor {
	return newStmtExecutor(ex)
}

func newStmtExecutor(ex Executor) *stmtExecutor {
	return &stmtExecutor{ex: ex, stmts: make(map[string]*sql.Stmt)}
}
