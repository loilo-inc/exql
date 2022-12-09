package exql

import (
	"context"
	"database/sql"
	"reflect"

	q "github.com/loilo-inc/exql/query"
	"golang.org/x/xerrors"
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

func (t *tx) Insert(structPtr interface{}) (sql.Result, error) {
	return t.s.Insert(structPtr)
}

func (t *tx) InsertContext(ctx context.Context, structPtr interface{}) (sql.Result, error) {
	return t.s.InsertContext(ctx, structPtr)
}

func (t *tx) QueryForInsert(structPtr interface{}) (q.Query, *reflect.Value, error) {
	return t.s.QueryForInsert(structPtr)
}

func (t *tx) Update(table string, set map[string]interface{}, where Clause) (sql.Result, error) {
	return t.s.Update(table, set, where)
}

func (t *tx) UpdateModel(ptr interface{}, where Clause) (sql.Result, error) {
	return t.s.UpdateModel(ptr, where)
}

func (t *tx) UpdateContext(ctx context.Context, table string, set map[string]interface{}, where Clause) (sql.Result, error) {
	return t.s.UpdateContext(ctx, table, set, where)
}

func (t *tx) UpdateModelContext(ctx context.Context, ptr interface{}, where Clause) (sql.Result, error) {
	return t.s.UpdateModelContext(ctx, ptr, where)
}

func (t *tx) Delete(table string, where Clause) (sql.Result, error) {
	return t.s.Delete(table, where)
}

func (t *tx) DeleteContext(ctx context.Context, table string, where Clause) (sql.Result, error) {
	return t.s.DeleteContext(ctx, table, where)
}

func (t *tx) QueryForUpdateModel(ptr interface{}, where Clause) (q.Query, error) {
	return t.s.QueryForUpdateModel(ptr, where)
}

func (t *tx) Map(rows *sql.Rows, pointerOfStruct interface{}) error {
	return t.m.Map(rows, pointerOfStruct)
}

func (t *tx) MapMany(rows *sql.Rows, pointerOfSliceOfStruct interface{}) error {
	return t.m.MapMany(rows, pointerOfSliceOfStruct)
}

func (t *tx) Tx() *sql.Tx {
	return t.tx
}

func transaction(db *sql.DB, ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
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
		txErr = xerrors.Errorf("recovered: %s", p)
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
