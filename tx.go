package exql

import (
	"context"
	"database/sql"
	"fmt"
)

type Tx interface {
	dbtx
	Tx() *sql.Tx
}

type tx struct {
	*saver
	*finder
	*reflector
	tx *sql.Tx
}

func newTx(t *sql.Tx, reflector *reflector) *tx {
	return &tx{
		saver:     newSaver(t, reflector),
		finder:    newFinder(t, reflector),
		reflector: reflector,
		tx:        t,
	}
}

func (t *tx) Tx() *sql.Tx {
	return t.tx
}

func Transaction(db *sql.DB, ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	return transaction(noCacheReflector, db, ctx, opts, callback)
}

func transaction(reflector *reflector, db *sql.DB, ctx context.Context, opts *sql.TxOptions, callback func(tx Tx) error) error {
	sqlTx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	tx := newTx(sqlTx, reflector)
	var p any
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
