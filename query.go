package exql

import (
	"errors"
	"reflect"

	q "github.com/loilo-inc/exql/v2/query"
	"golang.org/x/xerrors"
)

func Where(str string, args ...any) q.Condition {
	return q.Cond(str, args...)
}

type ModelMetadata struct {
	TableName          string
	AutoIncrementField *reflect.Value
	PrimaryKeyColumns  []string
	PrimaryKeyValues   []any
	Values             q.KeyIterator[any]
}

func QueryForInsert(modelPtr Model) (q.Query, *reflect.Value, error) {
	m, err := AggregateModelMetadata(modelPtr)
	if err != nil {
		return nil, nil, err
	}
	b := q.NewBuilder()
	cols := q.Cols(m.Values.Keys()...)
	vals := q.Vals(m.Values.Values())
	b.Sprintf("INSERT INTO `%s`", modelPtr.TableName())
	b.Query("(:?) VALUES (:?)", cols, vals)
	return b.Build(), m.AutoIncrementField, nil
}

func QueryForBulkInsert[T Model](modelPtrs ...T) (q.Query, error) {
	if len(modelPtrs) == 0 {
		return nil, errors.New("empty list")
	}
	var head *ModelMetadata
	b := q.NewBuilder()
	vals := q.NewBuilder()
	for _, v := range modelPtrs {
		if data, err := AggregateModelMetadata(v); err != nil {
			return nil, err
		} else {
			if head == nil {
				head = data
			}
			vals.Query("(:?)", q.Vals(data.Values.Values()))
		}
	}
	b.Sprintf("INSERT INTO `%s`", head.TableName)
	b.Query("(:?) VALUES :?", q.Cols(head.Values.Keys()...), vals.Join(","))
	return b.Build(), nil
}

func AggregateModelMetadata(modelPtr Model) (*ModelMetadata, error) {
	if modelPtr == nil {
		return nil, xerrors.Errorf("pointer is nil")
	}
	objValue := reflect.ValueOf(modelPtr)
	objType := objValue.Type()
	if objType.Kind() != reflect.Ptr || objType.Elem().Kind() != reflect.Struct {
		return nil, xerrors.Errorf("object must be pointer of struct")
	}
	data := map[string]any{}
	// *User -> User
	objType = objType.Elem()
	exqlTagCount := 0
	var primaryKeyColumns []string
	var primaryKeyValues []any
	var autoIncrementField *reflect.Value
	for i := 0; i < objType.NumField(); i++ {
		f := objType.Field(i)
		if t, ok := f.Tag.Lookup("exql"); ok {
			tags, err := ParseTags(t)
			if err != nil {
				return nil, err
			}
			colName, ok := tags["column"]
			if !ok || colName == "" {
				return nil, xerrors.Errorf("column tag is not set")
			}
			exqlTagCount++
			if _, primary := tags["primary"]; primary {
				primaryKeyField := objValue.Elem().Field(i)
				primaryKeyColumns = append(primaryKeyColumns, colName)
				primaryKeyValues = append(primaryKeyValues, primaryKeyField.Interface())
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
		return nil, xerrors.Errorf("obj doesn't have exql tags in any fields")
	}

	if len(primaryKeyColumns) == 0 {
		return nil, xerrors.Errorf("table has no primary key")
	}

	tableName := modelPtr.TableName()
	if tableName == "" {
		return nil, xerrors.Errorf("empty table name")
	}
	return &ModelMetadata{
		TableName:          tableName,
		AutoIncrementField: autoIncrementField,
		PrimaryKeyColumns:  primaryKeyColumns,
		PrimaryKeyValues:   primaryKeyValues,
		Values:             q.NewKeyIterator(data),
	}, nil
}

func QueryForUpdateModel(
	updateStructPtr ModelUpdate,
	where q.Condition,
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

	tableName := updateStructPtr.UpdateTableName()
	if tableName == "" {
		return nil, xerrors.Errorf("empty table name")
	}
	b := q.NewBuilder()
	b.Sprintf("UPDATE `%s`", tableName)
	b.Query("SET :? WHERE :?", q.Set(values), where)
	return b.Build(), nil
}
