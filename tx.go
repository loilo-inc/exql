package exql

import (
	"context"
	"database/sql"

	"golang.org/x/xerrors"
)

type Tx interface {
	Saver
	Finder
	Mapper
	Tx() *sql.Tx
}

type tx struct {
	*saver
	*finder
	*mapper
	tx *sql.Tx
}

func newTx(t *sql.Tx) *tx {
	return &tx{saver: newSaver(t), finder: newFinder(t), mapper: &mapper{}, tx: t}
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
