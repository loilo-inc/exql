package exql

import (
	"context"

	"github.com/loilo-inc/exql/v2/query"
)

// Finder is an interface to execute select query and map rows into the destination.
type Finder interface {
	Find(q query.Query, destPtrOfStruct any) error
	FindContext(ctx context.Context, q query.Query, destPtrOfStruct any) error
	FindMany(q query.Query, destSlicePtrOfStruct any) error
	FindManyContext(ctx context.Context, q query.Query, destSlicePtrOfStruct any) error
}

type finder struct {
	ex Executor
}

func NewFinder(ex Executor) Finder {
	return newFinder(ex)
}

// Find implements Finder
func (f *finder) Find(q query.Query, destPtrOfStruct any) error {
	return f.FindContext(context.Background(), q, destPtrOfStruct)
}

// FindContext implements Finder
func (f *finder) FindContext(ctx context.Context, q query.Query, destPtrOfStruct any) error {
	if stmt, args, err := q.Query(); err != nil {
		return err
	} else if rows, err := f.ex.QueryContext(ctx, stmt, args...); err != nil {
		return err
	} else if err := MapRow(rows, destPtrOfStruct); err != nil {
		return err
	}
	return nil
}

// FindMany implements Finder
func (f *finder) FindMany(q query.Query, destSlicePtrOfStruct any) error {
	return f.FindManyContext(context.Background(), q, destSlicePtrOfStruct)
}

// FindManyContext implements Finder
func (f *finder) FindManyContext(ctx context.Context, q query.Query, destSlicePtrOfStruct any) error {
	if stmt, args, err := q.Query(); err != nil {
		return err
	} else if rows, err := f.ex.QueryContext(ctx, stmt, args...); err != nil {
		return err
	} else if err := MapRows(rows, destSlicePtrOfStruct); err != nil {
		return err
	}
	return nil
}

func newFinder(ex Executor) *finder {
	return &finder{ex: ex}
}
