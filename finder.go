package exql

import (
	"context"
	"database/sql"

	"github.com/loilo-inc/exql/v3/query"
)

// Finder is an interface to execute select query and map rows into the destination.
type Finder interface {
	Find(q query.Query, destPtrOfStruct any) error
	FindContext(ctx context.Context, q query.Query, destPtrOfStruct any) error
	FindMany(q query.Query, destSlicePtrOfStruct any) error
	FindManyContext(ctx context.Context, q query.Query, destSlicePtrOfStruct any) error
}

type finder struct {
	ex   Executor
	refl Reflector
}

// Find implements Finder
func (f *finder) Find(q query.Query, destPtrOfStruct any) error {
	return f.FindContext(context.Background(), q, destPtrOfStruct)
}

// FindContext implements Finder
func (f *finder) FindContext(ctx context.Context, q query.Query, destPtrOfStruct any) error {
	rows, err := executeQueryContext(f.ex, ctx, q)
	if err != nil {
		return err
	}
	return mapRow(f.refl, rows, destPtrOfStruct)
}

// FindMany implements Finder
func (f *finder) FindMany(q query.Query, destSlicePtrOfStruct any) error {
	return f.FindManyContext(context.Background(), q, destSlicePtrOfStruct)
}

// FindManyContext implements Finder
func (f *finder) FindManyContext(ctx context.Context, q query.Query, destSlicePtrOfStruct any) error {
	rows, err := executeQueryContext(f.ex, ctx, q)
	if err != nil {
		return err
	}
	return mapRows(f.refl, rows, destSlicePtrOfStruct)
}

func executeQueryContext(ex Executor, ctx context.Context, q query.Query) (*sql.Rows, error) {
	if stmt, args, err := q.Query(); err != nil {
		return nil, err
	} else if rows, err := ex.QueryContext(ctx, stmt, args...); err != nil {
		return nil, err
	} else {
		return rows, nil
	}
}

// NewFinder creates a new Finder with the given Executor.
func NewFinder(ex Executor) Finder {
	return newFinder(ex, noCacheReflector)
}

func newFinder(ex Executor, refl Reflector) *finder {
	return &finder{ex: ex, refl: refl}
}
