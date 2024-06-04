package exql

import (
	"context"
	"database/sql"

	"github.com/loilo-inc/exql/v2/query"
	"golang.org/x/xerrors"
)

type Tx interface {
	Saver
	Finder
	Mapper
	Tx() *sql.Tx
}

type tx struct {
	s  *saver
	f  *finder
	tx *sql.Tx
}

func newTx(t *sql.Tx) *tx {
	return &tx{s: newSaver(t), f: newFinder(t), tx: t}
}

func (t *tx) Insert(modelPtr Model) (sql.Result, error) {
	return t.s.Insert(modelPtr)
}

func (t *tx) InsertContext(ctx context.Context, modelPtr Model) (sql.Result, error) {
	return t.s.InsertContext(ctx, modelPtr)
}

func (t *tx) Update(table string, set map[string]interface{}, where query.Condition) (sql.Result, error) {
	return t.s.Update(table, set, where)
}

func (t *tx) UpdateModel(ptr ModelUpdate, where query.Condition) (sql.Result, error) {
	return t.s.UpdateModel(ptr, where)
}

func (t *tx) UpdateContext(ctx context.Context, table string, set map[string]interface{}, where query.Condition) (sql.Result, error) {
	return t.s.UpdateContext(ctx, table, set, where)
}

func (t *tx) UpdateModelContext(ctx context.Context, ptr ModelUpdate, where query.Condition) (sql.Result, error) {
	return t.s.UpdateModelContext(ctx, ptr, where)
}

func (t *tx) Delete(table string, where query.Condition) (sql.Result, error) {
	return t.s.Delete(table, where)
}

func (t *tx) DeleteContext(ctx context.Context, table string, where query.Condition) (sql.Result, error) {
	return t.s.DeleteContext(ctx, table, where)
}

func (d *tx) Exec(query query.Query) (sql.Result, error) {
	return d.s.Exec(query)
}

func (d *tx) ExecContext(ctx context.Context, query query.Query) (sql.Result, error) {
	return d.s.ExecContext(ctx, query)
}

func (d *tx) Query(query query.Query) (*sql.Rows, error) {
	return d.s.Query(query)
}

func (d *tx) QueryContext(ctx context.Context, query query.Query) (*sql.Rows, error) {
	return d.s.QueryContext(ctx, query)
}

func (d *tx) QueryRow(query query.Query) (*sql.Row, error) {
	return d.s.QueryRow(query)
}

func (d *tx) QueryRowContext(ctx context.Context, query query.Query) (*sql.Row, error) {
	return d.s.QueryRowContext(ctx, query)
}

// Find implements DB
func (t *tx) Find(q query.Query, destPtrOfStruct any) error {
	return t.f.Find(q, destPtrOfStruct)
}

// FindContext implements DB
func (t *tx) FindContext(ctx context.Context, q query.Query, destPtrOfStruct any) error {
	return t.f.FindContext(ctx, q, destPtrOfStruct)
}

// FindMany implements DB
func (t *tx) FindMany(q query.Query, destSlicePtrOfStruct any) error {
	return t.f.FindMany(q, destSlicePtrOfStruct)
}

// FindManyContext implements DB
func (t *tx) FindManyContext(ctx context.Context, q query.Query, destSlicePtrOfStruct any) error {
	return t.f.FindManyContext(ctx, q, destSlicePtrOfStruct)
}

// Deprecated: Use Find or MapRow/MapRows. It will be removed in next version.
func (t *tx) Map(rows *sql.Rows, destPtr any) error {
	return MapRow(rows, destPtr)
}

// Deprecated: Use FindContext or MapRow/MapRows. It will be removed in next version.
func (t *tx) MapMany(rows *sql.Rows, destSlicePtr any) error {
	return MapRows(rows, destSlicePtr)
}

func (t *tx) Tx() *sql.Tx {
	return t.tx
}

func Transaction(db *sql.DB, ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	sqlTx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	tx := newTx(sqlTx)
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
