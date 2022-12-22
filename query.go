package exql

import (
	"reflect"

	q "github.com/loilo-inc/exql/v2/query"
	"golang.org/x/xerrors"
)

func Where(str string, args ...any) q.Condition {
	return q.NewCondition(str, args...)
}

func WhereEx(m map[string]any) q.Condition {
	return q.NewKeyValueCondition(m)
}

func WhereAnd(list ...q.Condition) q.Condition {
	return q.ConditionAnd(list...)
}

func WhereOr(list ...q.Condition) q.Condition {
	return q.ConditionOr(list...)
}

func QueryForInsert(modelPtr Model) (q.Query, *reflect.Value, error) {
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

	tableName := modelPtr.TableName()
	if tableName == "" {
		return nil, nil, xerrors.Errorf("empty table name")
	}
	return &q.Insert{
			Into:   tableName,
			Values: data,
		},
		autoIncrementField,
		nil
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
	return &q.Update{
		Table: tableName,
		Where: where,
		Set:   values,
	}, nil
}
