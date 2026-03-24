package exql

import (
	"context"

	"github.com/loilo-inc/exql/v3/query"
)

// GenericFinder is generic version of Finder. It provides type-safe methods to execute Find query and map rows into the destination struct or slice of struct.
type GenericFinder[T Model] interface {
	Find(q query.Query) (*T, error)
	FindContext(ctx context.Context, q query.Query) (*T, error)
	FindMany(q query.Query) ([]*T, error)
	FindManyContext(ctx context.Context, q query.Query) ([]*T, error)
}

type genericFinder[T Model] struct {
	ex   Executor
	refl Reflector
}

// NewGenericFinder creates a new GenericFinder with the given Executor and Reflector.
func NewGenericFinder[T Model](ex Executor, db DB) GenericFinder[T] {
	return newGenericFinder[T](ex, db)
}

func newGenericFinder[T Model](ex Executor, refl Reflector) *genericFinder[T] {
	return &genericFinder[T]{ex: ex, refl: refl}
}

// Find implements GenericFinder
func (f *genericFinder[T]) Find(q query.Query) (*T, error) {
	return f.FindContext(context.Background(), q)
}

// FindContext implements GenericFinder
func (f *genericFinder[T]) FindContext(ctx context.Context, q query.Query) (*T, error) {
	rows, err := executeQueryContext(f.ex, ctx, q)
	if err != nil {
		return nil, err
	}
	return mapRowGeneric[T](f.refl, rows)
}

// FindMany implements GenericFinder
func (f *genericFinder[T]) FindMany(q query.Query) ([]*T, error) {
	return f.FindManyContext(context.Background(), q)
}

// FindManyContext implements GenericFinder
func (f *genericFinder[T]) FindManyContext(ctx context.Context, q query.Query) ([]*T, error) {
	rows, err := executeQueryContext(f.ex, ctx, q)
	if err != nil {
		return nil, err
	}
	return mapRowsGeneric[T](f.refl, rows)
}
