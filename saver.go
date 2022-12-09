//go:generate mockgen -source $GOFILE -destination ./mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package exql

import (
	"context"
	"database/sql"
	"reflect"

	q "github.com/loilo-inc/exql/query"
)

type Saver interface {
	Insert(structPtr any) (sql.Result, error)
	InsertContext(ctx context.Context, structPtr any) (sql.Result, error)
	Update(table string, set map[string]any, where q.Stmt) (sql.Result, error)
	UpdateModel(updaterStructPtr any, where q.Stmt) (sql.Result, error)
	UpdateContext(ctx context.Context, table string, set map[string]any, where q.Stmt) (sql.Result, error)
	UpdateModelContext(ctx context.Context, updaterStructPtr any, where q.Stmt) (sql.Result, error)
	Delete(table string, where q.Stmt) (sql.Result, error)
	DeleteContext(ctx context.Context, table string, where q.Stmt) (sql.Result, error)
}

type saver struct {
	ex Executor
}

type SET map[string]any

func NewSaver(ex Executor) *saver {
	return &saver{ex: ex}
}

func (s *saver) Insert(modelPtr any) (sql.Result, error) {
	return s.InsertContext(context.Background(), modelPtr)
}

func (s *saver) InsertContext(ctx context.Context, modelPtr any) (sql.Result, error) {
	q, autoIncrField, err := QueryForInsert(modelPtr)
	if err != nil {
		return nil, err
	}
	stmt, args, err := q.Query()
	if err != nil {
		return nil, err
	}
	result, err := s.ex.ExecContext(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	if autoIncrField != nil {
		lid, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}
		kind := autoIncrField.Kind()
		if kind == reflect.Int64 {
			autoIncrField.Set(reflect.ValueOf(lid))
		} else if kind == reflect.Uint64 {
			autoIncrField.Set(reflect.ValueOf(uint64(lid)))
		}
	}
	return result, nil
}

func (s *saver) Update(
	table string,
	set map[string]any,
	where q.Stmt,
) (sql.Result, error) {
	return s.UpdateContext(context.Background(), table, set, where)
}

func (s *saver) UpdateContext(
	ctx context.Context,
	table string,
	set map[string]any,
	where q.Stmt,
) (sql.Result, error) {
	query := &q.Update{Table: table, Set: set, Where: where}
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.ExecContext(ctx, stmt, args...)
	}
}

func (s *saver) Delete(from string, where q.Stmt) (sql.Result, error) {
	return s.DeleteContext(context.Background(), from, where)
}

func (s *saver) DeleteContext(ctx context.Context, from string, where q.Stmt) (sql.Result, error) {
	q := &q.Delete{From: from, Where: where}
	if stmt, args, err := q.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.ExecContext(ctx, stmt, args...)
	}
}

func (s *saver) UpdateModel(
	ptr any,
	where q.Stmt,
) (sql.Result, error) {
	return s.UpdateModelContext(context.Background(), ptr, where)
}

func (s *saver) UpdateModelContext(
	ctx context.Context,
	ptr any,
	where q.Stmt,
) (sql.Result, error) {
	q, err := QueryForUpdateModel(ptr, where)
	if err != nil {
		return nil, err
	}
	stmt, args, err := q.Query()
	if err != nil {
		return nil, err
	}
	return s.ex.ExecContext(ctx, stmt, args...)
}
