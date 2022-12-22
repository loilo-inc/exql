//go:generate mockgen -source $GOFILE -destination ./mocks/mock_$GOPACKAGE/$GOFILE -package mock_$GOPACKAGE
package exql

import (
	"context"
	"database/sql"
	"reflect"

	q "github.com/loilo-inc/exql/query"
)

type Saver interface {
	Insert(structPtr Model) (sql.Result, error)
	InsertContext(ctx context.Context, structPtr Model) (sql.Result, error)
	Update(table string, set map[string]any, where q.Condition) (sql.Result, error)
	UpdateModel(updaterStructPtr ModelUpdate, where q.Condition) (sql.Result, error)
	UpdateContext(ctx context.Context, table string, set map[string]any, where q.Condition) (sql.Result, error)
	UpdateModelContext(ctx context.Context, updaterStructPtr ModelUpdate, where q.Condition) (sql.Result, error)
	Delete(table string, where q.Condition) (sql.Result, error)
	DeleteContext(ctx context.Context, table string, where q.Condition) (sql.Result, error)
	Exec(query q.Query) (sql.Result, error)
	ExecContext(ctx context.Context, query q.Query) (sql.Result, error)
	Query(query q.Query) (*sql.Rows, error)
	QueryContext(ctx context.Context, query q.Query) (*sql.Rows, error)
	QueryRow(query q.Query) (*sql.Row, error)
	QueryRowContext(ctx context.Context, query q.Query) (*sql.Row, error)
}

type saver struct {
	ex Executor
}

func NewSaver(ex Executor) Saver {
	return &saver{ex: ex}
}

func (s *saver) Insert(modelPtr Model) (sql.Result, error) {
	return s.InsertContext(context.Background(), modelPtr)
}

func (s *saver) InsertContext(ctx context.Context, modelPtr Model) (sql.Result, error) {
	q, autoIncrField, err := QueryForInsert(modelPtr)
	if err != nil {
		return nil, err
	}
	result, err := s.ExecContext(ctx, q)
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
	where q.Condition,
) (sql.Result, error) {
	return s.UpdateContext(context.Background(), table, set, where)
}

func (s *saver) UpdateContext(
	ctx context.Context,
	table string,
	set map[string]any,
	where q.Condition,
) (sql.Result, error) {
	return s.ExecContext(ctx, &q.Update{Table: table, Set: set, Where: where})
}

func (s *saver) Delete(from string, where q.Condition) (sql.Result, error) {
	return s.DeleteContext(context.Background(), from, where)
}

func (s *saver) DeleteContext(ctx context.Context, from string, where q.Condition) (sql.Result, error) {
	return s.ExecContext(ctx, &q.Delete{From: from, Where: where})
}

func (s *saver) UpdateModel(
	ptr ModelUpdate,
	where q.Condition,
) (sql.Result, error) {
	return s.UpdateModelContext(context.Background(), ptr, where)
}

func (s *saver) UpdateModelContext(
	ctx context.Context,
	ptr ModelUpdate,
	where q.Condition,
) (sql.Result, error) {
	q, err := QueryForUpdateModel(ptr, where)
	if err != nil {
		return nil, err
	}
	return s.ExecContext(ctx, q)
}

func (s *saver) Exec(query q.Query) (sql.Result, error) {
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.Exec(stmt, args...)
	}
}

func (s *saver) ExecContext(ctx context.Context, query q.Query) (sql.Result, error) {
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.ExecContext(ctx, stmt, args...)
	}
}

func (s *saver) Query(query q.Query) (*sql.Rows, error) {
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.Query(stmt, args...)
	}
}

func (s *saver) QueryContext(ctx context.Context, query q.Query) (*sql.Rows, error) {
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.QueryContext(ctx, stmt, args...)
	}
}

func (s *saver) QueryRow(query q.Query) (*sql.Row, error) {
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.QueryRow(stmt, args...), nil
	}
}

func (s *saver) QueryRowContext(ctx context.Context, query q.Query) (*sql.Row, error) {
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.QueryRowContext(ctx, stmt, args...), nil
	}
}
