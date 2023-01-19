package exdriver

import (
	"context"
	"database/sql/driver"
)

type Connector struct {
	dsn   string
	dr    driver.Driver
	hooks *HookList
}

type HookDelegate interface {
	QueryHook() QueryHook
}

// NewConnector returns wrapped driver.Connector for hooking queries.
func NewConnector(dr driver.Driver, dsn string) *Connector {
	return &Connector{
		dr:    dr,
		dsn:   dsn,
		hooks: &HookList{},
	}
}

func (c *Connector) Connect(context.Context) (driver.Conn, error) {
	conn, err := c.dr.Open(c.dsn)
	if err != nil {
		return nil, err
	}
	return &connection{Conn: conn, hook: c.hooks}, nil
}

func (c *Connector) Driver() driver.Driver {
	return c.dr
}

func (c *Connector) Hooks() *HookList {
	return c.hooks
}

// Implements:
// - driver.Conn
// - driver.ConnBeginTx
// - driver.ConnPrepareContext
// - driver.Pinger
// - driver.ExecerContext
// - driver.QueryerContext
// - driver.SessionResetter
// - driver.NamedValueChecker
// - driver.Validater
type connection struct {
	driver.Conn
	hook QueryHook
}

func (c *connection) Begin() (driver.Tx, error) {
	panic("*sql.Conn never call this")
}

func (c *connection) Close() error {
	return c.Conn.Close()
}

func (c *connection) Prepare(query string) (driver.Stmt, error) {
	stmt, err := c.Conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &statement{stmt: stmt, q: query, hook: c.hook}, nil
}

func (c *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	cb := c.Conn.(driver.ConnBeginTx)
	tx, err := cb.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	c.hook.HookQuery(ctx, "BEGIN", nil)
	return &transaction{tx: tx, hook: c.hook}, nil
}

func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	pc := c.Conn.(driver.ConnPrepareContext)
	stmt, err := pc.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &statement{stmt: stmt, q: query, hook: c.hook}, nil
}

func (c *connection) Ping(ctx context.Context) error {
	pi := c.Conn.(driver.Pinger)
	return pi.Ping(ctx)
}

func (c *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	ex := c.Conn.(driver.ExecerContext)
	res, err := ex.ExecContext(ctx, query, args)
	if err != nil {
		return nil, err
	}
	c.hook.HookQuery(ctx, query, args)
	return res, nil
}

func (c *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	ex := c.Conn.(driver.QueryerContext)
	res, err := ex.QueryContext(ctx, query, args)
	if err != nil {
		return nil, err
	}
	c.hook.HookQuery(ctx, query, args)
	return res, nil
}

func (c *connection) ResetSession(ctx context.Context) error {
	ses := c.Conn.(driver.SessionResetter)
	return ses.ResetSession(ctx)
}

func (c *connection) CheckNamedValue(v *driver.NamedValue) error {
	nm := c.Conn.(driver.NamedValueChecker)
	return nm.CheckNamedValue(v)
}

func (c *connection) IsValid() bool {
	vl := c.Conn.(driver.Validator)
	return vl.IsValid()
}

type statement struct {
	stmt driver.Stmt
	q    string
	hook QueryHook
}

func (s *statement) Close() error {
	return s.stmt.Close()
}

func (s *statement) NumInput() int {
	return s.stmt.NumInput()
}

func (s *statement) Exec(args []driver.Value) (driver.Result, error) {
	panic("*sql.Stmt never call this")
}

func (s *statement) Query(args []driver.Value) (driver.Rows, error) {
	panic("*sql.Stmt never call this")
}

func (s *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	ex := s.stmt.(driver.StmtExecContext)
	res, err := ex.ExecContext(ctx, args)
	if err != nil {
		return nil, err
	}
	s.hook.HookQuery(ctx, s.q, args)
	return res, nil
}

func (s *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	qr := s.stmt.(driver.StmtQueryContext)
	res, err := qr.QueryContext(ctx, args)
	if err != nil {
		return nil, err
	}
	s.hook.HookQuery(ctx, s.q, args)
	return res, nil
}

type transaction struct {
	tx   driver.Tx
	hook QueryHook
}

func (t *transaction) Commit() error {
	if err := t.tx.Commit(); err != nil {
		return err
	}
	t.hook.HookQuery(context.Background(), "COMMIT", nil)
	return nil
}

func (t *transaction) Rollback() error {
	if err := t.tx.Rollback(); err != nil {
		return err
	}
	t.hook.HookQuery(context.Background(), "ROLLBACK", nil)
	return nil
}
