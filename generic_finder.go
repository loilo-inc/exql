package exql

import (
	"context"

	"github.com/loilo-inc/exql/v3/query"
)

// GenericFinder is generic version of Finder. It provides type-safe methods to execute Find query and map rows into the destination struct or slice of struct.
type GenericFinder[T any] interface {
	Find(q query.Query) (*T, error)
	FindContext(ctx context.Context, q query.Query) (*T, error)
	FindMany(q query.Query) ([]*T, error)
	FindManyContext(ctx context.Context, q query.Query) ([]*T, error)
}

type genericFinder[T any] struct {
	ex   Executor
	refl Reflector
}

func NewGenericFinder[T any](ex Executor, refl Reflector) GenericFinder[T] {
	return newGenericFinder[T](ex, refl)
}

func newGenericFinder[T any](ex Executor, refl Reflector) *genericFinder[T] {
	return &genericFinder[T]{ex: ex, refl: refl}
}

// Find implements GenericFinder
func (f *genericFinder[T]) Find(q query.Query) (*T, error) {
	return f.FindContext(context.Background(), q)
}

// FindContext implements GenericFinder
func (f *genericFinder[T]) FindContext(ctx context.Context, q query.Query) (*T, error) {
	if stmt, args, err := q.Query(); err != nil {
		return nil, err
	} else if rows, err := f.ex.QueryContext(ctx, stmt, args...); err != nil {
		return nil, err
	} else if result, err := mapRowGeneric[T](f.refl, rows); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}

// FindMany implements GenericFinder
func (f *genericFinder[T]) FindMany(q query.Query) ([]*T, error) {
	return f.FindManyContext(context.Background(), q)
}

// FindManyContext implements GenericFinder
func (f *genericFinder[T]) FindManyContext(ctx context.Context, q query.Query) ([]*T, error) {
	if stmt, args, err := q.Query(); err != nil {
		return nil, err
	} else if rows, err := f.ex.QueryContext(ctx, stmt, args...); err != nil {
		return nil, err
	} else if result, err := mapRowsGeneric[T](f.refl, rows); err != nil {
		return nil, err
	} else {
		return result, nil
	}
}
