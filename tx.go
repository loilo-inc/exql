package exql

import (
	"context"
	"database/sql"
	"fmt"

	q "github.com/loilo-inc/exql/v2/query"
)

type Tx interface {
	Saver
	Mapper
	Tx() *sql.Tx
}

type tx struct {
	s  Saver
	m  Mapper
	tx *sql.Tx
}

func (t *tx) Insert(modelPtr Model) (sql.Result, error) {
	return t.s.Insert(modelPtr)
}

func (t *tx) InsertContext(ctx context.Context, modelPtr Model) (sql.Result, error) {
	return t.s.InsertContext(ctx, modelPtr)
}

func (t *tx) Update(table string, set map[string]interface{}, where q.Condition) (sql.Result, error) {
	return t.s.Update(table, set, where)
}

func (t *tx) UpdateModel(ptr ModelUpdate, where q.Condition) (sql.Result, error) {
	return t.s.UpdateModel(ptr, where)
}

func (t *tx) UpdateContext(ctx context.Context, table string, set map[string]interface{}, where q.Condition) (sql.Result, error) {
	return t.s.UpdateContext(ctx, table, set, where)
}

func (t *tx) UpdateModelContext(ctx context.Context, ptr ModelUpdate, where q.Condition) (sql.Result, error) {
	return t.s.UpdateModelContext(ctx, ptr, where)
}

func (t *tx) Delete(table string, where q.Condition) (sql.Result, error) {
	return t.s.Delete(table, where)
}

func (t *tx) DeleteContext(ctx context.Context, table string, where q.Condition) (sql.Result, error) {
	return t.s.DeleteContext(ctx, table, where)
}

func (d *tx) Exec(query q.Query) (sql.Result, error) {
	return d.s.Exec(query)
}

func (d *tx) ExecContext(ctx context.Context, query q.Query) (sql.Result, error) {
	return d.s.ExecContext(ctx, query)
}

func (d *tx) Query(query q.Query) (*sql.Rows, error) {
	return d.s.Query(query)
}

func (d *tx) QueryContext(ctx context.Context, query q.Query) (*sql.Rows, error) {
	return d.s.QueryContext(ctx, query)
}

func (d *tx) QueryRow(query q.Query) (*sql.Row, error) {
	return d.s.QueryRow(query)
}

func (d *tx) QueryRowContext(ctx context.Context, query q.Query) (*sql.Row, error) {
	return d.s.QueryRowContext(ctx, query)
}

func (t *tx) Map(rows *sql.Rows, pointerOfStruct interface{}) error {
	return t.m.Map(rows, pointerOfStruct)
}

func (t *tx) MapMany(rows *sql.Rows, pointerOfSliceOfStruct interface{}) error {
	return t.m.MapMany(rows, pointerOfSliceOfStruct)
}

func (t *tx) AfterHook() *HookList {
	return t.s.AfterHook()
}

func (t *tx) BeforeHook() *HookList {
	return t.s.BeforeHook()
}

func (t *tx) Tx() *sql.Tx {
	return t.tx
}

func Transaction(db *sql.DB, ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	sqlTx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	tx := &tx{tx: sqlTx, s: NewSaver(sqlTx), m: NewMapper()}
	var p interface{}
	txErr := func() error {
		defer func() {
			p = recover()
		}()
		return callback(tx)
	}()
	if p != nil {
		txErr = fmt.Errorf("recovered: %s", p)
	}
	if txErr != nil {
		if err := sqlTx.Rollback(); err != nil {
			return err
		}
		return txErr
	} else if err := sqlTx.Commit(); err != nil {
		return err
	}
	return nil
}
