package exql

import (
	"fmt"
	"reflect"

	q "github.com/loilo-inc/exql/v3/query"
)

func Where(str string, args ...any) q.Condition {
	return q.Cond(str, args...)
}

func QueryForInsert(modelPtr Model) (q.Query, *reflect.Value, error) {
	return queryForInsert(noCacheReflector, modelPtr)
}

func queryForInsert(refl Reflector, modelPtr Model) (q.Query, *reflect.Value, error) {
	ms, err := refl.GetSchema(modelPtr)
	if err != nil {
		return nil, nil, err
	}
	v, err := ms.aggregateModelValue(modelPtr)
	if err != nil {
		return nil, nil, err
	}
	b := q.NewBuilder()
	cols := q.Cols(v.values.Keys()...)
	vals := q.Vals(v.values.Values())
	b.Sprintf("INSERT INTO `%s`", v.tableName)
	b.Query("(:?) VALUES (:?)", cols, vals)
	return b.Build(), v.autoIncrementField, nil
}

func QueryForBulkInsert[T Model](modelPtrs ...T) (q.Query, error) {
	return queryForBulkInsert(noCacheReflector, modelPtrs...)
}

func queryForBulkInsert[T Model](refl Reflector, modelPtrs ...T) (q.Query, error) {
	if len(modelPtrs) == 0 {
		return nil, fmt.Errorf("empty list")
	}
	var head *modelValue
	b := q.NewBuilder()
	vals := q.NewBuilder()
	for _, v := range modelPtrs {
		ms, err := refl.GetSchema(v)
		if err != nil {
			return nil, err
		}
		if data, err := ms.aggregateModelValue(v); err != nil {
			return nil, err
		} else {
			if head == nil {
				head = data
			}
			vals.Query("(:?)", q.Vals(data.values.Values()))
		}
	}
	b.Sprintf("INSERT INTO `%s`", head.tableName)
	b.Query("(:?) VALUES :?", q.Cols(head.values.Keys()...), vals.Join(","))
	return b.Build(), nil
}

func QueryForUpdateModel(
	updateStructPtr ModelUpdate,
	where q.Condition,
) (q.Query, error) {
	if updateStructPtr == nil {
		return nil, errModelNil
	}
	objValue := reflect.ValueOf(updateStructPtr)
	objType := objValue.Type()
	if objType.Kind() != reflect.Pointer || objType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("must be pointer of struct")
	}
	objType = objType.Elem()
	values := make(map[string]any)
	if objType.NumField() == 0 {
		return nil, fmt.Errorf("struct has no field")
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
			return nil, fmt.Errorf("tag must include column")
		} else {
			colName = col
		}
		if f.Type.Kind() != reflect.Pointer {
			return nil, fmt.Errorf("field must be pointer")
		}
		fieldValue := objValue.Elem().Field(i)
		if !fieldValue.IsNil() {
			values[colName] = fieldValue.Elem().Interface()
		}
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("no value for update")
	}

	tableName := updateStructPtr.UpdateTableName()
	if tableName == "" {
		return nil, fmt.Errorf("empty table name")
	}
	b := q.NewBuilder()
	b.Sprintf("UPDATE `%s`", tableName)
	b.Query("SET :? WHERE :?", q.Set(values), where)
	return b.Build(), nil
}
