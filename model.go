package exql

import (
	"fmt"
	"reflect"

	q "github.com/loilo-inc/exql/v3/query"
	"github.com/loilo-inc/exql/v3/util"
)

type modelSchema struct {
	autoIncrementField *int
	primaryKeyFields   []int
	updatableFields    []int
	fields             *util.SyncMap[string, int]
	columns            *util.SyncMap[int, string]
}

type modelValue struct {
	tableName          string
	autoIncrementField *reflect.Value
	values             q.KeyIterator[any]
}

func aggregateFields(t reflect.Type) (*modelSchema, error) {
	fields := &util.SyncMap[string, int]{}
	columns := &util.SyncMap[int, string]{}
	exqlTagCount := 0
	var updatableFields []int
	var primaryKeyFields []int
	var autoIncrementField *int
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag, ok := f.Tag.Lookup("exql")
		if !ok {
			continue
		}
		if f.Type.Kind() == reflect.Pointer {
			return nil, fmt.Errorf("struct field must not be a pointer: %s %s", f.Type.Name(), f.Type.Kind())
		}
		tags, err := ParseTags(tag)
		if err != nil {
			return nil, err
		}
		colName, ok := tags["column"]
		if !ok || colName == "" {
			return nil, fmt.Errorf("column tag is not set")
		}
		fields.Store(colName, i)
		columns.Store(i, colName)
		exqlTagCount++
		_, primary := tags["primary"]
		if primary {
			primaryKeyFields = append(primaryKeyFields, i)
		}
		_, autoIncrement := tags["auto_increment"]
		if autoIncrement {
			autoIncrementField = &i
		}
		if !autoIncrement {
			updatableFields = append(updatableFields, i)
		}
	}

	if exqlTagCount == 0 {
		return nil, fmt.Errorf("obj doesn't have exql tags in any fields")
	}

	if len(primaryKeyFields) == 0 {
		return nil, fmt.Errorf("table has no primary key")
	}

	return &modelSchema{
		autoIncrementField: autoIncrementField,
		primaryKeyFields:   primaryKeyFields,
		updatableFields:    updatableFields,
		fields:             fields,
		columns:            columns,
	}, nil
}

func (ms *modelSchema) aggregateModelValue(
	modelPtr Model,
) (*modelValue, error) {
	if modelPtr == nil {
		return nil, errModelNil
	}
	objValue := reflect.ValueOf(modelPtr)
	objType := objValue.Type()
	if objType.Kind() != reflect.Pointer || objType.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("object must be pointer of struct")
	}
	// *User -> User
	objType = objType.Elem()
	var autoIncrementField *reflect.Value
	if ms.autoIncrementField != nil {
		f := objValue.Elem().Field(*ms.autoIncrementField)
		autoIncrementField = &f
	}
	var data = map[string]any{}
	for _, idx := range ms.updatableFields {
		f := objValue.Elem().Field(idx)
		col, _ := ms.columns.Load(idx)
		data[col] = f.Interface()
	}
	tableName := modelPtr.TableName()
	if tableName == "" {
		return nil, fmt.Errorf("empty table name")
	}
	return &modelValue{
		tableName:          tableName,
		autoIncrementField: autoIncrementField,
		values:             q.NewKeyIterator(data),
	}, nil
}
