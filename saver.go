package exql

import (
	"context"
	"database/sql"
	"reflect"

	q "github.com/loilo-inc/exql/query"
	"golang.org/x/xerrors"
)

type Saver interface {
	Insert(structPtr any) (sql.Result, error)
	InsertContext(ctx context.Context, structPtr any) (sql.Result, error)
	QueryForInsert(structPtr any) (q.Query, *reflect.Value, error)
	Update(table string, set map[string]any, where Clause) (sql.Result, error)
	UpdateModel(updaterStructPtr any, where Clause) (sql.Result, error)
	UpdateContext(ctx context.Context, table string, set map[string]any, where Clause) (sql.Result, error)
	UpdateModelContext(ctx context.Context, updaterStructPtr any, where Clause) (sql.Result, error)
	QueryForUpdateModel(updateStructPtr any, where Clause) (q.Query, error)
	Delete(table string, where Clause) (sql.Result, error)
	DeleteContext(ctx context.Context, table string, where Clause) (sql.Result, error)
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
	q, autoIncrField, err := s.QueryForInsert(modelPtr)
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
	where Clause,
) (sql.Result, error) {
	return s.UpdateContext(context.Background(), table, set, where)
}

func (s *saver) UpdateContext(
	ctx context.Context,
	table string,
	set map[string]any,
	where Clause,
) (sql.Result, error) {
	query := &q.Update{Table: table, Set: set, Where: where}
	if stmt, args, err := query.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.ExecContext(ctx, stmt, args...)
	}
}

func (s *saver) Delete(from string, where Clause) (sql.Result, error) {
	return s.DeleteContext(context.Background(), from, where)
}

func (s *saver) DeleteContext(ctx context.Context, from string, where Clause) (sql.Result, error) {
	q := &q.Delete{From: from, Where: where}
	if stmt, args, err := q.Query(); err != nil {
		return nil, err
	} else {
		return s.ex.ExecContext(ctx, stmt, args...)
	}
}

func (s *saver) QueryForInsert(modelPtr any) (q.Query, *reflect.Value, error) {
	if modelPtr == nil {
		return nil, nil, xerrors.Errorf("pointer is nil")
	}
	objValue := reflect.ValueOf(modelPtr)
	objType := objValue.Type()
	if objType.Kind() != reflect.Ptr || objType.Elem().Kind() != reflect.Struct {
		return nil, nil, xerrors.Errorf("object must be pointer of struct")
	}
	data := map[string]any{}
	// *User -> User
	objType = objType.Elem()
	exqlTagCount := 0
	var primaryKeyFields []*reflect.Value
	var autoIncrementField *reflect.Value
	for i := 0; i < objType.NumField(); i++ {
		f := objType.Field(i)
		if t, ok := f.Tag.Lookup("exql"); ok {
			tags, err := ParseTags(t)
			if err != nil {
				return nil, nil, err
			}
			colName, ok := tags["column"]
			if !ok || colName == "" {
				return nil, nil, xerrors.Errorf("column tag is not set")
			}
			exqlTagCount++
			if _, primary := tags["primary"]; primary {
				primaryKeyField := objValue.Elem().Field(i)
				primaryKeyFields = append(primaryKeyFields, &primaryKeyField)
			}
			if _, autoIncrement := tags["auto_increment"]; autoIncrement {
				field := objValue.Elem().Field(i)
				autoIncrementField = &field
				// Not include auto_increment field in insert query
				continue
			}
			data[colName] = objValue.Elem().Field(i).Interface()
		}
	}
	if exqlTagCount == 0 {
		return nil, nil, xerrors.Errorf("obj doesn't have exql tags in any fields")
	}

	if len(primaryKeyFields) == 0 {
		return nil, nil, xerrors.Errorf("table has no primary key")
	}

	getTableName := objValue.MethodByName("TableName")
	if !getTableName.IsValid() {
		return nil, nil, xerrors.Errorf("obj doesn't implement TableName() method")
	}
	tableName := getTableName.Call(nil)[0]
	if tableName.Type().Kind() != reflect.String {
		return nil, nil, xerrors.Errorf("wrong implementation of TableName()")
	}
	return &q.Insert{
			Into:   tableName.String(),
			Values: data,
		},
		autoIncrementField,
		nil
}

func (s *saver) UpdateModel(
	ptr any,
	where Clause,
) (sql.Result, error) {
	return s.UpdateModelContext(context.Background(), ptr, where)
}

func (s *saver) UpdateModelContext(
	ctx context.Context,
	ptr any,
	where Clause,
) (sql.Result, error) {
	q, err := s.QueryForUpdateModel(ptr, where)
	if err != nil {
		return nil, err
	}
	stmt, args, err := q.Query()
	if err != nil {
		return nil, err
	}
	return s.ex.ExecContext(ctx, stmt, args...)
}

func (s *saver) QueryForUpdateModel(
	updateStructPtr any,
	where Clause,
) (q.Query, error) {
	if updateStructPtr == nil {
		return nil, xerrors.Errorf("pointer is nil")
	}
	objValue := reflect.ValueOf(updateStructPtr)
	objType := objValue.Type()
	if objType.Kind() != reflect.Ptr || objType.Elem().Kind() != reflect.Struct {
		return nil, xerrors.Errorf("must be pointer of struct")
	}
	objType = objType.Elem()
	values := make(map[string]any)
	if objType.NumField() == 0 {
		return nil, xerrors.Errorf("struct has no field")
	}

	for i := 0; i < objType.NumField(); i++ {
		f := objType.Field(i)
		tag, ok := f.Tag.Lookup("exql")
		if !ok {
			continue
		}
		var colName string
		if tags, err := ParseTags(tag); err != nil {
			return nil, err
		} else if col, ok := tags["column"]; !ok {
			return nil, xerrors.Errorf("tag must include column")
		} else {
			colName = col
		}
		if f.Type.Kind() != reflect.Ptr {
			return nil, xerrors.Errorf("field must be pointer")
		}
		fieldValue := objValue.Elem().Field(i)
		if !fieldValue.IsNil() {
			values[colName] = fieldValue.Elem().Interface()
		}
	}
	if len(values) == 0 {
		return nil, xerrors.Errorf("no value for update")
	}

	getTableName := objValue.MethodByName("ForTableName")
	if !getTableName.IsValid() {
		return nil, xerrors.Errorf("obj doesn't implement ForTableName() method")
	}
	tableName := getTableName.Call(nil)[0]
	if tableName.Type().Kind() != reflect.String {
		return nil, xerrors.Errorf("wrong implementation of ForTableName()")
	}
	return &q.Update{
		Table: tableName.String(),
		Where: where,
		Set:   values,
	}, nil
}
